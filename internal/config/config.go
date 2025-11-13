package config

type Config struct {
	Address      string
	BaseShortURL string
}

func GetConfig() *Config {
	envs := parseEnvs()
	flags := parseFlags()

	currentAdress := flags.Address
	if envs.Address != "" {
		currentAdress = envs.Address
	}
	currentBaseURL := flags.BaseShortURL
	if envs.BaseShortURL != "" {
		currentBaseURL = envs.BaseShortURL
	}

	return &Config{
		Address:      currentAdress,
		BaseShortURL: currentBaseURL,
	}
}