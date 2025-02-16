package parser

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

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
