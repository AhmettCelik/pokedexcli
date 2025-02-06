package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type cliCommand struct {
	name        string
	description string
	callback    func() error
}

func cleanInput(text string) []string {
	text = strings.TrimSpace(text)
	text = strings.ToLower(text)
	words := strings.Fields(text)
	return words
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	commands := map[string]cliCommand{}

	commandExit := func() error {
		fmt.Println("Closing the Pokedex... Goodbye")
		os.Exit(0)
		return nil
	}

	commandHelp := func() error {
		if len(commands) == 0 {
			return fmt.Errorf("There are no commands founds available")
		}
		fmt.Println("Welcome to the Pokedex!\nUsage:")
		fmt.Println("")
		for _, cmd := range commands {
			fmt.Printf("%s: %s\n", cmd.name, cmd.description)
		}
		return nil
	}

	commands["exit"] = cliCommand{
		name:        "exit",
		description: "Exit the Pokedex",
		callback:    commandExit,
	}

	commands["help"] = cliCommand{
		name:        "help",
		description: "Displays a help message",
		callback:    commandHelp,
	}

	for {
		fmt.Print("Pokedex > ")
		scanner.Scan()
		input := scanner.Text()
		words := cleanInput(input)
		if len(words) == 1 {
			cmd, exists := commands[words[0]]
			if exists {
				cmd.callback()
			} else {
				fmt.Println("Unknown command")
			}
		}
	}
}
