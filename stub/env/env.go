package env

import "os"

type Environment struct {
	ResponseFilePath  string
	ServerAddr        string
	ServerPort        string
	ControlServerAddr string
	ControlServerPort string
	ManagedServerType string
	LogLevel          string
}

func ReadENV() Environment {
	var f Environment
	f.ResponseFilePath = os.Getenv("RESPONSE_FILE_PATH")
	f.ControlServerAddr = os.Getenv("CONTROL_SERVER_ADDR")
	f.ControlServerPort = os.Getenv("CONTROL_SERVER_PORT")
	f.ServerAddr = os.Getenv("SERVER_ADDR")
	f.ServerPort = os.Getenv("SERVER_PORT")
	f.LogLevel = os.Getenv("LOG_LEVEL")
	f.ManagedServerType = os.Getenv("MANAGED_SERVER_TYPE")

	return f
}
