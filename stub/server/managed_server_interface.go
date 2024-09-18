package server

type ManagedServer interface {
	Init()
	Start()
	Stop()
}
