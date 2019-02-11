package server

import (
	"encoding/json"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
)

type response struct {
	Status   int    `json:"status"`
	Message  string `json:"message"`
	Feedback string `json:"feedback,omitempty"`
}

// ErrorHandler for errors
func ErrorHandler(w http.ResponseWriter, r *http.Request, code int, feedback ...string) {
	var f = strings.Join(feedback, "\n")

	w.Header().Set("Content-Type", "application/json; charset=utf8")
	w.WriteHeader(code)

	if code != http.StatusNotFound {
		log.Debugf("Request error: %v (%v)", code, http.StatusText(code))
	}

	if err := json.NewEncoder(w).Encode(response{
		Status:   code,
		Message:  http.StatusText(code),
		Feedback: f,
	}); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Error(err)
	}
}
