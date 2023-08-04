package main

import (
	"context"
	"fmt"

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
