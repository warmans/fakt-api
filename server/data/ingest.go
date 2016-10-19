package data

import (
	"log"
	"time"

	"sync"

	"errors"
	"fmt"

	"github.com/warmans/dbr"
	"github.com/warmans/fakt-api/server/data/service/common"
	"github.com/warmans/fakt-api/server/data/service/event"
	"github.com/warmans/fakt-api/server/data/service/performer"
	"github.com/warmans/fakt-api/server/data/service/venue"
	"github.com/warmans/fakt-api/server/data/source"
)

type Ingest struct {
	DB              *dbr.Session
	UpdateFrequency time.Duration
	EventVisitors   []common.EventVisitor
	Crawlers        []source.Crawler

	EventService     *event.EventService
	VenueService     *venue.VenueService
	PerformerService *performer.PerformerService
}

func (i *Ingest) Run() {
	for {
		wg := sync.WaitGroup{}
		for _, c := range i.Crawlers {
			wg.Add(1)
			go func(c source.Crawler) {

				defer wg.Done()
				log.Printf("%T | Crawling...", c)
				events, err := c.Crawl(source.MustMakeTimeLocation("Europe/Berlin"))
				if err != nil {
					log.Printf("%T | Failed failed crawling: %s", c, err.Error())
					return
				}

				log.Printf("%T | Discovered %d events", c, len(events))
				for _, event := range events {
					//append the source to all events
					event.Source = c.Name()
					if err := i.Ingest(event); err != nil {
						log.Printf("%T | Failed to ingest event: %s", c, err.Error())
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
		for _, performer := range event.Performers {
			err = i.PerformerService.PerformerMustExist(tr, performer)
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

func (s *Ingest) Cleanup() {
	res, err := s.DB.Exec(`UPDATE event SET deleted=1 WHERE date(date) < date('now') AND deleted=0`)
	if err != nil {
		log.Printf("Cleaned failed: %s", err.Error())
		return
	}

	affected, _ := res.RowsAffected()
	log.Printf("Cleaned up %d rows", affected)
}
