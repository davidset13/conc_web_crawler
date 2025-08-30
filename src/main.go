package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/sync/semaphore"
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

	queue := NewCrawlerQueue()

	client := &http.Client{}

	URL_channel := make(chan string, 1024)

	sem := semaphore.NewWeighted(10)

	notify := make(chan struct{}, 1)
	select {
	case notify <- struct{}{}:
	default:
	}

	for len(queue.visited) < 100 {
		<-notify
		for i := 0; i < queue.queue.Len(); i++ {
			url := queue.Dequeue()
			URL_channel <- url
		}
	}

	wg.Add(1)
	go func() {
		for url_ := range URL_channel {
			u := url_
			wg.Add(1)
			go func() {
				err := sem.Acquire(context.Background(), 1)
				if err != nil {
					return
				}
				defer sem.Release(1)
				defer wg.Done()

				req, err := http.NewRequest("GET", u, nil)
				if err != nil {
					return
				}

				req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) "+
					"AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0 Safari/537.36")

				req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

				req.Header.Set("Accept-Language", "en-US,en;q=0.9")

				resp, err := client.Do(req)
				if err != nil {
					return
				}

				defer resp.Body.Close()

				fmt.Println("Status:", resp.Status)

				doc, err := goquery.NewDocumentFromReader(resp.Body)
				if err != nil {
					return
				}

				txt := doc.Find("body").Text()

				writer.ch <- Record{URL: u, Text: txt, FetchedAt: time.Now().Format(time.RFC3339)}

				base, _ := url.Parse(u)

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
					queue.Enqueue(abs.String())
					select {
					case notify <- struct{}{}:
					default:
					}
				})
			}()
		}
		wg.Done()
	}()

	for _, seedURL := range seedUrls {
		queue.Enqueue(seedURL)
	}

	close(writer.ch)
	wg.Wait()
}
