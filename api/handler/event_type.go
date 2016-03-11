package handler

import (
	"net/http"
	"log"
	"github.com/warmans/stressfaktor-api/api/common"
	"github.com/warmans/stressfaktor-api/data/store"
	"golang.org/x/net/context"
)

func NewEventTypeHandler(eventStore *store.Store) common.CtxHandler {
	return &EventTypeHandler{eventStore: eventStore}
}

type EventTypeHandler struct {
	eventStore *store.Store
}

func (h *EventTypeHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request, ctx context.Context) {
	defer r.Body.Close()
	events, err := h.eventStore.FindEventTypes()
	if err != nil {
		log.Print(err.Error())
		common.SendResponse(rw, &common.Response{Status: http.StatusInternalServerError, Payload: nil})
		return
	}
	common.SendResponse(rw, &common.Response{Status: http.StatusOK, Payload: events})
}
