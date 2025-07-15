package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

// Config 配置结构体
type Config struct {
	ChatModels ChatModelsConfig `yaml:"chat_models"`
	Settings   SettingsConfig   `yaml:"settings"`
}

// ChatModelsConfig 聊天模型配置
type ChatModelsConfig struct {
	Platform string       `yaml:"platform"`
	Configs  OpenAIConfig `yaml:"configs"`
}

// OpenAIConfig OpenAI配置
type OpenAIConfig struct {
	APIKey       string `yaml:"api_key"`
	ProxyURL     string `yaml:"proxy_url"`
	DefaultModel string `yaml:"default_model"`
	SummaryModel string `yaml:"summary_model"`
}

// SettingsConfig 设置配置
type SettingsConfig struct {
	MaxTokens int    `yaml:"max_tokens"`
	Stream    bool   `yaml:"stream"`
	Timeout   string `yaml:"timeout"`
}

// LoadConfig 加载配置文件
func LoadConfig(filename string) (*Config, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	// 环境变量覆盖配置文件
	if envKey := os.Getenv("OPENAI_API_KEY"); envKey != "" {
		config.ChatModels.Configs.APIKey = envKey
	}
	if envProxy := os.Getenv("OPENAI_PROXY_URL"); envProxy != "" {
		config.ChatModels.Configs.ProxyURL = envProxy
	}
	if envModel := os.Getenv("OPENAI_DEFAULT_MODEL"); envModel != "" {
		config.ChatModels.Configs.DefaultModel = envModel
	}

	return &config, nil
}
