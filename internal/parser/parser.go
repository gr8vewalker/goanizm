package parser

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

type Result struct {
	Name  string
	Link  string
	Cover string
}

type search_entry struct {
	Title         string `json:"info_title"`
	Slug          string `json:"info_slug"`
	Poster        string `json:"info_poster"`
	OriginalTitle string `json:"info_titleoriginal"`
	EnglishTitle  string `json:"info_titleenglish"`
	OtherNames    string `json:"info_othernames"`
}

func Search(query string) ([]Result, error) {
	resp, err := http.Get("https://anizm.net/getAnimeListForSearch")
	if err != nil {
		return nil, fmt.Errorf("Couldn't make a search: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Couldn't make a search: %w", err)
	}

	var entries []search_entry
	if err := json.Unmarshal(body, &entries); err != nil {
		return nil, fmt.Errorf("Couldn't make a search: %w", err)
	}

	var results []Result

	query = strings.ToLower(query)
	for _, entry := range entries {
		if strings.Contains(strings.ToLower(entry.OriginalTitle), query) ||
			strings.Contains(strings.ToLower(entry.EnglishTitle), query) ||
			strings.Contains(strings.ToLower(entry.OtherNames), query) {

			results = append(results, Result{
				Name:  entry.Title,
				Link:  "https://anizm.net/" + entry.Slug,
				Cover: "https://anizm.net/storage/pcovers/" + entry.Poster,
			})
		}
	}

	return results, nil
}

type Anime struct {
	Name        string
	Link        string
	Cover       string
	Genre       []string
	Studio      string
	Description string
	Episodes    []Episode
}

type Episode struct {
	Name string
	Link string
}

func Details(result Result) (Anime, error) {
	resp, err := http.Get(result.Link)
	if err != nil {
		return Anime{}, fmt.Errorf("Couldn't get details: %w", err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return Anime{}, fmt.Errorf("Couldn't get details: %w", err)
	}

	anime := Anime{
		Name:        doc.Find("h2.anizm_pageTitle > a").First().Text(),
		Link:        result.Link,
		Cover:       "https://anizm.net" + doc.Find("div.infoPosterImg > img").First().AttrOr("src", ""),
		Studio:      doc.Find("div.anizm_boxContent > span.dataTitle:contains(Stüdyo) + span").First().Text(),
		Description: doc.Find("div.anizm_boxContent > div.infoDesc").First().Text(),
	}

	doc.Find("span.dataValue > span.tag > span.label").Each(func(i int, s *goquery.Selection) {
		anime.Genre = append(anime.Genre, s.Text())
	})

	doc.Find("div.episodeListTabContent div a").Each(func(i int, s *goquery.Selection) {
		anime.Episodes = append(anime.Episodes, Episode{
			Name: strings.TrimSpace(s.Text()),
			Link: s.AttrOr("href", ""),
		})
	})

	return anime, nil
}

type Video struct {
}

func Videos(episode Episode) ([]Video, error) {
	resp, err := http.Get(episode.Link)
	if err != nil {
		return nil, fmt.Errorf("Couldn't get episode: %w", err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Couldn't parse episode: %w", err)
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	videos := []Video{}
	client := &http.Client{}
	noRedirectClient := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	parseFansub := func(s *goquery.Selection) {
		defer wg.Done()
		fansubName := strings.TrimSpace(s.Find(".title").First().Text())
		fansubLink := s.AttrOr("translator", "")

		req, err := http.NewRequest("GET", fansubLink, nil)
		if err != nil {
			log.Printf("[Ep: %v, Fansub: %v] Cannot create request for %v, skipping...\n", episode.Name, fansubName, fansubLink)
			return
		}
		req.Header.Add("Origin", "https://anizm.net")
		req.Header.Add("Referer", "https://anizm.net/")

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("[Ep: %v, Fansub: %v] Cannot perform request %v, skipping...\n", episode.Name, fansubName, fansubLink)
			return
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("[Ep: %v, Fansub: %v] Cannot read body from %v, skipping...\n", episode.Name, fansubName, fansubLink)
			return
		}

		var unmarshaled map[string]string
		if err := json.Unmarshal(body, &unmarshaled); err != nil {
			log.Printf("[Ep: %v, Fansub: %v] Cannot parse json from %v, skipping...\n", episode.Name, fansubName, fansubLink)
			return
		}

		fansubData := unmarshaled["data"]
		fansubDoc, err := goquery.NewDocumentFromReader(strings.NewReader(fansubData))
		if err != nil {
			log.Printf("[Ep: %v, Fansub: %v] Cannot parse fansub html data from %v, skipping...\n", episode.Name, fansubName, fansubLink)
			return
		}

		fansubDoc.Find(".videoPlayerButtons").Each(func(i int, s *goquery.Selection) {
			wg.Add(1)
			go func() {
				defer wg.Done()
				parsed := parseVideos(noRedirectClient, fansubName, strings.ReplaceAll(s.AttrOr("video", ""), "/video/", "/player/"))
				mu.Lock()
				videos = append(videos, parsed...)
				mu.Unlock()
			}()
		})
	}

	fansubs := doc.Find("div#fansec > a").EachIter()
	for _, fansub := range fansubs {
		wg.Add(1)
		go parseFansub(fansub)
	}

	// using go routines makes this %33 faster
	// it depends on internet connection etc. too but go routines is always faster.

	wg.Wait()

	return videos, nil
}

func parseVideos(noRedirectClient *http.Client, fansubName string, link string) []Video {
	req, err := http.NewRequest("GET", link, nil)
	if err != nil {
		log.Printf("[Parsing/Fansub: %v] Cannot create request for %v, skipping...\n", fansubName, link)
		return nil
	}
	req.Header.Add("Origin", "https://anizm.net")
	req.Header.Add("Referer", "https://anizm.net/")

	resp, err := noRedirectClient.Do(req)
	if err != nil {
		log.Printf("[Parsing/Fansub: %v] Cannot perform request %v, skipping...\n", fansubName, link)
		return nil
	}
	defer resp.Body.Close()

	playerLink := resp.Header["Location"][0]

	var videos []Video

	switch {
	case strings.Contains(playerLink, "anizmplayer.com"):
		// TODO: extractor.
	}

	return videos
}
