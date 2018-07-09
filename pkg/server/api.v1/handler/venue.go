package handler

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/warmans/fakt-api/pkg/server/api.v1/common"
	"github.com/warmans/fakt-api/pkg/server/api.v1/middleware"
	"github.com/warmans/fakt-api/pkg/server/data/service/venue"
	"github.com/warmans/route-rest/routes"
	"go.uber.org/zap"
)

func NewVenueHandler(ds *venue.VenueService) routes.RESTHandler {
	return &VenueHandler{ds: ds}
}

type VenueHandler struct {
	routes.DefaultRESTHandler
	ds *venue.VenueService
}

func (h *VenueHandler) HandleGetList(rw http.ResponseWriter, r *http.Request) {

	logger := middleware.MustGetLogger(r)

	venues, err := h.ds.FindVenues(venue.VenueFilterFromRequest(r))
	if err != nil {
		common.SendError(rw, err, logger)
		return
	}

	common.SendResponse(rw, &common.Response{Status: http.StatusOK, Payload: venues})
}

func (h *VenueHandler) HandleGet(rw http.ResponseWriter, r *http.Request) {

	logger, ok := r.Context().Value("logger").(*zap.Logger)
	if !ok {
		panic("Context must contain a logger")
	}

	vars := mux.Vars(r)
	venueID, err := strconv.Atoi(vars["venue_id"])
	if err != nil {
		common.SendError(rw, common.HTTPError{"Invalid venue ID", http.StatusBadRequest, err}, nil)
		return
	}

	f := &venue.VenueFilter{}
	f.IDs = []int64{int64(venueID)}
	f.PageSize = 1
	f.Page = 1

	venues, err := h.ds.FindVenues(f)
	if err != nil {
		common.SendError(rw, err, logger)
		return
	}

	if len(venues) < 1 {
		common.SendError(rw, common.HTTPError{"Venue not Found", http.StatusNotFound, err}, nil)
		return
	}

	common.SendResponse(rw, &common.Response{Status: http.StatusOK, Payload: venues[0]})
}
