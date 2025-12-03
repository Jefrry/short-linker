package config

import (
	"flag"
)

type Flags struct {
	Address      string
	BaseShortURL string
	DatabaseDsn  string
	JWTSecret    string
}

func parseFlags() *Flags {
	address := flag.String("a", "localhost:8080", "Server address and port")
	baseShortURL := flag.String("b", "http://localhost:8080", "Base URL for shortened links")
	databaseDsn := flag.String("d", "", "Database DSN")
	jwtSecret := flag.String("j", "super-secret-key", "JWT secret key")

	flag.Parse()

	return &Flags{
		Address:      *address,
		BaseShortURL: *baseShortURL,
		DatabaseDsn:  *databaseDsn,
		JWTSecret:    *jwtSecret,
	}
}
