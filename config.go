package main

import (
	"git.openpunk.com/CPunch/copywriter/util"
	"github.com/go-ini/ini"
)

type ConfigData struct {
	TrendingCategory string `ini:"trend"`
	CustomPrompt     string `ini:"custom"`
	ImageStylePrompt string `ini:"image"`
	TopicType        string `ini:"topicType"` // can be "trends" or "news"
}

const (
	DEFAULT_TRENDING_CATEGORY = "all"
	TOPIC_TYPE_TRENDS         = "trends"
	TOPIC_TYPE_NEWS           = "news"
)

func NewConfig(TrendingCategory, CustomPrompt, ImageStylePrompt, TopicType string) *ConfigData {
	return &ConfigData{
		TrendingCategory: TrendingCategory,
		CustomPrompt:     CustomPrompt,
		ImageStylePrompt: ImageStylePrompt,
		TopicType:        TopicType,
	}
}

func (config *ConfigData) LoadConfig(filename string) {
	util.Info("Loading config file '%s'...", filename)
	cfg, err := ini.Load(filename)
	if err != nil {
		util.Warning("Failed to load config file: %v", err)
		return
	}

	err = cfg.MapTo(&config)
	if err != nil {
		util.Fail("Failed to map config file: %v", err)
	}

	if config.TopicType != TOPIC_TYPE_NEWS && config.TopicType != TOPIC_TYPE_TRENDS {
		util.Warning("Invalid topic type '%s', defaulting to '%s'", config.TopicType, TOPIC_TYPE_TRENDS)
		config.TopicType = TOPIC_TYPE_TRENDS
	}
}
