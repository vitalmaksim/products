package config

import (
	"fmt"
	"os"
)

type Config struct {
	HTTPAddr    string
	DatabaseURL string
	SupplierAPI string
}

func Load() (Config, error) {
	cfg := Config{
		HTTPAddr:    os.Getenv("HTTPAddr"),
		DatabaseURL: os.Getenv("DatabaseURL"),
		SupplierAPI: os.Getenv("SupplierAPI"),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("database url missing")
	}

	if cfg.HTTPAddr == "" {
		return Config{}, fmt.Errorf("http addr url missing")
	}

	if cfg.SupplierAPI == "" {
		return Config{}, fmt.Errorf("supplier api url missing")
	}

	return cfg, nil
}
