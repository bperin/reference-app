package response

import (
	"encoding/json"
	"net/http"
)

type ErrorPayload struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description,omitempty"`
}

func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		_ = json.NewEncoder(w).Encode(data)
	}
}

func Error(w http.ResponseWriter, status int, errStr, description string) {
	JSON(w, status, ErrorPayload{
		Error:            errStr,
		ErrorDescription: description,
	})
}
