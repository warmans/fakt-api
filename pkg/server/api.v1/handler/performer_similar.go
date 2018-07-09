package handler

import (
	"net/http"

	"strconv"

	"github.com/gorilla/mux"
	"github.com/warmans/fakt-api/pkg/server/api.v1/common"
	"github.com/warmans/fakt-api/pkg/server/data/store/performer"
	"github.com/warmans/route-rest/routes"
	"github.com/warmans/fakt-api/pkg/server/api.v1/middleware"
)

func NewPerformerSimilarHandler(ds *performer.Store) routes.RESTHandler {
	return &PerformerSimilarHandler{ds: ds}
}

type PerformerSimilarHandler struct {
	routes.DefaultRESTHandler
	ds *performer.Store
}

func (h *PerformerSimilarHandler) HandleGetList(rw http.ResponseWriter, r *http.Request) {

	logger := middleware.MustGetLogger(r)

	vars := mux.Vars(r)
	eventId, err := strconv.Atoi(vars["performer_id"])
	if err != nil {
		common.SendError(rw, common.HTTPError{"Invalid performer_id", http.StatusBadRequest, err}, nil)
		return
	}

	similarPerformers, err := h.ds.FindSimilarPerformerIDs(int64(eventId))
	if err != nil {
		common.SendError(rw, err, logger)
		return
	}
	if len(similarPerformers) == 0 {
		common.SendResponse(rw, &common.Response{Status: http.StatusOK, Payload: []struct{}{}})
		return
	}

	f := performer.FilterFromRequest(r)
	f.IDs = similarPerformers

	performers, err := h.ds.FindPerformers(f)
	if err != nil {
		common.SendError(rw, err, logger)
		return
	}
	common.SendResponse(rw, &common.Response{Status: http.StatusOK, Payload: performers})
}
