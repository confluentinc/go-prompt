package prompt

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/confluentinc/go-prompt/internal/debug"
	runewidth "github.com/mattn/go-runewidth"
	"github.com/sourcegraph/go-lsp"
)

// Render to render prompt information from state of Buffer.
type Render struct {
	out                ConsoleWriter
	prefix             string
	livePrefixCallback func() (prefix string, useLivePrefix bool)
	breakLineCallback  func(*Document)
	title              string
	row                uint16
	col                uint16

	previousCursor int

	// colors,
	prefixTextColor              Color
	prefixBGColor                Color
	inputTextColor               Color
	inputBGColor                 Color
	previewSuggestionTextColor   Color
	previewSuggestionBGColor     Color
	suggestionTextColor          Color
	suggestionBGColor            Color
	diagnosticsTextColor         Color
	diagnosticsBGColor           Color
	diagnosticsDetailsTextColor  Color
	diagnosticsDetailsBGColor    Color
	selectedSuggestionTextColor  Color
	selectedSuggestionBGColor    Color
	descriptionTextColor         Color
	descriptionBGColor           Color
	selectedDescriptionTextColor Color
	selectedDescriptionBGColor   Color
	scrollbarThumbColor          Color
	scrollbarBGColor             Color
}

// Setup to initialize console output.
func (r *Render) Setup() {
	if r.title != "" {
		r.out.SetTitle(r.title)
		debug.AssertNoError(r.out.Flush())
	}
}

// getCurrentPrefix to get current prefix.
// If live-prefix is enabled, return live-prefix.
func (r *Render) getCurrentPrefix() string {
	if prefix, ok := r.livePrefixCallback(); ok {
		return prefix
	}
	return r.prefix
}

func (r *Render) renderPrefix() {
	r.out.SetColor(r.prefixTextColor, r.prefixBGColor, false)
	r.out.WriteStr(r.getCurrentPrefix())
	r.out.SetColor(DefaultColor, DefaultColor, false)
}

// TearDown to clear title and erasing.
func (r *Render) TearDown() {
	r.out.ClearTitle()
	r.out.EraseDown()
	debug.AssertNoError(r.out.Flush())
}

func (r *Render) prepareArea(lines int) {
	for i := 0; i < lines; i++ {
		r.out.ScrollDown()
	}
	for i := 0; i < lines; i++ {
		r.out.ScrollUp()
	}
}

// UpdateWinSize called when window size is changed.
func (r *Render) UpdateWinSize(ws *WinSize) {
	r.row = ws.Row
	r.col = ws.Col
}

// Render completions in the dropdown and returns the lenth that the cursor has to be moved back
func (r *Render) renderCompletion(completions *CompletionManager, cursorPos int) int {
	completionsSelectedIdx := completions.GetSelectedIdx()
	completionsVerticalScroll := completions.GetVerticalScroll()
	suggestions := completions.GetSuggestions()
	if len(suggestions) == 0 {
		return 0
	}
	prefix := r.getCurrentPrefix()
	formatted, width := formatSuggestions(
		suggestions,
		int(r.col)-runewidth.StringWidth(prefix)-1, // -1 means a width of scrollbar
	)
	// +1 means a width of scrollbar.
	width++

	windowHeight := len(formatted)
	if windowHeight > int(completions.max) {
		windowHeight = int(completions.max)
	}
	formatted = formatted[completionsVerticalScroll : completionsVerticalScroll+windowHeight]
	r.prepareArea(windowHeight)

	cursor := cursorPos
	x, _ := r.toPos(cursor)
	if x+width >= int(r.col) {
		cursor = r.backward(cursor, x+width-int(r.col))
	}

	contentHeight := len(suggestions)

	fractionVisible := float64(windowHeight) / float64(contentHeight)
	fractionAbove := float64(completionsVerticalScroll) / float64(contentHeight)

	scrollbarHeight := int(clamp(float64(windowHeight), 1, float64(windowHeight)*fractionVisible))
	scrollbarTop := int(float64(windowHeight) * fractionAbove)

	isScrollThumb := func(row int) bool {
		return scrollbarTop <= row && row <= scrollbarTop+scrollbarHeight
	}

	selected := completionsSelectedIdx - completionsVerticalScroll
	r.out.SetColor(White, Cyan, false)
	for i := 0; i < windowHeight; i++ {
		r.out.CursorDown(1)
		if i == selected {
			r.out.SetColor(r.selectedSuggestionTextColor, r.selectedSuggestionBGColor, true)
		} else {
			r.out.SetColor(r.suggestionTextColor, r.suggestionBGColor, false)
		}
		r.out.WriteStr(formatted[i].Text)

		if i == selected {
			r.out.SetColor(r.selectedDescriptionTextColor, r.selectedDescriptionBGColor, false)
		} else {
			r.out.SetColor(r.descriptionTextColor, r.descriptionBGColor, false)
		}
		r.out.WriteStr(formatted[i].Description)

		if isScrollThumb(i) {
			r.out.SetColor(DefaultColor, r.scrollbarThumbColor, false)
		} else {
			r.out.SetColor(DefaultColor, r.scrollbarBGColor, false)
		}
		r.out.WriteStr(" ")
		r.out.SetColor(DefaultColor, DefaultColor, false)

		r.lineWrap(cursor + width)
		r.backward(cursor+width, width) // we go back to the same position before doing r.out.CursorDown(1) above - completions are rendered on the same pos below eacho other
	}

	if x+width >= int(r.col) {
		r.out.CursorForward(x + width - int(r.col))
	}

	r.out.SetColor(DefaultColor, DefaultColor, false)
	return int(r.col) * windowHeight // the number of characters to go up
}

// ClearScreen :: Clears the screen and moves the cursor to home
func (r *Render) ClearScreen() {
	r.out.EraseScreen()
	r.out.CursorGoTo(0, 0)
}

// Render renders to the console.
func (r *Render) Render(buffer *Buffer, lastKeyStroke Key, completion *CompletionManager, lexer *Lexer, diagnostics []lsp.Diagnostic) (tracedBackLines int) {
	// In situations where a pseudo tty is allocated (e.g. within a docker container),
	// window size via TIOCGWINSZ is not immediately available and will result in 0,0 dimensions.
	if r.col == 0 {
		return 0
	}
	defer func() { debug.AssertNoError(r.out.Flush()) }()

	prefix := r.getCurrentPrefix()
	line := buffer.Text()

	// Down, ControlN
	traceBackLines := r.previousCursor / int(r.col) // calculate number of lines we had before

	// if the new buffer is empty and we are not browsing the history using the Down/controlDown keys
	// then we reset the traceBackLines to 0 since there's nothing to trace back/erase.
	if len(line) == 0 && lastKeyStroke != ControlDown && lastKeyStroke != Down {
		traceBackLines = 0
	}
	debug.Log(fmt.Sprintln(line))
	debug.Log(fmt.Sprintln(traceBackLines))

	// prepare area by getting the end position the console cursor will be at after rendering
	cursorEndPos := r.getCursorEndPos(prefix+line, 0)

	// Clear screen
	r.clear(r.previousCursor)

	// Render new text
	r.renderPrefix()
	r.out.SetColor(DefaultColor, DefaultColor, false)
	// if diagnostics is on, we have to redefine lexer here
	r.renderLine(line, lexer, diagnostics)
	r.out.SetColor(DefaultColor, DefaultColor, false)

	// At this point the rendering is done and the cursor has moved to its end position we calculated earlier.
	// We now need to find out where the console cursor would be if it had the same position as the buffer cursor.
	translatedBufferCursorPos := r.getCursorEndPos(prefix+buffer.Document().TextBeforeCursor(), 0)
	cursorPos := r.move(cursorEndPos, translatedBufferCursorPos)

	// If suggestion is select for preview
	if suggest, ok := completion.GetSelectedSuggestion(); ok {
		cursorPos = r.backward(cursorPos, runewidth.StringWidth(buffer.Document().GetWordBeforeCursorUntilSeparator(completion.wordSeparator)))

		r.out.SetColor(r.previewSuggestionTextColor, r.previewSuggestionBGColor, false)
		r.out.WriteStr(suggest.Text)
		r.out.SetColor(DefaultColor, DefaultColor, false)

		rest := buffer.Document().TextAfterCursor()

		if lexer.IsEnabled {
			processed := lexer.Process(rest)

			var s = rest

			for _, v := range processed {
				a := strings.SplitAfter(s, v.Text)
				s = strings.TrimPrefix(s, a[0])

				r.out.SetColor(v.Color, r.inputBGColor, false)
				r.out.WriteStr(a[0])
			}
		} else {
			r.out.WriteStr(rest)
		}
		cursorPosBehindSuggestion := cursorPos + runewidth.StringWidth(suggest.Text)
		cursorEndPosWithInsertedSuggestion := r.getCursorEndPos(suggest.Text+rest, cursorPos)
		r.out.SetColor(DefaultColor, DefaultColor, false)

		cursorPos = r.move(cursorEndPosWithInsertedSuggestion, cursorPosBehindSuggestion)
	}

	// Render completions - We have to store completionLen to move back the cursor to the right position after rendering the completion or completion + diagnostics
	completionLen := r.renderCompletion(completion, cursorPos)

	// Render dianostics messages - showing error detail at the bottom of the screen.
	cursorPos = r.renderDiagnosticsMsg(cursorPos, buffer.Document().cursorPosition, completionLen, diagnostics)

	r.previousCursor = cursorPos
	return traceBackLines
}

func diagnosticsDetail(diagnostics []lsp.Diagnostic, maxCol int) string {
	var messages []string

	for _, diagnostic := range diagnostics {
		if len(diagnostic.Message) > 0 {
			rest := maxCol - len(diagnostic.Message)%maxCol
			message := diagnostic.Message + strings.Repeat(" ", rest)
			messages = append(messages, message)
		}
	}

	return "\n" + strings.Join(messages, "")
}

func hasDiagnostic(pos int, diagnostics []lsp.Diagnostic) bool {
	for _, diagnostic := range diagnostics {
		start := diagnostic.Range.Start.Character
		end := diagnostic.Range.End.Character

		if pos >= start && pos <= end {
			return true
		}
	}

	return false
}

// Render diagnostics and returns the length that the cursor has to be moved back
func (r *Render) renderDiagnosticsMsg(cursorPos, documentPos, completionLen int, diagnostics []lsp.Diagnostic) int {
	if hasDiagnostic(documentPos, diagnostics) {
		diagnosticsText := diagnosticsDetail(diagnostics, int(r.col))
		// Why we do -1 here: This is a trick due to the fact the the terminal cursor is lazy and will only create a new line if you write something.
		// So even though we filled the whole line with empty spaces, at the last line, the cursor won't automatically jump to the next line.
		// We adjust the cursor position doing -1 because that's the actual position of the terminal cursor.
		cursorEndPosWithInsertedDiagnostics := r.getCursorEndPos(diagnosticsText, cursorPos) - 1
		r.out.SetColor(White, r.diagnosticsDetailsBGColor, false)

		r.out.WriteStr(diagnosticsText)
		r.out.SetColor(White, DefaultColor, false)
		return r.move(cursorEndPosWithInsertedDiagnostics+completionLen, cursorPos)
	}

	return r.move(cursorPos+completionLen, cursorPos)
}

func (r *Render) renderDiagnostic(word string) {
	if len(word) < 1 {
		return
	}

	// Is first char whitespace
	if strings.HasPrefix(word, " ") {
		r.out.SetColor(DefaultColor, r.inputBGColor, false)
		r.out.WriteStr(" ")
		word = strings.TrimPrefix(word, " ")
	}

	// Is last char whitespace
	traillingWhitespace := false
	if len(word) > 1 && strings.HasSuffix(word, " ") {
		traillingWhitespace = true
		word = strings.TrimSuffix(word, " ")
	}

	r.out.SetColor(r.diagnosticsTextColor, r.diagnosticsBGColor, false)
	r.out.WriteStr(word)

	if traillingWhitespace {
		r.out.SetColor(DefaultColor, r.inputBGColor, false)
		r.out.WriteStr(" ")
	}
}

func (r *Render) renderLine(line string, lexer *Lexer, diagnostics []lsp.Diagnostic) {
	if lexer != nil && lexer.IsEnabled {
		processed := lexer.Process(line)
		var s = line
		pos := 0
		for _, v := range processed {
			a := strings.SplitAfter(s, v.Text)
			s = strings.TrimPrefix(s, a[0])

			if hasDiagnostic(pos, diagnostics) {
				r.renderDiagnostic(a[0])
			} else {
				r.out.SetColor(v.Color, r.inputBGColor, false)
				r.out.WriteStr(a[0])
			}

			pos += len(a[0])
		}
	} else {
		r.out.SetColor(r.inputTextColor, r.inputBGColor, false)
		r.out.WriteStr(line)
	}
}

// BreakLine to break line.
func (r *Render) BreakLine(buffer *Buffer, lexer *Lexer) {
	// Erasing and Render
	r.clear(r.getCursorEndPos(r.getCurrentPrefix()+buffer.Document().TextBeforeCursor(), 0))

	r.renderPrefix()

	if lexer.IsEnabled {
		processed := lexer.Process(buffer.Document().Text + "\n")

		var s = buffer.Document().Text + "\n"

		for _, v := range processed {
			a := strings.SplitAfter(s, v.Text)
			s = strings.TrimPrefix(s, a[0])

			r.out.SetColor(v.Color, r.inputBGColor, false)
			r.out.WriteStr(a[0])
		}
	} else {
		r.out.SetColor(r.inputTextColor, r.inputBGColor, false)
		r.out.WriteStr(buffer.Document().Text + "\n")
	}

	r.out.SetColor(DefaultColor, DefaultColor, false)

	debug.AssertNoError(r.out.Flush())
	if r.breakLineCallback != nil {
		r.breakLineCallback(buffer.Document())
	}

	r.previousCursor = 0
}

func (r *Render) getCursorEndPos(text string, startPos int) int {
	lines := strings.SplitAfter(text, "\n")
	cursor := startPos
	for _, line := range lines {
		filledCols := runewidth.StringWidth(line)
		cursor += filledCols
		if len(line) > 0 && line[len(line)-1:] == "\n" {
			remainingChars := int(r.col) - (cursor % int(r.col))
			cursor += remainingChars
		}
	}
	return cursor
}

// clear erases the screen from a beginning of input
// even if there is line break which means input length exceeds a window's width.
func (r *Render) clear(cursor int) {
	r.move(cursor, 0)
	r.out.EraseDown()
}

// backward moves cursor to backward from a current cursor position
// regardless there is a line break.
func (r *Render) backward(from, n int) int {
	return r.move(from, from-n)
}

// move moves cursor to specified position from the beginning of input
// even if there is a line break.
func (r *Render) move(from, to int) int {
	fromX, fromY := r.toPos(from)
	toX, toY := r.toPos(to)

	r.out.CursorUp(fromY - toY)
	r.out.CursorBackward(fromX - toX)
	return to
}

// toPos returns the relative position from the beginning of the string.
func (r *Render) toPos(cursor int) (x, y int) {
	col := int(r.col)
	return cursor % col, cursor / col
}

func (r *Render) lineWrap(cursor int) {
	if runtime.GOOS != "windows" && cursor > 0 && cursor%int(r.col) == 0 {
		r.out.WriteRaw([]byte{'\n'})
	}
}

func clamp(high, low, x float64) float64 {
	switch {
	case high < x:
		return high
	case x < low:
		return low
	default:
		return x
	}
}
