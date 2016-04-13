package handler

import (
	"net/http"
	"log"
	"github.com/warmans/stressfaktor-api/api/common"
	"github.com/warmans/stressfaktor-api/data/store"
	"golang.org/x/net/context"
)

func NewEventTypeHandler(ds *store.Store) common.CtxHandler {
	return &EventTypeHandler{ds: ds}
}

type EventTypeHandler struct {
	ds *store.Store
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
