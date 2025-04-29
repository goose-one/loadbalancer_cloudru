package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Server struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

type Backend struct {
	Host   string `yaml:"host"`
	Port   string `yaml:"port"`
	Scheme string `yaml:"scheme"`
}

type Logger struct {
	Level string `yaml:"level"`
}

type Database struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Username string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"db_name"`
}

type LoadBalancer struct {
	TimeHealthCheck    int    `yaml:"time_check"`
	EndpointHealtCheck string `yaml:"endpoint_health_check"`
}

type RateLimiter struct {
	MaxRequests int `yaml:"default_capacity"`
	Interval    int `yaml:"default_interaval"`
}
type Config struct {
	Server       Server       `yaml:"service"`
	Logger       Logger       `yaml:"logger"`
	DB           Database     `yaml:"db"`
	Backends     []Backend    `yaml:"backends"`
	RateLimiter  RateLimiter  `yaml:"rate_limiter"`
	LoadBalancer LoadBalancer `yaml:"load_balancer"`
}

func LoadConfig(filename string) (*Config, error) {
	fileClean := filepath.Clean(filename)
	f, err := os.Open(fileClean)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	config := &Config{}
	if err := yaml.NewDecoder(f).Decode(config); err != nil {
		return nil, err
	}

	return config, nil
}
