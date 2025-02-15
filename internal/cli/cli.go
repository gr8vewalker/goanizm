package cli

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

func MoveUp(n int) {
	fmt.Printf("\033[%dA", n)
}

func ClearLine() {
	fmt.Print("\033[2K\r")
}

func CarriageReturn() {
	fmt.Print("\r")
}

func ReadInteger() (int, error) {
	return ReadIntegerFiltered(nil)
}

func ReadIntegerFiltered(filter func(int) bool) (int, error) {
	var read int

	for {
		str := ReadLine()

		i, err := strconv.Atoi(str)

		if err != nil || (filter != nil && !filter(i)) { // try until a good integer input comes.
			continue
		}
		read = i
		break
	}

	return read, nil
}

func ReadLine() string {
	in := bufio.NewReader(os.Stdin)
	line, err := in.ReadString('\n')
	if err != nil {
		log.Fatalln("Cannot read line from input", err)
	}
	return strings.TrimSpace(line)
}
