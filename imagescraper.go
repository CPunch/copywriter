package main

import (
	"io"
	"math/rand"
	"net/http"
	"strings"

	"github.com/gocolly/colly"
)

const (
	USER_AGENT = "Mozilla/5.0 (iPhone; CPU iPhone OS 16_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.5 Mobile/15E148 Safari/604.1"
)

var (
	IMAGE_EXTENSIONS = []string{".jpg", ".jpeg", ".png", ".gif"}
)

func validateURL(url string) bool {
	// // validate that image source is allowed
	// validExtension := false
	// for _, ext := range IMAGE_EXTENSIONS {
	// 	if strings.HasSuffix(url, ext) {
	// 		validExtension = true
	// 		break
	// 	}
	// }

	// // embedded images are cool too, but we need to make sure they're not a logo or something lol
	// if strings.Contains(url, "data:image") && len(url) > 100 {
	// 	return true
	// }

	// if !validExtension {
	// 	return false
	// }

	// make request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false
	}

	// set useragent
	req.Header.Set("User-Agent", USER_AGENT)

	// make the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	// check status code
	if resp.StatusCode != 200 {
		return false
	}

	// response body too big or too small? helps keep logos and stuff out too
	if resp.ContentLength > 10000000 || resp.ContentLength < 5000 {
		return false
	}

	// Read the response body into a byte slice
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		Fail("Failed to read response body: %v", err)
	}

	// check the mime type of the response body
	if !strings.Contains(http.DetectContentType(body), "image") {
		return false
	}

	return true
}

// we scrape various image sites for images based on the search query
func doImageSearch(searchQuery string) []string {
	scrapedImages := []string{}

	// make our search query url friendly
	searchString := strings.Replace(searchQuery, " ", "-", -1)
	searchString = strings.Replace(searchString, "\"", "", -1)

	c := colly.NewCollector()
	c.UserAgent = USER_AGENT
	c.AllowURLRevisit = true
	c.DisableCookies()

	// scrape all images from a page
	c.OnHTML("img[src]", func(e *colly.HTMLElement) {
		src := e.Attr("src")
		if src != "" && validateURL(src) {
			// add the image to our list of scraped images
			scrapedImages = append(scrapedImages, src)
		}
	})

	c.OnHTML("img[data-src]", func(e *colly.HTMLElement) {
		src := e.Attr("data-src")
		if src != "" && validateURL(src) {
			// add the image to our list of scraped images
			scrapedImages = append(scrapedImages, src)
		}
	})

	c.OnResponse(func(r *colly.Response) {
		Info("Visited %s", r.Request.URL.String())
	})

	// some sites have different search query formats
	// pexelsQuery := strings.Replace(searchString, "-", "%20", -1)
	stocSnapQuery := strings.Replace(searchString, "-", "+", -1)

	// c.Visit("https://unsplash.com/s/" + searchString)
	// c.Visit("https://burst.shopify.com/photos/search?utf8=%E2%9C%93&q=" + searchString + "&button=")
	// c.Visit("https://www.pexels.com/search/" + pexelsQuery + "/")
	// c.Visit("https://www.flickr.com/search/?text=" + pexelsQuery)
	c.Visit("https://www.google.com/images?q=" + stocSnapQuery)
	c.Visit("https://stocksnap.io/search/" + stocSnapQuery)

	return scrapedImages
}

func getImageUrl(query string) string {
	imgs := doImageSearch(query)
	Info("Found %d images", len(imgs))

	// TODO: maybe ask GPT to select the best one ?
	indx := rand.Intn(len(imgs))
	return imgs[indx]
}
