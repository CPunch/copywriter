package main

import (
	"fmt"
	"testing"
)

func TestGetPopularTrends(t *testing.T) {
	fmt.Println(getPopularTrends("all"))
}

// func TestImageMeta(t *testing.T) {
// 	prompt := `In an unexpected press conference that rocketed around the world, President Joe Biden did the unthinkable — he confirmed the existence of aliens. Without a hint of subtext or metaphoric implication, President Biden simply laid out the fact: extraterrestrial beings are real, and they've been in contact with Earth.`
// 	query := generateResponse(ResponseOptions{
// 		MaxTokens: 25,
// 		Prompt:    fmt.Sprintf("%s\n---\nsearch query to for a relevant image that matches the tone for the above text: ", prompt),
// 		UseGPT4:   false,
// 		Clean:     true,
// 	})

// 	fmt.Println(query)
// }
