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
		HTTPAddr:    os.Getenv("HTTP_ADDR"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		SupplierAPI: os.Getenv("SUPPLIER_URL"),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL missing")
	}

	if cfg.HTTPAddr == "" {
		return Config{}, fmt.Errorf("HTTP_ADDR missing")
	}

	if cfg.SupplierAPI == "" {
		return Config{}, fmt.Errorf("SUPPLIER_URL missing")
	}

	return cfg, nil
}
