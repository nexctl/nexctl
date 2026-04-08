package response

import (
	"encoding/json"
	"net/http"

	"github.com/nexctl/nexctl/server/pkg/errcode"
)

// Envelope is the standard API response shape.
type Envelope struct {
	Code    errcode.Code `json:"code"`
	Message string       `json:"message"`
	Data    any          `json:"data,omitempty"`
}

// WriteOK writes a successful API response.
func WriteOK(w http.ResponseWriter, data any) {
	write(w, http.StatusOK, Envelope{Code: errcode.OK, Message: "ok", Data: data})
}

// WriteCreated writes a created API response.
func WriteCreated(w http.ResponseWriter, data any) {
	write(w, http.StatusCreated, Envelope{Code: errcode.OK, Message: "ok", Data: data})
}

// WriteError writes an error API response.
func WriteError(w http.ResponseWriter, status int, code errcode.Code, message string) {
	write(w, status, Envelope{Code: code, Message: message})
}

func write(w http.ResponseWriter, status int, payload Envelope) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
