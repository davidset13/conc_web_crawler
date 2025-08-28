package main

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"os"
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

func CreateJSONWriter(path string) (*JSONChannels, error) {
	channels := &JSONChannels{
		ch:   make(chan Record),
		errc: make(chan error),
	}

	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}

	gz := gzip.NewWriter(f)
	bufw := bufio.NewWriter(gz)
	enc := json.NewEncoder(bufw)

	go func() {
		defer func() {
			_ = bufw.Flush()
			_ = gz.Close()
			_ = f.Close()
			close(channels.errc)
		}()

		for r := range channels.ch {
			if err := enc.Encode(&r); err != nil {
				channels.errc <- err
				return
			}
		}

		channels.errc <- nil
	}()

	return channels, nil
}
