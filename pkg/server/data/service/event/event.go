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
	"github.com/warmans/fakt-api/pkg/server/data/service/common"
	"github.com/warmans/fakt-api/pkg/server/data/service/performer"
	"github.com/warmans/fakt-api/pkg/server/data/service/utag"
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
	UTags             []string  `json:"utag"`
	Tags              []string  `json:"tag"`
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

	//limit to events with only this user tags
	if r.Form.Get("utag") == "" {
		f.UTags = make([]string, 0, 0)
	} else {
		f.UTags = strings.Split(r.Form.Get("utag"), ",")
	}

	//only this tag
	if r.Form.Get("tag") == "" {
		f.Tags = make([]string, 0, 0)
	} else {
		f.Tags = strings.Split(r.Form.Get("tag"), ",")
	}

	//additionally only look for tags from a specific user
	f.UTagUser = r.Form.Get("tag_user")
}

type EventService struct {
	DB               *dbr.Session
	UTagService      *utag.UTagService
	PerformerService *performer.PerformerService
}

func (s *EventService) EventMustExist(tr *dbr.Tx, event *common.Event) error {

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

	//finally append the performers
	for _, perfs := range event.Performers {
		if perfs.ID == 0 {
			continue
		}
		//make the association
		_, err := tr.Exec("REPLACE INTO event_performer (event_id, performer_id) VALUES (?, ?)", event.ID, perfs.ID)
		if err != nil {
			return err
		}
	}

	//and the tags...
	if err := s.StoreEventTags(tr, event.ID, event.Tags); err != nil {
		return err
	}

	return nil
}

func (s *EventService) StoreEventTags(tr *dbr.Tx, eventID int64, tags []string) error {
	//handle tags
	if _, err := tr.Exec("DELETE FROM event_tag WHERE event_id = ?", eventID); err != nil {
		return fmt.Errorf("Failed to clear existing tags due to error: %s", err.Error())
	}

	for _, tag := range tags {

		var tagId int64
		tag = strings.ToLower(tag)

		err := s.DB.QueryRow("SELECT id FROM tag WHERE tag = ?", tag).Scan(&tagId)
		if err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("Failed to find tag id for %s because %s", tag, err.Error())
		}
		if tagId == 0 {
			res, err := tr.Exec("INSERT OR IGNORE INTO tag (tag) VALUES (?)", tag)
			if err != nil {
				return fmt.Errorf("Failed to insert tag %s because %s", tag, err.Error())
			}
			tagId, err = res.LastInsertId()
			if err != nil {
				return fmt.Errorf("Failed to get inserted tag id because %s", err.Error())
			}
		}

		if _, err := tr.Exec("INSERT OR IGNORE INTO event_tag (event_id, tag_id) VALUES (?, ?)", eventID, tagId); err != nil {
			return fmt.Errorf("Failed to insert event_tag relationship (event: %d, tag: %s, tagId: %d) because %s", eventID, tag, tagId, err.Error())
		}
	}
	return nil
}

func (s *EventService) FindEventTags(eventID int64) ([]string, error) {

	tags := []string{}

	res, err := s.DB.Query("SELECT coalesce(t.tag, '') FROM event_tag et LEFT JOIN tag t ON et.tag_id = t.id WHERE et.event_id = ?", eventID)
	if err != nil {
		if err == sql.ErrNoRows {
			return tags, nil
		}
		return tags, fmt.Errorf("Failed to fetch tags at query because %s", err.Error())
	}

	for res.Next() {
		tag := ""
		if err := res.Scan(&tag); err != nil {
			return tags, err
		}
		tags = append(tags, tag)
	}

	return tags, nil
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

func (s *EventService) FindSimilarEventIDs(eventID int64) ([]int64, error) {

	similarEvents := []int64{}

	res, err := s.DB.Query(`
		SELECT
			ep2.event_id AS similar_event_id
		FROM event_performer ep1
		LEFT JOIN performer_tag pt1 ON ep1.performer_id = pt1.performer_id
		LEFT JOIN performer_tag pt2 ON pt1.tag_id = pt2.tag_id
		LEFT JOIN event_performer ep2 ON pt2.performer_id = ep2.performer_id
		WHERE ep1.event_id = ? AND ep2.event_id != ?
		GROUP BY ep2.event_id
		ORDER BY SUM(1) DESC
	`, eventID, eventID)
	if err != nil {
		if err == sql.ErrNoRows {
			return similarEvents, nil
		}
		return similarEvents, fmt.Errorf("Failed to find similar events for id %d. Reason: %s", eventID, err.Error())
	}

	for res.Next() {
		var curEventID int64
		if err := res.Scan(&curEventID); err != nil {
			return similarEvents, fmt.Errorf("Failed to find similar events for id %d. Reason: %s", eventID, err.Error())
		}
		similarEvents = append(similarEvents, curEventID)
	}

	return similarEvents, nil
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

	if filter.PageSize != 0 {
		q.Limit(uint64(filter.PageSize)).Offset(uint64((filter.PageSize * page) - filter.PageSize))
	}

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

	if len(filter.Tags) > 0 {
		q.LeftJoin("event_tag", "event.id = event_tag.event_id")
		q.Where("event_tag.tag_id IN ?", filter.Tags)
	}

	if len(filter.UTags) > 0 {
		q.LeftJoin("event_user_tag", "event.id = event_user_tag.event_id")
		q.Where("event_user_tag.tag_id IN ?", filter.UTags)
	}

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
				events = append(events, curEvent)
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
				pf := &performer.PerformerFilter{}
				pf.IDs = performerIDs
				if curEvent.Performers, err = s.PerformerService.FindPerformers(pf); err != nil {
					return nil, err
				}
			}

			//append the user tags
			if curEvent.UTags, err = s.UTagService.FindEventUTags(curEvent.ID, &common.UTagsFilter{}); err != nil {
				return nil, err
			}

			//append the tags
			if curEvent.Tags, err = s.FindEventTags(curEvent.ID); err != nil {
				return nil, err
			}

		}
	}

	if curEvent.ID != 0 {
		events = append(events, curEvent)
	}
	return events, nil
}
