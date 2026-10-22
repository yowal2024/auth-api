// Reconstructed from config.Load and config.applyEnv disassembly.
package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"strconv"
)

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config %s: %w", path, err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	applyEnv(&cfg)
	return &cfg, nil
}

func applyEnv(c *Config) {
	for name, target := range map[string]*string{
		"AUTH_API_ADDR": &c.Server.Addr, "AUTH_API_MODE": &c.Server.Mode,
		"AUTH_API_MYSQL_DSN": &c.MySQL.DSN, "AUTH_API_REDIS_ADDR": &c.Redis.Addr,
		"AUTH_API_REDIS_PASSWORD": &c.Redis.Password,
	} {
		if value := os.Getenv(name); value != "" {
			*target = value
		}
	}
	for name, target := range map[string]*int{
		"AUTH_API_REDIS_DB": &c.Redis.DB, "AUTH_API_SESSION_TTL": &c.Session.TTLSec,
		"AUTH_API_SESSION_ABSOLUTE_TTL": &c.Session.AbsoluteTTLSec,
	} {
		if value, err := strconv.Atoi(os.Getenv(name)); err == nil {
			*target = value
		}
	}
}
