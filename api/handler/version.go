package handler

import (
	"net/http"
	"github.com/warmans/stressfaktor-api/api/common"
)

func NewVersionHandler(version string) http.Handler {
	return &VersionHandler{version: version}
}

type VersionHandler struct {
	version string
}

func (h *VersionHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	common.SendResponse(
		rw,
		&common.Response{
			Status: http.StatusOK,
			Payload: map[string]string{"version": h.version},
		},
	)
}
