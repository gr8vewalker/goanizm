package main

import (
	"log"
	"slices"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/gr8vewalker/goanizm/internal/cli"
	"github.com/gr8vewalker/goanizm/internal/parser"
)

func main() {
	results := search()
	anime := selectAndDetail(results)
	selectedEpisodes := selectEpisodes(anime.Episodes)
	color.Red("Selected %v episodes", len(selectedEpisodes))
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

func selectEpisodes(episodes []parser.Episode) []int {
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

	return selectedEpisodes
}
