package config

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"
)

const DefaultConfigPath = "/data/config.json"

func LoadConfig() (*Config, error) {
	cfg := DefaultConfig()

	if err := loadFromFile(cfg); err != nil {
		loadFromEnv(cfg)
	} else {
		loadFromEnv(cfg)
	}

	if cfg.JWT.Secret == "" {
		cfg.JWT.Secret = GenerateJWTSecret()
	}

	return cfg, nil
}

func LoadConfigFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	cfg := DefaultConfig()
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config file: %w", err)
	}

	return cfg, nil
}

func loadFromFile(cfg *Config) error {
	path := os.Getenv("FASMAIL_CONFIG_PATH")
	if path == "" {
		path = DefaultConfigPath
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, cfg)
}

func loadFromEnv(cfg *Config) {
	if v := os.Getenv("FASMAIL_PORT"); v != "" {
		cfg.Server.Port = v
	}
	if v := os.Getenv("FASMAIL_DB_HOST"); v != "" {
		cfg.Database.Host = v
	}
	if v := os.Getenv("FASMAIL_DB_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			cfg.Database.Port = port
		}
	}
	if v := os.Getenv("FASMAIL_DB_USER"); v != "" {
		cfg.Database.User = v
	}
	if v := os.Getenv("FASMAIL_DB_PASSWORD"); v != "" {
		cfg.Database.Password = v
	}
	if v := os.Getenv("FASMAIL_DB_NAME"); v != "" {
		cfg.Database.DBName = v
	}
	if v := os.Getenv("FASMAIL_DB_SSLMODE"); v != "" {
		cfg.Database.SSLMode = v
	}
	if v := os.Getenv("FASMAIL_REDIS_ADDR"); v != "" {
		cfg.Redis.Addr = v
	}
	if v := os.Getenv("FASMAIL_REDIS_PASSWORD"); v != "" {
		cfg.Redis.Password = v
	}
	if v := os.Getenv("FASMAIL_JWT_SECRET"); v != "" {
		cfg.JWT.Secret = v
	}
	if v := os.Getenv("FASMAIL_JWT_ACCESS_TTL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.JWT.AccessTokenTTL = d
		}
	}
	if v := os.Getenv("FASMAIL_JWT_REFRESH_TTL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.JWT.RefreshTokenTTL = d
		}
	}
}

func SaveConfig(cfg *Config, path string) error {
	if path == "" {
		path = DefaultConfigPath
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.MkdirAll("/data", 0755); err != nil {
		dir := os.TempDir()
		path = dir + "/config.json"
		if err2 := os.MkdirAll(dir, 0755); err2 != nil {
			return fmt.Errorf("create config dir: %w", err2)
		}
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}

	return nil
}

func GenerateJWTSecret() string {
	b := make([]byte, 64)
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Sprintf("failed to generate JWT secret: %v", err))
	}
	return hex.EncodeToString(b)
}

func GeneratePassword(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Sprintf("failed to generate password: %v", err))
	}
	return hex.EncodeToString(b)[:length]
}

func ConfigFileExists() bool {
	path := os.Getenv("FASMAIL_CONFIG_PATH")
	if path == "" {
		path = DefaultConfigPath
	}
	_, err := os.Stat(path)
	return err == nil
}
