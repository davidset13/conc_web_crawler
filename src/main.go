package main

import (
	"log"
	"sync"
)

func main() {
	main_wg := &sync.WaitGroup{}
	writer_wg := &sync.WaitGroup{}
	crawler := NewCrawlerQueue(1000, 1000000, 10000, main_wg)
	JSONWriter, err := CreateJSONWriter("data.jsonl.gz", writer_wg, crawler)
	if err != nil {
		log.Fatalf("Error creating JSON writer: %v", err)
	}

	crawler.Run(seedURLs, JSONWriter)

	main_wg.Wait()

	writer_wg.Wait()
}
