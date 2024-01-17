package main

import (
	"math/rand"
	"strings"
	"time"

	"github.com/confluentinc/go-prompt"
	"github.com/sourcegraph/go-lsp"
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
		if _, ok := specialSplitTokens[char]; ok {
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

/* This outputs all words in the line with their respective color */
func Lexer(line string) []prompt.LexerElement {
	lexerWords := []prompt.LexerElement{}

	if line == "" {
		return lexerWords
	}

	words := splitWithSeparators(line)

	for _, word := range words {
		element := prompt.LexerElement{Text: word}
		if strings.ToLower(word) == "select" {
			element.Color = prompt.Yellow
		}

		lexerWords = append(lexerWords, element)
	}

	return lexerWords
}

func main() {
	p := prompt.New(nil, completer,
		prompt.OptionTitle("sql-prompt"),
		prompt.OptionHistory([]string{"SELECT * FROM users;"}),
		prompt.OptionPrefixTextColor(prompt.Yellow),
		prompt.OptionPreviewSuggestionTextColor(prompt.Blue),
		prompt.OptionSelectedSuggestionBGColor(prompt.LightGray),
		prompt.OptionSuggestionBGColor(prompt.DarkGray),
		prompt.OptionSetLexer(Lexer), // We set the lexer so that we can see that diagnostics highlighting takes precedence if it is set
	)

	mockDiagnostic := lsp.Diagnostic{
		Range: lsp.Range{
			Start: lsp.Position{Line: 0, Character: 0},
			End:   lsp.Position{Line: 0, Character: rand.Intn(10)},
		},
		Severity: 1,
		Code:     "1234",
		Source:   "mock source",
		Message:  "Error: this is a lsp diagnostic",
	}

	p.SetDiagnostics([]lsp.Diagnostic{mockDiagnostic})

	// We highlight the first x (0-10) characters of the first line every 5 seconds
	go func() {
		for {
			time.Sleep(5 * time.Second)

			mockDiagnostic := lsp.Diagnostic{
				Range: lsp.Range{
					Start: lsp.Position{Line: 0, Character: 0},
					End:   lsp.Position{Line: 0, Character: rand.Intn(10)},
				},
				Severity: 1,
				Code:     "1234",
				Source:   "mock source",
				Message:  "Error: this is a lsp diagnostic",
			}

			p.SetDiagnostics([]lsp.Diagnostic{mockDiagnostic})
		}
	}()

	p.Input()
}
