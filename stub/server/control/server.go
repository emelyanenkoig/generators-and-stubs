package control

type ManagedServerInterface interface {
	InitManagedServer(cs *ControlServer)  // Инициализация сервера
	StartManagedServer(cs *ControlServer) // Запуск сервера
	SetRunning(v bool)
	IsRunning() bool
}

func (cs *ControlServer) InitManagedServer() {
	cs.ManagedServer.InitManagedServer(cs) // Инициализация сервера
}

func (cs *ControlServer) StartManagedServer() {
	cs.ManagedServer.StartManagedServer(cs) // Запуск сервера
}
