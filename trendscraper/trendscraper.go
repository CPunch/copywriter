package trendscraper

import (
	"context"
	"fmt"
	"math/rand"
	"strings"

	"git.openpunk.com/CPunch/copywriter/util"
	"github.com/groovili/gogtrends"
)

func ScrapePopularTrends(category string) (title, article string, _ error) {
	util.Info("Scraping google trends in category '%s'...", category)
	stories, err := gogtrends.Realtime(context.Background(), "en-US", "US", category)
	if err != nil {
		return "", "", err
	}

	var trends []string
	for _, trend := range stories {
		trends = append(trends, trend.Articles[0].Title+" - "+trend.Articles[0].Snippet)
	}

	if len(trends) > 10 {
		trends = trends[:10]
	}

	resp, err := util.GenerateResponse(util.ResponseOptions{
		MaxTokens: 100,
		Prompt:    fmt.Sprintf("%s\n---\nWrite some keywords for the above articles: ", strings.Join(trends, "\n")),
		UseGPT4:   false,
		Clean:     false,
	})
	if err != nil {
		return "", "", err
	}

	return fmt.Sprintf("The following is a list of topics that readers might be interested in:\n%s", resp), "", nil
}

func ScrapeRealtimeNews(category string) (title, article string, _ error) {
	util.Info("Scraping stories in category '%s'...", category)
	stories, err := gogtrends.Realtime(context.Background(), "en-US", "US", category)
	if err != nil {
		// Fail("Failed to scrape google trends: %s", err.Error())
		return "", "", err
	}

	story := stories[rand.Intn(len(stories))]
	articles := story.Articles
	if len(articles) > 3 {
		articles = articles[:3]
	}

	title = "The following is a list of articles related to the topic:\n"

	var context string
	for _, article := range articles {
		content, err := util.ScrapeArticle(article.URL)
		if err != nil { // just skip the article
			util.Warning("Failed to scrape %s: %s", article.URL, err.Error())
			continue
		}
		context += fmt.Sprintf("# %s\n%s\n\n", article.Title, content)
		title += fmt.Sprintf("%s\n", article.Title)
		// ctx.Keywords = append(ctx.Keywords, article.Title)
	}

	util.Info("Summarizing context...")
	resp, err := util.SummarizeText(context)
	if err != nil {
		return "", "", err
	}

	article = fmt.Sprintf("An article summary related to the article is given below:\n%s\nRelated Keywords: %s\n", resp, story.Title)
	return
}
