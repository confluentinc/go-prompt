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

func TestDiagnosticsOnArrowKeys(t *testing.T) {
	const selectAFrom = "select a from"
	scenarios := []struct {
		previousText     string
		nextText         string
		linesToTraceBack int
		lastKey          Key
	}{
		{previousText: selectAFrom, nextText: selectAFrom, linesToTraceBack: 0, lastKey: Up},
		{previousText: selectAFrom, nextText: selectAFrom, linesToTraceBack: 0, lastKey: Left},
		{previousText: selectAFrom, nextText: selectAFrom, linesToTraceBack: 0, lastKey: Right},
		{previousText: selectAFrom, nextText: selectAFrom, linesToTraceBack: 0, lastKey: Down},
		{previousText: selectAFrom, nextText: selectAFrom, linesToTraceBack: 0, lastKey: Escape},
	}

	buf := make([]byte, 1024)
	writer := VT100Writer{buffer: buf}
	posixWriter := &PosixWriter{writer, 0}
	diagnostics := []lsp.Diagnostic{{
		Range: lsp.Range{
			Start: lsp.Position{Line: 0, Character: 0},
			End:   lsp.Position{Line: 0, Character: 10},
		},
		Message: "mock message",
	}}

	r := &Render{
		diagnostics:        diagnostics,
		out:                posixWriter,
		livePrefixCallback: func() (string, bool) { return "", false },
		col:                100,
		row:                100,
	}

	for idx, s := range scenarios {
		fmt.Printf("Testing scenario: %v\n", idx)
		b := NewBuffer()
		b.InsertText(s.nextText, false, true)
		r.previousCursor = r.getCursorEndPos(s.previousText, 0)

		r.Render(b, s.previousText, s.lastKey, NewCompletionManager(emptyCompleter, 0), nil)
		require.NotNil(t, r.diagnostics)
	}
}

func TestDiagnosticsNilOnTextChange(t *testing.T) {
	const selectAFrom = "select a from"
	scenarios := []struct {
		previousText     string
		nextText         string
		linesToTraceBack int
		lastKey          Key
	}{
		{previousText: selectAFrom, nextText: selectAFrom, linesToTraceBack: 0, lastKey: Enter},
		{previousText: selectAFrom, nextText: "select a fro", linesToTraceBack: 0, lastKey: Enter},
	}

	buf := make([]byte, 1024)
	writer := VT100Writer{buffer: buf}
	posixWriter := &PosixWriter{writer, 0}
	diagnostics := []lsp.Diagnostic{{
		Range: lsp.Range{
			Start: lsp.Position{Line: 0, Character: 0},
			End:   lsp.Position{Line: 0, Character: 10},
		},
		Message: "mock message",
	}}

	r := &Render{
		diagnostics:        diagnostics,
		out:                posixWriter,
		livePrefixCallback: func() (string, bool) { return "", false },
		col:                100,
		row:                100,
	}

	for idx, s := range scenarios {
		fmt.Printf("Testing scenario: %v\n", idx)
		b := NewBuffer()
		b.InsertText(s.nextText, false, true)
		r.previousCursor = r.getCursorEndPos(s.previousText, 0)

		r.Render(b, s.previousText, s.lastKey, NewCompletionManager(emptyCompleter, 0), nil)
	}
	require.Nil(t, r.diagnostics)
}

func TestDiagnosticsAlwaysNil(t *testing.T) {
	const selectAFrom = "select a from"
	scenarios := []struct {
		previousText     string
		nextText         string
		linesToTraceBack int
		lastKey          Key
	}{
		{previousText: selectAFrom, nextText: selectAFrom, linesToTraceBack: 0, lastKey: Up},
		{previousText: selectAFrom, nextText: selectAFrom, linesToTraceBack: 0, lastKey: Left},
		{previousText: selectAFrom, nextText: selectAFrom, linesToTraceBack: 0, lastKey: Right},
		{previousText: selectAFrom, nextText: selectAFrom, linesToTraceBack: 0, lastKey: Down},
		{previousText: selectAFrom, nextText: selectAFrom, linesToTraceBack: 0, lastKey: Escape},
		{previousText: selectAFrom, nextText: "another text", linesToTraceBack: 0, lastKey: Escape},
	}

	buf := make([]byte, 1024)
	writer := VT100Writer{buffer: buf}
	posixWriter := &PosixWriter{writer, 0}

	r := &Render{
		diagnostics:        nil,
		out:                posixWriter,
		livePrefixCallback: func() (string, bool) { return "", false },
		col:                100,
		row:                100,
	}

	for idx, s := range scenarios {
		fmt.Printf("Testing scenario: %v\n", idx)
		b := NewBuffer()
		b.InsertText(s.nextText, false, true)
		r.previousCursor = r.getCursorEndPos(s.previousText, 0)

		r.Render(b, s.previousText, s.lastKey, NewCompletionManager(emptyCompleter, 0), nil)
		require.Nil(t, r.diagnostics)
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
	require.True(t, hasDiagnostic(5, diagnostics))
	require.True(t, hasDiagnostic(25, diagnostics))

	// Test on the boundaries
	require.True(t, hasDiagnostic(0, diagnostics))
	require.True(t, hasDiagnostic(10, diagnostics))
	require.True(t, hasDiagnostic(20, diagnostics))
	require.True(t, hasDiagnostic(30, diagnostics))

	// Test outside of range
	require.False(t, hasDiagnostic(-1, diagnostics))
	require.False(t, hasDiagnostic(11, diagnostics))
	require.False(t, hasDiagnostic(19, diagnostics))
	require.False(t, hasDiagnostic(31, diagnostics))
}
