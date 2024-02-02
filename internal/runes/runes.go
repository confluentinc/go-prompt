package runes

import "github.com/samber/lo"

// IndexNotByte is similar with strings.IndexByte but showing the opposite behavior.
func IndexOfNot(x []rune, c byte) int {
	i := lo.LastIndexOf(x, ' ')

	if i == -1 && i+1 < len(x) {
		return i + 1
	}

	return -1
}

// LastIndexNotByte is similar with strings.LastIndexByte but showing the opposite behavior.
func LastIndexNotByte(s string, c byte) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] != c {
			return i
		}
	}
	return -1
}
