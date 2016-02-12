package data

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
	"log"
	"github.com/warmans/stressfaktor-api/entity"
	"github.com/warmans/stressfaktor-api/data/source/sfaktor"
)

type Ingest struct {
	DB              *sql.DB
	UpdateFrequency time.Duration
	Stressfaktor    *sfaktor.Crawler
	EventVisitors  []entity.EventVisitor
}

func (i *Ingest) Run() {
	for {
		for _, event := range i.Stressfaktor.Crawl() {
			if err := i.Ingest(event); err != nil {
				log.Printf("Failed to ingest event: %s", err.Error())
			}
		}
		time.Sleep(i.UpdateFrequency)
	}
}

//todo: update records as well as inserting new ones
func (i *Ingest) Ingest(event *entity.Event) error {

	//process record
	for _, v := range i.EventVisitors {
		v.Visit(event)
	}

	//update DB
	tx, err := i.DB.Begin()
	if err != nil {
		return err
	}

	err = func(tr *sql.Tx) error {
		//ensure venue exists and has an ID
		err = i.venueMustExist(tr, event.Venue)
		if err != nil {
			return err
		}

		//get/create the main event record
		err = tr.QueryRow("SELECT id FROM event WHERE venue_id=? AND date=?", event.Venue.ID, event.Date.Format(entity.DATE_FORMAT_SQL)).Scan(&event.ID)
		if err != nil && err != sql.ErrNoRows {
			return err
		}
		if event.ID == 0 {
			//now insert the main record
			res, err := tr.Exec(
				"INSERT INTO event (venue_id, date, type, description) VALUES (?, ?, ?, ?)",
				event.Venue.ID,
				event.Date.Format(entity.DATE_FORMAT_SQL),
				event.Type,
				event.Description,
			)
			if err != nil {
				return err
			}
			event.ID, err = res.LastInsertId()
			if err != nil {
				return err
			}
		}

		//finally append the performers
		for _, performer := range event.Performers {

			err = i.performerMustExist(tr, performer)
			if err != nil {
				return err
			}

			//make the association
			_, err := tr.Exec("REPLACE INTO event_performer (event_id, performer_id) VALUES (?, ?)", event.ID, performer.ID)
			if err != nil {
				return err
			}
		}
		return nil
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

func (i *Ingest) venueMustExist(tr *sql.Tx, venue *entity.Venue) error {
	//get the venue ID if it exists
	if venue.ID == 0 {
		err := tr.QueryRow("SELECT id FROM venue WHERE name=? AND address=?", venue.Name, venue.Address).Scan(&venue.ID)
		if err != nil && err != sql.ErrNoRows {
			return err
		}
	}
	//still no venue ID... create it
	if venue.ID == 0 {
		res, err := tr.Exec(
			"INSERT INTO venue (name, address) VALUES (?, ?)",
			venue.Name,
			venue.Address,
		)
		if err != nil {
			return err
		}
		venue.ID, err = res.LastInsertId()
		if err != nil {
			return err
		}
	}
	return nil
}

func (i *Ingest) performerMustExist(tr *sql.Tx, performer *entity.Performer) error {

	if performer.ID != 0 {
		return nil
	}

	//get/create the performer
	err := tr.QueryRow("SELECT id FROM performer WHERE name=? AND genre=?", performer.Name, performer.Genre).Scan(&performer.ID)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if performer.ID == 0 {
		res, err := tr.Exec(
			"INSERT INTO performer (name, genre, home, listen_url) VALUES (?, ?, ?, ?)",
			performer.Name,
			performer.Genre,
			performer.Home,
			performer.ListenURL,
		)
		if err != nil {
			return err
		}

		performer.ID, err = res.LastInsertId()
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Ingest) Cleanup() {
	res, err := s.DB.Exec(`UPDATE event SET deleted=1 WHERE date < $1 AND deleted=0`, time.Now().Format(entity.DATE_FORMAT_SQL))
	if err != nil {
		log.Printf("Cleaned failed: %s", err.Error())
		return
	}

	affected, _ := res.RowsAffected()
	log.Printf("Cleaned up %d rows", affected)
}
