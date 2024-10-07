package managed

import (
	"gns/stub/server/managed/entities"
	"time"
)

type ManagedServerInterface interface {
	InitManagedServer() // Инициализация сервера
	RunManagedServer()  // Запуск сервера
	IsRunning() bool
	SetRunning(v bool)
	GetConfig() entities.ServerConfig
	SetConfig(config entities.ServerConfig) error
	GetTimeSinceStart() time.Time
	GetReqSinceStart() uint
}

const (
	ServerTypeFastHTTP = "fasthttp"
	ServerTypeGin      = "gin"

	HTTP10 = "HTTP/1.0"
	HTTP11 = "HTTP/1.1"
	HTTP20 = "HTTP/2.0"
)
