package audio

import (
	"bytes"
	_ "embed"
	"io"
	"os"
)

// defaultMp3 is the loop that ships inside the binary, so piper works on a
// machine with no audio files on it.
//
//go:embed assets/default.mp3
var defaultMp3 []byte

// openSound returns the sample to play. An empty path means the embedded loop.
func openSound(path string) (io.ReadCloser, error) {
	if path == "" {
		return readSeekCloser{bytes.NewReader(defaultMp3)}, nil
	}
	return os.Open(path)
}

// readSeekCloser makes the embedded bytes look like a file to the decoder.
// io.NopCloser would not do: go-mp3 type-asserts its reader to io.Seeker, and
// looping is a seek back to the start.
type readSeekCloser struct {
	*bytes.Reader
}

func (readSeekCloser) Close() error { return nil }
