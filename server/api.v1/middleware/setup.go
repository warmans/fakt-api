package middleware

import (
	"log"
	"net/http"

	"github.com/warmans/stressfaktor-api/server/api.v1/common"
)

func AddSetup(nextHandler http.Handler) http.Handler {
	return &SetupMiddleware{next: nextHandler}
}

type SetupMiddleware struct {
	next    http.Handler
	headers map[string]string
}

func (m *SetupMiddleware) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	//always parse the form
	if err := r.ParseForm(); err != nil {
		common.SendError(rw, common.HTTPError{"Invalid request data", http.StatusBadRequest, nil}, false)
		return
	}
	//always close the body
	defer func() {
		if err := r.Body.Close(); err != nil {
			//can't really do much once the headers have already been sent. Meh!
			log.Printf("Failed to close HTTP response body: %s", err)
		}
	}()
	m.next.ServeHTTP(rw, r)
}
