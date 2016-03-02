package handler

import (
	"net/http"
	"github.com/warmans/stressfaktor-api/api/common"
	"log"
	"strings"
	"strconv"
	"github.com/warmans/stressfaktor-api/data/store"
)

func NewVenueHandler(eventStore *store.Store) http.Handler {
	return &VenueHandler{eventStore: eventStore}
}

type VenueHandler struct {
	eventStore *store.Store
}

func (h *VenueHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	r.ParseForm()

	//query to filter
	filter := &store.VenueFilter{VenueIDs: make([]int, 0)}
	if venue := r.Form.Get("venue"); venue != "" {
		for _, idStr := range strings.Split(venue, ",") {
			if idInt, err := strconv.Atoi(idStr); err == nil {
				filter.VenueIDs = append(filter.VenueIDs, idInt)
			}
		}
	}

	filter.Name = r.Form.Get("name")

	venues, err := h.eventStore.FindVenues(filter)
	if err != nil {
		log.Print(err.Error())
		common.SendResponse(rw, &common.Response{Status: http.StatusInternalServerError, Payload: nil})
		return
	}

	common.SendResponse(rw, &common.Response{Status: http.StatusOK, Payload: venues})
}
