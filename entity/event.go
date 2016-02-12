package entity

import (
	"database/sql"
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

func (e *Event) Accept(visitor EventVisitor) {
	visitor.Visit(e)
}

type Venue struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Address string `json:"address"`
	Events  []*Event `json:"event,omitempty"`
}

type VenueFilter struct {
	VenueIDs []int     `json:"venues"`
}

type Performer struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Genre     string `json:"genre"`
	Home      string `json:"home"`
	ListenURL string `json:"listen_url"`
	Events    []*Event `json:"event,omitempty"`
}

type EventFilter struct {
	EventIDs    []int     `json:"events"`
	DateFrom    time.Time `json:"from_date"`
	DateTo      time.Time `json:"to_date"`
	VenueIDs    []int     `json:"venues"`
	Types       []string  `json:"types"`
	ShowDeleted bool       `json:"show_deleted"`
}

type EventStore struct {
	DB *sql.DB
}

func (s *EventStore) FindEvents(filter *EventFilter) ([]*Event, error) {

	q := Sql{}
	q.Select("e.id", "e.date", "e.type", "e.description", "v.id", "v.name", "v.address", "p.id", "p.name", "p.genre", "p.home", "p.listen_url")
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
	q.SetOrder("e.date", "e.id", "v.id", "p.id")

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
		var eType, eDescription, vName, vAddress, pName, pGenre, pHome, pListen string
		var eDate time.Time

		result.Scan(&eID, &eDate, &eType, &eDescription, &vID, &vName, &vAddress, &pID, &pName, &pGenre, &pHome, &pListen)

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
			ListenURL: pListen,
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
