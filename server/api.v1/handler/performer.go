package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/go-kit/kit/log"
	"github.com/warmans/fakt-api/server/api.v1/common"
	"github.com/warmans/fakt-api/server/data/service/performer"
	"github.com/warmans/route-rest/routes"
)

func NewPerformerHandler(ds *performer.PerformerService) routes.RESTHandler {
	return &PerformerHandler{ds: ds}
}

type PerformerHandler struct {
	routes.DefaultRESTHandler
	ds *performer.PerformerService
}

func (h *PerformerHandler) HandleGetList(rw http.ResponseWriter, r *http.Request) {

	logger, ok := r.Context().Value("logger").(log.Logger)
	if !ok {
		panic("Context must contain a logger")
	}

	//query to filter
	filter := &performer.PerformerFilter{
		Name: r.Form.Get("name"),
		Genre: r.Form.Get("genre"),
		Home: r.Form.Get("home"),
		PerformerID: make([]int, 0),
	}
	if ids := r.Form.Get("ids"); ids != "" {
		for _, idStr := range strings.Split(ids, ",") {
			if idInt, err := strconv.Atoi(idStr); err == nil {
				filter.PerformerID = append(filter.PerformerID, idInt)
			}
		}
	}

	performers, err := h.ds.FindPerformers(filter)
	if err != nil {
		common.SendError(rw, err, logger)
		return
	}

	common.SendResponse(rw, &common.Response{Status: http.StatusOK, Payload: performers})
}
