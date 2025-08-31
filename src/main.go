package main

import (
	"log"
	"sync"
)

func main() {
	collective_wg := &sync.WaitGroup{}
	JSONWriter, err := CreateJSONWriter("data.json.gz", collective_wg)
	if err != nil {
		log.Fatalf("Error creating JSON writer: %v", err)
	}

	crawler := NewCrawlerQueue(10, 1024, 10, collective_wg, JSONWriter)

	crawler.Run(seedURLs)

	collective_wg.Wait()
}
