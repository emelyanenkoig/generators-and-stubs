package server

import (
	"github.com/valyala/fasthttp"
	"log"
)

type FastHTTPServer struct {
	cs *ControlServer
}

func NewFastHTTPServer(cs *ControlServer) ManagedServerInterface {
	return &FastHTTPServer{cs: cs}
}

func (s *FastHTTPServer) Init() error {
	// FastHTTP не требует явной инициализации, просто запускается.
	return nil
}

func (s *FastHTTPServer) Start() error {
	log.Println("Managed Server is starting on port 8080 (fasthttp)...")
	if err := fasthttp.ListenAndServe(":8080", func(ctx *fasthttp.RequestCtx) {
		s.cs.RouteHandler(ctx.Response.BodyWriter(), ctx.Request.Body())
	}); err != nil {
		return err
	}
	return nil
}
