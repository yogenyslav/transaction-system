package config

import (
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Postgres struct {
		User     string `env:"POSTGRES_USER" env-default:"dbuser"`
		Password string `env:"POSTGRES_PASSWORD" env-default:"dbpass"`
		Host     string `env:"POSTGRES_HOST" env-default:"localhost"`
		Db       string `env:"POSTGRES_DB" env-default:"db"`
		Port     int    `env:"POSTGRES_PORT" env-default:"5432"`
	}
	Rabbit struct {
		User     string `env:"RABBIT_USER" env-default:"guest"`
		Password string `env:"RABBIT_PASSWORD" env-default:"guest"`
		Host     string `env:"RABBIT_HOST" env-default:"localhost"`
		Port     int    `env:"RABBIT_PORT" env-default:"5672"`
	}
	Server struct {
		Port int `env:"SERVER_PORT" env-default:"9999"`
	}
}

func MustNewConfig(path string) *Config {
	cfg := &Config{}
	errs := make([]error, 3)
	errs[0] = cleanenv.ReadConfig(path, &cfg.Postgres)
	errs[1] = cleanenv.ReadConfig(path, &cfg.Rabbit)
	errs[2] = cleanenv.ReadConfig(path, &cfg.Server)
	for _, err := range errs {
		if err != nil {
			panic(err)
		}
	}
	return cfg
}

func (cfg *Config) WithDbHost(host string) *Config {
	cfg.Postgres.Host = host
	return cfg
}
