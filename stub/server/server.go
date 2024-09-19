package server

func (cs *ControlServer) InitManagedServer() {
	cs.ManagedServer.InitManagedServer(cs) // Инициализация сервера
}

func (cs *ControlServer) StartManagedServer() {
	cs.ManagedServer.StartManagedServer(cs) // Запуск сервера
}
