package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/warmans/fakt-api/server/api.v1/common"

	"context"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
	"github.com/warmans/fakt-api/server/data"
	"github.com/warmans/fakt-api/server/data/service/event"
	"github.com/warmans/route-rest/routes"
)

func NewEventHandler(ds *event.EventService) routes.RESTHandler {
	return &EventHandler{es: ds}
}

type EventHandler struct {
	es     *event.EventService
	ingest *data.Ingest

	routes.DefaultRESTHandler
}

func (h *EventHandler) HandleGet(rw http.ResponseWriter, r *http.Request) {

	logger, ok := r.Context().Value("logger").(log.Logger)
	if !ok {
		panic("Context must contain a logger")
	}

	vars := mux.Vars(r)
	eventId, err := strconv.Atoi(vars["event_id"])
	if err != nil {
		common.SendError(rw, common.HTTPError{"Invalid eventID", http.StatusBadRequest, err}, nil)
		return
	}

	events, err := h.es.FindEvents(&event.EventFilter{EventIDs: []int{eventId}})
	if err != nil {
		common.SendError(rw, err, logger)
		return
	}

	if len(events) < 1 {
		common.SendError(rw, common.HTTPError{"Event not found", http.StatusNotFound, err}, nil)
		return
	}

	common.SendResponse(rw, &common.Response{Status: http.StatusOK, Payload: events})
}

func (h *EventHandler) HandleGetList(rw http.ResponseWriter, r *http.Request) {

	logger, ok := r.Context().Value("logger").(log.Logger)
	if !ok {
		panic("Context must contain a logger")
	}

	events, err := h.es.FindEvents(h.filterFromRequest(r, r.Context()))
	if err != nil {
		common.SendError(rw, err, logger)
		return
	}
	common.SendResponse(rw, &common.Response{Status: http.StatusOK, Payload: events})
}



func (h *EventHandler) filterFromRequest(r *http.Request, ctx context.Context) *event.EventFilter {

	filter := &event.EventFilter{
		EventIDs:          make([]int, 0),
		VenueIDs:          make([]int64, 0),
		Types:             make([]string, 0),
		ShowDeleted:       false,
		LoadPerformerTags: false,
	}

	if eventIds := r.Form.Get("ids"); eventIds != "" {
		for _, idStr := range strings.Split(eventIds, ",") {
			if idInt, err := strconv.Atoi(idStr); err == nil {
				filter.EventIDs = append(filter.EventIDs, idInt)
			}
		}
	}
	if from := r.Form.Get("from"); from != "" {
		if dateFrom, err := time.Parse("2006-01-02", from); err == nil {
			filter.DateFrom = dateFrom
		}
	}
	if to := r.Form.Get("to"); to != "" {
		if dateTo, err := time.Parse("2006-01-02", to); err == nil {
			filter.DateTo = dateTo
		}
	}
	if dateRelative := r.Form.Get("date_relative"); dateRelative != "" {
		filter.DateFrom, filter.DateTo = common.GetRelativeDateRange(dateRelative)
	}

	//if nodate range is specified limit 1 month
	if filter.DateFrom.IsZero() && filter.DateTo.IsZero() {
		filter.DateTo = time.Now().Add(time.Hour * 24 * 7 * 4)
	}

	//only allow max 3 months of data to be returned
	maxDate := time.Now().Add(time.Hour * 24 * 7 * 4 * 3)
	if filter.DateTo.After(maxDate) {
		filter.DateTo = maxDate
	}

	if venue := r.Form.Get("venue"); venue != "" {
		for _, idStr := range strings.Split(venue, ",") {
			if idInt, err := strconv.Atoi(idStr); err == nil {
				filter.VenueIDs = append(filter.VenueIDs, int64(idInt))
			}
		}
	}
	if tpe := r.Form.Get("type"); tpe != "" {
		for _, typeStr := range strings.Split(tpe, ",") {
			filter.Types = append(filter.Types, typeStr)
		}
	}

	if deleted := r.Form.Get("deleted"); deleted == "1" || deleted == "true" {
		filter.ShowDeleted = true
	}

	if perfTags := r.Form.Get("performer_tags"); perfTags == "1" || perfTags == "true" {
		filter.LoadPerformerTags = true
	}

	//limit to events with only these tags
	filter.Tag = r.Form.Get("tag")

	//additionally only look for tags from a specific user
	filter.TagUser = r.Form.Get("tag_user")

	return filter
}
