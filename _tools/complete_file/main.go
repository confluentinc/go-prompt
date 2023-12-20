package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/confluentinc/go-prompt"
	"github.com/confluentinc/go-prompt/completer"
)

var filePathCompleter = completer.FilePathCompleter{
	IgnoreCase: true,
	Filter: func(entry os.DirEntry) bool {
		return entry.IsDir() || strings.HasSuffix(entry.Name(), ".go")
	},
}

func executor(in string) {
	fmt.Println("Your input: " + in)
}

func completerFunc(d prompt.Document) []prompt.Suggest {
	t := d.GetWordBeforeCursor()
	if strings.HasPrefix(t, "--") {
		return []prompt.Suggest{
			{"--foo", ""},
			{"--bar", ""},
			{"--baz", ""},
		}
	}
	return filePathCompleter.Complete(d)
}

func main() {
	p := prompt.New(
		executor,
		completerFunc,
		prompt.OptionPrefix(">>> "),
		prompt.OptionCompletionWordSeparator(completer.FilePathCompletionSeparator),
	)
	p.Run()
}
