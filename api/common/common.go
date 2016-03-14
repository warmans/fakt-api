package common

import (
	"net/http"
	"encoding/json"
	"log"
	"fmt"
	"github.com/gorilla/sessions"
	"golang.org/x/net/context"
	"github.com/warmans/stressfaktor-api/data/store"
)

type Response struct {
	Status  int         `json:"status"`
	Payload interface{} `json:"payload"`
	Message string      `json:"message"`
}

func SendResponse(rw http.ResponseWriter, response *Response) {
	rw.Header().Add("Content-Type", "application/json; charset=utf-8")
	rw.WriteHeader(response.Status)

	jsonEncoder := json.NewEncoder(rw)
	jsonEncoder.Encode(response)
}

type HTTPError struct {
	Msg     string
	Status  int
	LastErr error
}

func (e HTTPError) Error() string {
	return fmt.Sprintf("%s (caused by: %s)", e.Msg, e.LastErr.Error())
}


func SendError(rw http.ResponseWriter, err error, writeToLog bool) {
	if writeToLog {
		log.Print(err.Error())
	}

	code := 500
	message := "An error occured"
	switch err.(type){
	case HTTPError:
		//assume HTTP error messages are safe to show to the user
		message = fmt.Sprintf("%s (%s)", err.(HTTPError).Msg, http.StatusText(err.(HTTPError).Status))
		code = err.(HTTPError).Status
	}

	SendResponse(rw, &Response{code, nil, message})
}

type CtxHandler interface {
	ServeHTTP(rw http.ResponseWriter, r *http.Request, ctx context.Context)
}

func AddCtx(handler CtxHandler, sess sessions.Store, auth *store.AuthStore, restrict bool) http.Handler {
	return &CtxMiddleware{NextHandler: handler, SessionStore: sess, AuthStore: auth}
}

type CtxMiddleware struct {
	NextHandler    CtxHandler
	SessionStore   sessions.Store
	AuthStore      *store.AuthStore
	RestrictAccess bool
}

func (m *CtxMiddleware) ServeHTTP(rw http.ResponseWriter, r *http.Request) {

	ctx := context.Background()

	sess, err := m.SessionStore.Get(r, "login")
	if err != nil {
		log.Printf("Failed to get session: %s", err.Error())
		m.NextHandler.ServeHTTP(rw, r, ctx)
		return
	}

	userId, found := sess.Values["userId"]
	if found == false || userId == nil || userId.(int64) < 1 {
		if m.RestrictAccess {
			SendError(rw, HTTPError{"Access Denied", http.StatusForbidden, nil}, false)
			return
		}
		m.NextHandler.ServeHTTP(rw, r, ctx)
		return
	}

	user, err := m.AuthStore.GetUser(userId.(int64))
	if err == nil && user != nil {
		ctx = context.WithValue(ctx, "user", user)
	} else {
		if m.RestrictAccess {
			SendError(rw, HTTPError{"Access Denied", http.StatusForbidden, nil}, false)
			return
		}
	}

	m.NextHandler.ServeHTTP(rw, r, ctx)
}