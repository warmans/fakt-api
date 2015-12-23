package entity

import (
	"time"
	"database/sql"
	"regexp"
	"log"
	"strings"
	"errors"
	"fmt"
)

const DATE_FORMAT_SQL = "2006-01-02 15:04:05.999999999-07:00"

type Event struct {
	ID          int64        `json:"id"`
	Date        time.Time    `json:"date"`
	Venue       *Venue       `json:"venue"`
	Type        string       `json:"type"`
	Description string       `json:"description"`
	Performers  []*Performer `json:"performer"`
}

func (e *Event) GuessPerformers() {

	//reset performer list
	e.Performers = make([]*Performer, 0)

	re := regexp.MustCompile(`"[^"]+"[\s]+?\([^(]+\)`)
	space := regexp.MustCompile(`\"[\s]+\(`)

	result := re.FindAllString(e.Description, -1)
	for _, raw := range result {
		parts := space.Split(raw, -1)
		if len(parts) != 2 {
			log.Printf("%s did not have enough parts", raw)
			continue;
		}
		perf := &Performer{
			Name: strings.Trim(parts[0], `" `),
			Genre: strings.Trim(parts[1], "() "),
		}
		e.Performers = append(e.Performers, perf)
	}
	log.Printf("%v", e.Performers)
}

type Venue struct {
	ID      int64     `json:"id"`
	Name    string    `json:"name"`
	Address string    `json:"address"`
}

type Performer struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Genre string `json:"genre"`
	URI   string `json:"uri"`
}

type EventFilter struct {
	EventIDs	 []int		`json:"events"`
	DateFrom     time.Time  `json:"from_date"`
	DateTo       time.Time  `json:"to_date"`
	VenueIDs     []int      `json:"venues"`
	PerformerIDs []int      `json:"performers"`
}

type EventStore struct {
	DB *sql.DB
}

func (s *EventStore) Initialize() error {
	_, err := s.DB.Exec(`
		CREATE TABLE IF NOT EXISTS event (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			venue_id INTEGER,
			date DATETIME,
			type TEXT NULL,
			description TEXT NULL
		);
	`)
	if err != nil {
		return err
	}
	_, err = s.DB.Exec(`
		CREATE TABLE IF NOT EXISTS venue (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT,
			address TEXT NULL
		);
	`)
	if err != nil {
		return err
	}
	_, err = s.DB.Exec(`
		CREATE TABLE IF NOT EXISTS performer (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT,
			genre TEXT,
			uri TEXT
		);
	`)
	if err != nil {
		return err
	}
	_, err = s.DB.Exec(`
		CREATE TABLE IF NOT EXISTS event_performer (
			event_id INTEGER,
			performer_id INTEGER
		);
		CREATE INDEX IF NOT EXISTS unique_assoc ON event_performer (event_id, performer_id)
	`)
	if err != nil {
		return err
	}

	return err
}

func (s *EventStore) Find(filter *EventFilter) ([]*Event, error) {

	filterSql, filterValues := getFilterSql(filter)
	fullSql := `
		SELECT
			e.id,
			e.date,
			e.type,
			e.description,
			v.id,
			v.name,
			v.address,
			p.id,
			p.name,
			p.genre,
			p.uri
		FROM event e
		LEFT JOIN venue v ON e.venue_id = v.id
		LEFT JOIN event_performer ep ON e.id = ep.event_id
		LEFT JOIN performer p ON ep.performer_id = p.id`+
		filterSql + `
		ORDER BY e.id, v.id, p.id ASC`

	result, err :=  s.DB.Query(fullSql, filterValues...)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	defer result.Close()

	events := make([]*Event, 0)
	curEvent := &Event{}

	for result.Next() {

		if err := result.Err(); err != nil {
			return nil, err
		}

		e_id := 0
		e_date := time.Time{}
		e_type := ""
		e_description := ""
		v_id := 0
		v_name := ""
		v_address := ""
		p_id := 0
		p_name := ""
		p_genre := ""
		p_uri :=  ""

		result.Scan(
			&e_id,
			&e_date,
			&e_type,
			&e_description,
			&v_id,
			&v_name,
			&v_address,
			&p_id,
			&p_name,
			&p_genre,
			&p_uri,
		)

		if curEvent.ID != int64(e_id) {

			//append to result set
			if curEvent.ID != 0 {
				events = append(events, curEvent)
			}

			//new current event
			curEvent = &Event{
				ID: int64(e_id),
				Date: e_date,
				Type: e_type,
				Description: e_description,
				Venue: &Venue{
					ID: int64(v_id),
					Name: v_name,
					Address: v_address,
				},
				Performers: make([]*Performer, 0),
			}
		}

		curPerformer := &Performer{
			ID: int64(p_id),
			Name: p_name,
			Genre: p_genre,
			URI: p_uri,
		}
		if curPerformer.ID != 0 {
			curEvent.Performers = append(curEvent.Performers, curPerformer)
		}
	}

	if curEvent.ID != 0 {
		events = append(events, curEvent)
	}

	return events, nil
}

func getFilterSql(ef *EventFilter) (string, []interface{}) {
	values := make([]interface{}, 0)
	sql := make([]string, 0)

	//event IDs
	if len(ef.EventIDs) > 0 {
		sql = append(sql, "e.id IN ("+(strings.TrimRight(strings.Repeat("?,", len(ef.VenueIDs)), ","))+")")
		for _, val := range ef.EventIDs {
			values = append(values, val)
		}
	}

	//date range
	if !ef.DateFrom.IsZero() {
		sql = append(sql, "e.date >= ?")
		values = append(values, ef.DateFrom.Format(DATE_FORMAT_SQL))
	}
	if !ef.DateTo.IsZero() {
		sql = append(sql, "e.date < ?")
		values = append(values, ef.DateTo.Format(DATE_FORMAT_SQL))
	}

	//venue
	if len(ef.VenueIDs) > 0 {
		sql = append(sql, "v.id IN ("+(strings.TrimRight(strings.Repeat("?,", len(ef.VenueIDs)), ","))+")")
		for _, val := range ef.VenueIDs {
			values = append(values, val)
		}
	}

	//performer
	if len(ef.PerformerIDs) > 0 {
		sql = append(sql, "p.id IN ("+(strings.TrimRight(strings.Repeat("?,", len(ef.VenueIDs)), ","))+")")
		for _, val := range ef.PerformerIDs {
			values = append(values, val)
		}
	}

	return strings.Join(sql, " AND "), values
}

func (s *EventStore) Upsert(event *Event) error {

	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}

	err = func(tr *sql.Tx) error {
		//ensure venue exists and has an ID
		err = s.venueMustExist(tr, event.Venue)
		if err != nil {
			return err
		}

		//get/create the main event record
		err = tr.QueryRow("SELECT id FROM event WHERE venue_id=? AND date=?", event.Venue.ID, event.Date.Format(DATE_FORMAT_SQL)).Scan(&event.ID)
		if err != nil && err != sql.ErrNoRows {
			return err
		}
		if event.ID == 0 {
			//now insert the main record
			res, err := tr.Exec(
				"INSERT INTO event (venue_id, date, type, description) VALUES (?, ?, ?, ?)",
				event.Venue.ID,
				event.Date.Format(DATE_FORMAT_SQL),
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

			err = s.performerMustExist(tr, performer)
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

func (s *EventStore) venueMustExist(tr *sql.Tx, venue *Venue) error {
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

func (s *EventStore) performerMustExist(tr *sql.Tx, performer *Performer) error {

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
			"INSERT INTO performer (name, genre, uri) VALUES (?, ?, ?)",
			performer.Name,
			performer.Genre,
			performer.URI,
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
