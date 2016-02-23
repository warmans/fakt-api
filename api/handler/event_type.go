package handler

import (
	"net/http"
	"github.com/warmans/stressfaktor-api/entity"
	"log"
	"github.com/warmans/stressfaktor-api/api/common"
)

func NewEventTypeHandler(eventStore *entity.EventStore) http.Handler {
	return &EventTypeHandler{eventStore: eventStore}
}

type EventTypeHandler struct {
	eventStore *entity.EventStore
}

func (h *EventTypeHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	events, err := h.eventStore.FindEventTypes()
	if err != nil {
		log.Print(err.Error())
		common.SendResponse(rw, &common.Response{Status: http.StatusInternalServerError, Payload: nil})
		return
	}
	common.SendResponse(rw, &common.Response{Status: http.StatusOK, Payload: events})
}
