package main

import (
	"fmt"
	"strings"
)

func cleanInput(text string) []string {
	text = strings.TrimSpace(text)
	text = strings.ToLower(text)
	words := strings.Fields(text)
	return words
}

func main() {
	fmt.Println("Hello, World!")
}

