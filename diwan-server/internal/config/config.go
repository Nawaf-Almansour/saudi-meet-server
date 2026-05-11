package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server    ServerConfig    `yaml:"server"`
	Database  DatabaseConfig  `yaml:"database"`
	Redis     RedisConfig     `yaml:"redis"`
	Nats      NatsConfig      `yaml:"nats"`
	PlugNmeet PlugNmeetConfig `yaml:"plugnmeet"`
}

type ServerConfig struct {
	Port      int    `yaml:"port"`
	JWTSecret string `yaml:"jwt_secret"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
	SSLMode  string `yaml:"sslmode"`
}

type RedisConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type NatsConfig struct {
	URL string `yaml:"url"`
}

type PlugNmeetConfig struct {
	Host       string `yaml:"host"`
	APIKey     string `yaml:"api_key"`
	APISecret  string `yaml:"api_secret"`
	ClientHost string `yaml:"client_host"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
