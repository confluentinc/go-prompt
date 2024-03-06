package prompt

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestFileInputParser(t *testing.T) {
	file, err := os.Create("input.txt")
	require.NoError(t, err)
	defer cleanup(file)

	err = os.Setenv(EnvVarInputFile, file.Name())
	require.NoError(t, err)

	p, err := New(nil, nil, OptionTitle("sql-prompt"),
		OptionPrefixTextColor(Yellow),
		OptionPreviewSuggestionTextColor(Blue),
		OptionSelectedSuggestionBGColor(LightGray),
		OptionSuggestionBGColor(DarkGray),
	)
	require.NoError(t, err)

	p.(*Prompt).renderer.out = NoopWriter{}

	go func() {
		time.Sleep(2 * time.Second)
		require.NoError(t, simulateUserInput(file, "first input"))
	}()
	actual := p.Input()
	require.Equal(t, "first input", actual)

	go func() {
		time.Sleep(2 * time.Second)
		require.NoError(t, simulateUserInput(file, "second input"))
	}()
	actual = p.Input()
	require.Equal(t, "second input", actual)
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

func simulateUserInput(file *os.File, textToType string) error {
	_, err := file.WriteString(textToType)
	if err != nil {
		return err
	}
	time.Sleep(100 * time.Millisecond)
	_, err = file.WriteString("\n")
	if err != nil {
		return err
	}
	return nil
}
