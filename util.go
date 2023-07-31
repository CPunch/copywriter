package main

import (
	"context"
	"fmt"
	"os"
	"time"

	openai "github.com/sashabaranov/go-openai"
)

const (
	MAX_CHAT_RETRY = 5
)

func getEnv(key string, fallback string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	return value
}

func getTimeString() string {
	dt := time.Now()
	return fmt.Sprintf(dt.Format("2006-01-02"))
}

type ResponseOptions struct {
	MaxTokens int
	Prompt    string
	UseGPT4   bool
}

func generateResponse(options ResponseOptions) string {
	var model string
	if options.UseGPT4 {
		model = openai.GPT4
	} else {
		model = openai.GPT3Dot5Turbo
	}

	client := openai.NewClient(getEnv("OPENAI_API_KEY", ""))

	var err error
	var resp openai.ChatCompletionResponse
	for i := 0; i < MAX_CHAT_RETRY; i++ {
		resp, err = client.CreateChatCompletion(
			context.Background(),
			openai.ChatCompletionRequest{
				Model:       model,
				MaxTokens:   options.MaxTokens,
				Temperature: 1,
				Messages: []openai.ChatCompletionMessage{
					{
						Role:    openai.ChatMessageRoleSystem,
						Content: options.Prompt,
					},
				},
			},
		)

		if err != nil {
			// try again but sleep for a bit
			time.Sleep(1 * time.Second)
			continue
		}

		return resp.Choices[0].Message.Content
	}

	panic(fmt.Sprintf("ChatCompletion error: %v\n", err))
}
