package config

import (
	"os"

	"gopkg.in/yaml.v3" //nolint: depguard
)

type Config struct {
	Logger  LoggerConf  `yaml:"logger"`
	HTTP    HTTPConf    `yaml:"http"`
	GRPC    GRPCConf    `yaml:"grpc"`
	Storage StorageConf `yaml:"storage"`
	DB      DBConf      `yaml:"db"`
}

type LoggerConf struct {
	Level string `yaml:"level"`
}

type HTTPConf struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

type GRPCConf struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

type StorageConf struct {
	Type string `yaml:"type"`
}

type DBConf struct {
	DSN string `yaml:"dsn"`
}

func NewConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil { //nolint: typecheck
		return Config{}, err
	}
	return cfg, nil
}
