package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func main() {

	var wg *sync.WaitGroup = &sync.WaitGroup{}

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current working directory: %v", err)
	}
	path := filepath.Join(cwd, "web_scrape.json.gz")
	writer, err := CreateJSONWriter(path, wg)
	if err != nil {
		log.Fatalf("Failed to create JSON writer: %v", err)
	}

	fmt.Println("JSON Writer Created", writer)

	pageURL := seedUrls[0]

	client := &http.Client{}

	req, err := http.NewRequest("GET", pageURL, nil)
	if err != nil {
		log.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) "+
		"AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0 Safari/537.36")

	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Failed to get page: %v", err)
	}

	defer resp.Body.Close()

	fmt.Println("Status:", resp.Status)

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatalf("Failed to parse page: %v", err)
	}

	txt := doc.Find("body").Text()
	writer.ch <- Record{URL: pageURL, Text: txt, FetchedAt: time.Now().Format(time.RFC3339)}

	base, _ := url.Parse(pageURL)

	links := make([]string, 5)
	doc.Find("a[href]").Each(func(index int, item *goquery.Selection) {
		if index >= 5 {
			return
		}

		href, _ := item.Attr("href")
		if href == "" {
			return
		}

		u, err := url.Parse(href)
		if err != nil {
			return
		}

		abs := base.ResolveReference(u)
		links[index] = abs.String()
	})

	writer.ch <- Record{URL: pageURL, Text: txt, FetchedAt: time.Now().Format(time.RFC3339)}
	close(writer.ch)
	wg.Wait()
	fmt.Println(links)
}
