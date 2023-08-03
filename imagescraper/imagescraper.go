package imagescraper

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

/*
 In the future, I want to scrape the images and dump them to a file. currently we just return the url,
 but we really should be downloading the image and saving, that way in case google decides to remove the
 image, we still have it.
*/

func validateURL(url string) bool {
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
		return false
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

func GetImageUrl(query string) string {
	imgs := doImageSearch(query)

	// TODO: replace this :(
	if len(imgs) == 0 {
		return "https://www.rd.com/wp-content/uploads/2020/11/GettyImages-889552354-e1606774439626.jpg"
	}

	// TODO: maybe ask GPT to select the best one ?
	indx := rand.Intn(len(imgs))
	return imgs[indx]
}
