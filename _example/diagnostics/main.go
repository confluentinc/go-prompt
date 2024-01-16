package main

import (
	"strings"
	"time"

	"github.com/confluentinc/go-prompt"
	"github.com/sourcegraph/go-lsp"
)

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

		element.Text = word

		lexerWords = append(lexerWords, element)
	}

	return lexerWords
}

func main() {
	p := prompt.New(nil, nil,
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
			End:   lsp.Position{Line: 0, Character: 10},
		},
		Severity: 1,
		Code:     "1234",
		Source:   "mock source",
		Message:  "mock message",
	}

	// We highlight the first 10 characters of the first line every 5 seconds
	go func() {
		for true {
			time.Sleep(5 * time.Second)
			p.SetDiagnostics([]lsp.Diagnostic{mockDiagnostic})
		}
	}()

	p.Input()
}
