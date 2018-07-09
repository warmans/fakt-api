package handler

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/warmans/fakt-api/pkg/server/api.v1/common"
	"github.com/warmans/fakt-api/pkg/server/api.v1/middleware"
	"github.com/warmans/fakt-api/pkg/server/data/store/event"
	"github.com/warmans/fakt-api/pkg/server/data/store/venue"
	"github.com/warmans/route-rest/routes"
)

func NewVenueEventHandler(es *event.Store, vs *venue.Store) routes.RESTHandler {
	return &VenueEventHandler{events: es, venues: vs}
}

type VenueEventHandler struct {
	routes.DefaultRESTHandler
	events *event.Store
	venues *venue.Store
}

func (h *VenueEventHandler) HandleGetList(rw http.ResponseWriter, r *http.Request) {

	logger := middleware.MustGetLogger(r)

	vars := mux.Vars(r)
	venueID, err := strconv.Atoi(vars["venue_id"])
	if err != nil {
		common.SendError(rw, common.HTTPError{"Invalid performerID", http.StatusBadRequest, err}, nil)
		return
	}

	filter := event.FilterFromRequest(r)
	filter.VenueIDs = []int64{int64(venueID)}

	events, err := h.events.FindEvents(filter)
	if err != nil {
		common.SendError(rw, err, logger)
		return
	}

	common.SendResponse(
		rw,
		&common.Response{
			Status:  http.StatusOK,
			Payload: events,
		},
	)
}
