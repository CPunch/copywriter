package main

import (
	"fmt"
	"testing"
)

// func TestGetPopularTrends(t *testing.T) {
// 	fmt.Println(getPopularTrends("all"))
// }

func TestSEOContext(t *testing.T) {
	title, article, err := scrapePopularTrends("m")
	if err != nil {
		t.Error(err)
	}

	fmt.Println(title)
	fmt.Println(article)
}

// func TestGenNewBlogTitle(t *testing.T) {
// 	trends := getPopularTrends("b")

// 	for i := 0; i < MAX_RETRY; i++ { // just in case gpt is a DUMBASS; i don't wanna burn a million dollars
// 		title, err := generateResponse(ResponseOptions{
// 			MaxTokens: 40,
// 			Prompt: fmt.Sprintf(
// 				"\nThe following is a list of topics:\n%s\n\nAn example of a short, simple and eye-catching title of an article that fits with some of these topics: ",
// 				strings.Join(trends, "\n"),
// 			),
// 			UseGPT4: false,
// 			Clean:   true,
// 		})
// 		if err != nil {
// 			t.Error(err)
// 		}

// 		// no titles?? try again
// 		if len(title) == 0 {
// 			continue
// 		}

// 		Success("Title: %s", title)
// 		return
// 	}
// }

// func TestImageMeta(t *testing.T) {
// 	prompt := `In an unexpected press conference that rocketed around the world, President Joe Biden did the unthinkable — he confirmed the existence of aliens. Without a hint of subtext or metaphoric implication, President Biden simply laid out the fact: extraterrestrial beings are real, and they've been in contact with Earth.`
// 	query, err := generateResponse(ResponseOptions{
// 		MaxTokens: 30,
// 		Prompt:    fmt.Sprintf("%s\n---\nWrite a short one sentence image prompt that fits the above text: ", prompt),
// 		UseGPT4:   true,
// 		Clean:     true,
// 	})
// 	if err != nil {
// 		t.Error(err)
// 	}

// 	fmt.Println(query)
// }
