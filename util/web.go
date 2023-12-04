package util

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gocolly/colly"
)

const (
	USER_AGENT = "Mozilla/5.0 (iPhone; CPU iPhone OS 16_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.5 Mobile/15E148 Safari/604.1"
)

// convert url to markdown
func ScrapeArticle(url string) (string, error) {
	Info("Scraping article '%s'...", url)

	md := ""
	c := colly.NewCollector()
	c.UserAgent = USER_AGENT
	c.AllowURLRevisit = true
	c.DisableCookies()

	// scrape all paragraphs from a page
	c.OnHTML("p", func(e *colly.HTMLElement) {
		md += "\n" + e.Text
	})

	c.Visit(url)
	return md, nil
}

type DownloadOptions struct {
	URL      string
	FilePath string
	Header   http.Header
}

func DownloadToFile(args DownloadOptions) error {
	Info("Downloading %s to '%s'...", args.URL, args.FilePath)

	req, err := http.NewRequest("GET", args.URL, nil)
	if err != nil {
		return err
	}

	// add headers
	req.Header = args.Header

	// make the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// check status code
	if resp.StatusCode != 200 {
		return fmt.Errorf("Bad status code: %d", resp.StatusCode)
	}

	// write response body to file
	f, err := os.Create(args.FilePath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
