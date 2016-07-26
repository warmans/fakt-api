package handler

import (
	"net/http"

	"github.com/warmans/stressfaktor-api/server/api.v1/common"
)

func NewVersionHandler(version string) http.Handler {
	return &VersionHandler{version: version}
}

type VersionHandler struct {
	version string
}

func (h *VersionHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	common.SendResponse(
		rw,
		&common.Response{
			Status:  http.StatusOK,
			Payload: map[string]string{"version": h.version},
		},
	)
}
