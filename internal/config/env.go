package config

import "github.com/caarlos0/env/v11"

type Envs struct {
	Address      string `env:"SERVER_ADDRESS"`
	BaseShortURL string `env:"BASE_URL"`
}

func parseEnvs() *Envs {
	var cfg Envs
	err := env.Parse(&cfg)
	if err != nil {
		panic(err)
	}

	return &cfg
}