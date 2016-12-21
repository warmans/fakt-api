package handler

import (
	"net/http"
	"strconv"

	"github.com/warmans/fakt-api/server/api.v1/common"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
	"github.com/warmans/fakt-api/server/data/service/event"
	"github.com/warmans/route-rest/routes"
)

func NewEventHandler(ds *event.EventService) routes.RESTHandler {
	return &EventHandler{es: ds}
}

type EventHandler struct {
	es *event.EventService
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

	events, err := h.es.FindEvents(event.EventFilterFromRequest(r))
	if err != nil {
		common.SendError(rw, err, logger)
		return
	}
	common.SendResponse(rw, &common.Response{Status: http.StatusOK, Payload: events})
}
