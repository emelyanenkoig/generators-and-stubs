package server

// ManagedServerInterface - интерфейс для управляемых серверов.
type ManagedServerInterface interface {
	Init() error
	Start() error
}
