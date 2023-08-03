package main

import (
	"fmt"
	"testing"
)

// func TestGetPopularTrends(t *testing.T) {
// 	fmt.Println(getPopularTrends("all"))
// }

func TestImageMeta(t *testing.T) {
	prompt := `In an unexpected press conference that rocketed around the world, President Joe Biden did the unthinkable — he confirmed the existence of aliens. Without a hint of subtext or metaphoric implication, President Biden simply laid out the fact: extraterrestrial beings are real, and they've been in contact with Earth.`
	query, err := generateResponse(ResponseOptions{
		MaxTokens: 30,
		Prompt:    fmt.Sprintf("%s\n---\nWrite a short one sentence image prompt that fits the above text: ", prompt),
		UseGPT4:   true,
		Clean:     true,
	})
	if err != nil {
		t.Error(err)
	}

	fmt.Println(query)
}
