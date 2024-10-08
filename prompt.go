package prompt

import (
	"bytes"
	"os"
	"time"

	"github.com/confluentinc/go-prompt/internal/debug"
	"github.com/sourcegraph/go-lsp"
)

// Executor is called when user input something text.
type Executor func(string)

// ExitChecker is called after user input to check if prompt must stop and exit go-prompt Run loop.
// User input means: selecting/typing an entry, then, if said entry content matches the ExitChecker function criteria:
// - immediate exit (if breakline is false) without executor called
// - exit after typing <return> (meaning breakline is true), and the executor is called first, before exit.
// Exit means exit go-prompt (not the overall Go program)
type ExitChecker func(in string, breakline bool) bool

// Completer should return the suggest item from Document.
type Completer func(Document) []Suggest

// StatementTerminatorCb should return whether statement in buffer has been terminated
type StatementTerminatorCb func(lastKeyStroke Key, buffer *Buffer) bool

type IPrompt interface {
	Run()
	Input() string
	ClearScreen()
	SetConsoleParser(ConsoleParser)
	Buffer() *Buffer
	Renderer() *Render
	History() *History
	Lexer() *Lexer
	CompletionManager() *CompletionManager
	AddKeyBindings(...KeyBind)
	AddASCIICodeBindings(...ASCIICodeBind)
	SetKeyBindMode(KeyBindMode)
	SetCompletionOnDown(bool)
	SetExitChecker(ExitChecker)
	SetStatementTerminatorCb(StatementTerminatorCb)
	SetDiagnostics(diagnostics []lsp.Diagnostic)
}

// Prompt is core struct of go-prompt.
type Prompt struct {
	in                    ConsoleParser
	buf                   *Buffer
	prevText              string
	lastKey               Key
	renderer              *Render
	executor              Executor
	history               *History
	diagnostics           []lsp.Diagnostic
	lexer                 *Lexer
	completion            *CompletionManager
	keyBindings           []KeyBind
	ASCIICodeBindings     []ASCIICodeBind
	keyBindMode           KeyBindMode
	completionOnDown      bool
	exitChecker           ExitChecker
	statementTerminatorCb StatementTerminatorCb
	skipTearDown          bool
}

// Exec is the struct contains user input context.
type Exec struct {
	input string
}

// Run starts prompt.
func (p *Prompt) Run() {
	p.skipTearDown = false
	defer debug.Teardown()
	debug.Log("start prompt")
	p.setUp()
	defer p.tearDown()

	if p.completion.showAtStart {
		p.completion.Update(*p.buf.Document())
	}

	p.Render()

	bufCh := make(chan []byte, 128)
	stopReadBufCh := make(chan struct{})
	go p.readBuffer(bufCh, stopReadBufCh)

	exitCh := make(chan int)
	winSizeCh := make(chan *WinSize)
	stopHandleSignalCh := make(chan struct{})
	go p.handleSignals(exitCh, winSizeCh, stopHandleSignalCh)

	for {
		select {
		case b := <-bufCh:
			if shouldExit, e := p.feed(b); shouldExit {
				p.renderer.BreakLine(p.buf, p.lexer)
				stopReadBufCh <- struct{}{}
				stopHandleSignalCh <- struct{}{}
				return
			} else if e != nil {
				// Stop goroutine to run readBuffer function
				stopReadBufCh <- struct{}{}
				stopHandleSignalCh <- struct{}{}

				// Unset raw mode
				// Reset to Blocking mode because returned EAGAIN when still set non-blocking mode.
				debug.AssertNoError(p.in.TearDown())
				p.executor(e.input)

				p.completion.Update(*p.buf.Document())

				p.Render()

				if p.exitChecker != nil && p.exitChecker(e.input, true) {
					p.skipTearDown = true
					return
				}
				// Set raw mode
				debug.AssertNoError(p.in.Setup())
				go p.readBuffer(bufCh, stopReadBufCh)
				go p.handleSignals(exitCh, winSizeCh, stopHandleSignalCh)
			} else {

				p.completion.Update(*p.buf.Document())
				p.Render()
			}
		case w := <-winSizeCh:
			p.renderer.UpdateWinSize(w)
			p.Render()
		case code := <-exitCh:
			p.renderer.BreakLine(p.buf, p.lexer)
			p.tearDown()
			os.Exit(code)
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}
}

// Input just returns user input text.
func (p *Prompt) Input() string {
	defer debug.Teardown()
	debug.Log("start prompt")
	p.setUp()
	defer p.tearDown()

	if p.completion.showAtStart {
		p.completion.Update(*p.buf.Document())
	}

	p.Render()
	bufCh := make(chan []byte, 128)
	stopReadBufCh := make(chan struct{})
	go p.readBuffer(bufCh, stopReadBufCh)

	completionCh := make(chan bool)
	exitCh := make(chan int)
	winSizeCh := make(chan *WinSize)
	stopHandleSignalCh := make(chan struct{})
	go p.handleSignals(exitCh, winSizeCh, stopHandleSignalCh)

	for {
		select {
		case b := <-bufCh:
			if shouldExit, e := p.feed(b); shouldExit {
				p.renderer.BreakLine(p.buf, p.lexer)
				stopReadBufCh <- struct{}{}
				stopHandleSignalCh <- struct{}{}
				return ""
			} else if e != nil {
				// Stop goroutine to run readBuffer function
				stopReadBufCh <- struct{}{}
				stopHandleSignalCh <- struct{}{}
				return e.input
			} else {
				document := *p.buf.Document()
				// we don't want to trigger completions again while navigating existing completions
				if !p.completion.Completing() {
					go func() {
						p.completion.Update(document)
						completionCh <- true
					}()
				}
				p.Render()
			}
		case w := <-winSizeCh:
			p.renderer.UpdateWinSize(w)
			p.Render()
		case code := <-exitCh:
			p.renderer.BreakLine(p.buf, p.lexer)
			p.tearDown()
			os.Exit(code)
		case <-completionCh:
			p.Render()
		}
	}
}

// ClearScreen :: Clears the screen
func (p *Prompt) ClearScreen() {
	p.renderer.ClearScreen()
}

func (p *Prompt) SetConsoleParser(parser ConsoleParser) {
	p.in = parser
}

func (p *Prompt) Buffer() *Buffer {
	return p.buf
}

func (p *Prompt) Renderer() *Render {
	return p.renderer
}

func (p *Prompt) History() *History {
	return p.history
}

func (p *Prompt) Lexer() *Lexer {
	return p.lexer
}

func (p *Prompt) CompletionManager() *CompletionManager {
	return p.completion
}

func (p *Prompt) AddKeyBindings(keyBindings ...KeyBind) {
	p.keyBindings = append(p.keyBindings, keyBindings...)
}

func (p *Prompt) AddASCIICodeBindings(asciiCodeBindings ...ASCIICodeBind) {
	p.ASCIICodeBindings = append(p.ASCIICodeBindings, asciiCodeBindings...)
}

func (p *Prompt) SetKeyBindMode(keyBindMode KeyBindMode) {
	p.keyBindMode = keyBindMode
}

func (p *Prompt) SetCompletionOnDown(completionOnDown bool) {
	p.completionOnDown = completionOnDown
}

func (p *Prompt) SetExitChecker(exitChecker ExitChecker) {
	p.exitChecker = exitChecker
}

func (p *Prompt) SetStatementTerminatorCb(statementTerminatorCb StatementTerminatorCb) {
	p.statementTerminatorCb = statementTerminatorCb
}

func (p *Prompt) SetDiagnostics(diagnostics []lsp.Diagnostic) {
	p.diagnostics = diagnostics
	p.prevText = p.buf.Text()
	p.Render()
}

func (p *Prompt) ClearDiagnosticsOnTextChange() {
	//  If the user writes something, we clear diagnostics (highlights and error shown) because the ranges might be outdated
	if p.buf.Text() != p.prevText {
		p.diagnostics = nil
	}
}

func (p *Prompt) Render() {
	p.ClearDiagnosticsOnTextChange()
	p.renderer.Render(p.buf, p.lastKey, p.completion, p.lexer, p.diagnostics)
}

func (p *Prompt) feed(b []byte) (shouldExit bool, exec *Exec) {
	key := GetKey(b)
	p.prevText = p.buf.Text()
	// We store the last key stroke pressed to p.lastKey in the render to understand what was the last action taken.
	// For example: if the last action was going to the next erase, we want to erase the statement
	// that was in the buffer. If the last statement was sent, we want to just print a new empty buffer
	// and not erase the last statement. This could also be used for other functionalities in the future.
	p.lastKey = key
	p.buf.lastKeyStroke = key
	// completion
	completing := p.completion.Completing()
	p.handleCompletionKeyBinding(key, completing)

	switch key {
	case Enter, ControlJ, ControlM, AltEnter:
		if p.statementTerminatorCb == nil || !p.statementTerminatorCb(p.buf.lastKeyStroke, p.buf) {
			p.buf.NewLine(false)
		} else {
			p.renderer.BreakLine(p.buf, p.lexer)
			exec = &Exec{input: p.buf.Text()}
			p.buf = NewBuffer()
			if exec.input != "" {
				p.history.Add(exec.input)
			}
		}
	case ControlC:
		p.renderer.BreakLine(p.buf, p.lexer)
		p.buf = NewBuffer()
		p.history.Clear()
	case Up, ControlP:
		if !completing { // Don't use p.completion.Completing() because it takes double operation when switch to selected=-1.
			if p.buf.HasPrevLine() {
				// this is a multiline buffer
				// move the cursor up by one line
				p.buf.CursorUp(1)
			} else if newBuf, changed := p.history.Older(p.buf); changed {
				p.prevText = p.buf.Text()
				p.buf = newBuf
			}

			return
		}
	case Down, ControlN:
		if !completing { // Don't use p.completion.Completing() because it takes double operation when switch to selected=-1.
			if p.buf.HasNextLine() {
				p.buf.CursorDown(1)
			} else if newBuf, changed := p.history.Newer(p.buf); changed {

				p.prevText = p.buf.Text()
				p.buf = newBuf
			}
			return
		}
	case ControlD:
		if p.buf.Text() == "" {
			shouldExit = true
			return
		}
	case NotDefined:
		if p.handleASCIICodeBinding(b) {
			return
		}
		// After handling custom key bindings we need to sanitize the input of any
		// special characters that mess with the rendering (e.g. the escape char)
		cleanedInput := RemoveASCIISequences(b)
		p.buf.InsertText(string(cleanedInput), false, true)

		// By pressing anykey which isn't mapped we again show completions if they were hidden (by pressing escape)
		p.renderer.hideCompletion = false
	}

	shouldExit = p.handleKeyBinding(key)
	return
}

// Wheter or not we'll enter completions when the user presses down. We only navigate into completions if there's no new line below(multiline buffer) and history is not active
// (we're not browsing history with arros).
func (p *Prompt) completeOnDown() bool {
	return p.completionOnDown && !p.history.HasNewer() && !p.buf.HasNextLine()
}

func (p *Prompt) handleCompletionKeyBinding(key Key, completing bool) {
	switch key {
	case Down:
		if completing || p.completeOnDown() {
			p.completion.Next()
		}
	case Tab, ControlI:
		p.completion.Next()
	case Up:
		if completing {
			p.completion.Previous()
		}
	case BackTab:
		p.completion.Previous()
	case Escape:
		p.completion.Reset()
		p.renderer.hideCompletion = true
	default:
		if s, ok := p.completion.GetSelectedSuggestion(); ok {
			w := p.buf.Document().GetWordBeforeCursorUntilSeparator(p.completion.wordSeparator)
			if w != "" {
				p.buf.DeleteBeforeCursor(len([]rune(w)))
			}
			p.buf.InsertText(s.Text, false, true)
		}
		p.completion.Reset()
	}
}

func (p *Prompt) handleKeyBinding(key Key) bool {
	shouldExit := false
	for i := range commonKeyBindings {
		kb := commonKeyBindings[i]
		if kb.Key == key {
			kb.Fn(p.buf)
		}
	}

	if p.keyBindMode == EmacsKeyBind {
		for i := range emacsKeyBindings {
			kb := emacsKeyBindings[i]
			if kb.Key == key {
				kb.Fn(p.buf)
			}
		}
	}

	// Custom key bindings
	for i := range p.keyBindings {
		kb := p.keyBindings[i]
		if kb.Key == key {
			kb.Fn(p.buf)
		}
	}
	if p.exitChecker != nil && p.exitChecker(p.buf.Text(), false) {
		shouldExit = true
	}
	return shouldExit
}

func (p *Prompt) handleASCIICodeBinding(b []byte) bool {
	checked := false
	for _, kb := range p.ASCIICodeBindings {
		if bytes.Equal(kb.ASCIICode, b) {
			kb.Fn(p.buf)
			checked = true
		}
	}
	return checked
}

func (p *Prompt) readBuffer(bufCh chan []byte, stopCh chan struct{}) {
	debug.Log("start reading buffer")
	for {
		select {
		case <-stopCh:
			debug.Log("stop reading buffer")
			return
		default:
			if b, err := p.in.Read(); err == nil && !(len(b) == 1 && b[0] == 0) {
				bufCh <- b
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func (p *Prompt) setUp() {
	debug.AssertNoError(p.in.Setup())
	p.renderer.Setup()
	p.renderer.UpdateWinSize(p.in.GetWinSize())
}

func (p *Prompt) tearDown() {
	if !p.skipTearDown {
		debug.AssertNoError(p.in.TearDown())
	}
	p.renderer.TearDown()
}
