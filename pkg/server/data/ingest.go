package data

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/warmans/dbr"
	"github.com/warmans/fakt-api/pkg/server/data/service/common"
	"github.com/warmans/fakt-api/pkg/server/data/service/event"
	"github.com/warmans/fakt-api/pkg/server/data/service/performer"
	"github.com/warmans/fakt-api/pkg/server/data/service/venue"
	"github.com/warmans/fakt-api/pkg/server/data/source"
	"go.uber.org/zap"
)

type Ingest struct {
	DB               *dbr.Session
	UpdateFrequency  time.Duration
	EventVisitors    []common.EventVisitor
	Crawlers         []source.Crawler
	timezone         *time.Location

	EventService     *event.EventService
	VenueService     *venue.VenueService
	PerformerService *performer.PerformerService

	Logger           *zap.Logger
}

func (i *Ingest) Run() {
	for {
		wg := sync.WaitGroup{}
		for _, c := range i.Crawlers {
			wg.Add(1)
			go func(c source.Crawler) {

				defer wg.Done()
				logger := i.Logger.With(zap.String("crawler", fmt.Sprintf("%T", c)))

				logger.Info("crawling...")
				events, err := c.Crawl()
				if err != nil {
					logger.Error("Failed failed crawling", zap.Error(err))
					return
				}

				logger.Info(fmt.Sprintf("Discovered %d events", len(events)))
				for _, ev := range events {
					//append the source to all events
					ev.Source = c.Name()
					if err := i.Ingest(ev); err != nil {
						logger.Error("Failed to ingest event", zap.Error(err))
					}
				}
			}(c)
		}
		wg.Wait()

		i.Cleanup()
		time.Sleep(i.UpdateFrequency)
	}
}

func (i *Ingest) Ingest(event *common.Event) error {
	//pre-process record
	for _, v := range i.EventVisitors {
		v.Visit(event)
	}

	tx, err := i.DB.Begin()
	if err != nil {
		return err
	}

	err = func(tr *dbr.Tx) error {

		//event must have an existing venue
		if err := i.VenueService.VenueMustExist(tr, event.Venue); err != nil {
			return err
		}

		//performers should also exist before event is created
		for _, perf := range event.Performers {
			err = i.PerformerService.PerformerMustExist(tr, perf)
			if err != nil {
				return err
			}
		}

		return i.EventService.EventMustExist(tr, event)
	}(tx)

	if err == nil {
		if err := tx.Commit(); err != nil {
			return err
		}
	} else {
		if txerr := tx.Rollback(); txerr != nil {
			return errors.New(fmt.Sprintf("%s -> %s", err, txerr))
		}
		return err
	}

	return nil
}

func (i *Ingest) Cleanup() {
	res, err := i.DB.Exec(`UPDATE event SET deleted=1 WHERE date(date) < date('now') AND deleted=0`)
	if err != nil {
		i.Logger.Error("Cleaned failed", zap.Error(err))
		return
	}

	affected, _ := res.RowsAffected()
	i.Logger.Info(fmt.Sprintf("Cleaned up %d rows", affected))
}
