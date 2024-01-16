package main

import (
	"fmt"
	"strings"

	"github.com/confluentinc/go-prompt"
)

func completer(in prompt.Document) []prompt.Suggest {
	s := []prompt.Suggest{
		{Text: "users", Description: "Store the username and age"},
		{Text: "articles", Description: "Store the article text posted by user"},
		{Text: "comments", Description: "Store the text commented to articles"},
		{Text: "groups", Description: "Combine users with specific rules"},
	}
	return prompt.FilterHasPrefix(s, in.GetWordBeforeCursor(), true)
}

var SpecialSplitTokens = map[int32]uint8{
	'\t': 1,
	'\n': 1,
	'\v': 1,
	'\f': 1,
	'\r': 1,
	' ':  1,
	';':  1,
	'=':  1,
	'<':  1,
	'>':  1,
	',':  1,
}

func splitWithSeparators(line string) []string {
	words := []string{}
	word := ""

	for _, char := range line {
		if _, ok := SpecialSplitTokens[char]; ok {
			if word != "" {
				words = append(words, word)
			}
			words = append(words, string(char))
			word = ""
		} else {
			word += string(char)
		}
	}
	if word != "" {
		words = append(words, word)
	}
	return words
}

/* This outputs words all characters in the line with their respective color */
func Lexer(line string) []prompt.LexerElement {
	lexerWords := []prompt.LexerElement{}

	if line == "" {
		return lexerWords
	}

	words := splitWithSeparators(line)

	for _, word := range words {
		element := prompt.LexerElement{}

		if strings.ToLower(word) == "select" {
			element.Color = prompt.Yellow
		}

		// We have to maintain the spaces between words if not the last word
		element.Text = word

		lexerWords = append(lexerWords, element)
	}

	return lexerWords
}

func main() {
	in := prompt.Input(">>> ", completer,
		prompt.OptionTitle("sql-prompt"),
		prompt.OptionHistory([]string{"SELECT * FROM users;"}),
		prompt.OptionPrefixTextColor(prompt.Yellow),
		prompt.OptionPreviewSuggestionTextColor(prompt.Blue),
		prompt.OptionSelectedSuggestionBGColor(prompt.LightGray),
		prompt.OptionSuggestionBGColor(prompt.DarkGray),
		prompt.OptionSetLexer(Lexer))
	fmt.Println("Your input: " + in)
}
