package main

import (
	"context"
	_ "expvar"
	"flag"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"

	"github.com/henvic/ctxsignal"
	"github.com/henvic/kubeapply/server"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

var params = server.Params{}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	flag.Parse()

	var debug = (os.Getenv("DEBUG") != "")

	if debug {
		log.SetLevel(log.DebugLevel)
	}

	if debug || params.ExposeDebug {
		go profiler()
	}

	ctx, cancel := ctxsignal.WithTermination(context.Background())
	defer cancel()

	if err := server.Start(ctx, params); err != nil {
		log.Fatal(err)
	}
}

func profiler() {
	// let expvar and pprof be exposed here indirectly through http.DefaultServeMux
	log.Info("Exposing expvar and pprof on localhost:8081")
	log.Fatal(http.ListenAndServe("localhost:8081", nil))
}

func init() {
	flag.StringVar(&params.Address, "addr", "127.0.0.1:8080", "Serving address")
	flag.BoolVar(&params.ExposeDebug, "expose-debug", true, "Expose debugging tools over HTTP (on port 8081)")
}
