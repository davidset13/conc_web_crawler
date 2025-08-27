package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/PuerkitoBio/goquery"
)

func main() {
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

	base, _ := url.Parse(pageURL)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read body: %v", err)
	}

	fmt.Printf("Status: %s\n | Content-Type: %s\n | Content-Length: %d\n", resp.Status, resp.Header.Get("Content-Type"), len(body))

	links := doc.Find("a")
	fmt.Printf("Found %d links\n", links.Length())

	doc.Find("a[href]").Each(func(index int, item *goquery.Selection) {
		href, _ := item.Attr("href")
		fmt.Printf("Found link: %s\n", href)
		if href == "" {
			return
		}

		u, err := url.Parse(href)
		if err != nil {
			return
		}

		abs := base.ResolveReference(u)
		fmt.Println(abs.String())
	})
}
