package config

import "flag"

type Flags struct {
	Address      string
	BaseShortURL string
}

func ParseFlags() *Flags {
	address := flag.String("a", "localhost:8080", "Server address and port")
	baseShortURL := flag.String("b", "http://localhost:8080", "Base URL for shortened links")

	flag.Parse()

	return &Flags{
		Address:      *address,
		BaseShortURL: *baseShortURL,
	}
}