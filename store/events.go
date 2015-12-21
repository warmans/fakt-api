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
		)
	`)
	if err != nil {
		return err
	}
	_, err = s.DB.Exec(`
		CREATE TABLE venu (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT,
			address TEXT NULL
		)
	`)

	return err
}

func (s *EventStore) Upsert(event *entity.Event) {
	s.replace(event)
}

func (s *EventStore) replace(event *entity.Event) error {

	tr, err := s.DB.Begin()
	if err != nil {
		return err
	}
	_, err = tr.Exec(
		"REPLACE INTO venue (id, name, address) VALUES (?, ?, ?)",
		event.Venue.ID,
		event.Venue.Name,
		event.Venue.Address,
	)
	if err != nil {
		return err
	}

	_, err = tr.Exec(
		"REPLACE INTO event (id, venue_id, date, type, description) VALUES (?, ?, ?, ?, ?)",
		event.ID,
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
