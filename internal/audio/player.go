// Package audio plays piper's music: it loops a sample through the system's
// audio device and fades it in and out.
package audio

import (
	"sync"
	"time"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/mp3"
	"github.com/gopxl/beep/v2/speaker"
)

// sampleRate is what the device is opened at. Samples in another rate are
// resampled to it, so the device is opened once at a rate known up front.
const sampleRate = beep.SampleRate(44100)

// Options describes one playback: which file, how loud, and how long the
// fades take.
type Options struct {
	// Sound is the file to loop. Empty plays the sample built into the binary.
	Sound   string
	Volume  float64
	FadeIn  time.Duration
	FadeOut time.Duration
}

// Player is a sample looping in the background. Stop fades it out.
type Player struct {
	track   *track
	drained chan struct{} // closed when the fade-in goroutine returns
	file    beep.StreamSeekCloser
	fadeOut time.Duration
	stop    sync.Once
}

// Start opens the sample, starts looping it and fades it in. It returns as soon
// as the device is playing; the fade itself runs in the background so the
// caller can get back to watching its command.
func Start(o Options) (*Player, error) {
	reader, err := openSound(o.Sound)
	if err != nil {
		return nil, err
	}

	song, format, err := mp3.Decode(reader)
	if err != nil {
		reader.Close()
		return nil, err
	}

	// Loop2 with no options repeats forever; it needs the source to report a
	// length, which is why the reader has to be seekable.
	looped, err := beep.Loop2(song)
	if err != nil {
		song.Close()
		return nil, err
	}

	var streamer beep.Streamer = looped
	if format.SampleRate != sampleRate {
		streamer = beep.Resample(4, format.SampleRate, sampleRate, streamer)
	}

	t := newTrack(streamer, o.Volume)

	if err := openSpeaker(); err != nil {
		song.Close()
		return nil, err
	}

	p := &Player{
		track:   t,
		drained: make(chan struct{}),
		file:    song,
		fadeOut: o.FadeOut,
	}

	speaker.Play(t)

	go func() {
		defer close(p.drained)
		t.fadeTo(1, o.FadeIn)
	}()

	return p, nil
}

// Stop fades the music out and silences it. It is safe to call more than once.
func (p *Player) Stop() {
	p.stop.Do(func() {
		// Let a fade-in that is still running finish first, so two ramps are
		// not driving the envelope at the same time.
		<-p.drained
		p.track.fadeTo(0, p.fadeOut)
		p.track.stopped.Store(true)

		// Clear waits for any in-flight Stream call to finish, so the decoder
		// is not read while its file is closed.
		speaker.Clear()
		p.file.Close()
	})
}

var (
	speakerOnce sync.Once
	speakerErr  error
)

// openSpeaker initialises the audio device, which beep refuses to do twice.
func openSpeaker() error {
	speakerOnce.Do(func() {
		speakerErr = speaker.Init(sampleRate, sampleRate.N(time.Second/10))
	})
	return speakerErr
}
