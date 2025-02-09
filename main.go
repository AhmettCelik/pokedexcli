package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
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

type specificLocationAreaResponse struct {
	ID                int    `json:"id"`
	Name              string `json:"name"`
	PokemonEncounters []struct {
		Pokemon struct {
			Name string `json:"name"`
		} `json:"pokemon"`
	} `json:"pokemon_encounters"`
}

type Pokemon struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	BaseExperience int    `json:"base_experience"`
	Height         int    `json:"height"`
	Weight         int    `json:"weight"`
	Types          []struct {
		Slot int `json:"slot"`
		Type struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"type"`
	} `json:"types"`
	Stats []struct {
		BaseStat int `json:"base_stat"`
		Effort   int `json:"effort"`
		Stat     struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"stat"`
	} `json:"stats"`
}

type cliCommand struct {
	name        string
	description string
	callback    func(c *config, name string) error
}

type User struct {
	Pokedex map[string]Pokemon
}

func cleanInput(text string) []string {
	text = strings.TrimSpace(text)
	text = strings.ToLower(text)
	words := strings.Fields(text)
	return words
}

func (c *config) printAreaNames(data []byte) error {
	locationAreaResponse := locationAreaResponse{}

	err := json.Unmarshal(data, &locationAreaResponse)
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

func (c *config) printPokemonNames(data []byte) error {
	specificLocationAreaResponse := specificLocationAreaResponse{}

	err := json.Unmarshal(data, &specificLocationAreaResponse)
	if err != nil {
		return err
	}

	fmt.Println()
	for _, pokemonEncounter := range specificLocationAreaResponse.PokemonEncounters {
		fmt.Println(pokemonEncounter.Pokemon.Name)
	}
	fmt.Println()

	return nil
}

func (c *config) printPokemonIsCatched(data []byte, name string, u *User) error {
	pokemonStruct := Pokemon{}

	err := json.Unmarshal(data, &pokemonStruct)
	if err != nil {
		return err
	}

	K := 100.0

	successProb := 1 / (1 + float64(pokemonStruct.BaseExperience)/K)
	randomValue := rand.Float64()

	if randomValue < successProb {
		fmt.Println()
		fmt.Printf("%s was caught!", name)
		u.Pokedex[name] = pokemonStruct
		fmt.Println()
	} else {
		fmt.Println()
		fmt.Printf("%s escaped!", name)
		fmt.Println()
	}

	return nil
}

var pokeCache *pokecache.Cache
var user User

func main() {
	pokeCache = pokecache.NewCache(5 * time.Minute)
	scanner := bufio.NewScanner(os.Stdin)
	commands := map[string]cliCommand{}
	user.Pokedex = map[string]Pokemon{}

	commandExit := func(c *config, name string) error {
		fmt.Println("")
		fmt.Println("Closing the Pokedex... Goodbye!")
		os.Exit(0)
		return nil
	}

	commandHelp := func(c *config, name string) error {
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

	commandMap := func(c *config, name string) error {
		url := c.next
		if url == "" {
			url = "https://pokeapi.co/api/v2/location-area/"
		}

		if cachedData, ok := pokeCache.Get(url); ok {
			err := c.printAreaNames(cachedData)
			return err
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

		pokeCache.Add(url, body)

		c.printAreaNames(body)

		return nil
	}

	commandExplore := func(c *config, name string) error {
		url := "https://pokeapi.co/api/v2/location-area/" + name

		if cachedData, ok := pokeCache.Get(url); ok {
			c.printPokemonNames(cachedData)
			return nil
		}

		res, err := http.Get(url)
		if err != nil {
			return err
		}
		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}

		if res.StatusCode > 299 {
			return fmt.Errorf("Response failed with status code: %d and\nbody: %s\n", res.StatusCode, body)
		}

		pokeCache.Add(url, body)

		c.printPokemonNames(body)

		return nil
	}

	commandCatch := func(c *config, name string) error {
		fmt.Println()
		fmt.Printf("Throwing a Pokeball at %s...\n", name)

		url := "https://pokeapi.co/api/v2/pokemon/" + name

		if cachedData, ok := pokeCache.Get(url); ok {
			c.printPokemonIsCatched(cachedData, name, &user)
			return nil
		}

		res, err := http.Get(url)
		if err != nil {
			return err
		}
		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}

		if res.StatusCode > 299 {
			return fmt.Errorf("Response failed with status code: %d and\nbody: %s\n", res.StatusCode, body)
		}

		pokeCache.Add(url, body)

		c.printPokemonIsCatched(body, name, &user)

		return nil
	}

	commandInspect := func(c *config, name string) error {
		userPokedex := user.Pokedex
		pokemon, exists := userPokedex[name]

		if !exists {
			fmt.Println()
			fmt.Println("you have not caught that pokemon")
			fmt.Println()
			return nil
		}

		fmt.Printf("Name: %s\n", pokemon.Name)
		fmt.Printf("Height: %d\n", pokemon.Height)
		fmt.Printf("Weight: %d\n", pokemon.Weight)

		fmt.Println("Stats:")
		for _, s := range pokemon.Stats {
			fmt.Printf("  -%s: %d\n", s.Stat.Name, s.BaseStat)
		}

		fmt.Println("Types:")
		for _, t := range pokemon.Types {
			fmt.Printf("  - %s\n", t.Type.Name)
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

	commands["map"] = cliCommand{
		name:        "map",
		description: "Displays location areas",
		callback:    commandMap,
	}

	commands["explore"] = cliCommand{
		name:        "explore",
		description: "Displays a list of all the Pokemon located there",
		callback:    commandExplore,
	}

	commands["catch"] = cliCommand{
		name:        "catch",
		description: "Catch a pokemon",
		callback:    commandCatch,
	}

	commands["inspect"] = cliCommand{
		name:        "inspect",
		description: "Inspect one of your pokemons details",
		callback:    commandInspect,
	}

	c := config{}

	for {
		fmt.Print("Pokedex > ")
		scanner.Scan()
		input := scanner.Text()
		words := cleanInput(input)
		if len(words) == 1 {
			cmd, exists := commands[words[0]]

			if !exists {
				fmt.Println("Unknown command")
				return
			}

			cmd.callback(&c, "")
		} else {
			cmd, exists := commands[words[0]]
			args := words[1:]

			if !exists {
				fmt.Println("Unknown command")
				return
			}

			cmd.callback(&c, args[0])
		}
	}
}
