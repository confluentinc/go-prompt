package prompt

import (
	"os"
)

// FileInputParser is a ConsoleParser implementation for POSIX environment.
type FileInputParser struct {
	filepath string
	file     *os.File
	offset   int64
}

// Setup should be called before starting input
func (t *FileInputParser) Setup() error {
	var err error
	t.file, err = os.OpenFile(t.filepath, os.O_RDONLY, 0)
	return err
}

// TearDown should be called after stopping input
func (t *FileInputParser) TearDown() error {
	return t.file.Close()
}

// Read returns byte array.
func (t *FileInputParser) Read() ([]byte, error) {
	buf := make([]byte, maxReadBytes)
	n, err := t.file.ReadAt(buf, t.offset)
	if err != nil {
		// we need to check this because if the whole file content fits into the buffer,
		// we will get an EOF error, even though this is expected and not an error
		if err.Error() != "EOF" || n == 0 {
			return []byte{}, err
		}
	}
	t.offset += int64(n)
	return buf[:n], nil
}

// GetWinSize returns WinSize object to represent width and height of terminal.
func (t *FileInputParser) GetWinSize() *WinSize {
	// return a fixed size
	return &WinSize{
		Row: 50,
		Col: 100,
	}
}

var _ ConsoleParser = &FileInputParser{}

// NewFileInputParser returns a ConsoleParser object to read from a file.
func NewFileInputParser(filepath string) *FileInputParser {
	return &FileInputParser{
		filepath: filepath,
	}
}
