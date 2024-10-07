package main

import (
	"gns/stub/env"
	"gns/stub/log"
	"gns/stub/server/control"
	"runtime"
	"sync"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	f := env.ReadENV()

	logger := log.InitLogger(f.LogLevel)
	defer logger.Sync()

	controlServer := control.NewControlServer(f)

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
