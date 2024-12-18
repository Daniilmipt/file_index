package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	General struct {
		ErrorRate         float64 `yaml:"error_rate"`
		ThreadCountIndex  int64   `yaml:"thread_count_index"`
		ThreadCountSearch int64   `yaml:"thread_count_search"`
	}
}

func InitConfig(configPathEnv string) *Config {
	configFilePath := os.Getenv(configPathEnv)
	if configFilePath == "" {
		panic(fmt.Errorf("environment variable %s is not set or empty", configPathEnv))
	}

	file, err := os.Open(configFilePath)
	if err != nil {
		panic(fmt.Errorf("failed to open config file: %v", err))
	}
	defer file.Close()

	// Create an instance of Config
	var config Config

	// Decode the YAML file into the Config struct
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		panic(fmt.Errorf("failed to parse config file: %v", err))
	}

	return &config
}
