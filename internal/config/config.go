package config

import (
	"fmt"
	"os"

	"github.com/caarlos0/env/v10"
	"github.com/joho/godotenv"
)

type Config struct {
	Env      string   `env:"SURGE_ENV,required"`
	Port     int      `env:"SURGE_PORT,required"`
	LogLevel string   `env:"SURGE_LOG_LEVEL" envDefault:"info"`
	Postgres Postgres `envPrefix:"POSTGRES_"`
	Redis    Redis    `envPrefix:"REDIS_"`
}

type Postgres struct {
	Host     string `env:"HOST,required"`
	Port     int    `env:"PORT,required"`
	SSLMode  string `env:"SSLMODE,required"`
	User     string `env:"USER,required"`
	Password string `env:"PASSWORD,required"`
	DB       string `env:"DB,required"`
}

type Redis struct {
	Host     string `env:"HOST,required"`
	Port     int    `env:"PORT,required"`
	Password string `env:"PASSWORD,required"`
	DB       int    `env:"DB,required"`
}

func Load() (*Config, error) {
	if os.Getenv("SURGE_ENV") != "production" {
		_ = godotenv.Load(".env.local")
	}

	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}
