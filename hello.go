package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/sashabaranov/go-openai"
	"io"
	"log"
	"time"
)

func main() {
	// 加载配置文件
	config, err := LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 验证必要的配置
	if config.ChatModels.Configs.APIKey == "" {
		log.Fatal("错误：API密钥不能为空，请在config.yaml中设置或使用环境变量OPENAI_API_KEY")
	}

	fmt.Printf("使用配置:\n")
	fmt.Printf("- 平台: %s\n", config.ChatModels.Platform)
	fmt.Printf("- 代理URL: %s\n", config.ChatModels.Configs.ProxyURL)
	fmt.Printf("- 默认模型: %s\n", config.ChatModels.Configs.DefaultModel)
	fmt.Printf("- 最大Token: %d\n", config.Settings.MaxTokens)
	fmt.Println("")

	// 创建自定义配置的OpenAI客户端
	clientConfig := openai.DefaultConfig(config.ChatModels.Configs.APIKey)
	clientConfig.BaseURL = config.ChatModels.Configs.ProxyURL

	c := openai.NewClientWithConfig(clientConfig)
	ctx := context.Background()

	req := openai.ChatCompletionRequest{
		Model:     config.ChatModels.Configs.DefaultModel,
		MaxTokens: config.Settings.MaxTokens,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: "Who is Isaac Newton?",
			},
		},
		Stream: config.Settings.Stream,
	}
	stream, err := c.CreateChatCompletionStream(ctx, req)
	if err != nil {
		fmt.Printf("ChatCompletionStream error: %v\n", err)
		return
	}
	defer stream.Close()

	fmt.Printf("Stream response: ")
	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			fmt.Println("\nStream finished")
			return
		}

		if err != nil {
			fmt.Printf("\nStream error: %v\n", err)
			return
		}

		// fmt.Printf(response.Choices[0].Delta.Content)
		fmt.Println("message：", time.Now(), response.Choices[0].Delta.Content)
	}
}
