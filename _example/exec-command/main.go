package main

import (
	"os"
	"os/exec"

	"github.com/confluentinc/go-prompt"
)

func executor(t string) {
	if t == "bash" {
		cmd := exec.Command("bash")
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
	}
	return
}

func completer(t prompt.Document) []prompt.Suggest {
	return []prompt.Suggest{
		{Text: "bash"},
	}
}

func main() {
	p, _ := prompt.New(
		executor,
		completer,
	)
	p.Run()
}
