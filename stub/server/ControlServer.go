package server

import (
	"github.com/gin-gonic/gin"
	"gns/stub/env"
	"log"
)

func (cs *ControlServer) InitControlServer(f env.Environment) {
	r := gin.Default()

	r.GET("/rest/api/v1/server/config", cs.GetControlServerConfig)
	r.POST("/rest/api/v1/server/config", cs.UpdateControlServerConfig)

	r.DELETE("/rest/api/v1/server/config", cs.DeleteControlServerConfig)
	r.POST("/rest/api/v1/server/start", cs.StartControlServer)
	r.POST("/rest/api/v1/server/stop", cs.StopControlServer)
	r.GET("/rest/api/v1/server/status", cs.StatusControlServer)

	log.Println("Starting control server: 8062")
	if err := r.Run(":8062"); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}

func NewControlServer(serverType string) *ControlServer {
	cs := &ControlServer{
		RRobinIndex: map[string]int{},
	}
	switch serverType {
	case "fasthttp":
		//cs.ManagedServer = &FastHTTPServer{}
	case "gin":
		cs.ManagedServer = &GinServer{}
	default:
		cs.ManagedServer = &NetHTTPServer{}
	}
	return cs
}
