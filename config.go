package main

import (
	"github.com/go-ini/ini"
)

type ConfigData struct {
	TrendingCategory string `ini:"trend"`
	CustomPrompt     string `ini:"custom"`
	ImageStylePrompt string `ini:"image"`
}

func NewConfig(TrendingCategory, CustomPrompt, ImageStylePrompt string) *ConfigData {
	return &ConfigData{
		TrendingCategory: TrendingCategory,
		CustomPrompt:     CustomPrompt,
		ImageStylePrompt: ImageStylePrompt,
	}
}

func (config *ConfigData) LoadConfig(filename string) {
	Info("Loading config file '%s'...", filename)
	cfg, err := ini.Load(filename)
	if err != nil {
		Warning("Failed to load config file: %v", err)
		return
	}

	err = cfg.MapTo(&config)
	if err != nil {
		Fail("Failed to map config file: %v", err)
	}
}
