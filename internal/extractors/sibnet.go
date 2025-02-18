package extractors

import (
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func Sibnet(client *http.Client, link string, fansubName string) []Video {
	var videos []Video

	resp, err := http.Get(link)
	if err != nil {
		log.Printf("[Extractor/Sibnet] Cannot perform request %v, skipping...\n", link)
		return nil
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Printf("[Extractor/Sibnet] Cannot read/parse body from %v, skipping...\n", link)
		return nil
	}

	scriptData := doc.Find("script").FilterFunction(func(i int, s *goquery.Selection) bool {
		return strings.Contains(s.Text(), "player.src")
	}).Text()
	slug := scriptData[strings.Index(scriptData, "player.src"):]
	slug = slug[strings.IndexRune(slug, '"')+1:]
	slug = slug[:strings.IndexRune(slug, '"')]

	var videoLink string
	if strings.Contains(slug, "http") {
		videoLink = slug
	} else {
		u, err := url.Parse(link)
		if err != nil {
			log.Printf("[Extractor/Sibnet] Cannot parse url: %v, skipping...\n", link)
			return nil
		}

		videoLink = "https://" + u.Host + slug
	}

	video := Video{
		Link:    videoLink,
		Name:    "[" + fansubName + "] Sibnet",
		Headers: make(http.Header),
	}

	video.Headers.Add("Referer", link)

	videos = append(videos, video)

	return videos
}
