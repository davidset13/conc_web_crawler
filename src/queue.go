package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/sync/semaphore"
)

type CrawlerQueue struct {
	client      *http.Client
	sem         *semaphore.Weighted
	workCh      chan string
	wg          *sync.WaitGroup
	mu          sync.Mutex
	visited     map[string]struct{}
	closeOnce   sync.Once
	maxVisits   int
	JSONWriter  *JSONChannels
	concurrency int
}

func NewCrawlerQueue(concurrency int, frontierCap int, maxVisits int, wg *sync.WaitGroup, JSONWriter *JSONChannels) *CrawlerQueue {
	return &CrawlerQueue{
		client:      &http.Client{},
		sem:         semaphore.NewWeighted(int64(concurrency)),
		workCh:      make(chan string, frontierCap),
		wg:          wg,
		visited:     make(map[string]struct{}),
		maxVisits:   maxVisits,
		JSONWriter:  JSONWriter,
		concurrency: concurrency,
	}
}

func (q *CrawlerQueue) Enqueue(element string) {
	q.mu.Lock()
	if _, ok := q.visited[element]; ok {
		q.mu.Unlock()
		return
	}
	q.visited[element] = struct{}{}
	n := len(q.visited)
	q.mu.Unlock()

	if n >= q.maxVisits {
		q.closeOnce.Do(func() {
			close(q.workCh)
		})
		return
	}

	q.wg.Add(1)
	select {
	case q.workCh <- element:

	default:
		q.wg.Done()
		return
	}

}

func (q *CrawlerQueue) Work() {
	for u := range q.workCh {
		func(u string) {
			defer q.wg.Done()

			if err := q.sem.Acquire(context.Background(), 1); err != nil {
				return
			}
			defer q.sem.Release(1)

			req, err := http.NewRequest("GET", u, nil)
			if err != nil {
				return
			}

			req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) "+
				"AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0 Safari/537.36")

			req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

			req.Header.Set("Accept-Language", "en-US,en;q=0.9")

			resp, err := q.client.Do(req)
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

			q.JSONWriter.ch <- Record{URL: u, Text: txt, FetchedAt: time.Now().Format(time.RFC3339)}

			base, _ := url.Parse(u)
			doc.Find("a[href]").Each(func(index int, item *goquery.Selection) {
				if index >= 5 {
					return
				}

				href, _ := item.Attr("href")

				if href == "" {
					return
				}

				rel, err := url.Parse(href)
				if err != nil {
					return
				}

				abs := base.ResolveReference(rel).String()

				q.Enqueue(abs)
			})
		}(u)
	}
}

func (q *CrawlerQueue) Run(seedURLs []string) {

	for i := 0; i < q.concurrency; i++ {
		go q.Work()
	}

	for _, seedURL := range seedURLs {
		q.Enqueue(seedURL)
	}

	go func() {
		q.wg.Wait()
		q.closeOnce.Do(func() {
			close(q.workCh)
			close(q.JSONWriter.ch)
		})
	}()
}
