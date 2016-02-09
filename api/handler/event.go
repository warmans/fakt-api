package handler

import (
	"net/http"
	"github.com/warmans/stressfaktor-api/entity"
	"log"
	"github.com/warmans/stressfaktor-api/api/common"
	"strings"
	"strconv"
	"time"
)

func NewEventHandler(eventStore *entity.EventStore) http.Handler {
	return &EventHandler{eventStore: eventStore}
}

type EventHandler struct {
	eventStore *entity.EventStore
}

func (h *EventHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {

	defer r.Body.Close()
	events, err := h.eventStore.Find(h.filterFromRequest(r))
	if err != nil {
		log.Print(err.Error())
		http.Error(rw, "Failed", http.StatusInternalServerError)
		return
	}

	common.SendResponse(rw, &common.Response{Status: 200, Payload: events})
}

func (h *EventHandler) filterFromRequest(r *http.Request) *entity.EventFilter {
	r.ParseForm()

	filter := &entity.EventFilter{
		EventIDs: make([]int, 0),
		VenueIDs: make([]int, 0),
		PerformerIDs: make([]int, 0),
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
				filter.VenueIDs = append(filter.VenueIDs, idInt)
			}
		}
	}
	if performer := r.Form.Get("performer"); performer != "" {
		for _, idStr := range strings.Split(performer, ",") {
			if idInt, err := strconv.Atoi(idStr); err == nil {
				filter.PerformerIDs = append(filter.PerformerIDs, idInt)
			}
		}
	}

	return filter
}
