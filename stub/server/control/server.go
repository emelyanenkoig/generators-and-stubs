package control

type ManagedServerInterface interface {
	InitManagedServer(cs *ControlServer) // Инициализация сервера
	RunManagedServer(cs *ControlServer)  // Запуск сервера
	SetRunning(v bool)
	IsRunning() bool
}
