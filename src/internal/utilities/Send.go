package utilities

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Success bool     `json:"success"`
	Message string   `json:"message,omitempty"`
	Data    any      `json:"data,omitempty"`
	Module  string   `json:"module,omitempty"`
	Errors  []string `json:"errors,omitempty"`
}

func SendSuccess(w http.ResponseWriter, data any, message string, code int) {
	response := Response{
		Success: true,
		Data:    data,
		Message: message,
	}
	sendJSON(w, response, code)
}

func SendError(w http.ResponseWriter, errors []string, module string, code int) {
	response := Response{
		Success: false,
		Module:  module,
		Errors:  errors,
	}
	sendJSON(w, response, code)
}

func sendJSON(w http.ResponseWriter, response any, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(response)
}
