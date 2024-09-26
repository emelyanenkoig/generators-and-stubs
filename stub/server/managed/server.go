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
	SetConfig(config entities.ServerConfig)
	GetTimeSinceStart() time.Time
	GetReqSinceStart() uint
}

const (
	ServerTypeFastHTTP = "fasthttp"
	ServerTypeGin      = "gin"
)
