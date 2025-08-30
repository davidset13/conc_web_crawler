package main

import (
	"net/http"
	"sync"

	"golang.org/x/sync/semaphore"
)

type CrawlerQueue struct {
	client  *http.Client
	sem     *semaphore.Weighted
	workCh  chan string
	wg      *sync.WaitGroup
	mu      sync.Mutex
	visited map[string]struct{}
}

func NewCrawlerQueue(concurrency int, frontierCap int) *CrawlerQueue {
	return &CrawlerQueue{
		client:  &http.Client{},
		sem:     semaphore.NewWeighted(int64(concurrency)),
		workCh:  make(chan string, frontierCap),
		wg:      &sync.WaitGroup{},
		visited: make(map[string]struct{}),
	}
}

func (q *CrawlerQueue) Enqueue(element string) {
	q.mu.Lock()
	if _, ok := q.visited[element]; ok {
		q.mu.Unlock()
		return
	}
	q.visited[element] = struct{}{}
	q.mu.Unlock()

	q.wg.Add(1)
	q.workCh <- element
}

func (q *CrawlerQueue) Dequeue() string {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.IsEmpty() {
		panic("queue is empty")
	}
	element := q.queue.Front()
	q.queue.Remove(element)
	return element.Value.(string)
}

func (q *CrawlerQueue) IsEmpty() bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.queue.Len() == 0
}
