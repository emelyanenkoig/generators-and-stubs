package control

import "time"

type ManagedServerInterface interface {
	InitManagedServer() // Инициализация сервера
	RunManagedServer()  // Запуск сервера
	IsRunning() bool
	SetRunning(v bool)
	GetConfig() ServerConfig
	SetConfig(config ServerConfig)
	GetTimeSinceStart() time.Time
	GetReqSinceStart() uint
}
