package main

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Logger LoggerConf `yaml:"logger"`
	RMQ    RMQConf    `yaml:"rmq"`
}

type LoggerConf struct {
	Level string `yaml:"level"`
}

type RMQConf struct {
	DSN              string `yaml:"dsn"`
	Exchange         string `yaml:"exchange"`
	Queue            string `yaml:"queue"`
	RoutingKey       string `yaml:"routingKey"`
	StatusQueue      string `yaml:"statusQueue"`
	StatusRoutingKey string `yaml:"statusRoutingKey"`
}

func NewConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	var cfg Config
	if err := yaml.Unmarshal([]byte(os.ExpandEnv(string(data))), &cfg); err != nil { //nolint:typecheck
		return Config{}, err
	}
	return cfg, nil
}
