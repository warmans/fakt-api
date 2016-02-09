package common

import (
	"net/http"
	"encoding/json"
)

type Response struct {
	Status  int         `json:"status"`
	Payload interface{} `json:"payload"`
}

func SendResponse(rw http.ResponseWriter, response *Response) {
	rw.Header().Add("Content-Type", "application/json; charset=utf-8")
	rw.WriteHeader(response.Status)

	jsonEncoder := json.NewEncoder(rw)
	jsonEncoder.Encode(response)
}
