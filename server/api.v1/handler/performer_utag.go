package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/warmans/fakt-api/server/api.v1/common"
	dcom "github.com/warmans/fakt-api/server/data/service/common"
	"github.com/warmans/fakt-api/server/data/service/utag"
	"github.com/warmans/route-rest/routes"
	"github.com/warmans/fakt-api/server/api.v1/middleware"
)

func NewPerformerUTagHandler(ds *utag.UTagService) routes.RESTHandler {
	return &PerformerUTagHandler{ds: ds}
}

type PerformerUTagHandler struct {
	routes.DefaultRESTHandler
	ds     *utag.UTagService
}

func (h *PerformerUTagHandler) HandleGetList(rw http.ResponseWriter, r *http.Request) {

	logger := middleware.MustGetLogger(r)

	vars := mux.Vars(r)
	performerID, err := strconv.Atoi(vars["performer_id"])
	if err != nil {
		common.SendError(rw, common.HTTPError{"Invalid performerID", http.StatusBadRequest, err}, nil)
		return
	}

	//then get all tags for the event
	tags, err := h.ds.FindPerformerUTags(int64(performerID), &dcom.UTagsFilter{Username: r.Form.Get("username")})
	if err != nil && err != sql.ErrNoRows {
		common.SendError(rw, common.HTTPError{"Failed to get tags", http.StatusInternalServerError, err}, logger)
		return
	}

	common.SendResponse(
		rw,
		&common.Response{
			Status:  http.StatusOK,
			Payload: tags,
		},
	)
}

func (h *PerformerUTagHandler) HandlePost(rw http.ResponseWriter, r *http.Request) {

	logger := middleware.MustGetLogger(r)

	vars := mux.Vars(r)
	performerID, err := strconv.Atoi(vars["performer_id"])
	if err != nil {
		common.SendError(rw, common.HTTPError{"Invalid performerID", http.StatusBadRequest, err}, nil)
		return
	}

	usr, err := middleware.Restrict(r)
	if err != nil {
		common.SendError(rw, err, nil)
		return
	}

	payload := make([]string, 0)
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		common.SendError(rw, common.HTTPError{"Invalid payload", http.StatusBadRequest, nil}, nil)
		return
	}

	if err := h.ds.StorePerformerUTags(int64(performerID), usr.ID, payload); err != nil {
		common.SendError(rw, common.HTTPError{"Failed to save tags", http.StatusInternalServerError, err}, logger)
		return
	}

	h.HandleGetList(rw, r)
}

func (h *PerformerUTagHandler) HandleDelete(rw http.ResponseWriter, r *http.Request) {

	logger := middleware.MustGetLogger(r)

	vars := mux.Vars(r)
	performerID, err := strconv.Atoi(vars["performer_id"])
	if err != nil {
		common.SendError(rw, common.HTTPError{"Invalid performerID", http.StatusBadRequest, err}, nil)
		return
	}

	usr, err := middleware.Restrict(r)
	if err != nil {
		common.SendError(rw, err, nil)
		return
	}

	payload := make([]string, 0)
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		common.SendError(rw, common.HTTPError{"Invalid payload", http.StatusBadRequest, nil}, nil)
		return
	}

	if err := h.ds.RemovePerformerUTags(int64(performerID), usr.ID, payload); err != nil {
		common.SendError(rw, common.HTTPError{"Failed to save tags", http.StatusInternalServerError, err}, logger)
		return
	}

	h.HandleGetList(rw, r)
}
