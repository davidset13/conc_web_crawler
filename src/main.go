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

	crawler := NewCrawlerQueue(10, 1024, 10, main_wg, JSONWriter, writer_wg)

	crawler.Run(seedURLs)

	main_wg.Wait()

	writer_wg.Wait()
}
