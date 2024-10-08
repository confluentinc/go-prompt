//go:build !windows

package prompt

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/sourcegraph/go-lsp"
	"github.com/stretchr/testify/require"
)

func emptyCompleter(in Document) []Suggest {
	return []Suggest{}
}

func TestFormatCompletion(t *testing.T) {
	scenarioTable := []struct {
		scenario      string
		completions   []Suggest
		prefix        string
		suffix        string
		expected      []Suggest
		maxWidth      int
		expectedWidth int
	}{
		{
			scenario: "",
			completions: []Suggest{
				{Text: "select"},
				{Text: "from"},
				{Text: "insert"},
				{Text: "where"},
			},
			prefix: " ",
			suffix: " ",
			expected: []Suggest{
				{Text: " select "},
				{Text: " from   "},
				{Text: " insert "},
				{Text: " where  "},
			},
			maxWidth:      20,
			expectedWidth: 8,
		},
		{
			scenario: "",
			completions: []Suggest{
				{Text: "select", Description: "select description"},
				{Text: "from", Description: "from description"},
				{Text: "insert", Description: "insert description"},
				{Text: "where", Description: "where description"},
			},
			prefix: " ",
			suffix: " ",
			expected: []Suggest{
				{Text: " select ", Description: " select description "},
				{Text: " from   ", Description: " from description   "},
				{Text: " insert ", Description: " insert description "},
				{Text: " where  ", Description: " where description  "},
			},
			maxWidth:      40,
			expectedWidth: 28,
		},
	}

	for _, s := range scenarioTable {
		ac, width := formatSuggestions(s.completions, s.maxWidth)
		if !reflect.DeepEqual(ac, s.expected) {
			t.Errorf("Should be %#v, but got %#v", s.expected, ac)
		}
		if width != s.expectedWidth {
			t.Errorf("Should be %#v, but got %#v", s.expectedWidth, width)
		}
	}
}

func TestBreakLineCallback(t *testing.T) {
	var i int
	r := &Render{
		prefix:                       "> ",
		out:                          &PosixWriter{},
		livePrefixCallback:           func() (string, bool) { return "", false },
		prefixTextColor:              Blue,
		prefixBGColor:                DefaultColor,
		inputTextColor:               DefaultColor,
		inputBGColor:                 DefaultColor,
		previewSuggestionTextColor:   Green,
		previewSuggestionBGColor:     DefaultColor,
		suggestionTextColor:          White,
		suggestionBGColor:            Cyan,
		selectedSuggestionTextColor:  Black,
		selectedSuggestionBGColor:    Turquoise,
		descriptionTextColor:         Black,
		descriptionBGColor:           Turquoise,
		selectedDescriptionTextColor: White,
		selectedDescriptionBGColor:   Cyan,
		scrollbarThumbColor:          DarkGray,
		scrollbarBGColor:             Cyan,
		col:                          1,
	}
	b := NewBuffer()
	l := NewLexer()
	r.BreakLine(b, l)

	if i != 0 {
		t.Errorf("i should initially be 0, before applying a break line callback")
	}

	r.breakLineCallback = func(doc *Document) {
		i++
	}
	r.BreakLine(b, l)
	r.BreakLine(b, l)
	r.BreakLine(b, l)

	if i != 3 {
		t.Errorf("BreakLine callback not called, i should be 3")
	}
}

func TestLinesToTracebackRender(t *testing.T) {
	scenarios := []struct {
		previousText     string
		nextText         string
		linesToTraceBack int
		lastKey          Key
	}{
		{previousText: "select..", nextText: "", linesToTraceBack: 0, lastKey: Enter},
		{previousText: "select.. \n from.. \n where..", nextText: "", linesToTraceBack: 0, lastKey: Enter},
		{previousText: "select.. \n from.. \n where..", nextText: "select..", linesToTraceBack: 2, lastKey: Tab},
		{previousText: "select.. \n from.. \n where..", nextText: "select.. \n from.. \n where field = 2", linesToTraceBack: 2, lastKey: Tab},
		{previousText: "select.. \n from.. \n where..", nextText: "select.. \n from.. \n where field = 2", linesToTraceBack: 2, lastKey: Right},
		{previousText: "select.. \n from.. ", nextText: "previous statement", linesToTraceBack: 1, lastKey: Up},
		{previousText: "select.. \n from.. ", nextText: "next statement", linesToTraceBack: 1, lastKey: Down},
		{previousText: "select.. \n from.. ", nextText: "next statement", linesToTraceBack: 1, lastKey: ControlDown},
		{previousText: "select.. \n from.. ", nextText: "", linesToTraceBack: 1, lastKey: Down},
		{previousText: "select.. \n from.. ", nextText: "", linesToTraceBack: 1, lastKey: ControlDown},
	}

	r := &Render{
		prefix:                       "> ",
		out:                          &PosixWriter{},
		livePrefixCallback:           func() (string, bool) { return "", false },
		prefixTextColor:              Blue,
		prefixBGColor:                DefaultColor,
		inputTextColor:               DefaultColor,
		inputBGColor:                 DefaultColor,
		previewSuggestionTextColor:   Green,
		previewSuggestionBGColor:     DefaultColor,
		suggestionTextColor:          White,
		suggestionBGColor:            Cyan,
		selectedSuggestionTextColor:  Black,
		selectedSuggestionBGColor:    Turquoise,
		descriptionTextColor:         Black,
		descriptionBGColor:           Turquoise,
		selectedDescriptionTextColor: White,
		selectedDescriptionBGColor:   Cyan,
		scrollbarThumbColor:          DarkGray,
		scrollbarBGColor:             Cyan,
		col:                          100,
		row:                          100,
	}

	for idx, s := range scenarios {
		fmt.Printf("Testing scenario: %v\n", idx)
		b := NewBuffer()
		b.InsertText(s.nextText, false, true)
		l := NewLexer()

		r.previousCursor = r.getCursorEndPos(s.previousText, 0)
		tracedBackLines := r.Render(b, s.lastKey, NewCompletionManager(emptyCompleter, 0), l, nil)
		require.Equal(t, s.linesToTraceBack, tracedBackLines)
	}
}

func TestGetCursorEndPosition(t *testing.T) {
	r := &Render{
		prefix:                       "> ",
		out:                          &PosixWriter{},
		livePrefixCallback:           func() (string, bool) { return "", false },
		prefixTextColor:              Blue,
		prefixBGColor:                DefaultColor,
		inputTextColor:               DefaultColor,
		inputBGColor:                 DefaultColor,
		previewSuggestionTextColor:   Green,
		previewSuggestionBGColor:     DefaultColor,
		suggestionTextColor:          White,
		suggestionBGColor:            Cyan,
		selectedSuggestionTextColor:  Black,
		selectedSuggestionBGColor:    Turquoise,
		descriptionTextColor:         Black,
		descriptionBGColor:           Turquoise,
		selectedDescriptionTextColor: White,
		selectedDescriptionBGColor:   Cyan,
		scrollbarThumbColor:          DarkGray,
		scrollbarBGColor:             Cyan,
		col:                          5,
		row:                          10,
	}

	scenarios := []struct {
		text                 string
		startPos             int
		expectedCursorEndPos int
	}{
		{text: "abc", startPos: 0, expectedCursorEndPos: 3},
		{text: "abcd", startPos: 0, expectedCursorEndPos: 4},
		{text: "abcde", startPos: 0, expectedCursorEndPos: 5},
		{text: "abc\n", startPos: 0, expectedCursorEndPos: 5},
		{text: "\nabc", startPos: 0, expectedCursorEndPos: 8},
		{text: "\nabcde", startPos: 0, expectedCursorEndPos: 10},
		{text: "\nabcdeabcde", startPos: 0, expectedCursorEndPos: 15},
		{text: "abc\n\n", startPos: 0, expectedCursorEndPos: 10},
		{text: "abc\nd", startPos: 0, expectedCursorEndPos: 6},
		{text: "ab\nc", startPos: 0, expectedCursorEndPos: 6},
		{text: "ab\n\nc", startPos: 0, expectedCursorEndPos: 11},
		{text: "ab\n\nc", startPos: 2, expectedCursorEndPos: 11},
		{text: "ab\n\nc", startPos: 3, expectedCursorEndPos: 16},
		{text: "ab\n\ncdefghijk", startPos: 3, expectedCursorEndPos: 24},
	}

	for idx, s := range scenarios {
		fmt.Printf("Testing scenario: %v\n", idx)
		actualEndPos := r.getCursorEndPos(s.text, s.startPos)
		require.Equal(t, s.expectedCursorEndPos, actualEndPos)
	}

}

func TestDiagnosticsDetail(t *testing.T) {
	// Test with multiple diagnostics
	diagnostics := []lsp.Diagnostic{
		{Message: "Error 1"},
		{Message: "Error 2"},
		{Message: "Error 3"},
	}

	expected := "\nError 1   Error 2   Error 3   "
	actual := diagnosticsDetail(diagnostics, 10, 10)
	require.Equal(t, expected, actual)

	// Test with a single diagnostic
	diagnostics = []lsp.Diagnostic{
		{Message: "Single Error"},
	}

	expected = "\nSingle Error        "
	actual = diagnosticsDetail(diagnostics, 10, 10)
	require.Equal(t, expected, actual)

	// Test with multiple diagnostics with multiline messages
	diagnostics = []lsp.Diagnostic{
		{Message: "Error 1\nNextline"},
		{Message: "Error 2\nNextline"},
		{Message: "Error 3\nNextline"},
	}

	expected = "\nError 1   Nextline  Error 2   Nextline  Error 3   Nextline  "
	actual = diagnosticsDetail(diagnostics, 10, 10)
	require.Equal(t, expected, actual)

	// Test with a single diagnostic  with multiline messages
	diagnostics = []lsp.Diagnostic{
		{Message: "A long error\nNextline"},
	}

	expected = "\nA long error        Nextline  " // Here the first error overflow the column width (12 chars and 10 columns)
	actual = diagnosticsDetail(diagnostics, 10, 10)
	require.Equal(t, expected, actual)

	// Test with no diagnostics
	diagnostics = []lsp.Diagnostic{}

	expected = ""
	actual = diagnosticsDetail(diagnostics, 10, 10)
	require.Equal(t, expected, actual)

	require.Equal(t, expected, actual)

	// Test truncate text if too diagnostics overflow the row count
	diagnostics = []lsp.Diagnostic{
		{Message: "Row 1"},
		{Message: "Row 2"},
		{Message: "Row 3"},
		{Message: "Row 4"},
		{Message: "Row 5"},
		{Message: "Row 6"},
	}

	expected = "\nRow 1     Row 2     Row 3     Row 4     Row 5  ..."
	actual = diagnosticsDetail(diagnostics, 5, 10)
	require.Equal(t, expected, actual)

	// Test truncate text if messages wil lead to too many rows
	diagnostics = []lsp.Diagnostic{
		{Message: "Row 1\nRow 2"},
		{Message: "Row 3\nRow 4"},
		{Message: "Row 5\nRow 6"},
	}

	expected = "\nRow 1     Row 2     Row 3     Row 4     Row 5  ..."
	actual = diagnosticsDetail(diagnostics, 5, 10)
	require.Equal(t, expected, actual)
}

func TestHasDiagnostic(t *testing.T) {
	diagnostics := []lsp.Diagnostic{
		{
			Range: lsp.Range{
				Start: lsp.Position{Line: 0, Character: 0},
				End:   lsp.Position{Line: 0, Character: 10},
			},
		},
		{
			Range: lsp.Range{
				Start: lsp.Position{Line: 0, Character: 20},
				End:   lsp.Position{Line: 0, Character: 30},
			},
		},
	}

	// Test within range
	require.True(t, hasDiagnostic(0, 5, diagnostics))
	require.True(t, hasDiagnostic(0, 25, diagnostics))

	// Test on the boundaries
	require.True(t, hasDiagnostic(0, 0, diagnostics))
	require.True(t, hasDiagnostic(0, 10, diagnostics))
	require.True(t, hasDiagnostic(0, 20, diagnostics))
	require.True(t, hasDiagnostic(0, 30, diagnostics))

	// Test multibytes chars
	stmt := "ああああああああああ"
	pos := len([]rune(stmt))
	require.True(t, hasDiagnostic(0, pos, diagnostics))
	require.False(t, hasDiagnostic(0, pos+1, diagnostics))

	// Test outside of range
	require.False(t, hasDiagnostic(0, -1, diagnostics))
	require.False(t, hasDiagnostic(0, 11, diagnostics))
	require.False(t, hasDiagnostic(0, 19, diagnostics))
	require.False(t, hasDiagnostic(0, 31, diagnostics))

	// Test for different line (1)
	// Test within range
	require.False(t, hasDiagnostic(1, 5, diagnostics))
	require.False(t, hasDiagnostic(1, 25, diagnostics))

	// Test on the boundaries
	require.False(t, hasDiagnostic(1, 0, diagnostics))
	require.False(t, hasDiagnostic(1, 10, diagnostics))
	require.False(t, hasDiagnostic(1, 20, diagnostics))
	require.False(t, hasDiagnostic(1, 30, diagnostics))

	// Test outside of range
	require.False(t, hasDiagnostic(1, -1, diagnostics))
	require.False(t, hasDiagnostic(1, 11, diagnostics))
	require.False(t, hasDiagnostic(1, 19, diagnostics))
	require.False(t, hasDiagnostic(1, 31, diagnostics))

}
