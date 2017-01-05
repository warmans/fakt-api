package handler

import (
	"net/http"
	"strconv"

	"github.com/warmans/fakt-api/server/api.v1/common"

	"github.com/gorilla/mux"
	"github.com/warmans/fakt-api/server/data/service/event"
	"github.com/warmans/route-rest/routes"
	"github.com/warmans/fakt-api/server/api.v1/middleware"
)

func NewEventHandler(ds *event.EventService) routes.RESTHandler {
	return &EventHandler{es: ds}
}

type EventHandler struct {
	es *event.EventService
	routes.DefaultRESTHandler
}

func (h *EventHandler) HandleGet(rw http.ResponseWriter, r *http.Request) {

	logger := middleware.MustGetLogger(r)

	vars := mux.Vars(r)
	eventId, err := strconv.Atoi(vars["event_id"])
	if err != nil {
		common.SendError(rw, common.HTTPError{"Invalid eventID", http.StatusBadRequest, err}, nil)
		return
	}

	f := &event.EventFilter{}
	f.IDs = []int64{int64(eventId)}
	f.PageSize = 1
	f.Page = 1

	events, err := h.es.FindEvents(f)
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

	logger := middleware.MustGetLogger(r)

	events, err := h.es.FindEvents(event.EventFilterFromRequest(r))
	if err != nil {
		common.SendError(rw, err, logger)
		return
	}
	common.SendResponse(rw, &common.Response{Status: http.StatusOK, Payload: events})
}
