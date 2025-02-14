package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/gr8vewalker/goanizm/internal/parser"
)

func readLine() string {
	in := bufio.NewReader(os.Stdin)
	line, err := in.ReadString('\n')
	if err != nil {
		log.Fatalln("Cannot read line from input", err)
	}
	return strings.TrimSpace(line)
}

func main() {
	color.Cyan("What do you want to download?")
	query := readLine()
	results, err := parser.Search(query)
	if err != nil {
		log.Fatalln("Cannot done a search", err)
	}

	color.Cyan("Select an anime: ")
	for index, result := range results {
		color.Magenta("%v - %v\n", index, result.Name)
	}

	var selectionIndex uint
	fmt.Scanln(&selectionIndex)

	selected := results[selectionIndex]

	anime, err := parser.Details(selected)
	color.Cyan("Selected anime: %v", color.MagentaString(anime.Name))
}
