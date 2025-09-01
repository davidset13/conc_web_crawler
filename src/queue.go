package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
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
	maxVisits   int32
	JSONWriter  *JSONChannels
	concurrency int
	processed   atomic.Int32
	ctx         context.Context
	cancel      context.CancelFunc
}

func NewCrawlerQueue(concurrency int, frontierCap int, maxVisits int, wg *sync.WaitGroup, JSONWriter *JSONChannels) *CrawlerQueue {
	ctx, cancel := context.WithCancel(context.Background())
	return &CrawlerQueue{
		client:      &http.Client{},
		sem:         semaphore.NewWeighted(int64(concurrency)),
		workCh:      make(chan string, frontierCap),
		wg:          wg,
		visited:     make(map[string]struct{}),
		maxVisits:   int32(maxVisits),
		JSONWriter:  JSONWriter,
		concurrency: concurrency,
		processed:   atomic.Int32{},
		ctx:         ctx,
		cancel:      cancel,
	}
}

func (q *CrawlerQueue) Enqueue(element string) {
	select {
	case <-q.ctx.Done():
		return
	default:
	}

	q.mu.Lock()
	if _, ok := q.visited[element]; ok {
		q.mu.Unlock()
		return
	}
	n := int32(len(q.visited))

	if n >= q.maxVisits {
		q.mu.Unlock()
		return
	}

	q.visited[element] = struct{}{}
	q.mu.Unlock()

	q.wg.Add(1)
	select {
	case q.workCh <- element:
	case <-q.ctx.Done():
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

			req = req.WithContext(q.ctx)
			req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) "+
				"AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0 Safari/537.36")

			req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

			req.Header.Set("Accept-Language", "en-US,en;q=0.9")

			resp, err := q.client.Do(req)
			if err != nil {
				return
			} else if resp.StatusCode >= 400 {
				return
			}
			defer resp.Body.Close()

			fmt.Println("Status:", resp.Status)

			doc, err := goquery.NewDocumentFromReader(resp.Body)
			if err != nil {
				return
			}

			txt := doc.Find("body").Text()

			select {
			case q.JSONWriter.ch <- Record{URL: u, Text: txt, FetchedAt: time.Now().Format(time.RFC3339)}:
			case <-q.ctx.Done():
				return
			}

			if q.ctx.Err() != nil {
				return
			}

			base, _ := url.Parse(u)
			doc.Find("a[href]").Each(func(index int, item *goquery.Selection) {
				if index >= 100 {
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

			m := q.processed.Add(1)
			if m > q.maxVisits {
				q.cancel()
				return
			}

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
		close(q.workCh)
		q.closeOnce.Do(func() {
			close(q.JSONWriter.ch)
			fmt.Println("Writer wait done")
		})
	}()
}
