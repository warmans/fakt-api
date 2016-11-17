package common

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/jinzhu/now"
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

//GetRelativeDateRange takes e.g. this weekend and returns the start and end date in SQL format
func GetRelativeDateRange(name string) (time.Time, time.Time) {

	//end days a few hours past midnight since e.g. 1am Saturday should still count as Friday night
	switch strings.ToLower(name) {
	case "this week":
		return time.Now(), now.EndOfSunday().Add(time.Hour * 4)
	case "this weekend":
		return now.BeginningOfWeek().Add(time.Hour * 24 * 5), now.EndOfSunday().Add(time.Hour * 4)
	case "tomorrow":
		return now.BeginningOfDay().Add(time.Hour * 24), now.EndOfDay().Add(time.Hour * 28)
	default:
		//unknown values including "today" get today-ish
		return now.BeginningOfDay(), now.EndOfDay().Add(time.Hour * 4)
	}
}
