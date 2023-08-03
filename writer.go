package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"git.openpunk.com/CPunch/copywriter/imagescraper"
)

const (
	MAX_RETRY = 5
)

type BlogWriter struct {
	config    *Config
	outDir    string
	Title     string
	Content   string // markdown with injected images
	Tags      string
	Author    string
	Thumbnail string
}

func (bw *BlogWriter) genImageAboutMeta(prompt string) string {
	query := generateResponse(ResponseOptions{
		MaxTokens: 25,
		Prompt:    fmt.Sprintf("%s\n---\n search query to find a relevant image for the above text: ", prompt),
		UseGPT4:   false,
	})

	return imagescraper.GetImageUrl(query)
}

func (bw *BlogWriter) populateImages(content string) string {
	Info("Scraping for images...")
	lines := strings.Split(content, "\n")

	// strip lines
	for i := range lines {
		lines[i] = strings.TrimSpace(lines[i])
	}

	for i := 1; i < len(lines); i++ {
		if i%5 == 0 {
			// inject image
			imgURL := bw.genImageAboutMeta(strings.Join(lines[i-2:i], "\n"))
			img := fmt.Sprintf("\n![](%s)", imgURL)

			// insert line
			lines = append(lines[:i], append([]string{img}, lines[i:]...)...)
		}
	}

	return strings.Join(lines, "\n")
}

func NewBlogWriter(config *Config, outDir string) *BlogWriter {
	return &BlogWriter{
		config: config,
		outDir: outDir,
	}
}

func (bw *BlogWriter) genBlogTitle() string {
	Info("Generating blog title...")
	trends := getPopularTrends(bw.config.TrendingCategory)

	for i := 0; i < MAX_RETRY; i++ { // just in case gpt is a DUMBASS; i don't wanna burn a million dollars
		titles := generateResponse(ResponseOptions{
			MaxTokens: 40,
			Prompt:    fmt.Sprintf("%s\nThe following is a list of topics:\n%s\n\nAn example of a short, creative and eye-catching title of a blog post that matches these topics: ", bw.config.CustomPrompt, strings.Join(trends, "\n")),
			UseGPT4:   false,
		})

		titles = strings.ReplaceAll(titles, "\"", "")

		// split titles by '\n'
		titleList := strings.Split(titles, "\n")

		// no titles?? try again
		if len(titleList) == 0 {
			continue
		}

		return titleList[0]
	}

	Fail("Failed to create title!") // this calls os.exit, so the following return is just to fix golang warnings
	return ""
}

func (bw *BlogWriter) genBlogTags() string {
	Info("Generating tags...")
	for i := 0; i < MAX_RETRY; i++ { // just in case gpt is a DUMBASS; i don't wanna burn a million dollars
		tagString := generateResponse(ResponseOptions{
			MaxTokens: 50,
			Prompt:    fmt.Sprintf("%s\n\nTags as a json array with only 1 word each, max 5:\n", bw.Content),
			UseGPT4:   false,
		})

		tagString = strings.ReplaceAll(tagString, "```", "")

		// try to unmarshal the tags, if it fails, try again!
		var tags []string
		if err := json.Unmarshal([]byte(tagString), &tags); err != nil {
			continue
		}

		return tagString
	}

	Warning("GPT failed to generate any valid tags")
	return "[]"
}

func (bw *BlogWriter) genBlogContent() string {
	Info("Generating blog post contents...")
	markdown := generateResponse(ResponseOptions{
		MaxTokens: 5000,
		Prompt:    fmt.Sprintf("%s\nThe following is a blog post written in markdown about %s, minimum 1000 words. Use '##' for section headings. *DO NOT INCLUDE THE TITLE*:\n", bw.config.CustomPrompt, bw.Title),
		UseGPT4:   true,
	})

	// inject images
	return bw.populateImages(markdown)
}

func (bw *BlogWriter) genHeaders() string {
	return fmt.Sprintf("---\ntitle: \"%s\"\nauthor: \"%s\"\ndate: \"%s\"\ntags: %s\nimage: \"%s\"\n---\n\n", bw.Title, bw.Author, getTimeString(), bw.Tags, bw.Thumbnail)
}

// passing an empty string "" will force us to generate the title using google trends
func (bw *BlogWriter) setTitle(title string) {
	if title == "" {
		title = bw.genBlogTitle()
	}

	Info("Title: '%s'...", title)
	bw.Title = title
}

func (bw *BlogWriter) WritePost(title string) string {
	bw.setTitle(title)
	bw.Content = bw.genBlogContent()
	bw.Tags = bw.genBlogTags()
	bw.Thumbnail = bw.genImageAboutMeta(bw.Title)
	bw.Author = "Mason Coleman"

	header := bw.genHeaders()
	fullPost := fmt.Sprintf("%s\n%s", header, bw.Content)

	Success("Generated post!")
	return fullPost
}
