package config

import (
	"gopkg.in/yaml.v3"
	"os"
	"time"
)

type Config struct {
	Server struct {
		Port string `yaml:"port"`
	} `yaml:"server"`

	Storage struct {
		TTL      time.Duration `yaml:"ttl"`
		RedisKey string        `yaml:"redis_key"`
		DumpFile string        `yaml:"dump_file"`
	} `yaml:"storage"`

	Logger struct {
		LogFile string `yaml:"log_file"`
	} `yaml:"logger"`
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
