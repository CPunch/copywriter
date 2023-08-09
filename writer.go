package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
	"unicode"

	"git.openpunk.com/CPunch/copywriter/imagescraper"
	"git.openpunk.com/CPunch/copywriter/replicate"
	"git.openpunk.com/CPunch/copywriter/trendscraper"
	"git.openpunk.com/CPunch/copywriter/util"
)

const (
	MAX_RETRY = 5
)

type BlogWriter struct {
	config     *ConfigData
	outDir     string
	imageCount int
	maxImages  int
	TitleCtx   string
	ArticleCtx string
	Title      string
	Content    string // markdown with injected images
	Tags       string
	Author     string
	Thumbnail  string
}

func genBlogFilePath(title string) string {
	// strip any non-alphanumeric characters
	title = strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsNumber(r) || unicode.IsSpace(r) {
			return r
		}
		return -1
	}, title)

	title = strings.TrimSpace(title)
	title = strings.ReplaceAll(title, " ", "-")
	title = strings.ToLower(title)

	return title
}

func NewBlogWriter(config *ConfigData) *BlogWriter {
	return &BlogWriter{
		config:     config,
		imageCount: 0,
	}
}

// builds the output directory for the blog writer
func (bw *BlogWriter) setupOutDir(outDir string) {
	dirPath := path.Join(outDir, genBlogFilePath(bw.Title))
	if err := os.MkdirAll(dirPath, 0777); err != nil {
		util.Fail("Failed to create directory '%s': %v", dirPath, err)
	}
	bw.outDir = dirPath
}

func (bw *BlogWriter) getNextFile() (fileName, filePath string) {
	bw.imageCount++
	fileName = fmt.Sprintf("file_%d.jpg", bw.imageCount)
	filePath = path.Join(bw.outDir, fileName)
	return
}

// generate or scrapes the web for the query.
// returns the filename of the downloaded image
// in the outDir
func (bw *BlogWriter) genImage(query string) (string, error) {
	if bw.config.ImageStylePrompt != "" {
		query = query + " " + strings.TrimSpace(bw.config.ImageStylePrompt)
	}

	util.Info("Generating image for query '%s'...", query)

	fileName, filePath := bw.getNextFile()
	header := make(http.Header)
	var url string

	// check if REPLICATE_API_KEY is in our environment, if it's not we'll fallback to our image scraper
	if token := util.GetEnv("REPLICATE_API_KEY", ""); token != "" {
		util.Info("Using replicate.ai to generate image...")

		rc := replicate.NewClient(token)
		header = rc.Header

		var err error
		url, err = rc.MakePrediction(query)
		if err != nil {
			return "", fmt.Errorf("Failed to generate image: %v", err)
		}
	} else {
		util.Info("Using image scraper to grab an image...")

		url = imagescraper.GetImageUrl(query)
		header.Set("User-Agent", util.USER_AGENT)
	}

	// download image
	if err := util.DownloadToFile(util.DownloadOptions{
		URL:      url,
		FilePath: filePath,
		Header:   header,
	}); err != nil {
		return "", fmt.Errorf("Failed to download image: %v", err)
	}

	return fileName, nil
}

func (bw *BlogWriter) genImageAboutMeta(prompt string) (img string, query string, err error) {
	query, err = util.GenerateResponse(util.ResponseOptions{
		MaxTokens: 30,
		Prompt:    fmt.Sprintf("%s\n---\nWrite a short one sentence prompt for an image that fits the above text: Image of ", prompt),
		UseGPT4:   true,
		Clean:     true,
	})
	if err != nil {
		return "", "", fmt.Errorf("Failed to generate image: %v", err)
	}

	img, err = bw.genImage(query)
	return
}

func (bw *BlogWriter) populateImages(content string) (string, error) {
	util.Info("Populating images...")
	lines := strings.Split(content, "\n")

	// look for '![]('
	for i := 0; i < len(lines); i++ {
		if strings.Contains(lines[i], "![](") {
			imgPrompt := strings.ReplaceAll(lines[i], "![](", "")
			imgPrompt = strings.ReplaceAll(imgPrompt, ")", "")

			// inject image
			imgURL, err := bw.genImage(imgPrompt)
			if err != nil {
				return "", fmt.Errorf("Failed to generate image: %v", err)
			}

			lines[i] = fmt.Sprintf("\n![](%s)", imgURL)
		}

		// gpt sometimes writes this at the end of the content, so just remove everything after
		if lines[i] == "---" {
			lines = lines[:i]
			break
		}
	}

	return strings.Join(lines, "\n"), nil
}

func (bw *BlogWriter) genBlogTitle() (string, error) {
	util.Info("Generating blog title...")

	title, err := util.GenerateResponse(util.ResponseOptions{
		MaxTokens: 40,
		Prompt: fmt.Sprintf(
			"Write a short, simple and eye-catching title of an article that readers of the following text would be interested in:\n\n%s\nThe following is a list of articles the reader is interested in:\n\n%s\n\nTitle: ",
			bw.config.CustomPrompt, bw.TitleCtx,
		),
		UseGPT4: false,
		Clean:   true,
	})
	if err != nil {
		return "", err
	}

	// no title?
	if len(title) == 0 {
		return "", fmt.Errorf("Failed to create title!")
	}

	return title, nil
}

func (bw *BlogWriter) genBlogTags() (string, error) {
	util.Info("Generating tags...")
	for i := 0; i < MAX_RETRY; i++ { // just in case gpt is a DUMBASS; i don't wanna burn a million dollars
		tagString, err := util.GenerateResponse(util.ResponseOptions{
			MaxTokens: 50,
			Prompt:    fmt.Sprintf("%s\n\nTags as a json array with only 1 word each, max 5:\n", bw.Content),
			UseGPT4:   false,
		})
		if err != nil {
			return "", err
		}

		tagString = strings.ReplaceAll(tagString, "```", "")

		// try to unmarshal the tags, if it fails, try again!
		var tags []string
		if err := json.Unmarshal([]byte(tagString), &tags); err != nil {
			continue
		}

		return tagString, nil
	}

	util.Warning("GPT failed to generate any valid tags")
	return "[]", nil
}

func (bw *BlogWriter) genBlogContent() (string, error) {
	thumb, thumbnailQuery, err := bw.genImageAboutMeta(bw.Title)
	if err != nil {
		return "", fmt.Errorf("Failed to generate thumbnail: %v", err)
	}
	bw.Thumbnail = thumb

	util.Info("Generating blog post contents...")
	markdown, err := util.GenerateResponse(util.ResponseOptions{
		MaxTokens: 5000,
		Prompt: fmt.Sprintf(
			"%s\n%s\nWrite an interesting and informative 1000 word article that readers would find relevant written in markdown. Use '##' for section headings. Mark where you would insert an image using '![](<DESCRIPTION OF IMAGE>)'.\n---\n\n## %s\n\n![](%s)\n",
			bw.config.CustomPrompt, bw.ArticleCtx, bw.Title, thumbnailQuery,
		),
		UseGPT4: true,
		Clean:   false,
	})
	if err != nil {
		return "", err
	}

	// inject images
	return bw.populateImages(markdown)
}

func (bw *BlogWriter) genHeaders() string {
	return fmt.Sprintf(
		"---\ntitle: \"%s\"\nauthor: \"%s\"\ndate: \"%s\"\ntags: %s\nimage: \"%s\"\n---\n",
		bw.Title, bw.Author, util.GetTimeString(), bw.Tags, bw.Thumbnail,
	)
}

func (bw *BlogWriter) genTopicCtx() (err error) {
	if bw.config.TopicType == TOPIC_TYPE_NEWS {
		bw.TitleCtx, bw.ArticleCtx, err = trendscraper.ScrapeRealtimeNews(bw.config.TrendingCategory)
		return
	}

	bw.TitleCtx, bw.ArticleCtx, err = trendscraper.ScrapePopularTrends(bw.config.TrendingCategory)
	return
}

// passing an empty string "" will generate the title using the selected topic type
func (bw *BlogWriter) setTitle(title string) error {
	if title == "" {
		err := bw.genTopicCtx()
		if err != nil {
			return err
		}

		title, err = bw.genBlogTitle()
		if err != nil {
			return err
		}
	}

	util.Info("Title: '%s'...", title)
	bw.Title = title
	return nil
}

func (bw *BlogWriter) WritePost() error {
	var err error
	bw.Content, err = bw.genBlogContent()
	if err != nil {
		return fmt.Errorf("Failed to generate blog content: %v", err)
	}

	bw.Tags, err = bw.genBlogTags()
	if err != nil {
		return fmt.Errorf("Failed to generate blog tags: %v", err)
	}
	bw.Author = "Mason Coleman"

	header := bw.genHeaders()
	fullPost := fmt.Sprintf("%s\n%s", header, bw.Content)
	util.Success("Generated post!")

	// write hugo markdown file
	outFile := path.Join(bw.outDir, "index.md")
	util.Info("Writing to file '%s'...", outFile)
	if err := os.WriteFile(outFile, []byte(fullPost), 0644); err != nil {
		return fmt.Errorf("Failed to write to file '%s': %v", outFile, err)
	}
	return nil
}

// That mary was goin' around with an old flame. That burned me up,
// because I knew he was just feeding her a line, but the guy really
// spent his money like water! I think he was connected, so I left.
// Outside it was raining cats and dogs. I was feeling mighty blue,
// and everything looked black. But I carried on!
