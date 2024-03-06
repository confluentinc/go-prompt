package main

import (
	"fmt"
	"github.com/confluentinc/go-prompt"
	"os"
	"time"
)

func main() {
	file, err := os.Create("input.txt")
	if err != nil {
		fmt.Printf("Error creating file: %v\n", err)
		return
	}
	defer cleanup(file)

	err = os.Setenv(prompt.EnvVarInputFile, file.Name())
	if err != nil {
		fmt.Printf("Error setting env var: %v\n", err)
		return
	}

	go func() {
		time.Sleep(1 * time.Second)
		simulateUserInput(file, "test input")
	}()

	p, err := prompt.New(nil, nil, prompt.OptionTitle("sql-prompt"),
		prompt.OptionPrefixTextColor(prompt.Yellow),
		prompt.OptionPreviewSuggestionTextColor(prompt.Blue),
		prompt.OptionSelectedSuggestionBGColor(prompt.LightGray),
		prompt.OptionSuggestionBGColor(prompt.DarkGray),
	)
	if err != nil {
		fmt.Printf("Error creating prompt: %v\n", err)
		return
	}

	fmt.Println("Your input: " + p.Input())

	go func() {
		time.Sleep(1 * time.Second)
		simulateUserInput(file, "test input 2")
	}()

	fmt.Println("Your input: " + p.Input())
}

func cleanup(file *os.File) {
	err := file.Close()
	if err != nil {
		fmt.Printf("failed to close file: %v\n", err)
	}
	err = os.Remove(file.Name())
	if err != nil {
		fmt.Printf("failed to remove file: %v\n", err)
	}
}

func simulateUserInput(file *os.File, textToType string) {
	for _, char := range []rune(textToType) {
		file.WriteString(string(char))
		time.Sleep(100 * time.Millisecond)
	}
	file.WriteString("\n")
}
