package main

import (
	"os"

	"gopkg.in/yaml.v3" //nolint: depguard
)

type Config struct {
	Logger    LoggerConf    `yaml:"logger"`
	Storage   StorageConf   `yaml:"storage"`
	DB        DBConf        `yaml:"db"`
	RMQ       RMQConf       `yaml:"rmq"`
	Scheduler SchedulerConf `yaml:"scheduler"`
}

type LoggerConf struct {
	Level string `yaml:"level"`
}

type StorageConf struct {
	Type string `yaml:"type"`
}

type DBConf struct {
	DSN string `yaml:"dsn"`
}

type RMQConf struct {
	DSN        string `yaml:"dsn"`
	Exchange   string `yaml:"exchange"`
	Queue      string `yaml:"queue"`
	RoutingKey string `yaml:"routingKey"`
}

type SchedulerConf struct {
	ScanInterval string `yaml:"scanInterval"`
	Retention    string `yaml:"retention"`
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
