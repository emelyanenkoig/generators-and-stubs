package config

import (
	"encoding/json"
	"os"
)

type RequestConfig struct {
	URL      string            `json:"url"`
	Method   string            `json:"method"`
	Headers  map[string]string `json:"headers"`
	Body     string            `json:"body"`
	Threads  int               `json:"threads"`
	Duration int               `json:"duration"` // Duration in seconds
}

type Config struct {
	Requests []RequestConfig `json:"requests"`
}

func LoadConfig(filePath string) (*Config, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	var config Config
	err = decoder.Decode(&config)
	return &config, err
}
