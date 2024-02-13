package prompt

func dummyExecutor(in string) {}

// Input get the input data from the user and return it.
func Input(prefix string, completer Completer, opts ...Option) string {
	pt, err := New(dummyExecutor, completer)
	if err != nil {
		return ""
	}

	pt.Renderer().prefixTextColor = DefaultColor
	pt.Renderer().prefix = prefix

	for _, opt := range opts {
		if err := opt(pt); err != nil {
			return ""
		}
	}
	return pt.Input()
}

// Choose to the shortcut of input function to select from string array.
// Deprecated: Maybe anyone want to use this.
func Choose(prefix string, choices []string, opts ...Option) string {
	completer := newChoiceCompleter(choices, FilterHasPrefix)
	pt, err := New(dummyExecutor, completer)
	if err != nil {
		return ""
	}
	pt.Renderer().prefixTextColor = DefaultColor
	pt.Renderer().prefix = prefix

	for _, opt := range opts {
		if err := opt(pt); err != nil {
			panic(err)
		}
	}
	return pt.Input()
}

func newChoiceCompleter(choices []string, filter Filter) Completer {
	s := make([]Suggest, len(choices))
	for i := range choices {
		s[i] = Suggest{Text: choices[i]}
	}
	return func(x Document) []Suggest {
		return filter(s, x.GetWordBeforeCursor(), true)
	}
}
