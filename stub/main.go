package main

import (
	"fmt"
	"gns/stub/env"
	"gns/stub/server"
	"log"
	"runtime"
	"sync"
)

// TODO добавить логгирование
// TODO Сделать отключаемый и включаемый logger из ENV
// + проверка на ошибочный путь
func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	f := env.ReadENV()

	controlServer := server.NewControlServer(f.ManagedServerType)

	// Загружаем начальную конфигурацию сервера
	err := controlServer.LoadServerConfig(fmt.Sprintf("%s", f.ResponseFilePath))
	if err != nil {
		log.Fatalf("Error loading Server Config: %v", err)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		controlServer.InitManagedServer()
		controlServer.StartManagedServer()
		wg.Done()
	}()
	controlServer.InitControlServer(f)

	wg.Wait()
}
