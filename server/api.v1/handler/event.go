package handler

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/warmans/resty"
	"github.com/warmans/stressfaktor-api/server/api.v1/common"
	"github.com/warmans/stressfaktor-api/server/data"
	"github.com/warmans/stressfaktor-api/server/data/store"
	"golang.org/x/net/context"
)

func NewEventHandler(ds *store.Store) resty.RESTHandler {
	return &EventHandler{ds: ds}
}

type EventHandler struct {
	resty.DefaultRESTHandler
	ds     *store.Store
	ingest *data.Ingest
}

func (h *EventHandler) HandleGet(rw http.ResponseWriter, r *http.Request, ctx context.Context) {
	events, err := h.ds.FindEvents(h.filterFromRequest(r, ctx))
	if err != nil {
		log.Print(err.Error())
		common.SendResponse(rw, &common.Response{Status: http.StatusInternalServerError, Payload: nil})
		return
	}
	common.SendResponse(rw, &common.Response{Status: http.StatusOK, Payload: events})
}

func (h *EventHandler) filterFromRequest(r *http.Request, ctx context.Context) *store.EventFilter {

	filter := &store.EventFilter{
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
