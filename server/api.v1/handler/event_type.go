package handler

import (
	"net/http"

	"github.com/warmans/ctxhandler"
	"github.com/warmans/fakt-api/server/api.v1/common"
	"golang.org/x/net/context"
	"github.com/warmans/fakt-api/server/data/service/event"
	"github.com/go-kit/kit/log"
)

func NewEventTypeHandler(ds *event.EventService) ctxhandler.CtxHandler {
	return &EventTypeHandler{ds: ds}
}

type EventTypeHandler struct {
	ds *event.EventService
}

func (h *EventTypeHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request, ctx context.Context) {

	logger, ok := ctx.Value("logger").(log.Logger)
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
