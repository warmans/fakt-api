package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/go-kit/kit/log"
	"github.com/warmans/fakt-api/server/api.v1/common"
	"github.com/warmans/fakt-api/server/data/service/venue"
)

func NewVenueHandler(ds *venue.VenueService) http.Handler {
	return &VenueHandler{ds: ds}
}

type VenueHandler struct {
	ds *venue.VenueService
}

func (h *VenueHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {

	logger, ok := r.Context().Value("logger").(log.Logger)
	if !ok {
		panic("Context must contain a logger")
	}

	//query to filter
	filter := &venue.VenueFilter{VenueIDs: make([]int, 0), Name: r.Form.Get("name")}
	if ven := r.Form.Get("venue"); ven != "" {
		for _, idStr := range strings.Split(ven, ",") {
			if idInt, err := strconv.Atoi(idStr); err == nil {
				filter.VenueIDs = append(filter.VenueIDs, idInt)
			}
		}
	}

	venues, err := h.ds.FindVenues(filter)
	if err != nil {
		common.SendError(rw, err, logger)
		return
	}

	common.SendResponse(rw, &common.Response{Status: http.StatusOK, Payload: venues})
}
