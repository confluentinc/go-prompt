package prompt

import (
	"strings"
	"sync"

	"github.com/confluentinc/go-prompt/internal/debug"
	runewidth "github.com/mattn/go-runewidth"
)

const (
	shortenSuffix = "..."
	leftPrefix    = " "
	leftSuffix    = " "
	rightPrefix   = " "
	rightSuffix   = " "
)

// Suggest is printed when completing.
type Suggest struct {
	Text        string
	Description string
}

// CompletionManager manages which suggestion is now selected.
type CompletionManager struct {
	selected  int // -1 means nothing one is selected.
	tmp       []Suggest
	max       uint16
	completer Completer

	verticalScroll int
	wordSeparator  string
	showAtStart    bool

	mu sync.RWMutex
}

// GetSelectedSuggestion returns the selected item.
func (c *CompletionManager) GetSelectedSuggestion() (s Suggest, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.selected == -1 {
		return Suggest{}, false
	} else if c.selected < -1 || c.selected >= len(c.tmp) {
		debug.Assert(false, "must not reach here")
		c.selected = -1
		c.verticalScroll = 0
		return Suggest{}, false
	}

	return c.tmp[c.selected], true
}

// GetSuggestions returns the list of suggestion.
func (c *CompletionManager) GetSuggestions() []Suggest {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.tmp
}

// GetSelectedIdx returns the idx of the current selection
func (c *CompletionManager) GetSelectedIdx() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.selected
}

// GetVerticalScroll returns the current value of verticalScroll.
func (c *CompletionManager) GetVerticalScroll() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.verticalScroll
}

// Reset to select nothing.
func (c *CompletionManager) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.reset()
}

func (c *CompletionManager) reset() {
	c.selected = -1
	c.verticalScroll = 0
	c.tmp = []Suggest{}
}

// Update to update the suggestions.
func (c *CompletionManager) Update(in Document) {
	if c.completer == nil {
		return
	}

	updatedSuggestions := c.completer(in)

	c.mu.Lock()
	defer c.mu.Unlock()

	c.reset()
	c.tmp = updatedSuggestions
}

// Previous to select the previous suggestion item.
func (c *CompletionManager) Previous() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.verticalScroll == c.selected && c.selected > 0 {
		c.verticalScroll--
	}
	c.selected--
	c.update()
}

// Next to select the next suggestion item.
func (c *CompletionManager) Next() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.verticalScroll+int(c.max)-1 == c.selected {
		c.verticalScroll++
	}
	c.selected++
	c.update()
}

// Completing returns whether the CompletionManager selects something one.
func (c *CompletionManager) Completing() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.selected != -1
}

func (c *CompletionManager) update() {
	max := int(c.max)
	lenSuggestions := len(c.tmp)
	if lenSuggestions < max {
		max = lenSuggestions
	}

	if c.selected >= lenSuggestions {
		c.reset()
	} else if c.selected < -1 {
		c.selected = lenSuggestions - 1
		c.verticalScroll = lenSuggestions - max
	}
}

func deleteBreakLineCharacters(s string) string {
	s = strings.Replace(s, "\n", "", -1)
	s = strings.Replace(s, "\r", "", -1)
	return s
}

func formatTexts(o []string, max int, prefix, suffix string) (new []string, width int) {
	l := len(o)
	n := make([]string, l)

	lenPrefix := runewidth.StringWidth(prefix)
	lenSuffix := runewidth.StringWidth(suffix)
	lenShorten := runewidth.StringWidth(shortenSuffix)
	min := lenPrefix + lenSuffix + lenShorten
	for i := 0; i < l; i++ {
		o[i] = deleteBreakLineCharacters(o[i])

		w := runewidth.StringWidth(o[i])
		if width < w {
			width = w
		}
	}

	if width == 0 {
		return n, 0
	}
	if min >= max {
		return n, 0
	}
	if lenPrefix+width+lenSuffix > max {
		width = max - lenPrefix - lenSuffix
	}

	for i := 0; i < l; i++ {
		x := runewidth.StringWidth(o[i])
		if x <= width {
			spaces := strings.Repeat(" ", width-x)
			n[i] = prefix + o[i] + spaces + suffix
		} else if x > width {
			x := runewidth.Truncate(o[i], width, shortenSuffix)
			// When calling runewidth.Truncate("您好xxx您好xxx", 11, "...") returns "您好xxx..."
			// But the length of this result is 10. So we need fill right using runewidth.FillRight.
			n[i] = prefix + runewidth.FillRight(x, width) + suffix
		}
	}
	return n, lenPrefix + width + lenSuffix
}

func formatSuggestions(suggests []Suggest, max int) (new []Suggest, width int) {
	num := len(suggests)
	new = make([]Suggest, num)

	left := make([]string, num)
	for i := 0; i < num; i++ {
		left[i] = suggests[i].Text
	}
	right := make([]string, num)
	for i := 0; i < num; i++ {
		right[i] = suggests[i].Description
	}

	left, leftWidth := formatTexts(left, max, leftPrefix, leftSuffix)
	if leftWidth == 0 {
		return []Suggest{}, 0
	}
	right, rightWidth := formatTexts(right, max-leftWidth, rightPrefix, rightSuffix)

	for i := 0; i < num; i++ {
		new[i] = Suggest{Text: left[i], Description: right[i]}
	}
	return new, leftWidth + rightWidth
}

// NewCompletionManager returns initialized CompletionManager object.
func NewCompletionManager(completer Completer, max uint16) *CompletionManager {
	return &CompletionManager{
		selected:  -1,
		max:       max,
		completer: completer,

		verticalScroll: 0,
	}
}
