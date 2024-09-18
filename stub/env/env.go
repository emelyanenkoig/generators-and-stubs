package env

import "os"

type Environment struct {
	ResponseFilePath  string
	ServerPort        string
	ControlServerPort string
	Addr              string
	LogLevel          string
	ServerType        string
}

func ReadENV() Environment {
	var f Environment
	f.ResponseFilePath = os.Getenv("RESPONSE_FILE_PATH")
	f.ControlServerPort = os.Getenv("CONTROL_SERVER_PORT")
	f.ServerPort = os.Getenv("SERVER_PORT")
	f.Addr = os.Getenv("ADDR")
	f.LogLevel = os.Getenv("LOG_LEVEL")
	f.ServerType = os.Getenv("SERVER_TYPE")

	return f
}
