package handler

import (
	"net/http"

	"github.com/warmans/fakt-api/pkg/server/api.v1/common"
	"github.com/warmans/fakt-api/pkg/server/data/service/event"
	"github.com/warmans/route-rest/routes"
	"github.com/gorilla/mux"
	"strconv"
	"github.com/warmans/fakt-api/pkg/server/api.v1/middleware"
)

func NewEventSimilarHandler(ds *event.EventService) routes.RESTHandler {
	return &EventSimilarHandler{ds: ds}
}

type EventSimilarHandler struct {
	routes.DefaultRESTHandler
	ds *event.EventService
}

func (h *EventSimilarHandler) HandleGetList(rw http.ResponseWriter, r *http.Request) {

	logger := middleware.MustGetLogger(r)

	vars := mux.Vars(r)
	eventId, err := strconv.Atoi(vars["event_id"])
	if err != nil {
		common.SendError(rw, common.HTTPError{"Invalid eventID", http.StatusBadRequest, err}, nil)
		return
	}

	similarEvents, err := h.ds.FindSimilarEventIDs(int64(eventId))
	if err != nil {
		common.SendError(rw, err, logger)
		return
	}
	if len(similarEvents) == 0 {
		common.SendResponse(rw, &common.Response{Status: http.StatusOK, Payload: []struct{}{}})
		return
	}

	f := event.EventFilterFromRequest(r)
	f.IDs = similarEvents

	events, err := h.ds.FindEvents(f)
	if err != nil {
		common.SendError(rw, err, logger)
		return
	}
	common.SendResponse(rw, &common.Response{Status: http.StatusOK, Payload: events})
}
