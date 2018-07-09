package handler

import (
	"net/http"

	"github.com/warmans/fakt-api/pkg/server/api.v1/common"

	"github.com/warmans/fakt-api/pkg/server/data/service/tag"
	"github.com/warmans/route-rest/routes"
	"github.com/warmans/fakt-api/pkg/server/api.v1/middleware"
)

func NewTagHandler(ts *tag.TagService) routes.RESTHandler {
	return &TagHandler{tags: ts}
}

type TagHandler struct {
	routes.DefaultRESTHandler
	tags *tag.TagService
}

func (h *TagHandler) HandleGetList(rw http.ResponseWriter, r *http.Request) {

	logger := middleware.MustGetLogger(r)

	tags, err := h.tags.FindTags(tag.TagFilterFromRequest(r))
	if err != nil {
		common.SendError(rw, err, logger)
		return
	}
	common.SendResponse(rw, &common.Response{Status: http.StatusOK, Payload: tags})
}
