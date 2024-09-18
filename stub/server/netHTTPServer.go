package server

import (
	"log"
	"net/http"
)

type NetHTTPServer struct {
	cs *ControlServer
}

func NewNetHTTPServer(cs *ControlServer) ManagedServerInterface {
	return &NetHTTPServer{cs: cs}
}

func (s *NetHTTPServer) Init() error {
	s.cs.Server.Server = &http.Server{
		Addr:    ":8080",
		Handler: http.HandlerFunc(s.cs.RouteHandler),
	}
	return nil
}

func (s *NetHTTPServer) Start() error {
	log.Println("Managed Server is starting on port 8080 (net/http)...")
	if err := s.cs.Server.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}
