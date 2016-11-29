package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/go-kit/kit/log"
	"github.com/warmans/fakt-api/server/api.v1/common"
	"github.com/warmans/fakt-api/server/data/service/performer"
)

func NewPerformerHandler(ds *performer.PerformerService) http.Handler {
	return &PerformerHandler{ds: ds}
}

type PerformerHandler struct {
	ds *performer.PerformerService
}

func (h *PerformerHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {

	logger, ok := r.Context().Value("logger").(log.Logger)
	if !ok {
		panic("Context must contain a logger")
	}

	//query to filter
	filter := &performer.PerformerFilter{PerformerID: make([]int, 0)}
	if venue := r.Form.Get("performer"); venue != "" {
		for _, idStr := range strings.Split(venue, ",") {
			if idInt, err := strconv.Atoi(idStr); err == nil {
				filter.PerformerID = append(filter.PerformerID, idInt)
			}
		}
	}
	filter.Name = r.Form.Get("name")
	filter.Genre = r.Form.Get("genre")
	filter.Home = r.Form.Get("home")

	performers, err := h.ds.FindPerformers(filter)
	if err != nil {
		common.SendError(rw, err, logger)
		return
	}

	common.SendResponse(rw, &common.Response{Status: http.StatusOK, Payload: performers})
}
