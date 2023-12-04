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
	UseLong   bool // if gpt4, will use 32k model. if gpt3, will use the 16k model
	/*
		Clean: removes any non alphanumeric characters, and trims
		whitespace. first \n is marked as EOF, so it will only return
		the first line
	*/
	Clean                 bool
	CleanKeepPunctuations bool
}

func init() {
	client = openai.NewClient(GetEnv("OPENAI_API_KEY", ""))
}

func GenerateResponse(args ResponseOptions) (string, error) {
	var model string
	if args.UseGPT4 {
		if args.UseLong {
			model = openai.GPT432K
		} else {
			model = openai.GPT4
		}
	} else {
		if args.UseLong {
			model = openai.GPT3Dot5Turbo16K
		} else {
			model = openai.GPT3Dot5Turbo
		}
	}

	// Info("Generating response with prompt:\n%s", args.Prompt)

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
			// not recoverable errors
			if err == openai.ErrChatCompletionInvalidModel ||
				err == openai.ErrChatCompletionStreamNotSupported ||
				err == openai.ErrCompletionRequestPromptTypeNotSupported ||
				err == openai.ErrCompletionStreamNotSupported ||
				err == openai.ErrCompletionUnsupportedModel {
				return "", err
			}

			timeToSleep := 2 * time.Second
			if strings.Contains(err.Error(), "Rate limit reached") {
				timeToSleep = 5 * time.Second
			}

			// try again but sleep for a bit
			time.Sleep(timeToSleep)
			continue
		}

		// clean up text
		text := resp.Choices[0].Message.Content
		if args.Clean {
			text = strings.Map(func(r rune) rune {
				if unicode.IsLetter(r) || unicode.IsNumber(r) || unicode.IsSpace(r) {
					return r
				}

				if args.CleanKeepPunctuations {
					if r == '.' || r == ',' || r == '!' || r == '?' ||
						r == ':' || r == ';' || r == '-' {
						return r
					}
				}
				return -1
			}, text)

			text = strings.Split(text, "\n")[0]
			text = strings.TrimSpace(text)
		}

		return text, nil
	}

	return "", fmt.Errorf("ChatCompletion error: %v", err)
}

func SummarizeText(text string) (string, error) {
	size := 8096
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
			MaxTokens: 6000,
			UseGPT4:   false,
			UseLong:   true,
			Prompt:    fmt.Sprintf("Summarize the following text while retaining all relevant information:\n\n%s\n%s\n\nSummary:", summary, chunk),
		})
		if err != nil {
			return "", err
		}
	}

	Info("Summary: %s", summary)
	return summary, nil
}
