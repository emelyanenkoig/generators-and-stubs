package server

type ManagedServerInterface interface {
	InitManagedServer(cs *ControlServer)  // Инициализация сервера
	StartManagedServer(cs *ControlServer) // Запуск сервера
	SetRunning(v bool)
	IsRunning() bool
}
