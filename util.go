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

	"github.com/gocolly/colly"
	openai "github.com/sashabaranov/go-openai"
)

const (
	MAX_CHAT_RETRY = 5
	USER_AGENT     = "Mozilla/5.0 (iPhone; CPU iPhone OS 16_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.5 Mobile/15E148 Safari/604.1"
)

func getEnv(key string, fallback string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	return value
}

// convert url to markdown
func scrapeArticle(url string) (string, error) {
	Info("Scraping article '%s'...", url)

	md := ""
	c := colly.NewCollector()
	c.UserAgent = USER_AGENT
	c.AllowURLRevisit = true
	c.DisableCookies()

	// scrape all images from a page
	c.OnHTML("p", func(e *colly.HTMLElement) {
		md += "\n" + e.Text
	})

	c.Visit(url)
	return md, nil
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

	// Warning("GPT Prompt:\n%s", args.Prompt)

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

		// Info("Got response: %s", text)
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
		summary, err = generateResponse(ResponseOptions{
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
