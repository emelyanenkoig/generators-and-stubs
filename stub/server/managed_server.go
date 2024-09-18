package server

import "net/http"

type ManagedServerNetHttp struct {
	server http.Server
}

func NewManagedServerNetHttp(server http.Server) *ManagedServerNetHttp {
	return &ManagedServerNetHttp{server: server}
}

func (s *ManagedServerNetHttp) Init() {
	err := s.server.ListenAndServe()
	if err != nil {
		return
	}
}
