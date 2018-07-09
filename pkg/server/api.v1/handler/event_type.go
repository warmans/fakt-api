package handler

import (
	"net/http"

	"github.com/warmans/fakt-api/pkg/server/api.v1/common"
	"github.com/warmans/fakt-api/pkg/server/api.v1/middleware"
	"github.com/warmans/fakt-api/pkg/server/data/store/event"
	"github.com/warmans/route-rest/routes"
)

func NewEventTypeHandler(ds *event.Store) routes.RESTHandler {
	return &EventTypeHandler{ds: ds}
}

type EventTypeHandler struct {
	routes.DefaultRESTHandler
	ds *event.Store
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
