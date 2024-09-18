package server

//
//import (
//	"fmt"
//	"github.com/gorilla/mux"
//	"gns/stub/env"
//	"log"
//	"net/http"
//	"time"
//)
//
//func (cs *ControlServer) InitControlServer(f env.Environment) {
//
//	// Set up router for the control Server
//	controlRouter := mux.NewRouter()
//
//	// Server configuration endpoints
//	controlRouter.HandleFunc("/rest/api/v1/server/config", cs.UpdateServerConfigHandler).Methods(http.MethodPost)
//	controlRouter.HandleFunc("/rest/api/v1/server/config", cs.DeleteServerConfigHandler).Methods(http.MethodDelete)
//	controlRouter.HandleFunc("/rest/api/v1/server/start", cs.StartServerHandler).Methods(http.MethodPost)
//	controlRouter.HandleFunc("/rest/api/v1/server/stop", cs.StopServerHandler).Methods(http.MethodPost)
//	controlRouter.HandleFunc("/rest/api/v1/server/status", cs.StatusServerHandler).Methods(http.MethodGet)
//
//	// Set up HTTP control Server
//	cs.ControlServer = &http.Server{
//		Addr:           fmt.Sprintf(":%s", f.ControlServerPort),
//		Handler:        controlRouter,
//		ReadTimeout:    10 * time.Second,
//		WriteTimeout:   10 * time.Second,
//		IdleTimeout:    120 * time.Second,
//		MaxHeaderBytes: 1 << 20,
//	}
//
//	// Start the control Server
//	log.Println("Control Server is starting on port 8062...")
//	log.Fatal(cs.ControlServer.ListenAndServe())
//
//}
//
//func (cs *ControlServer) InitManagedServer() error {
//	switch cs.ServerType {
//	case "gin":
//		cs.Server = NewGinServer(cs)
//	case "fasthttp":
//		//cs.Server = NewFastHTTPServer(cs)
//	default:
//		cs.Server = NewNetHTTPServer(cs) // по умолчанию net/http
//	}
//	return cs.Server.Init()
//}
//
//func (cs *ControlServer) StartManagedServer() error {
//	return cs.Server.Start()
//}
