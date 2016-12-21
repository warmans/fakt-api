package event

import (
	"database/sql"
	"fmt"
	"time"

	"net/http"
	"strconv"
	"strings"

	"github.com/warmans/dbr"
	"github.com/warmans/dbr/dialect"
	"github.com/warmans/fakt-api/server/data/service/common"
	"github.com/warmans/fakt-api/server/data/service/performer"
	"github.com/warmans/fakt-api/server/data/service/utag"
)

const DefaultPageSize = 50

func EventFilterFromRequest(r *http.Request) *EventFilter {
	f := &EventFilter{}
	f.Populate(r)
	return f
}

type EventFilter struct {

	common.CommonFilter

	DateFrom          time.Time `json:"from_date"`
	DateTo            time.Time `json:"to_date"`
	DateRelative      string    `json:"date_relative"`
	VenueIDs          []int64   `json:"venues"`
	Types             []string  `json:"types"`
	ShowDeleted       bool      `json:"show_deleted"`
	UTag              string    `json:"utag"`
	UTagUser          string    `json:"utag_user"`
	LoadPerformerTags bool      `json:"load_performer_tags"`
	Source            string    `json:"source"`
}

func (f *EventFilter) Populate(r *http.Request) {

	f.CommonFilter.Populate(r)

	if from := r.Form.Get("from"); from != "" {
		if dateFrom, err := time.Parse("2006-01-02", from); err == nil {
			f.DateFrom = dateFrom
		}
	}
	if to := r.Form.Get("to"); to != "" {
		if dateTo, err := time.Parse("2006-01-02", to); err == nil {
			f.DateTo = dateTo
		}
	}
	if dateRelative := r.Form.Get("date_relative"); dateRelative != "" {
		f.DateFrom, f.DateTo = common.GetRelativeDateRange(dateRelative)
	}

	//only allow max 3 months of data to be returned
	maxDate := time.Now().Add(time.Hour * 24 * 7 * 4 * 3)
	if f.DateTo.After(maxDate) {
		f.DateTo = maxDate
	}

	f.VenueIDs = make([]int64, 0)
	if venue := r.Form.Get("venue"); venue != "" {
		for _, idStr := range strings.Split(venue, ",") {
			if idInt, err := strconv.Atoi(idStr); err == nil {
				f.VenueIDs = append(f.VenueIDs, int64(idInt))
			}
		}
	}

	f.Types = make([]string, 0)
	if tpe := r.Form.Get("type"); tpe != "" {
		for _, typeStr := range strings.Split(tpe, ",") {
			f.Types = append(f.Types, typeStr)
		}
	}

	if deleted := r.Form.Get("deleted"); deleted == "1" || deleted == "true" {
		f.ShowDeleted = true
	}

	if perfTags := r.Form.Get("performer_tags"); perfTags == "1" || perfTags == "true" {
		f.LoadPerformerTags = true
	}

	//limit to events with only these tags
	f.UTag = r.Form.Get("tag")

	//additionally only look for tags from a specific user
	f.UTagUser = r.Form.Get("tag_user")
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
		err := tr.QueryRow("SELECT id FROM event WHERE venue_id=? AND date=?", event.Venue.ID, event.Date.Format(common.DateFormatSQL)).Scan(&event.ID)
		if err != nil && err != sql.ErrNoRows {
			return err
		}
	}

	//update/create the main event record
	if event.ID == 0 {

		res, err := tr.Exec(
			"INSERT INTO event (date, venue_id, type, description) VALUES (?, ?, ?, ?)",
			event.Date.Format(common.DateFormatSQL),
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

	//if no page is specified assume the first page
	page := filter.Page
	if page == 0 {
		page = 1
	}

	q := s.DB.Select(
		"event.id",
		"event.date",
		"event.type",
		"event.description",
		"coalesce(event.source, '')",
		"coalesce(venue.id, 0)",
		"venue.name",
		"venue.address",
		"coalesce(group_concat(event_performer.performer_id), '')",
	)
	q.From("event")
	q.LeftJoin("venue", "event.venue_id = venue.id")
	q.LeftJoin("event_performer", "event.id = event_performer.event_id")
	q.OrderBy("event.date").OrderBy("event.id").OrderBy("venue.id")
	q.GroupBy("event.id")
	q.Limit(uint64(filter.PageSize))
	q.Offset(uint64((page * filter.PageSize) - filter.PageSize))

	if len(filter.IDs) > 0 {
		q.Where("event.id IN ?", filter.IDs)
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
	q.Where("event.deleted = ?", common.IfOrInt(filter.ShowDeleted, 1, 0))

	sqlString, vals := q.ToSql()
	interpolated, err := dbr.InterpolateForDialect(sqlString, vals, dialect.SQLite3)
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
				if curEvent.HasUTag(filter.UTag, filter.UTagUser) {
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
		if curEvent.HasUTag(filter.UTag, filter.UTagUser) {
			events = append(events, curEvent)
		}
	}
	return events, nil
}
