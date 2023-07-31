package main

import (
	"context"

	"github.com/groovili/gogtrends"
)

func getPopularTrends() []string {
	stories, err := gogtrends.Realtime(context.Background(), "en-US", "US", "all")
	if err != nil {
		Fail("Failed to scrape google trends: %s", err.Error())
		return nil
	}

	var trends []string
	for _, trend := range stories {
		trends = append(trends, trend.Title)
	}

	if len(trends) > 5 {
		trends = trends[:5]
	}

	return trends
}
