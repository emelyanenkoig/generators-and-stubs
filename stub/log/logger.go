package log

import "go.uber.org/zap"

type Logger struct {
	*zap.Logger
}

func InitLogger(logLevel string) *zap.Logger {
	var logger *zap.Logger
	var err error

	switch logLevel {
	case "DEBUG":
		logger, err = zap.NewDevelopment()
	case "INFO":
		logger, err = zap.NewProduction()
	case "ERROR":
		logger, err = zap.NewProduction()
	case "WARN":
		logger, err = zap.NewProductionConfig().Build()
	default:
		logger, err = zap.NewProduction() // по умолчанию INFO уровень
	}

	if err != nil {
		panic("cannot initialize log")
	}
	return logger
}
