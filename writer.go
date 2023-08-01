package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

const (
	MAX_RETRY = 5
)

const (
	GEN_TITLE_MODE_PREVIOUS_TITLES = iota
	GEN_TITLE_MODE_TRENDS
)

var (
	GEN_TITLE_MODE = GEN_TITLE_MODE_TRENDS
)

func getListOfPreviousPosts() string {
	// get max last 10 posts
	posts := getPosts(10)
	var titles []string
	for _, post := range posts {
		titles = append(titles, post.Title)
	}

	return strings.Join(titles, "\n")
}

func genBlogTitleBasedOnPreviousTitles() string {
	for i := 0; i < MAX_RETRY; i++ { // just in case gpt is a DUMBASS; i don't wanna burn a million dollars
		titles := generateResponse(ResponseOptions{
			MaxTokens: 25,
			Prompt:    fmt.Sprintf("The following is a list of titles of blog posts:\n%s", getListOfPreviousPosts()),
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

	Fail("Failed to create title!")
	return ""
}

func genBlogTitleBasedOnTrends() string {
	trends := getPopularTrends()

	for i := 0; i < MAX_RETRY; i++ { // just in case gpt is a DUMBASS; i don't wanna burn a million dollars
		titles := generateResponse(ResponseOptions{
			MaxTokens: 25,
			Prompt:    fmt.Sprintf("The following is a list of topics:\n%s\n\nWhat is an example of a title of a blog post that matches these trends: ", strings.Join(trends, "\n")),
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

	Fail("Failed to create title!")
	return ""
}

func GenBlogTitle() string {
	switch GEN_TITLE_MODE {
	case GEN_TITLE_MODE_PREVIOUS_TITLES:
		Info("Generating title based on previous titles...")
		return genBlogTitleBasedOnPreviousTitles()
	case GEN_TITLE_MODE_TRENDS:
		Info("Generating title based on trends...")
		return genBlogTitleBasedOnTrends()
	}

	Fail("Invalid GEN_TITLE_MODE: %v", GEN_TITLE_MODE)
	return ""
}

func genBlogTags(content string) string {
	for i := 0; i < MAX_RETRY; i++ { // just in case gpt is a DUMBASS; i don't wanna burn a million dollars
		tagString := generateResponse(ResponseOptions{
			MaxTokens: 50,
			Prompt:    fmt.Sprintf("%s\n\nTags as a json array with only 1 word each, max 5:", content),
			UseGPT4:   false,
		})

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

func genImageAboutMeta(prompt string) string {
	query := generateResponse(ResponseOptions{
		MaxTokens: 20,
		Prompt:    fmt.Sprintf("%s\n---\n search query to find a relevant image for the above text: ", prompt),
		UseGPT4:   false,
	})

	return getImageUrl(query)
}

func populateImages(content string) string {
	Info("Scraping for images...")
	lines := strings.Split(content, "\n")

	// strip lines
	for i := range lines {
		lines[i] = strings.TrimSpace(lines[i])
	}

	for i := 1; i < len(lines); i++ {
		if i%5 == 0 {
			// inject image
			imgURL := genImageAboutMeta(strings.Join(lines[i-2:i], "\n"))
			img := fmt.Sprintf("\n![](%s)", imgURL)

			// insert line
			lines = append(lines[:i], append([]string{img}, lines[i:]...)...)
		}
	}

	return strings.Join(lines, "\n")
}

func GenBlogPost(title string) string {
	defer func() {
		if e := recover(); e != nil {
			Fail("%s", e)
		}
	}()

	// sometimes gpt will generate a title with weird whitespace
	title = strings.TrimSpace(title)

	Info("Generating blog post contents...")
	markdown := generateResponse(ResponseOptions{
		MaxTokens: 5000,
		Prompt:    fmt.Sprintf("The following is a blog post written in markdown about %s, minimum 1000 words *DO NOT INCLUDE THE TITLE*:\n\n", title),
		UseGPT4:   true,
	})

	// inject images
	content := populateImages(markdown)
	thumbnail := genImageAboutMeta(title)
	tags := genBlogTags(markdown)
	author := "Mason Coleman"
	fullPost := fmt.Sprintf("---\ntitle: \"%s\"\nauthor: %s\ndate: %s\ntags: %s\nimage: \"%s\"\n---\n\n%s", title, author, getTimeString(), tags, thumbnail, content)

	Success("Generated post!")
	return fullPost
}
