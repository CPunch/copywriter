package main

import (
	"context"
	"fmt"
	"math/rand"
	"strings"

	"github.com/groovili/gogtrends"
)

func getPopularTrends(category string) []string {
	Info("Scraping google trends in category '%s'...", category)
	stories, err := gogtrends.Realtime(context.Background(), "en-US", "US", category)
	if err != nil {
		Fail("Failed to scrape google trends: %s", err.Error())
		return nil
	}

	var trends []string
	for _, trend := range stories {
		trends = append(trends, fmt.Sprintf("%s - %s", trend.Articles[0].Title, trend.Articles[0].Snippet))
	}

	if len(trends) > 5 {
		trends = trends[:5]
	}

	return trends
}

func scrapePopularTrends(category string) (title, article string, _ error) {
	Info("Scraping google trends in category '%s'...", category)
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

	return strings.Join(trends, "\n"), "", nil
}

func scrapeRealtimeNews(category string) (title, article string, _ error) {
	Info("Scraping stories in category '%s'...", category)
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

	var context string
	for _, article := range articles {
		content, err := scrapeArticle(article.URL)
		if err != nil { // just skip the article
			Warning("Failed to scrape %s: %s", article.URL, err.Error())
			continue
		}
		context += fmt.Sprintf("# %s\n%s\n\n", article.Title, content)
		title += fmt.Sprintf("%s\n", article.Title)
		// ctx.Keywords = append(ctx.Keywords, article.Title)
	}

	Info("Summarizing context...")
	resp, err := SummarizeText(context)
	if err != nil {
		return "", "", err
	}

	article = fmt.Sprintf("An article summary related to the article is given below:\n%s\nRelated Keywords: %s\n", resp, story.Title)
	return
}
