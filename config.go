package main

import (
	"github.com/go-ini/ini"
)

type Config struct {
	TrendingCategory string `ini:"trending_category"`
	CustomPrompt     string `ini:"custom_prompt"`
}

func LoadConfig(filename string) *Config {
	Info("Loading config file '%s'...", filename)
	cfg, err := ini.Load(filename)
	if err != nil {
		Fail("Failed to load config file: %v", err)
	}

	var config Config
	err = cfg.MapTo(&config)
	if err != nil {
		Fail("Failed to map config file: %v", err)
	}

	return &config
}
