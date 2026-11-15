package audio

import (
	"os"

	"github.com/gopxl/beep/v2/mp3"
)

// WriteDefaultSample writes the sample embedded in the binary to path. It
// exists so a test can put a real, decodable mp3 on disk without shipping a
// separate fixture that could drift from the embedded one.
func WriteDefaultSample(path string) error {
	return os.WriteFile(path, defaultMp3, 0o644)
}

// CanDecode reports whether the player would be able to decode path, without
// opening the audio device. A config that passes `piper sound` is one that
// Start will not reject for the file's sake.
func CanDecode(path string) error {
	reader, err := openSound(path)
	if err != nil {
		return err
	}
	defer reader.Close()

	song, _, err := mp3.Decode(reader)
	if err != nil {
		return err
	}
	return song.Close()
}
