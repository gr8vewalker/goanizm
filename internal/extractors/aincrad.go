package extractors

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/grafov/m3u8"
)

func Aincrad(client *http.Client, link string) []Video {
	hash := strings.Split(strings.Split(link, "video/")[1], "/")[0]
	form := strings.NewReader(url.Values{
		"hash": {hash},
		"r":    {"https://anizm.net"},
	}.Encode())
	req, err := http.NewRequest("POST", "https://anizmplayer.com/player/index.php?data="+hash+"&do=getVideo", form)
	if err != nil {
		log.Printf("[Extractor/Aincrad] Cannot create request for %v, skipping...\n", req.URL)
		return nil
	}

	req.Header.Add("Origin", "https://anizmplayer.com")
	req.Header.Add("Referer", link)
	req.Header.Add("X-Requested-With", "XMLHttpRequest")

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[Extractor/Aincrad] Cannot perform request %v, skipping...\n", req.URL)
		return nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[Extractor/Aincrad] Cannot read body from %v, skipping...\n", req.URL)
		return nil
	}

	var data map[string]json.RawMessage
	if err := json.Unmarshal(body, &data); err != nil {
		log.Printf("[Extractor/Aincrad] Cannot parse body from %v, skipping...\n", req.URL)
		return nil
	}

	
	var securedLink string
	if err := json.Unmarshal(data["securedLink"], &securedLink); err != nil {
		log.Printf("[Extractor/Aincrad] Cannot parse body from %v, skipping...\n", req.URL)
		return nil
	}

	return extractPlaylist(client, securedLink, link)
}

func extractPlaylist(client *http.Client, link string, referer string) []Video {
	req, err := http.NewRequest("POST", link, nil)
	if err != nil {
		log.Printf("[Extractor/Aincrad] Cannot create request for %v, skipping...\n", req.URL)
		return nil
	}

	req.Header.Add("Origin", "https://anizmplayer.com")
	req.Header.Add("Referer", referer)

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[Extractor/Aincrad] Cannot perform request %v, skipping...\n", req.URL)
		return nil
	}
	defer resp.Body.Close()
	p, listType, err := m3u8.DecodeFrom(resp.Body, true)
	if err != nil || listType != m3u8.MASTER {
		log.Printf("[Extractor/Aincrad] Couldn't parse m3u8 playlist from %v, skipping...\n", req.URL)
		return nil
	}
	masterPlaylist := p.(*m3u8.MasterPlaylist)

	var videos []Video

	for _, variant := range masterPlaylist.Variants {
		if variant != nil {
			video := Video{
				Link: variant.URI,
				Name: "Aincrad - " + variant.Resolution,
				Headers: make(http.Header),
			}
			for _, alternative := range variant.Alternatives {
				if alternative.GroupId == variant.Audio {
					video.Audio = alternative.URI
				}
			}
			video.Headers["Origin"] = []string{"https://anizm.net"}
			video.Headers["Referer"] = []string{link}
			videos = append(videos, video)
		}
	}

	return videos
}
