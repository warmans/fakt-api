package handler

import (
	"net/http"

	"github.com/warmans/fakt-api/server/api.v1/common"

	"github.com/go-kit/kit/log"
	"github.com/warmans/fakt-api/server/data/service/tag"
	"github.com/warmans/route-rest/routes"
)

func NewTagHandler(ts *tag.TagService) routes.RESTHandler {
	return &TagHandler{tags: ts}
}

type TagHandler struct {
	routes.DefaultRESTHandler
	tags *tag.TagService
}

func (h *TagHandler) HandleGetList(rw http.ResponseWriter, r *http.Request) {

	logger, ok := r.Context().Value("logger").(log.Logger)
	if !ok {
		panic("Context must contain a logger")
	}

	tags, err := h.tags.FindTags(tag.TagFilterFromRequest(r))
	if err != nil {
		common.SendError(rw, err, logger)
		return
	}
	common.SendResponse(rw, &common.Response{Status: http.StatusOK, Payload: tags})
}
