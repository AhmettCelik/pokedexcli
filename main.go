package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/AhmettCelik/pokedexcli/internal/pokecache"
)

type config struct {
	next     string
	previous string
}

type locationAreaResponse struct {
	Count    int    `json:"count"`
	Next     string `json:"next"`
	Previous string `json:"previous"`
	Areas    []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"results"`
}

type cliCommand struct {
	name        string
	description string
	callback    func(c *config) error
}

func cleanInput(text string) []string {
	text = strings.TrimSpace(text)
	text = strings.ToLower(text)
	words := strings.Fields(text)
	return words
}

var pokeCache *pokecache.Cache

func main() {
	pokeCache = pokecache.NewCache(5 * time.Minute)
	scanner := bufio.NewScanner(os.Stdin)
	commands := map[string]cliCommand{}

	commandExit := func(c *config) error {
		fmt.Println("")
		fmt.Println("Closing the Pokedex... Goodbye!")
		os.Exit(0)
		return nil
	}

	commandHelp := func(c *config) error {
		if len(commands) == 0 {
			return fmt.Errorf("There are no commands founds available")
		}
		fmt.Println("")
		fmt.Println("Welcome to the Pokedex!\nUsage:")
		fmt.Println("")
		for _, cmd := range commands {
			fmt.Printf("%s: %s\n", cmd.name, cmd.description)
		}
		return nil
	}

	commandMap := func(c *config) error {
		url := c.next
		if url == "" {
			url = "https://pokeapi.co/api/v2/location-area/"
		}

		res, err := http.Get(url)
		if err != nil {
			fmt.Println("Error getting response: ", err)
			return err
		}
		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		if err != nil {
			fmt.Println("Error reading body: ", err)
			return err
		}

		if res.StatusCode > 299 {
			return fmt.Errorf("Response failed with status code: %d and\nbody: %s\n", res.StatusCode, body)
		}

		locationAreaResponse := locationAreaResponse{}
		err = json.Unmarshal(body, &locationAreaResponse)
		if err != nil {
			return err
		}

		fmt.Println()
		for _, area := range locationAreaResponse.Areas {
			fmt.Println(area.Name)
		}
		fmt.Println()

		c.next = locationAreaResponse.Next
		c.previous = locationAreaResponse.Previous

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

	commands["map"] = cliCommand{
		name:        "map",
		description: "Displays a help message",
		callback:    commandMap,
	}

	c := config{}

	for {
		fmt.Print("Pokedex > ")
		scanner.Scan()
		input := scanner.Text()
		words := cleanInput(input)
		if len(words) == 1 {
			cmd, exists := commands[words[0]]
			if exists {
				cmd.callback(&c)
			} else {
				fmt.Println("Unknown command")
			}
		}
	}
}
