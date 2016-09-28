package handler

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/warmans/ctxhandler"
	"github.com/warmans/fakt-api/server/api.v1/common"
	"github.com/warmans/fakt-api/server/data/store"
	"golang.org/x/net/context"
)

func NewVenueHandler(ds *store.Store) ctxhandler.CtxHandler {
	return &VenueHandler{ds: ds}
}

type VenueHandler struct {
	ds *store.Store
}

func (h *VenueHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request, ctx context.Context) {

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

	venues, err := h.ds.FindVenues(filter)
	if err != nil {
		log.Print(err.Error())
		common.SendResponse(rw, &common.Response{Status: http.StatusInternalServerError, Payload: nil})
		return
	}

	common.SendResponse(rw, &common.Response{Status: http.StatusOK, Payload: venues})
}
