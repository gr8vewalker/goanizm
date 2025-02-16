package main

import (
	"log"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/gr8vewalker/goanizm/internal/cli"
	"github.com/gr8vewalker/goanizm/internal/extractors"
	"github.com/gr8vewalker/goanizm/internal/parser"
	ffmpeg_go "github.com/u2takey/ffmpeg-go"
)

type VideoDownloadInfo struct {
	Video extractors.Video
	Path  string
}

func main() {
	results := search()
	anime := selectAndDetail(results)
	selectedEpisodes := selectEpisodes(anime.Episodes)
	videosToDownload := selectVideos(selectedEpisodes)

	var wg sync.WaitGroup

	for _, video := range videosToDownload {
		wg.Add(1)
		go download(&wg, video)
	}

	wg.Wait()
}

func search() []parser.Result {
	color.Cyan("What do you want to download?")

	query := cli.ReadLine()
	results, err := parser.Search(query)

	if err != nil {
		log.Fatalln("Cannot done a search", err)
	}

	return results
}

func selectAndDetail(results []parser.Result) parser.Anime {
	color.Cyan("Select an anime: ")
	for index, result := range results {
		color.Magenta("%v - %v\n", index+1, result.Name)
	}

	selectionIndex, err := cli.ReadIntegerFiltered(func(i int) bool {
		return !(i < 1 || i > len(results))
	})

	if err != nil {
		log.Fatalln("Cannot do selection", err)
	}

	anime, err := parser.Details(results[selectionIndex-1])

	if err != nil {
		log.Fatalln("Cannot get details", err)
	}

	color.Cyan("Selected anime: %v", color.MagentaString(anime.Name))
	return anime
}

func selectEpisodes(episodes []parser.Episode) []parser.Episode {
	color.Cyan("Please select episode(s) to download.")
	color.Cyan("If you finished selecting enter 'yes'")

	var selectedEpisodes []int

	printEpisodeSelector := func(initialRun bool) {
		if !initialRun {
			cli.MoveUp(len(episodes) + 1)
			cli.CarriageReturn()
		}
		for index, episode := range episodes {
			status := " "
			if slices.Contains(selectedEpisodes, index+1) {
				status = "x"
			}
			color.Magenta("[%v] %v - %v", status, index+1, episode.Name)
		}
	}

	printEpisodeSelector(true)

	for {
		input := cli.ReadLine()

		if strings.EqualFold(input, "yes") {
			break
		}

		i, err := strconv.Atoi(input)

		if err != nil || i < 1 || i > len(episodes) {
			printEpisodeSelector(false)
			continue
		}

		if slices.Contains(selectedEpisodes, i) {
			index := slices.Index(selectedEpisodes, i)
			selectedEpisodes = slices.Delete(selectedEpisodes, index, index+1)
		} else {
			selectedEpisodes = append(selectedEpisodes, i)
		}

		printEpisodeSelector(false)
	}

	// convert to parser.Episode
	converted := []parser.Episode{}
	for _, i := range selectedEpisodes {
		converted = append(converted, episodes[i-1])
	}
	return converted
}

func selectVideos(selectedEpisodes []parser.Episode) []VideoDownloadInfo {
	var selectedVideos []VideoDownloadInfo

	for _, episode := range selectedEpisodes {
		videos, err := parser.Videos(episode)
		if err != nil {
			log.Fatalln("Cannot get videos", err)
		}

		color.Cyan("Select a video to download for '%v':", color.MagentaString(episode.Name))
		for i, video := range videos {
			color.Magenta("%v - %v", i+1, video.Name)
		}

		selection, err := cli.ReadIntegerFiltered(func(i int) bool {
			return !(i < 1 || i > len(videos))
		})

		if err != nil {
			log.Fatalln("Cannot do selection", err)
		}

		selectedVideo := videos[selection-1]

		color.Cyan("Specify a path to download the video:")
		path := cli.ReadLine()

		selectedVideos = append(selectedVideos, VideoDownloadInfo{
			Video: selectedVideo,
			Path:  path,
		})
	}

	return selectedVideos
}

func download(wg *sync.WaitGroup, video VideoDownloadInfo) {
	defer wg.Done()

	if video.Video.Audio == "" {
		err := extractors.DownloadFile(video.Path, video.Video.Link, video.Video.Headers)
		if err != nil {
			log.Printf("Cannot download %v - %v", video.Video.Name, err)
		}
	} else {
		// This means it's a playlist
		videotmp := video.Path + ".video.tmp"
		audiotmp := video.Path + ".audio.tmp"
		err := extractors.DownloadPlaylist(videotmp, video.Video.Link, video.Video.Headers)
		if err != nil {
			log.Printf("Cannot download %v - %v", video.Video.Name, err)
			return
		}
		err = extractors.DownloadPlaylist(audiotmp, video.Video.Audio, video.Video.Headers)
		if err != nil {
			log.Printf("Cannot download audio for %v - %v", video.Video.Name, err)
			return
		}

		in1 := ffmpeg_go.Input(videotmp).Video()
		in2 := ffmpeg_go.Input(audiotmp).Audio()
		err = ffmpeg_go.Output(
			[]*ffmpeg_go.Stream{in1, in2},
			video.Path,
			ffmpeg_go.KwArgs{
				"c": "copy",
			}).OverWriteOutput().ErrorToStdOut().Run()

		err = os.Remove(videotmp)
		if err != nil {
			log.Printf("Cannot remove temporary file for %v - %v", video.Video.Name, err)
		}
		err = os.Remove(audiotmp)
		if err != nil {
			log.Printf("Cannot remove temporary file for %v - %v", video.Video.Name, err)
		}
	}
}
