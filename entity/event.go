package entity

import (
	"log"
	"regexp"
	"strings"
	"time"
	"github.com/warmans/dbr"
	"github.com/warmans/dbr/dialect"
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

type EventFilter struct {
	EventIDs    []int     `json:"events"`
	DateFrom    time.Time `json:"from_date"`
	DateTo      time.Time `json:"to_date"`
	VenueIDs    []int64     `json:"venues"`
	Types       []string  `json:"types"`
	ShowDeleted bool      `json:"show_deleted"`
}

type Venue struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Address string `json:"address"`
	Events  []*Event `json:"event,omitempty"`
}

type VenueFilter struct {
	VenueIDs []int     `json:"venues"`
	Name     string    `json:"name"`
}

type Performer struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Genre     string `json:"genre"`
	Home      string `json:"home"`
	ListenURL string `json:"listen_url"`
	Events    []*Event `json:"event,omitempty"`
}

type PerformerFilter struct {
	PerformerID []int `json:"performers"`
	Name     string    `json:"name"`
	Genre    string    `json:"name"`
	Home     string    `json:"name"`
}

type EventStore struct {
	DB *dbr.Session
}

func (s *EventStore) FindEvents(filter *EventFilter) ([]*Event, error) {

	q := s.DB.Select("event.id", "event.date", "event.type", "event.description", "venue.id", "venue.name", "venue.address", "performer.id", "performer.name", "performer.genre", "performer.home", "performer.listen_url")
	q.From("event")
	q.LeftJoin("venue", "event.venue_id = venue.id")
	q.LeftJoin("event_performer", "event.id = event_performer.event_id")
	q.LeftJoin("performer", "event_performer.performer_id = performer.id")
	q.OrderBy("event.date").OrderBy("event.id").OrderBy("venue.id").OrderBy("performer.id")

	if len(filter.EventIDs) > 0 {
		q.Where("event.id IN ?", filter.EventIDs)
	}
	if len(filter.Types) > 0 {
		q.Where("event.type IN ?", filter.Types)
	}
	if len(filter.VenueIDs) > 0 {
		q.Where("venue.id IN ?", filter.VenueIDs)
	}
	if !filter.DateFrom.IsZero() {
		q.Where("event.date >= ?", filter.DateFrom)
	}
	if !filter.DateTo.IsZero() {
		q.Where("event.date < ?", filter.DateTo)
	}
	q.Where("event.deleted <= ?", IfOrInt(filter.ShowDeleted, 1, 0))

	// :/
	sql, vals := q.ToSql()
	interpolated, err := dbr.InterpolateForDialect(sql, vals, dialect.SQLite3)
	if err != nil {
		return nil, err
	}

	result, err := s.DB.Query(interpolated)
	if err != nil && err != dbr.ErrNotFound {
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

	q := s.DB.Select("v.id", "v.name", "v.address").
		 From("venue v").
		 OrderBy("v.name")

	if len(filter.VenueIDs) > 0 {
		q.Where("v.id IN ?", filter.VenueIDs)
	}
	if filter.Name != "" {
		q.Where("v.name = ?", filter.Name)
	}

	venues := make([]*Venue, 0)
	if _, err := q.Load(&venues); err != nil && err != dbr.ErrNotFound {
		return nil, err
	}

	for k, venue := range venues {

		//get event data for each venue
		events, err := s.FindEvents(&EventFilter{VenueIDs: []int64{venue.ID}})
		if err != nil {
			events = []*Event{}
		}

		//append event sans venue data
		for _, event := range events {
			event.Venue = nil
		}
		venues[k].Events = events
	}

	return venues, nil
}

func (s *EventStore) FindPerformers(filter *PerformerFilter) ([]*Performer, error) {

	q := s.DB.Select("id", "name", "genre", "home", "listen_url").
	From("performer p").
	OrderBy("p.name")

	if len(filter.PerformerID) > 0 {
		q.Where("p.id IN ?", filter.PerformerID)
	}
	if filter.Name != "" {
		q.Where("p.name = ?", filter.Name)
	}
	if filter.Home != "" {
		q.Where("p.home = ?", filter.Home)
	}
	if filter.Genre != "" {
		q.Where("p.genre = ?", filter.Genre)
	}
	performers := make([]*Performer, 0)
	if _, err := q.Load(&performers); err != nil && err != dbr.ErrNotFound {
		return nil, err
	}
	return performers, nil
}