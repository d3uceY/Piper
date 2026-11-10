package audio

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/mp3"
)

// fullScale is a streamer that hands out samples at 1.0, so any scaling the
// track applies is visible in the output.
func fullScale() beep.Streamer {
	return beep.StreamerFunc(func(samples [][2]float64) (int, bool) {
		for i := range samples {
			samples[i] = [2]float64{1, 1}
		}
		return len(samples), true
	})
}

func pull(t *testing.T, s beep.Streamer) ([][2]float64, bool) {
	t.Helper()
	samples := make([][2]float64, 8)
	n, ok := s.Stream(samples)
	return samples[:n], ok
}

func TestTrackAppliesVolumeAndEnvelope(t *testing.T) {
	tr := newTrack(fullScale(), 0.5)
	tr.setEnvelope(1)

	samples, ok := pull(t, tr)
	if !ok {
		t.Fatal("streamer reported it was done")
	}
	if got := samples[0][0]; got != 0.5 {
		t.Errorf("sample = %v, want 0.5 (volume 0.5 at full envelope)", got)
	}

	tr.setEnvelope(0.25)
	samples, _ = pull(t, tr)
	if got := samples[0][1]; got != 0.125 {
		t.Errorf("sample = %v, want 0.125 (volume 0.5 times envelope 0.25)", got)
	}
}

func TestTrackStartsSilent(t *testing.T) {
	tr := newTrack(fullScale(), 1)

	samples, _ := pull(t, tr)
	if got := samples[0][0]; got != 0 {
		t.Errorf("sample = %v before the fade in, want silence", got)
	}
}

// Returning false is how the track leaves the mixer, which is what makes it
// safe to close the decoder file afterwards.
func TestStoppedTrackReportsDone(t *testing.T) {
	tr := newTrack(fullScale(), 1)
	tr.setEnvelope(1)

	tr.stopped.Store(true)

	if n, ok := tr.Stream(make([][2]float64, 8)); ok || n != 0 {
		t.Errorf("Stream after stop = (%d, %v), want (0, false)", n, ok)
	}
}

func TestFadeToRampsUpToTheTarget(t *testing.T) {
	tr := newTrack(fullScale(), 1)

	start := time.Now()
	tr.fadeTo(1, 40*time.Millisecond)
	elapsed := time.Since(start)

	if got := tr.envelopeValue(); got != 1 {
		t.Errorf("envelope = %v after the fade, want 1", got)
	}
	if elapsed < 30*time.Millisecond {
		t.Errorf("fade took %v, want about 40ms", elapsed)
	}
}

func TestFadeToZeroDurationJumps(t *testing.T) {
	tr := newTrack(fullScale(), 1)

	tr.fadeTo(1, 0)

	if got := tr.envelopeValue(); got != 1 {
		t.Errorf("envelope = %v, want 1 without waiting", got)
	}
}

func TestFadeDownFromFullVolume(t *testing.T) {
	tr := newTrack(fullScale(), 1)
	tr.setEnvelope(1)

	tr.fadeTo(0, 30*time.Millisecond)

	if got := tr.envelopeValue(); got != 0 {
		t.Errorf("envelope = %v after the fade out, want 0", got)
	}
}

func TestStartReportsAMissingFile(t *testing.T) {
	if _, err := Start(Options{Sound: "does-not-exist.mp3"}); err == nil {
		t.Fatal("expected an error for a missing sample")
	}
}

// The embedded sample has to be a real, seekable mp3 in the binary: looping is
// a seek back to the start, and go-mp3 refuses to seek a reader it cannot seek.
func TestEmbeddedSampleDecodesAndSeeks(t *testing.T) {
	if len(defaultMp3) == 0 {
		t.Fatal("the embedded sample is empty")
	}

	rc, err := openSound("")
	if err != nil {
		t.Fatalf("open embedded: %v", err)
	}
	defer rc.Close()

	song, format, err := mp3.Decode(rc)
	if err != nil {
		t.Fatalf("the embedded sample is not decodable: %v", err)
	}
	defer song.Close()

	if format.SampleRate == 0 || format.NumChannels != 2 {
		t.Errorf("format = %v, want stereo at a real sample rate", format)
	}
	if song.Len() <= 0 {
		t.Fatal("the embedded sample has no length, so it cannot loop")
	}

	if err := song.Seek(0); err != nil {
		t.Fatalf("seek to start (what looping does): %v", err)
	}
}

func TestOpenSoundReadsAFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "loop.mp3")
	if err := os.WriteFile(path, defaultMp3, 0o644); err != nil {
		t.Fatal(err)
	}

	rc, err := openSound(path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer rc.Close()

	if _, _, err := mp3.Decode(rc); err != nil {
		t.Fatalf("decode: %v", err)
	}
}
