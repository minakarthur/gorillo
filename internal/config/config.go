package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Server   ServerConfig   `toml:"server"`
	Database DatabaseConfig `toml:"database"`
	Journal  JournalConfig  `toml:"journal"`
}

type ServerConfig struct {
	Port string `toml:"port"`
}

type DatabaseConfig struct {
	Host     string `toml:"host"`
	Port     string `toml:"port"`
	User     string `toml:"user"`
	Password string `toml:"password"`
	Name     string `toml:"name"`
}

type JournalConfig struct {
	RecentCount int `toml:"recent_count"`
}

func Load(path string) (*Config, error) {
	cfg := &Config{
		Server:   ServerConfig{Port: "8000"},
		Database: DatabaseConfig{Host: "127.0.0.1", Port: "3306", User: "root", Name: "gorillo"},
		Journal:  JournalConfig{RecentCount: 10},
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config %s: %w", path, err)
	}

	if err := toml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if cfg.Journal.RecentCount < 1 {
		cfg.Journal.RecentCount = 10
	}

	return cfg, nil
}

func (c *Config) DSN() string {
	d := c.Database
	return d.User + ":" + d.Password + "@tcp(" + d.Host + ":" + d.Port + ")/" + d.Name + "?parseTime=true&charset=utf8mb4"
}
