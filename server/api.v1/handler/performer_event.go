package handler

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/warmans/fakt-api/server/api.v1/common"
	"github.com/warmans/fakt-api/server/data/service/event"
	"github.com/warmans/fakt-api/server/data/service/performer"
	"github.com/warmans/route-rest/routes"
	"github.com/warmans/fakt-api/server/api.v1/middleware"
)

func NewPerformerEventHandler(es *event.EventService, ps *performer.PerformerService) routes.RESTHandler {
	return &PerformerEventHandler{events: es, performers: ps}
}

type PerformerEventHandler struct {
	routes.DefaultRESTHandler
	events     *event.EventService
	performers *performer.PerformerService
}

func (h *PerformerEventHandler) HandleGetList(rw http.ResponseWriter, r *http.Request) {

	logger := middleware.MustGetLogger(r)

	vars := mux.Vars(r)
	performerID, err := strconv.Atoi(vars["performer_id"])
	if err != nil {
		common.SendError(rw, common.HTTPError{"Invalid performerID", http.StatusBadRequest, err}, nil)
		return
	}

	eventIDs, err := h.performers.FindPerformerEventIDs(int64(performerID))
	if err != nil {
		common.SendError(rw, err, logger)
		return
	}

	if len(eventIDs)  == 0 {
		common.SendResponse(
			rw,
			&common.Response{
				Status:  http.StatusOK,
				Payload: []struct{}{},
			},
		)
		return
	}

	filter := event.EventFilterFromRequest(r)
	filter.IDs = eventIDs

	events, err := h.events.FindEvents(filter)
	if err != nil {
		common.SendError(rw, err, logger)
		return
	}

	common.SendResponse(
		rw,
		&common.Response{
			Status:  http.StatusOK,
			Payload: events,
		},
	)
}
