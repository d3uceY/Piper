package audio

import (
	"math"
	"sync/atomic"
	"time"

	"github.com/gopxl/beep/v2"
)

// track is what piper plays: the looping sample with two gains applied as the
// audio thread reads it. volume comes from the config and never changes;
// envelope is the fade, moved from another goroutine, so it is atomic rather
// than a plain float.
type track struct {
	src      beep.Streamer
	volume   float64
	envelope atomic.Uint64 // math.Float64bits, 0 is silent and 1 is full
	stopped  atomic.Bool
}

func newTrack(src beep.Streamer, volume float64) *track {
	return &track{src: src, volume: volume}
}

func (t *track) Stream(samples [][2]float64) (n int, ok bool) {
	// Reporting false is how a streamer leaves the mixer: once piper is done
	// the speaker stops pulling samples, so the decoder is never read again
	// and the file can be closed safely.
	if t.stopped.Load() {
		return 0, false
	}

	n, ok = t.src.Stream(samples)

	gain := t.volume * t.envelopeValue()
	if gain != 1 {
		for i := range samples[:n] {
			samples[i][0] *= gain
			samples[i][1] *= gain
		}
	}
	return n, ok
}

func (t *track) Err() error {
	return t.src.Err()
}

func (t *track) envelopeValue() float64 {
	return math.Float64frombits(t.envelope.Load())
}

func (t *track) setEnvelope(v float64) {
	t.envelope.Store(math.Float64bits(v))
}

// fadeTo moves the envelope to target over d, in steps small enough to sound
// smooth, and returns when the fade is over. A non-positive d jumps instead.
func (t *track) fadeTo(target float64, d time.Duration) {
	from := t.envelopeValue()
	if d <= 0 || from == target {
		t.setEnvelope(target)
		return
	}

	// One step per millisecond sounds smooth and costs nothing, but there is
	// no point in more than fifty steps for a long fade.
	steps := int(d / time.Millisecond)
	if steps < 1 {
		steps = 1
	}
	if steps > 50 {
		steps = 50
	}
	step := d / time.Duration(steps)

	ticker := time.NewTicker(step)
	defer ticker.Stop()

	for i := 1; i <= steps; i++ {
		<-ticker.C
		t.setEnvelope(from + (target-from)*float64(i)/float64(steps))
	}
}
