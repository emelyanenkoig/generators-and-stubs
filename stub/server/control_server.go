package server

import (
	"sync"
	"time"
)

type ControlServer struct {
	Config    ServerConfig
	Server    *ManagedServer
	StartTime time.Time
	ReqCount  int
	mu        sync.RWMutex
	TpsMu     sync.Mutex
}

// InitManagedServer инициализирует управляемый сервер на основе типа сервера
func (cs *ControlServer) InitManagedServer(serverType string) error {
	var managedServer ManagedServerInterface

	switch serverType {
	case "gin":
		managedServer = NewGinServer(cs)
	case "fasthttp":
		managedServer = NewFastHTTPServer(cs)
	default:
		managedServer = NewNetHTTPServer(cs)
	}

	if err := managedServer.Init(); err != nil {
		return err
	}

	go managedServer.Start()
	return nil
}
