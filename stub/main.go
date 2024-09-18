package main

import (
	"log"
	"os"
	"server"
)

func main() {
	controlServer := &server.ControlServer{}

	serverType := os.Getenv("SERVER_TYPE") // Извлекаем тип сервера из ENV
	if serverType == "" {
		serverType = "nethttp" // По умолчанию net/http
	}

	err := controlServer.InitManagedServer(serverType)
	if err != nil {
		log.Fatalf("Failed to initialize managed server: %v", err)
	}

	log.Println("Control server started successfully.")
}
