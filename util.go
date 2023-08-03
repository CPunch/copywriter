package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
	"unicode"

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

type DownloadOptions struct {
	URL      string
	FilePath string
	Header   http.Header
}

func downloadToFile(args DownloadOptions) error {
	Info("Downloading %s to '%s'...", args.URL, args.FilePath)

	req, err := http.NewRequest("GET", args.URL, nil)
	if err != nil {
		return err
	}

	// add headers
	req.Header = args.Header

	// make the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// check status code
	if resp.StatusCode != 200 {
		return fmt.Errorf("Bad status code: %d", resp.StatusCode)
	}

	// write response body to file
	f, err := os.Create(args.FilePath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func getTimeString() string {
	dt := time.Now()
	return fmt.Sprintf(dt.Format("2006-01-02 15:04:05"))
}

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

func generateResponse(args ResponseOptions) (string, error) {
	var model string
	if args.UseGPT4 {
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
