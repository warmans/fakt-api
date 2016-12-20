package handler

import (
	"net/http"
	"strconv"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
	"github.com/warmans/fakt-api/server/api.v1/common"
	"github.com/warmans/fakt-api/server/data/service/event"
	"github.com/warmans/fakt-api/server/data/service/venue"
	"github.com/warmans/route-rest/routes"
)

func NewVenueEventHandler(es *event.EventService, vs *venue.VenueService, logger log.Logger) routes.RESTHandler {
	return &VenueEventHandler{events: es, venues: vs, logger: logger}
}

type VenueEventHandler struct {
	routes.DefaultRESTHandler
	logger log.Logger
	events *event.EventService
	venues *venue.VenueService
}

func (h *VenueEventHandler) HandleGetList(rw http.ResponseWriter, r *http.Request) {

	logger, ok := r.Context().Value("logger").(log.Logger)
	if !ok {
		panic("Context must contain a logger")
	}

	vars := mux.Vars(r)
	venueID, err := strconv.Atoi(vars["venue_id"])
	if err != nil {
		common.SendError(rw, common.HTTPError{"Invalid performerID", http.StatusBadRequest, err}, nil)
		return
	}

	filter := event.EventFilterFromRequest(r)
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
