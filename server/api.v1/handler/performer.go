package handler

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/warmans/fakt-api/server/api.v1/common"
	"github.com/warmans/fakt-api/server/data/service/performer"
	"github.com/warmans/route-rest/routes"
	"github.com/warmans/fakt-api/server/api.v1/middleware"
)

func NewPerformerHandler(ds *performer.PerformerService) routes.RESTHandler {
	return &PerformerHandler{ds: ds}
}

type PerformerHandler struct {
	routes.DefaultRESTHandler
	ds *performer.PerformerService
}

func (h *PerformerHandler) HandleGetList(rw http.ResponseWriter, r *http.Request) {

	logger := middleware.MustGetLogger(r)

	performers, err := h.ds.FindPerformers(performer.PerformerFilterFromRequest(r))
	if err != nil {
		common.SendError(rw, err, logger)
		return
	}

	common.SendResponse(rw, &common.Response{Status: http.StatusOK, Payload: performers})
}

func (h *PerformerHandler) HandleGet(rw http.ResponseWriter, r *http.Request) {

	logger := middleware.MustGetLogger(r)

	vars := mux.Vars(r)
	performerID, err := strconv.Atoi(vars["performer_id"])
	if err != nil {
		common.SendError(rw, common.HTTPError{"Invalid performerID", http.StatusBadRequest, err}, nil)
		return
	}
	if performerID == 0 {
		common.SendError(rw, common.HTTPError{"Invalid performerID", http.StatusBadRequest, err}, nil)
		return
	}

	f := &performer.PerformerFilter{}
	f.IDs = []int64{int64(performerID)}
	f.PageSize = 1
	f.Page = 1

	performers, err := h.ds.FindPerformers(f)
	if err != nil {
		common.SendError(rw, err, logger)
		return
	}

	common.SendResponse(rw, &common.Response{Status: http.StatusOK, Payload: performers})
}
