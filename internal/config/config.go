package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Postgres PostgresConfig `yaml:"postgres"`
	JWT      JWTConfig      `yaml:"jwt"`
	Vault    VaultConfig    `yaml:"vault"`
	Site     SiteConfig     `yaml:"site"`
}

type ServerConfig struct {
	Host  string `yaml:"host"`
	Port  int    `yaml:"port"`
	Debug bool   `yaml:"debug"`
}

type PostgresConfig struct {
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	User         string `yaml:"user"`
	Password     string `yaml:"password"`
	DBName       string `yaml:"dbname"`
	SSLMode      string `yaml:"sslmode"`
	MaxIdleConns int    `yaml:"max_idle_conns"`
	MaxOpenConns int    `yaml:"max_open_conns"`
}

type JWTConfig struct {
	AccessSecret      string `yaml:"access_secret"`
	AccessExpireHours int    `yaml:"access_expire_hours"`
}

type VaultConfig struct {
	MasterKey string `yaml:"master_key"`
}

type SiteConfig struct {
	Name        string `yaml:"name"`
	BrandFA     string `yaml:"brand_fa"`
	Domain      string `yaml:"domain"`
	Currency    string `yaml:"currency"`
	SupportText string `yaml:"support_text"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// Environment variable overrides
	if host := os.Getenv("POSTGRES_HOST"); host != "" {
		cfg.Postgres.Host = host
	}
	if pass := os.Getenv("POSTGRES_PASSWORD"); pass != "" {
		cfg.Postgres.Password = pass
	}
	if db := os.Getenv("POSTGRES_DB"); db != "" {
		cfg.Postgres.DBName = db
	}
	if key := os.Getenv("VAULT_MASTER_KEY"); key != "" {
		cfg.Vault.MasterKey = key
	}
	if jwtSecret := os.Getenv("JWT_ACCESS_SECRET"); jwtSecret != "" {
		cfg.JWT.AccessSecret = jwtSecret
	}

	return &cfg, nil
}
