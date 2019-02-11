package server

import (
	"fmt"
	"net/http"
	"os/exec"

	log "github.com/sirupsen/logrus"
)

func handleVersion(w http.ResponseWriter, r *http.Request) {
	var cmd = exec.CommandContext(r.Context(), "kubectl", "version", "--client", "--output=json") // #nosec

	var v, err = cmd.CombinedOutput()

	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError)
		log.Errorf("cannot show version during request: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf8")

	if _, err = fmt.Fprint(w, string(v)); err != nil {
		log.Errorf("error responding /version request: %v", err)
	}
}
