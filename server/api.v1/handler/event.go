package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/warmans/fakt-api/server/api.v1/common"
	"github.com/warmans/resty"

	"github.com/go-kit/kit/log"
	"github.com/warmans/fakt-api/server/data"
	"github.com/warmans/fakt-api/server/data/service/event"
	"golang.org/x/net/context"
)

func NewEventHandler(ds *event.EventService) resty.RESTHandler {
	return &EventHandler{es: ds}
}

type EventHandler struct {
	resty.DefaultRESTHandler
	es     *event.EventService
	ingest *data.Ingest
}

func (h *EventHandler) HandleGet(rw http.ResponseWriter, r *http.Request, ctx context.Context) {

	logger, ok := ctx.Value("logger").(log.Logger)
	if !ok {
		panic("Context must contain a logger")
	}

	events, err := h.es.FindEvents(h.filterFromRequest(r, ctx))
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

	if event := r.Form.Get("event"); event != "" {
		for _, idStr := range strings.Split(event, ",") {
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
