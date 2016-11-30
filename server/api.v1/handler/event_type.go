package handler

import (
	"net/http"

	"github.com/go-kit/kit/log"
	"github.com/warmans/fakt-api/server/api.v1/common"
	"github.com/warmans/fakt-api/server/data/service/event"
	"github.com/warmans/route-rest/routes"
)

func NewEventTypeHandler(ds *event.EventService) routes.RESTHandler {
	return &EventTypeHandler{ds: ds}
}

type EventTypeHandler struct {
	routes.DefaultRESTHandler
	ds *event.EventService
}

func (h *EventTypeHandler) HandleGetList(rw http.ResponseWriter, r *http.Request) {

	logger, ok := r.Context().Value("logger").(log.Logger)
	if !ok {
		panic("Context must contain a logger")
	}

	events, err := h.ds.FindEventTypes()
	if err != nil {
		common.SendError(rw, err, logger)
		return
	}
	common.SendResponse(rw, &common.Response{Status: http.StatusOK, Payload: events})
}
