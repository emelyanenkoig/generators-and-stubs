package server

import (
	"fmt"
	"github.com/gorilla/mux"
	"gns/stub/env"
	"log"
	"net/http"
	"time"
)

func (cs *ControlServer) InitControlServer(f env.Environment) {

	// Set up router for the control Server
	controlRouter := mux.NewRouter()

	// Server configuration endpoints
	controlRouter.HandleFunc("/rest/api/v1/server/config", cs.UpdateServerConfigHandler).Methods(http.MethodPost)
	controlRouter.HandleFunc("/rest/api/v1/server/config", cs.DeleteServerConfigHandler).Methods(http.MethodDelete)
	controlRouter.HandleFunc("/rest/api/v1/server/start", cs.StartServerHandler).Methods(http.MethodPost)
	controlRouter.HandleFunc("/rest/api/v1/server/stop", cs.StopServerHandler).Methods(http.MethodPost)
	controlRouter.HandleFunc("/rest/api/v1/server/status", cs.StatusServerHandler).Methods(http.MethodGet)

	// Set up HTTP control Server
	cs.ControlServer = &http.Server{
		Addr:           fmt.Sprintf(":%s", f.ControlServerPort),
		Handler:        controlRouter,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    120 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	// Start the control Server
	log.Println("Control Server is starting on port 8062...")
	log.Fatal(cs.ControlServer.ListenAndServe())

}

func (cs *ControlServer) InitManagedServer() {

	// Set up router for the managed Server with middleware to control access
	managedRouter := mux.NewRouter()
	managedRouter.Use(cs.ServerAccessControlMiddleware) // Middleware to control access based on serverRunning state

	// Register routes dynamically based on the Server configuration
	// TODO потестить что хендлеры читает метод
	for _, pathConfig := range cs.Config.Paths {
		managedRouter.HandleFunc(pathConfig.Path, cs.RouteHandler).Methods(http.MethodGet)
		managedRouter.HandleFunc(pathConfig.Path, cs.RouteHandler).Methods(http.MethodPost)
		managedRouter.HandleFunc(pathConfig.Path, cs.RouteHandler).Methods(http.MethodDelete)
		managedRouter.HandleFunc(pathConfig.Path, cs.RouteHandler).Methods(http.MethodPut)
	}

	// Set up HTTP managed Server
	cs.Server.Server = &http.Server{
		Addr:           ":8080",
		Handler:        managedRouter,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
		IdleTimeout:    10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Println("Managed Server is starting on port 8080...")
	if err := cs.Server.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Could not listen on :8081: %v\n", err)
	}

}
