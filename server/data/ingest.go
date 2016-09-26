package data

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/warmans/dbr"
	"github.com/warmans/stressfaktor-api/server/data/source"
	"github.com/warmans/stressfaktor-api/server/data/store"
)

type Ingest struct {
	DB              *dbr.Session
	UpdateFrequency time.Duration
	EventVisitors   []store.EventVisitor
	Crawlers        []source.Crawler
}

func (i *Ingest) Run() {
	for {
		for _, c := range i.Crawlers {

			log.Printf("Crawling %T...", c)
			events, err := c.Crawl(source.MustMakeTimeLocation("Europe/Berlin"))
			if err != nil {
				log.Printf("Failed %T failed crawling: %s", c, err.Error())
				continue
			}


			for _, event := range events {
				log.Printf("Discovered %+v", *event)
				if err := i.Ingest(event); err != nil {
					log.Printf("Failed to ingest event: %s", err.Error())
				}
			}
		}

		i.Cleanup()
		time.Sleep(i.UpdateFrequency)
	}
}

func (i *Ingest) Ingest(event *store.Event) error {

	//sanity check incoming data
	if !event.IsValid() || event.Venue == nil || !event.Venue.IsValid() {
		return fmt.Errorf("Invalid event/venue was rejected at injest. Event was %+v venue was %+v", event, event.Venue)
	}

	//update DB
	tx, err := i.DB.Begin()
	if err != nil {
		return err
	}

	//pre-process record
	for _, v := range i.EventVisitors {
		v.Visit(event)
	}

	err = func(tr *dbr.Tx) error {

		//ensure venue exists and has an ID
		err = i.venueMustExist(tr, event.Venue)
		if err != nil {
			return err
		}

		//ensure all performers exist
		for _, performer := range event.Performers {
			err = i.performerMustExist(tr, performer)
			if err != nil {
				return err
			}
		}

		if event.ID == 0 {
			// if no ID was supplied try and find one based on the venue and date. i.e. assume two events cannot occur at the
			// same venue at the same time. Note that if either of these fields has been updated there will be no match
			// and the row will be processed as though it is a new record.
			err = tr.QueryRow("SELECT id FROM event WHERE venue_id=? AND date=?", event.Venue.ID, event.Date.Format(store.DATE_FORMAT_SQL)).Scan(&event.ID)
			if err != nil && err != sql.ErrNoRows {
				return err
			}
		}

		//update/create the main event record
		if event.ID == 0 {

			res, err := tr.Exec(
				"INSERT INTO event (date, venue_id, type, description) VALUES (?, ?, ?, ?)",
				event.Date.Format(store.DATE_FORMAT_SQL),
				event.Venue.ID,
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

		} else {
			// note that we cannot update the venue or date. Doing so will create a new event since these fields act
			// as a composite primary key.
			_, err := tr.Exec(
				"UPDATE event SET type=?, description=? WHERE id=?",
				event.Type,
				event.Description,
				event.ID,
			)
			if err != nil {
				return err
			}
		}

		//clear existing relationships (i.e. always use the most up-to-date listing)
		_, err := tr.Exec("DELETE FROM event_performer WHERE event_id=?", event.ID)
		if err != nil {
			return err
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

func (i *Ingest) venueMustExist(tr *dbr.Tx, venue *store.Venue) error {

	//get the venue ID if it exists
	if venue.ID == 0 {
		err := tr.QueryRow("SELECT id FROM venue WHERE name=?", venue.Name, venue.Address).Scan(&venue.ID)
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
	} else {
		_, err := tr.Exec(
			"UPDATE venue SET address=? WHERE id=?",
			venue.Address,
			venue.ID,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (i *Ingest) performerMustExist(tr *dbr.Tx, performer *store.Performer) error {

	if !performer.IsValid() {
		return nil
	}

	if performer.ID == 0 {
		//get/create the performer
		err := tr.QueryRow("SELECT id FROM performer WHERE name=? AND genre=?", performer.Name, performer.Genre).Scan(&performer.ID)
		if err != nil && err != sql.ErrNoRows {
			return err
		}
	}
	if performer.ID == 0 {
		res, err := tr.Exec(
			"INSERT INTO performer (name, info, genre, home, img, listen_url) VALUES (?, ?, ?, ?, ?, ?)",
			performer.Name,
			performer.Info,
			performer.Genre,
			performer.Home,
			performer.Img,
			performer.ListenURL,
		)
		if err != nil {
			return err
		}
		performer.ID, err = res.LastInsertId()
		if err != nil {
			return err
		}
	} else {
		_, err := tr.Exec(
			"UPDATE performer SET info=?, home=?, img=?, listen_url=? WHERE id=?",
			performer.Info,
			performer.Home,
			performer.Img,
			performer.ListenURL,
			performer.ID,
		)
		if err != nil {
			return err
		}
	}

	//clear existing relationships for extra data to allow links to be kept up-to-date
	_, err := tr.Exec("DELETE FROM performer_extra WHERE performer_id=?", performer.ID)
	if err != nil {
		return err
	}

	for _, link := range performer.Links {
		_, err := tr.Exec(
			"INSERT INTO performer_extra (performer_id, link, link_type, link_description) VALUES (?, ?, ?, ?)",
			performer.ID,
			link.URI,
			link.Type,
			link.Text,
		)
		if err != nil {
			log.Print("Failed to insert performer_extra: %s", err.Error())
			continue
		}
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
