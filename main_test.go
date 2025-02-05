package main

import (
	"reflect"
	"testing"
)

func TestCleanInput(t *testing.T) {
	cases := []struct {
		input    string
		expected []string
	}{
        // adding test cases
		{
			input:    "  hello world     ",
			expected: []string{"hello", "world"},
		},
		{
			input:    "hELlo WORLD     ",
			expected: []string{"hello", "world"},
		},
		{
			input:    "  Charmander BulBAsaUR      PIKACHU     ",
			expected: []string{"charmander", "bulbasaur", "pikachu"},
		},
		{
			input:    "  hello world   124! ^'4 a asdf  ",
			expected: []string{"hello", "world", "124!", "^'4", "a", "asdf"},
		},
	}

	// checking the cases
	for _, c := range cases {
		result := cleanInput(c.input)
		if !reflect.DeepEqual(len(result), len(c.expected)) {
			t.Errorf("For input %d, expected %d but got %d", len(c.input), len(c.expected), len(result))
		}
		for i := range result {
			word := result[i]
			expectedWord := c.expected[i]
			if !reflect.DeepEqual(word, expectedWord) {
				t.Errorf("For word %s, expected %s but got %s", c.input, expectedWord, word)
			}
		}
	}
}
