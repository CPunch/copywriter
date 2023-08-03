package main

import (
	"encoding/json"
	"fmt"
	"path"
	"strings"
	"unicode"

	"git.openpunk.com/CPunch/copywriter/imagescraper"
	"git.openpunk.com/CPunch/copywriter/replicate"
)

const (
	MAX_RETRY       = 5
	IMAGE_FREQUENCY = 10
)

type BlogWriter struct {
	config     *Config
	outDir     string
	imageCount int
	maxImages  int
	Title      string
	Content    string // markdown with injected images
	Tags       string
	Author     string
	Thumbnail  string
}

func NewBlogWriter(config *Config) *BlogWriter {
	return &BlogWriter{
		config:     config,
		imageCount: 0,
	}
}

func (bw *BlogWriter) setOutDir(outDir string) {
	bw.outDir = outDir
}

func (bw *BlogWriter) genImage(query string) string {
	query = strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsNumber(r) || unicode.IsSpace(r) {
			return r
		}
		return -1
	}, query)

	query = strings.Split(query, "\n")[0]
	query = strings.TrimSpace(query)

	Info("Generating image for query '%s'...", query)

	// check if REPLICATE_API_KEY is in our environment, if it's not we'll fallback to our image scraper
	var url string
	var headers []string
	if token := getEnv("REPLICATE_API_KEY", ""); token != "" {
		Info("Using replicate.ai to generate image...")

		rc := replicate.NewClient(token)

		var err error
		url, err = rc.MakePrediction(query)
		if err != nil {
			Fail("Failed to generate image: %v", err)
			return ""
		}

		headers = []string{"Authorization:Token " + token}
	} else {
		Info("Using image scraper to grab our image...")

		url = imagescraper.GetImageUrl(query)
		headers = []string{"User-Agent:Mozilla/5.0 (iPhone; CPU iPhone OS 16_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.5 Mobile/15E148 Safari/604.1"}
	}

	// download image
	bw.imageCount++
	fileName := fmt.Sprintf("image_%d.jpg", bw.imageCount)

	fmt.Printf("Downloading image to '%s'...\n", path.Join(bw.outDir, fileName))
	if err := downloadToFile(url, path.Join(bw.outDir, fileName), headers); err != nil {
		Fail("Failed to download image: %v", err)
		return ""
	}
	return fileName
}

func (bw *BlogWriter) genImageAboutMeta(prompt string) string {
	query := generateResponse(ResponseOptions{
		MaxTokens: 25,
		Prompt:    fmt.Sprintf("%s\n---\n search query to find a relevant image for the above text: ", prompt),
		UseGPT4:   false,
	})

	return bw.genImage(query)
}

func (bw *BlogWriter) populateImages(content string) string {
	Info("Populating images...")
	lines := strings.Split(content, "\n")

	// strip lines
	for i := range lines {
		lines[i] = strings.TrimSpace(lines[i])
	}

	for i := 1; i < len(lines); i++ {
		if i%IMAGE_FREQUENCY == 0 {
			// inject image
			imgURL := bw.genImageAboutMeta(strings.Join(lines[i-2:i], "\n"))
			img := fmt.Sprintf("\n![](%s)", imgURL)

			// insert line
			lines = append(lines[:i], append([]string{img}, lines[i:]...)...)
		}
	}

	return strings.Join(lines, "\n")
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

		return strings.TrimSpace(titleList[0])
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

func (bw *BlogWriter) WritePost() string {
	bw.Thumbnail = bw.genImageAboutMeta(bw.Title)
	bw.Content = bw.genBlogContent()
	bw.Tags = bw.genBlogTags()
	bw.Author = "Mason Coleman"

	header := bw.genHeaders()
	fullPost := fmt.Sprintf("%s\n%s", header, bw.Content)

	Success("Generated post!")
	return fullPost
}
