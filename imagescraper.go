package main

import (
	"math/rand"
	"net/http"
	"strings"

	"github.com/gocolly/colly"
)

const (
	USER_AGENT = "Mozilla/5.0 (Macintosh; Intel Mac OS X x.y; rv:42.0) Gecko/20100101 Firefox/42.0"
)

var (
	IMAGE_EXTENSIONS = []string{".jpg", ".jpeg", ".png", ".gif"}
)

func validateURL(url string) bool {
	// validate that image source is allowed
	validExtension := false
	for _, ext := range IMAGE_EXTENSIONS {
		if strings.HasSuffix(url, ext) {
			validExtension = true
			break
		}
	}

	if !validExtension {
		return false
	}

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

	// validate that the image is not too small
	if resp.ContentLength < 1000 {
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

	// some sites have different search query formats
	pexelsQuery := strings.Replace(searchString, "-", "%20", -1)
	stocSnapQuery := strings.Replace(searchString, "-", "+", -1)

	c.Visit("https://unsplash.com/s/" + searchString)
	c.Visit("https://burst.shopify.com/photos/search?utf8=%E2%9C%93&q=" + searchString + "&button=")
	c.Visit("https://www.pexels.com/search/" + pexelsQuery + "/")
	c.Visit("https://www.flickr.com/search/?text=" + pexelsQuery)
	c.Visit("http://www.google.com/images?q=" + stocSnapQuery)
	c.Visit("https://stocksnap.io/search/" + stocSnapQuery)

	return scrapedImages
}

func getImageUrl(query string) string {
	imgs := doImageSearch(query)

	// TODO: maybe ask GPT to select the best one ?
	indx := rand.Intn(len(imgs))
	return imgs[indx]
}
