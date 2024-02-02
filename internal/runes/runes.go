package runes

// IndexOfNot returns the index of the first rune in x that is not c
func IndexOfNot(x []rune, c rune) int {
	n := len(x)
	for i := 0; i < n; i++ {
		if x[i] != c {
			return i
		}
	}
	return -1
}
