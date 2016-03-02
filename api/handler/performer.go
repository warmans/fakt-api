package handler

import (
	"net/http"
	"github.com/warmans/stressfaktor-api/api/common"
	"log"
	"strings"
	"strconv"
	"github.com/warmans/stressfaktor-api/data/store"
)

func NewPerformerHandler(eventStore *store.Store) http.Handler {
	return &PerformerHandler{eventStore: eventStore}
}

type PerformerHandler struct {
	eventStore *store.Store
}

func (h *PerformerHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	r.ParseForm()

	//query to filter
	filter := &store.PerformerFilter{PerformerID: make([]int, 0)}
	if venue := r.Form.Get("performer"); venue != "" {
		for _, idStr := range strings.Split(venue, ",") {
			if idInt, err := strconv.Atoi(idStr); err == nil {
				filter.PerformerID = append(filter.PerformerID, idInt)
			}
		}
	}
	filter.Name = r.Form.Get("name")
	filter.Genre = r.Form.Get("genre")
	filter.Home = r.Form.Get("home")

	venues, err := h.eventStore.FindPerformers(filter)
	if err != nil {
		log.Print(err.Error())
		common.SendResponse(rw, &common.Response{Status: http.StatusInternalServerError, Payload: nil})
		return
	}

	common.SendResponse(rw, &common.Response{Status: http.StatusOK, Payload: venues})
}
