package parser

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gocolly/colly"
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
	c := colly.NewCollector()

	var anime Anime

	c.OnHTML("html", func(e *colly.HTMLElement) {
		anime.Name = e.ChildText("h2.anizm_pageTitle > a")
		anime.Link = result.Link
		anime.Cover = "https://anizm.net" + e.ChildAttr("div.infoPosterImg > img", "src")
		e.ForEach("span.dataValue > span.tag > span.label", func(i int, h *colly.HTMLElement) {
			anime.Genre = append(anime.Genre, h.Text)
		})
		anime.Studio = e.ChildText("div.anizm_boxContent > span.dataTitle:contains(Stüdyo) + span")
		anime.Description = e.ChildText("div.anizm_boxContent > div.infoDesc")
		e.ForEach("div.episodeListTabContent div a", func(i int, h *colly.HTMLElement) {
			anime.Episodes = append(anime.Episodes, Episode{
				Name: strings.TrimSpace(h.Text),
				Link: h.Attr("href"),
			})
		})
	})


	err := c.Visit(result.Link)

	if err != nil {
		return Anime{}, fmt.Errorf("Cannot get details of anime: %w", err)
	}

	return anime, nil
}
