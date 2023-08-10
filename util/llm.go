package util

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode"

	openai "github.com/sashabaranov/go-openai"
)

const (
	MAX_CHAT_RETRY = 5
)

var (
	client *openai.Client
)

type ResponseOptions struct {
	MaxTokens int
	Prompt    string
	UseGPT4   bool
	/*
		Clean: removes any non alphanumeric characters, and trims
		whitespace. first \n is marked as EOF, so it will only return
		the first line
	*/
	Clean bool
}

func init() {
	client = openai.NewClient(GetEnv("OPENAI_API_KEY", ""))
}

func GenerateResponse(args ResponseOptions) (string, error) {
	var model string
	if args.UseGPT4 {
		model = openai.GPT4
	} else {
		model = openai.GPT3Dot5Turbo
	}

	Info("Generating response with prompt:\n%s", args.Prompt)

	var err error
	var resp openai.ChatCompletionResponse
	for i := 0; i < MAX_CHAT_RETRY; i++ {
		resp, err = client.CreateChatCompletion(
			context.Background(),
			openai.ChatCompletionRequest{
				Model:       model,
				MaxTokens:   args.MaxTokens,
				Temperature: 1,
				Messages: []openai.ChatCompletionMessage{
					{
						Role:    openai.ChatMessageRoleSystem,
						Content: args.Prompt,
					},
				},
			},
		)

		if err != nil {
			// try again but sleep for a bit
			time.Sleep(2 * time.Second)
			continue
		}

		// clean up text
		text := resp.Choices[0].Message.Content
		if args.Clean {
			text = strings.Map(func(r rune) rune {
				if unicode.IsLetter(r) || unicode.IsNumber(r) || unicode.IsSpace(r) {
					return r
				}
				return -1
			}, text)

			text = strings.Split(text, "\n")[0]
			text = strings.TrimSpace(text)
		}

		return text, nil
	}

	return "", fmt.Errorf("ChatCompletion error: %v\n", err)
}

func SummarizeText(text string) (string, error) {
	size := 4096
	chunks := []string{}
	for i := 0; i < len(text); i += size {
		end := i + size
		if end > len(text) {
			end = len(text)
		}
		chunks = append(chunks, text[i:end])
	}

	var summary string
	var err error
	for _, chunk := range chunks {
		summary, err = GenerateResponse(ResponseOptions{
			MaxTokens: 2000,
			UseGPT4:   false,
			Prompt:    fmt.Sprintf("Summarize the following text:\n\n%s\n%s\n\nSummary:", summary, chunk),
		})
		if err != nil {
			return "", err
		}
	}

	Info("Summary: %s", summary)
	return summary, nil
}
