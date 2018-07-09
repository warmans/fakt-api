package handler

import (
	"net/http"

	"github.com/warmans/fakt-api/pkg/server/api.v1/common"
	"github.com/warmans/fakt-api/pkg/server/data/service/event"
	"github.com/warmans/route-rest/routes"
	"github.com/warmans/fakt-api/pkg/server/api.v1/middleware"
)

func NewEventTypeHandler(ds *event.EventService) routes.RESTHandler {
	return &EventTypeHandler{ds: ds}
}

type EventTypeHandler struct {
	routes.DefaultRESTHandler
	ds *event.EventService
}

func (h *EventTypeHandler) HandleGetList(rw http.ResponseWriter, r *http.Request) {

	logger := middleware.MustGetLogger(r)

	events, err := h.ds.FindEventTypes()
	if err != nil {
		common.SendError(rw, err, logger)
		return
	}
	common.SendResponse(rw, &common.Response{Status: http.StatusOK, Payload: events})
}
