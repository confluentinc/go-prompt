package prompt

import (
	"testing"

	"github.com/sourcegraph/go-lsp"
	"github.com/stretchr/testify/require"
)

func TestClearDiagnosticsOnTextChange(t *testing.T) {
	// Create a Prompt instance with buf.Text() != prevText
	b := NewBuffer()
	b.setText("new text")
	p := &Prompt{
		buf:      b,
		prevText: "old text",
		diagnostics: []lsp.Diagnostic{
			{Message: "Error 1"},
		},
	}

	p.ClearDiagnosticsOnTextChange()
	require.Nil(t, p.diagnostics)

	// Create a Prompt instance with buf.Text() == prevText
	b.setText("same text")
	p = &Prompt{
		buf:      b,
		prevText: "same text",
		diagnostics: []lsp.Diagnostic{
			{Message: "Error 1"},
		},
	}

	p.ClearDiagnosticsOnTextChange()
	require.NotNil(t, p.diagnostics)
}

func TestCompleteOnDown(t *testing.T) {
	p := &Prompt{
		completionOnDown: true,
		buf:              NewBuffer(),
		history:          &History{},
	}
	require.True(t, p.completeOnDown())

	// Do not complete because completionOnDown is false
	p = &Prompt{
		completionOnDown: false,
		buf:              NewBuffer(),
		history:          &History{},
	}
	require.False(t, p.completeOnDown())

	// Do not complete because history is active (select < len(tmp) - 1)
	p = &Prompt{
		completionOnDown: true,
		buf:              NewBuffer(),
		history: &History{
			tmp:      []string{"foo", "bar"},
			selected: 0,
		},
	}
	require.False(t, p.completeOnDown())

	// Do not complete because buf has next line that we can navigate to
	b := NewBuffer()
	b.InsertText("first line", false, false)
	b.NewLine(true)
	b.InsertText("second line", false, false)
	b.setCursorPosition(0)

	p = &Prompt{
		completionOnDown: true,
		buf:              b,
		history:          &History{},
	}
	require.False(t, p.completeOnDown())

	// Do not complete because all three conditions are false
	p = &Prompt{
		completionOnDown: false,
		buf:              b,
		history: &History{
			tmp:      []string{"foo", "bar"},
			selected: 0,
		},
	}
	require.False(t, p.completeOnDown())
}

func TestHandleCompletionKeyBinding(t *testing.T) {
	p := &Prompt{
		renderer: &Render{},
		completion: &CompletionManager{
			selected:       0,
			verticalScroll: 1,
			tmp: []Suggest{
				{Text: "foo", Description: "foo description"},
			},
		},
	}

	p.handleCompletionKeyBinding(Down, false)
	require.False(t, p.renderer.hideCompletion)
	require.Equal(t, p.completion.selected, 0)
	require.Equal(t, p.completion.verticalScroll, 1)
	require.Equal(t, len(p.completion.tmp), 1)

	p.handleCompletionKeyBinding(Escape, false)
	require.True(t, p.renderer.hideCompletion)
	require.Equal(t, p.completion.selected, -1)
	require.Equal(t, p.completion.verticalScroll, 0)
	require.Equal(t, len(p.completion.tmp), 0)
}

func TestFeedEscape(t *testing.T) {
	p := &Prompt{
		completionOnDown: true,
		buf:              NewBuffer(),
		history:          &History{},
		renderer:         &Render{},
		completion: &CompletionManager{
			selected:       0,
			verticalScroll: 1,
			tmp: []Suggest{
				{Text: "foo", Description: "foo description"},
			},
		},
	}
	require.False(t, p.renderer.hideCompletion)
	require.Equal(t, p.completion.selected, 0)
	require.Equal(t, p.completion.verticalScroll, 1)
	require.Equal(t, len(p.completion.tmp), 1)

	key := []byte{0x1b} // {Key: Escape, ASCIICode: []byte{0x1b}}
	p.feed(key)
	require.True(t, p.renderer.hideCompletion)
	require.Equal(t, p.completion.selected, -1)
	require.Equal(t, p.completion.verticalScroll, 0)
	require.Equal(t, len(p.completion.tmp), 0)

	key = []byte{0x40} // We need to send something not mapped, for example Space, ASCIICode 64, Hex 0x40
	p.feed(key)
	require.False(t, p.renderer.hideCompletion)
	require.Equal(t, p.completion.selected, -1)
	require.Equal(t, p.completion.verticalScroll, 0)
	require.Equal(t, len(p.completion.tmp), 0)
}
