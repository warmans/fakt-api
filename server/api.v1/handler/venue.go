package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/warmans/ctxhandler"
	"github.com/warmans/fakt-api/server/api.v1/common"
	"golang.org/x/net/context"
	"github.com/warmans/fakt-api/server/data/service/venue"
	"github.com/go-kit/kit/log"
)

func NewVenueHandler(ds *venue.VenueService) ctxhandler.CtxHandler {
	return &VenueHandler{ds: ds}
}

type VenueHandler struct {
	ds *venue.VenueService
}

func (h *VenueHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request, ctx context.Context) {

	logger, ok := ctx.Value("logger").(log.Logger)
	if !ok {
		panic("Context must contain a logger")
	}

	//query to filter
	filter := &venue.VenueFilter{VenueIDs: make([]int, 0), Name: r.Form.Get("name")}
	if venue := r.Form.Get("venue"); venue != "" {
		for _, idStr := range strings.Split(venue, ",") {
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
