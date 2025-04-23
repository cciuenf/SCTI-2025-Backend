package utilities

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Message string      `json:"message"`
	Payload interface{} `json:"data"`
}

func (r *Response) Send(w http.ResponseWriter, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	response := Response{
		Message: "",
		Payload: map[string]any{},
	}

	if r.Message != "" {
		response.Message = r.Message
	}

	if r.Payload != nil {
		response.Payload = r.Payload
	}

	json.NewEncoder(w).Encode(response)
}

func Send(w http.ResponseWriter, msg string, data interface{}, code int) {
	if data == nil {
		data = map[string]any{}
	}

	var r Response
	r.Message = msg
	r.Payload = data
	r.Send(w, code)
}
