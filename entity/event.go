package entity

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"
)

const DATE_FORMAT_SQL = "2006-01-02 15:04:05.999999999-07:00"

type Event struct {
	ID          int64        `json:"id"`
	Date        time.Time    `json:"date"`
	Venue       *Venue       `json:"venue,omitempty"`
	Type        string       `json:"type"`
	Description string       `json:"description"`
	Performers  []*Performer `json:"performer,omitempty"`
}

func (e *Event) GuessPerformers() {

	//reset performer list
	e.Performers = make([]*Performer, 0)

	re := regexp.MustCompile(`"[^"]+"[\s]+?\([^(]+\)`)
	spaceRe := regexp.MustCompile(`\"[\s]+\(`)
	fromRe := regexp.MustCompile(`(aus|from)\s+([^,\.\;]+)`)

	result := re.FindAllString(e.Description, -1)
	for _, raw := range result {
		parts := spaceRe.Split(raw, -1)
		if len(parts) != 2 {
			log.Printf("%s did not have enough parts", raw)
			continue
		}

		name := strings.Trim(parts[0], `" `)
		genre := strings.Trim(parts[1], "() ")

		//try and find a location in the genre
		home := ""
		if fromMatch := fromRe.FindStringSubmatch(genre); len(fromMatch) == 3 {
			//e.g. from Berlin, from, Berlin
			home = fromMatch[2]
		}

		perf := &Performer{
			Name:  name,
			Genre: genre,
			Home: home,
		}
		e.Performers = append(e.Performers, perf)
	}
}

type Venue struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Address string `json:"address"`
	Events []*Event `json:"event,omitempty"`
}

type VenueFilter struct {
	VenueIDs     []int     `json:"venues"`
}

type Performer struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Genre string `json:"genre"`
	Home  string `json:"home"`
	Events []*Event `json:"event,omitempty"`
}

type EventFilter struct {
	EventIDs     []int     `json:"events"`
	DateFrom     time.Time `json:"from_date"`
	DateTo       time.Time `json:"to_date"`
	VenueIDs     []int     `json:"venues"`
	Types        []string  `json:"types"`
	ShowDeleted  bool  	   `json:"show_deleted"`
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
			description TEXT NULL,
			deleted BOOLEAN DEFAULT 0
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
			home TEXT
		);
	`)
	if err != nil {
		return err
	}
	_, err = s.DB.Exec(`
		CREATE TABLE IF NOT EXISTS event_performer (
			event_id INTEGER,
			performer_id INTEGER,
			PRIMARY KEY (event_id, performer_id)
		);
	`)
	if err != nil {
		return err
	}

	return err
}

func (s *EventStore) Cleanup() {
	res, err := s.DB.Exec(`UPDATE event SET deleted=1 WHERE date < $1 AND deleted=0`, time.Now().Format(DATE_FORMAT_SQL))
	if err != nil {
		log.Printf("Cleaned failed: %s", err.Error())
		return
	}

	affected, _ := res.RowsAffected()
	log.Printf("Cleaned up %d rows", affected)
}

func (s *EventStore) FindEvents(filter *EventFilter) ([]*Event, error) {

	q := Sql{}
	q.Select("e.id", "e.date", "e.type", "e.description", "v.id", "v.name", "v.address", "p.id", "p.name", "p.genre", "p.home")
	q.From("event e")
	q.LeftJoin("venue v ON e.venue_id = v.id")
	q.LeftJoin("event_performer ep ON e.id = ep.event_id")
	q.LeftJoin("performer p ON ep.performer_id = p.id")
	q.WhereIntIn("e.id", filter.EventIDs...)
	q.WhereIntIn("v.id", filter.VenueIDs...)
	q.WhereStringIn("e.type", filter.Types...)
	q.WhereTime("e.date", ">=", filter.DateFrom)
	q.WhereTime("e.date", "<", filter.DateTo)
	q.WhereInt("e.deleted", "<=", IfOrInt(filter.ShowDeleted, 1, 0))
	q.SetOrder("v.name ASC")

	result, err := s.DB.Query(q.GetSQL(), q.GetValues()...)
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

		var eID, vID, pID int
		var eType, eDescription, vName,  vAddress, pName, pGenre, pHome string
		var eDate time.Time

		result.Scan(&eID, &eDate, &eType, &eDescription, &vID, &vName, &vAddress, &pID, &pName, &pGenre, &pHome)

		if curEvent.ID != int64(eID) {

			//append to result set
			if curEvent.ID != 0 {
				events = append(events, curEvent)
			}

			//new current event
			curEvent = &Event{
				ID:          int64(eID),
				Date:        eDate,
				Type:        eType,
				Description: eDescription,
				Venue: &Venue{
					ID:      int64(vID),
					Name:    vName,
					Address: vAddress,
				},
				Performers: make([]*Performer, 0),
			}
		}

		curPerformer := &Performer{
			ID:    int64(pID),
			Name:  pName,
			Genre: pGenre,
			Home:  pHome,
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

func (s *EventStore) FindVenues(filter *VenueFilter) ([]*Venue, error) {

	q := Sql{}
	q.Select("v.id", "v.name", "v.address")
	q.From("venue v")
	q.WhereIntIn("v.id", filter.VenueIDs...)
	q.SetOrder("v.name ASC")

	result, err := s.DB.Query(q.GetSQL(), q.GetValues()...)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	defer result.Close()

	venues := make([]*Venue, 0)

	for result.Next() {
		if err := result.Err(); err != nil {
			return nil, err
		}

		//get venue data
		venue := &Venue{Events: make([]*Event, 0)}
		if err := result.Scan(&venue.ID, &venue.Name, &venue.Address); err != nil {
			return venues, err
		}

		//get event data for each venue
		events, err := s.FindEvents(&EventFilter{VenueIDs: []int{int(venue.ID)}})
		if err != nil {
			events = []*Event{}
		}

		//append event sans venue data
		for _, event := range events {
			event.Venue = nil
		}
		venue.Events = events

		venues = append(venues, venue)
	}

	return venues, nil
}

func (s *EventStore) UpsertEvent(event *Event) error {

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
			"INSERT INTO performer (name, genre, home) VALUES (?, ?, ?)",
			performer.Name,
			performer.Genre,
			performer.Home,
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
