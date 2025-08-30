package main

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"os"
	"sync"
)

type Record struct {
	URL       string `json:"url"`
	Text      string `json:"text"`
	FetchedAt string `json:"fetched_at"`
}

type JSONChannels struct {
	ch   chan Record
	errc chan error
}

func CreateJSONWriter(path string, wg *sync.WaitGroup) (*JSONChannels, error) {
	channels := &JSONChannels{
		ch:   make(chan Record),
		errc: make(chan error, 1),
	}

	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}

	gz := gzip.NewWriter(f)
	bufw := bufio.NewWriter(gz)
	enc := json.NewEncoder(bufw)

	wg.Add(1)
	go func() {

		defer wg.Done()

		for r := range channels.ch {
			if err := enc.Encode(&r); err != nil {
				channels.errc <- err
				break
			}
		}

		bufw.Flush()
		gz.Close()
		f.Close()

		channels.errc <- nil
	}()

	return channels, nil
}
