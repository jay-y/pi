package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jay-y/pi/pkg/ai"
)

func GetEnvConfig() map[string]any {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current working directory:", err)
		return nil
	}
	envFile := filepath.Join(cwd, ".env")
	envJson, err := os.ReadFile(envFile)
	if err != nil {
		fmt.Println("Error reading .env file:", err)
		return nil
	}
	var envVars map[string]any
	err = json.Unmarshal(envJson, &envVars)
	if err != nil {
		fmt.Println("Error unmarshalling .env file:", err)
		return nil
	}
	return envVars
}

func main() {
	envVars := GetEnvConfig()
	if envVars == nil {
		fmt.Println("Error getting .env config")
		return
	}
	baseURL, ok := envVars["baseURL"].(string)
	if !ok {
		fmt.Println("baseUrl not found in .env file")
		return
	}
	// 注册内置 API 提供者
	ai.RegisterBuiltInApiProviders()
	ollamaModel := &ai.BaseModel{
		ID:   "qwen3-coder-next:q8_0",
		Name: "ollama/qwen3-coder-next:q8_0",
		// ID:            "qwen3.5:122b",
		// Name:          "ollama/qwen3.5:122b",
		API:           ai.ModelApi(ai.ApiOpenAICompletions),
		Provider:      ai.ModelProvider("ollama"),
		BaseURL:       baseURL,
		Reasoning:     false,
		Input:         []string{"text"},
		Cost:          ai.ModelCost{},
		ContextWindow: 128000,
		MaxTokens:     32000,
	}

	// 构建对话上下文
	cxt := ai.Context{
		SystemPrompt: "You are a helpful assistant.",
		Messages: []ai.Message{
			ai.NewUserMessage("你好！测试下对话速度～"),
		},
	}

	// 流式调用
	stream, err := ai.Stream(ollamaModel, cxt, &ai.StreamOptions{
		Ctx:    context.Background(),
		APIKey: "ollama",
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	for event := range stream.Events() {
		// jsonEvent, _ := json.Marshal(event)
		// fmt.Printf("Event: %s\n", string(jsonEvent))
		switch e := event.(type) {
		case *ai.AssistantMessageEventTextDelta:
			fmt.Print(e.Delta)
		}
	}

	// 获取最终结果
	if result := stream.Result(); result != nil {
		fmt.Printf("\nTotal tokens: %d in, %d out\n", result.Usage.Input, result.Usage.Output)
		fmt.Printf("Cost: $%.4f\n", result.Usage.Cost.Total)
	}

	// 完整调用（非流式）
	response, err := ai.Complete(ollamaModel, cxt, &ai.StreamOptions{
		Ctx:    context.Background(),
		APIKey: "ollama",
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	for _, block := range response.Content {
		if t, ok := block.(*ai.TextContentBlock); ok {
			fmt.Println(t.Text)
		}
	}
}
