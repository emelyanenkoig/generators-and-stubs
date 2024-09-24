package main

import (
	"fmt"
	"gns/stub/env"
	"gns/stub/server/control"
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

	controlServer := control.NewControlServer(f)

	err := controlServer.LoadServerConfig(fmt.Sprintf("%s", f.ResponseFilePath))
	if err != nil {
		log.Fatalf("Error loading Server Config: %v", err)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		controlServer.InitManagedServer()
		controlServer.RunManagedServer()
		wg.Done()
	}()
	controlServer.InitControlServer()
	controlServer.RunControlServer()

	wg.Wait()
}
