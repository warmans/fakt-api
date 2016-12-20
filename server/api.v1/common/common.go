package common

import (
	"encoding/json"
	"fmt"
	"net/http"

	"context"

	"github.com/go-kit/kit/log"
	"github.com/warmans/fakt-api/server/data/service/user"
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

func SendError(rw http.ResponseWriter, err error, logger log.Logger) {

	if logger != nil {
		logger.Log("msg", err)
	}

	code := 500
	message := "An error occured"
	switch err.(type) {
	case HTTPError:
		//assume HTTP error messages are safe to show to the user
		message = fmt.Sprintf("%s (%s)", err.(HTTPError).Msg, http.StatusText(err.(HTTPError).Status))
		code = err.(HTTPError).Status
	}

	SendResponse(rw, &Response{code, nil, message})
}

func Restrict(ctx context.Context) (*user.User, error) {
	usr := ctx.Value("user")
	if usr != nil {
		return nil, HTTPError{"Access Denied", http.StatusUnauthorized, nil}
	}
	usrType, ok := usr.(*user.User)
	if !ok {
		return nil, HTTPError{"Access Denied", http.StatusUnauthorized, nil}
	}
	return usrType, nil
}
