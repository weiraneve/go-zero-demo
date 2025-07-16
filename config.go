package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

type Config struct {
	ChatModels struct {
		Platform string `yaml:"platform"`
		Configs  struct {
			APIKey       string `yaml:"api_key"`
			ProxyURL     string `yaml:"proxy_url"`
			DefaultModel string `yaml:"default_model"`
		} `yaml:"configs"`
	} `yaml:"chat_models"`
	Settings struct {
		MaxTokens int  `yaml:"max_tokens"`
		Stream    bool `yaml:"stream"`
		Port      int  `yaml:"port"`
	} `yaml:"settings"`
}

func LoadConfig(filename string) (*Config, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	// 从环境变量获取API密钥（如果存在）
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		config.ChatModels.Configs.APIKey = apiKey
	}

	return &config, nil
}
