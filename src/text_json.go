package main

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"fmt"
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
		ch:   make(chan Record, 50000),
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

		var finalError error

		for r := range channels.ch {
			if err := enc.Encode(&r); err != nil {
				finalError = err
				fmt.Println("Error encoding record", err)
				break
			}
		}

		if err := bufw.Flush(); err != nil && finalError == nil {
			finalError = err
		}
		if err := gz.Close(); err != nil && finalError == nil {
			finalError = err
		}
		if err := f.Close(); err != nil && finalError == nil {
			finalError = err
		}

		select {
		case channels.errc <- finalError:
		default:
		}

		close(channels.errc)
	}()

	return channels, nil
}
