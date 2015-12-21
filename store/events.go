package store

import (
	"github.com/warmans/stressfaktor-api/entity"
	"database/sql"
)

type EventStore struct {
	DB *sql.DB
}

func (s *EventStore) Initialize() error {
	_, err := s.DB.Exec(`
		CREATE TABLE event (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			venue_id INTEGER,
			date DATETIME,
			type TEXT NULL,
			description TEXT NULL
		);
		CREATE INDEX event_unique ON event (venue_id, date)
	`)
	if err != nil {
		return err
	}
	_, err = s.DB.Exec(`
		CREATE TABLE venue (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT,
			address TEXT NULL
		);
		CREATE INDEX venue_unique ON venue (name, address)
	`)

	return err
}

func (s *EventStore) Upsert(event *entity.Event) error {
	return s.replace(event)
}

func (s *EventStore) replace(event *entity.Event) error {

	tr, err := s.DB.Begin()
	if err != nil {
		return err
	}

	//get the ID if it exists
	if event.Venue.ID == 0 {
		err := tr.QueryRow("SELECT id FROM venue WHERE name=? AND address=?", event.Venue.Name, event.Venue.Address).Scan(event.Venue.ID)
		if err != nil && err != sql.ErrNoRows {
			return err
		}
	}

	//still no ID... create it
	if event.Venue.ID == 0 {
		res, err := tr.Exec(
			"REPLACE INTO venue (name, address) VALUES (?, ?)",
			event.Venue.Name,
			event.Venue.Address,
		)
		if err != nil {
			return err
		}

		event.Venue.ID, err = res.LastInsertId()
		if err != nil {
			return err
		}
	}

	_, err = tr.Exec(
		"REPLACE INTO event (venue_id, date, type, description) VALUES (?, ?, ?, ?)",
		event.Venue.ID,
		event.Date.Format("2006-01-02 15:04:05.999999999-07:00"),
		event.Type,
		event.Description,
	)
	if err != nil {
		return err
	}

	if err := tr.Commit(); err != nil {
		return err
	}

	return nil
}
