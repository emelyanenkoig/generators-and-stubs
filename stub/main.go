package main

import (
	"fmt"
	"gns/stub/env"
	"gns/stub/log"
	"gns/stub/server/control"
	"go.uber.org/zap"
	"runtime"
	"sync"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	f := env.ReadENV()

	logger := log.InitLogger(f.LogLevel)
	defer logger.Sync()

	controlServer := control.NewControlServer(f)

	err := controlServer.LoadServerConfig(fmt.Sprintf("%s", f.ResponseFilePath))
	if err != nil {
		logger.Fatal("Error loading Server Config:", zap.Error(err))
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
