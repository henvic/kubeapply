package server

import (
	"encoding/json"
	"net"
	"net/http"
	"net/http/httputil"
	"os/exec"
	"strings"

	"github.com/henvic/kubeapply"
	"github.com/henvic/kubeapply/server/decoding"
	log "github.com/sirupsen/logrus"
)

// ApplyRequestBody for the apply endpoint.
type ApplyRequestBody struct {
	Command string                        `json:"command,omitempty"`
	Files   map[string]decoding.FileValue `json:"files,omitempty"`
	Flags   map[string]decoding.FlagValue `json:"flags,omitempty"`
}

// FlagsMap gets the flags on a map[string]string.
func (a *ApplyRequestBody) FlagsMap() map[string]string {
	var m = map[string]string{}

	for k, v := range a.Flags {
		m[k] = string(v)
	}

	return m
}

// FilesMap gets the files on a map[string]string.
func (a *ApplyRequestBody) FilesMap() map[string][]byte {
	var m = map[string][]byte{}

	for k, v := range a.Files {
		m[k] = v
	}

	return m
}

func handleApply(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		ErrorHandler(w, r, http.StatusMethodNotAllowed,
			"kubectl reference: https://kubernetes.io/docs/reference/generated/kubectl/kubectl-commands#apply")
		return
	}

	if t := r.Header.Get("Content-Type"); !strings.Contains(t, "application/json") {
		ErrorHandler(w, r, http.StatusNotAcceptable)
		return
	}

	var dump, err = httputil.DumpRequest(r, false)

	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError)
		log.Errorf("cannot dump request (remote IP: %v", r.RemoteAddr)
		return
	}

	arb := ApplyRequestBody{}

	if ed := json.NewDecoder(r.Body).Decode(&arb); ed != nil {
		ErrorHandler(w, r, http.StatusBadRequest)
		log.Debugf("bad request: %v", ed)
		return
	}

	a := &kubeapply.Apply{
		Subcommand: arb.Command,

		Flags: arb.FlagsMap(),
		Files: arb.FilesMap(),

		IP: filterIP(r.RemoteAddr),

		RequestDump: dump,
	}

	var resp kubeapply.Response

	log.Debugf("Preparing to run kubectl apply request from IP %v", r.RemoteAddr)

	resp, err = a.Run(r.Context())

	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			ErrorHandler(w, r, http.StatusInternalServerError)
			log.Errorf("error swapping process for request %s (PID %v): %v", resp.ID, ee.Pid(), err)
			return
		}

		log.Errorf("error executing request %s: %v", resp.ID, err)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf8")
	ee := json.NewEncoder(w).Encode(resp)

	if ee != nil {
		ErrorHandler(w, r, http.StatusInternalServerError)
		log.Errorf("cannot encode response for request %s: %v", resp.ID, ee)
		return
	}

	if err == nil {
		log.Infof("request %v fulfilled with success", resp.ID)
	}
}

func filterIP(remoteAddr string) string {
	ip, _, err := net.SplitHostPort(remoteAddr)

	if err != nil {
		return remoteAddr
	}

	return ip
}
