package event

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/warmans/dbr"
	"github.com/warmans/dbr/dialect"
	"github.com/warmans/fakt-api/server/data/service/common"
	"github.com/warmans/fakt-api/server/data/service/performer"
	"github.com/warmans/fakt-api/server/data/service/utag"
)

type EventFilter struct {
	EventIDs          []int     `json:"events"`
	DateFrom          time.Time `json:"from_date"`
	DateTo            time.Time `json:"to_date"`
	DateRelative      string    `json:"date_relative"`
	VenueIDs          []int64   `json:"venues"`
	Types             []string  `json:"types"`
	ShowDeleted       bool      `json:"show_deleted"`
	Tag               string    `json:"tag"`
	TagUser           string    `json:"tag_user"`
	LoadPerformerTags bool      `json:"load_performer_tags"`
	Source            string    `json:"source"`
}

type EventService struct {
	DB               *dbr.Session
	UTagService      *utag.UTagService
	PerformerService *performer.PerformerService
}

func (es *EventService) EventMustExist(tr *dbr.Tx, event *common.Event) error {

	//sanity check incoming data
	if !event.IsValid() || event.Venue == nil || !event.Venue.IsValid() {
		return fmt.Errorf("Invalid event/venue was rejected at injest. Event was %+v venue was %+v", event, event.Venue)
	}

	if event.ID == 0 {
		// if no ID was supplied try and find one based on the venue and date. i.e. assume two events cannot occur at the
		// same venue at the same time. Note that if either of these fields has been updated there will be no match
		// and the row will be processed as though it is a new record.
		err := tr.QueryRow("SELECT id FROM event WHERE venue_id=? AND date=?", event.Venue.ID, event.Date.Format(common.DATE_FORMAT_SQL)).Scan(&event.ID)
		if err != nil && err != sql.ErrNoRows {
			return err
		}
	}

	//update/create the main event record
	if event.ID == 0 {

		res, err := tr.Exec(
			"INSERT INTO event (date, venue_id, type, description) VALUES (?, ?, ?, ?)",
			event.Date.Format(common.DATE_FORMAT_SQL),
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

	//finally append the performers (
	for _, performer := range event.Performers {
		if performer.ID == 0 {
			continue
		}
		//make the association
		_, err := tr.Exec("REPLACE INTO event_performer (event_id, performer_id) VALUES (?, ?)", event.ID, performer.ID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *EventService) FindEventTypes() ([]string, error) {
	q := s.DB.
		Select("event.type").
		From("event").
		Where("event.deleted = 0").
		GroupBy("event.type").
		OrderDir("SUM(1)", false)

	types := make([]string, 0)
	if _, err := q.Load(&types); err != nil && err != dbr.ErrNotFound {
		return nil, err
	}
	return types, nil
}

//FindEvents is a slightly elaborate method to more efficiently fetch the majority of related event data
func (s *EventService) FindEvents(filter *EventFilter) ([]*common.Event, error) {

	q := s.DB.Select(
		"event.id",
		"event.date",
		"event.type",
		"event.description",
		"coalesce(event.source, '')",
		"coalesce(venue.id, 0)",
		"venue.name",
		"venue.address",
		"coalesce(group_concat(performer.id), '')",
	)
	q.From("event")
	q.LeftJoin("venue", "event.venue_id = venue.id")
	q.LeftJoin("event_performer", "event.id = event_performer.event_id")
	q.LeftJoin("performer", "event_performer.performer_id = performer.id")
	q.OrderBy("event.date").OrderBy("event.id").OrderBy("venue.id")
	q.GroupBy("event.id")

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
	if filter.Source != "" {
		q.Where("event.source = ?", filter.Source)
	}
	q.Where("event.deleted <= ?", common.IfOrInt(filter.ShowDeleted, 1, 0))

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

	events := make([]*common.Event, 0)
	curEvent := &common.Event{}

	for result.Next() {

		if err := result.Err(); err != nil {
			return nil, err
		}

		var eID, vID int
		var eType, eDescription, eSource, vName, vAddress, pIDs string
		var eDate time.Time

		err := result.Scan(&eID, &eDate, &eType, &eDescription, &eSource, &vID, &vName, &vAddress, &pIDs)
		if err != nil {
			return nil, err
		}

		if curEvent.ID != int64(eID) {

			//append to result set
			if curEvent.ID != 0 {
				if curEvent.HasTag(filter.Tag, filter.TagUser) {
					events = append(events, curEvent)
				}
			}

			//new current event
			curEvent = &common.Event{
				ID:          int64(eID),
				Date:        eDate,
				Type:        eType,
				Description: eDescription,
				Venue: &common.Venue{
					ID:      int64(vID),
					Name:    vName,
					Address: vAddress,
				},
				Source: eSource,
			}

			//append the performers
			if performerIDs := common.SplitConcatIDs(pIDs, ","); len(performerIDs) > 0 {
				if curEvent.Performers, err = s.PerformerService.FindPerformers(&performer.PerformerFilter{PerformerID: performerIDs}); err != nil {
					return nil, err
				}
			}

			//append the tags
			if curEvent.UTags, err = s.UTagService.FindEventUTags(curEvent.ID, &common.UTagsFilter{}); err != nil {
				return nil, err
			}
		}

	}

	if curEvent.ID != 0 {
		if curEvent.HasTag(filter.Tag, filter.TagUser) {
			events = append(events, curEvent)
		}
	}
	return events, nil
}
