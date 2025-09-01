package main

import (
	"log"
	"sync"
)

func main() {
	main_wg := &sync.WaitGroup{}
	writer_wg := &sync.WaitGroup{}
	JSONWriter, err := CreateJSONWriter("data.jsonl.gz", writer_wg)
	if err != nil {
		log.Fatalf("Error creating JSON writer: %v", err)
	}

	crawler := NewCrawlerQueue(100, 100000, 10000, main_wg, JSONWriter)

	crawler.Run(seedURLs)

	main_wg.Wait()

	writer_wg.Wait()
}
