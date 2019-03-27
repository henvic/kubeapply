package server

import (
	"encoding/json"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
)

type response struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Errors  string `json:"errors,omitempty"`
}

// ErrorHandler for errors.
// Use code = -1 to bypass writing header or trying to set Content-Type. Useful when headers are already sent.
func ErrorHandler(w http.ResponseWriter, r *http.Request, code int, errors ...string) {
	if code >= 0 {
		w.Header().Set("Content-Type", "application/json; charset=utf8")
		w.WriteHeader(code)
	}

	if code != http.StatusNotFound {
		log.Debugf("Request error: %v (%v)", code, http.StatusText(code))
	}

	if err := json.NewEncoder(w).Encode(response{
		Status:  code,
		Message: http.StatusText(code),
		Errors:  strings.Join(errors, "\n"),
	}); err != nil {
		log.Error(err)
	}
}
