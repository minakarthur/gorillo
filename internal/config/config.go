package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Journal  JournalConfig  `yaml:"journal"`
}

type ServerConfig struct {
	Port string `yaml:"port"`
}

type DatabaseConfig struct {
	Path string `yaml:"path"`
}

type JournalConfig struct {
	RecentCount int `yaml:"recent_count"`
}

func Load(path string) (*Config, error) {
	cfg := &Config{
		Server:   ServerConfig{Port: "8000"},
		Database: DatabaseConfig{Path: "gorillo.db"},
		Journal:  JournalConfig{RecentCount: 10},
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("read config %s: %w", path, err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if cfg.Journal.RecentCount < 1 {
		cfg.Journal.RecentCount = 10
	}

	return cfg, nil
}

func (c *Config) DSN() string {
	return c.Database.Path
}
