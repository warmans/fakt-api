package handler

import (
	"log"
	"net/http"

	"github.com/warmans/ctxhandler"
	"github.com/warmans/fakt-api/server/api.v1/common"
	"golang.org/x/net/context"
	"github.com/warmans/fakt-api/server/data/service/event"
)

func NewEventTypeHandler(ds *event.EventService) ctxhandler.CtxHandler {
	return &EventTypeHandler{ds: ds}
}

type EventTypeHandler struct {
	ds *event.EventService
}

func (h *EventTypeHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request, ctx context.Context) {
	events, err := h.ds.FindEventTypes()
	if err != nil {
		log.Print(err.Error())
		common.SendResponse(rw, &common.Response{Status: http.StatusInternalServerError, Payload: nil})
		return
	}
	common.SendResponse(rw, &common.Response{Status: http.StatusOK, Payload: events})
}
