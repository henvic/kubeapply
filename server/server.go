package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/errwrap"
	log "github.com/sirupsen/logrus"
)

// Params of the service
type Params struct {
	Address string

	ExposeDebug bool
}

// Start "kubectl apply" RESTful server
func Start(ctx context.Context, params Params) error {
	var s = &Server{}

	return s.Serve(ctx, params)
}

// Server for handling requests
type Server struct {
	ctx context.Context

	params Params

	http *http.Server
	ec   chan error
}

// Serve handlers
func (s *Server) Serve(ctx context.Context, params Params) error {
	s.ctx = ctx
	s.params = params

	mux := http.NewServeMux()
	mux.HandleFunc("/apply", handleApply)
	mux.HandleFunc("/version", handleVersion)

	s.http = &http.Server{
		Handler: mux,
	}

	return s.serve()
}

func getAddr(a string) string {
	l := strings.LastIndex(a, ":")

	if l == -1 && len(a) <= l {
		return a
	}

	return "http://localhost:" + a[l+1:]
}

// Serve HTTP requests
func (s *Server) serve() error {
	s.ec = make(chan error, 1)
	go s.listen()
	go s.waitShutdown()

	err := <-s.ec

	if err == http.ErrServerClosed {
		fmt.Println()
		log.Info("Server shutting down gracefully.")
		return nil
	}

	return err
}

func (s *Server) waitShutdown() {
	<-s.ctx.Done()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.http.Shutdown(ctx); err != nil && err != context.Canceled {
		s.ec <- errwrap.Wrapf("can't shutdown server properly: {{err}}", err)
	}
}

func (s *Server) listen() {
	l, err := net.Listen("tcp", s.params.Address)

	if err != nil {
		s.ec <- err
		return
	}

	log.Infof("Starting server on %v", getAddr(l.Addr().String()))

	err = s.http.Serve(l)

	s.ec <- err
}
