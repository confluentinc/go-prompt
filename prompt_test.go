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
