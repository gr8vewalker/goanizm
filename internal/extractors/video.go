package extractors

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"sync"

	"github.com/grafov/m3u8"
)

type Video struct {
	Link    string
	Name    string
	Headers http.Header
	Audio   string
}

// Utilities for downloading video:

var client = &http.Client{}

func DownloadFile(path string, link string, headers http.Header) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	req, err := http.NewRequest("GET", link, nil)
	if err != nil {
		return err
	}

	req.Header = headers

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func DownloadPlaylist(path string, link string, headers http.Header) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	req, err := http.NewRequest("GET", link, nil)
	if err != nil {
		return err
	}

	req.Header = headers

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	p, listType, err := m3u8.DecodeFrom(resp.Body, true)
	if err != nil {
		return err
	}

	if listType != m3u8.MEDIA {
		return nil // only download media ones
	}

	mediaPlaylist := p.(*m3u8.MediaPlaylist)

	var wg sync.WaitGroup
	individualBytes := make([][]byte, len(mediaPlaylist.Segments))

	segmentReader := func(i int, segment *m3u8.MediaSegment) {
		defer wg.Done()
		if segment == nil {
			return
		}
		req, err := http.NewRequest("GET", segment.URI, nil)
		if err != nil {
			return
		}

		resp, err := client.Do(req)
		if err != nil {
			return
		}
		defer resp.Body.Close()

		read, err := io.ReadAll(resp.Body)
		if err != nil {
			return
		}
		individualBytes[i] = read
	}

	counter := 0
	for i, segment := range mediaPlaylist.Segments {
		wg.Add(1)
		counter++
		go segmentReader(i, segment)
		if counter >= 8 {
			wg.Wait()
			counter = 0
		}
	}

	for _, data := range individualBytes {
		if data != nil {
			io.Copy(file, bytes.NewReader(data))
		}
	}

	return nil
}
