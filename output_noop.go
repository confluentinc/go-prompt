package prompt

type NoopWriter struct{}

func (w NoopWriter) WriteRaw(data []byte) {}

func (w NoopWriter) Write(data []byte) {}

func (w NoopWriter) WriteRawStr(data string) {}

func (w NoopWriter) WriteStr(data string) {}

func (w NoopWriter) Flush() error {
	return nil
}

func (w NoopWriter) EraseScreen() {}

func (w NoopWriter) EraseUp() {}

func (w NoopWriter) EraseDown() {}

func (w NoopWriter) EraseStartOfLine() {}

func (w NoopWriter) EraseEndOfLine() {}

func (w NoopWriter) EraseLine() {}

func (w NoopWriter) ShowCursor() {}

func (w NoopWriter) HideCursor() {}

func (w NoopWriter) CursorGoTo(row, col int) {}

func (w NoopWriter) CursorUp(n int) {}

func (w NoopWriter) CursorDown(n int) {}

func (w NoopWriter) CursorForward(n int) {}

func (w NoopWriter) CursorBackward(n int) {}

func (w NoopWriter) AskForCPR() {}

func (w NoopWriter) SaveCursor() {}

func (w NoopWriter) UnSaveCursor() {}

func (w NoopWriter) ScrollDown() {}

func (w NoopWriter) ScrollUp() {}

func (w NoopWriter) SetTitle(title string) {}

func (w NoopWriter) ClearTitle() {}

func (w NoopWriter) SetColor(fg, bg Color, bold bool) {}

func (w NoopWriter) SetDisplayAttributes(fg, bg Color, attrs ...DisplayAttribute) {}

var _ ConsoleWriter = NoopWriter{}
