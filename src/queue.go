package main

import (
	"container/list"
	"sync"
)

type CrawlerQueue struct {
	mu      sync.Mutex
	queue   *list.List
	visited map[string]struct{}
}

func NewCrawlerQueue() *CrawlerQueue {
	return &CrawlerQueue{queue: list.New()}
}

func (q *CrawlerQueue) Enqueue(element string) {
	q.mu.Lock()
	defer q.mu.Unlock()
	_, ok := q.visited[element]
	if !ok {
		q.visited[element] = struct{}{}
		q.queue.PushBack(element)
	}
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
