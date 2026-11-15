// Package cmd defines the piper command tree: the root command, which wraps
// another command, and the subcommands that manage piper itself.
package cmd

import (
	"errors"
	"fmt"
	"math"
	"os"
	"os/exec"
	"os/signal"
	"sync/atomic"
	"time"

	"github.com/d3uceY/Piper/internal/audio"
	"github.com/d3uceY/Piper/internal/config"
	"github.com/d3uceY/Piper/internal/version"
	"github.com/spf13/cobra"
)

// interruptGrace is how long an interrupted command gets to clean up by itself
// before piper kills it.
const interruptGrace = 2 * time.Second

// durationValue is a flag that reads a duration the same way the config file
// does, so --delay 10 and --delay 10s both mean ten seconds. A plain
// DurationVar would reject the first one.
type durationValue struct{ target *time.Duration }

func (d durationValue) String() string {
	if d.target == nil {
		return ""
	}
	return d.target.String()
}

func (d durationValue) Set(s string) error {
	v, err := config.ParseDuration(s)
	if err != nil {
		return err
	}
	*d.target = v
	return nil
}

// Type is what the flag package reports in help output.
func (d durationValue) Type() string { return "duration" }

// ExitError carries the wrapped command's exit status up to main, which
// repeats it, so `piper npm install` fails exactly when npm fails.
type ExitError struct{ Code int }

func (e *ExitError) Error() string {
	return fmt.Sprintf("command exited with status %d", e.Code)
}

// runFlags are the command-line overrides. Each one only takes effect when it
// was actually set, so the config file is the default for everything else.
type runFlags struct {
	delay   time.Duration
	sound   string
	volume  float64
	fadeIn  time.Duration
	fadeOut time.Duration
	silent  bool
}

// runSettings is one invocation's resolved settings.
type runSettings struct {
	config.Settings
	music bool
}

// newRootCmd builds the root command and attaches every subcommand.
func newRootCmd() *cobra.Command {
	var f runFlags

	root := &cobra.Command{
		Use:   "piper <command> [args...]",
		Short: "Run a command, playing music if it drags",
		Long: "piper runs a command and, if it is still going after a short delay,\n" +
			"loops music under it until it finishes.\n\n" +
			"Anything after the command name is passed to it untouched:\n\n" +
			"  piper npm install --save\n\n" +
			"piper's own flags go before the command name. A duration takes an\n" +
			"optional unit, so --delay 10 waits ten seconds.",
		Version:       version.Short(),
		Args:          cobra.ArbitraryArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			if err := f.validate(); err != nil {
				return err
			}
			s := f.resolve(cmd)
			return run(cmd, args, s)
		},
	}

	// Terse on the flag, detailed on `piper version`.
	root.SetVersionTemplate("piper {{.Version}}\n")

	flags := root.Flags()
	f.delay = 3 * time.Second
	f.fadeIn = 1500 * time.Millisecond
	f.fadeOut = 800 * time.Millisecond
	flags.Var(durationValue{&f.delay}, "delay", "how long the command may run before the music starts; 10 means 10s (config: delay)")
	flags.StringVar(&f.sound, "sound", "", "audio file to loop, empty for the sample built into piper (config: sound)")
	flags.Float64Var(&f.volume, "volume", 0.5, "music volume from 0 to 1 (config: volume)")
	flags.Var(durationValue{&f.fadeIn}, "fade-in", "how long the fade in takes; 10 means 10s (config: fade_in)")
	flags.Var(durationValue{&f.fadeOut}, "fade-out", "how long the fade out takes; 10 means 10s (config: fade_out)")
	flags.BoolVar(&f.silent, "silent", false, "never play music, whatever the config says")

	// Stop parsing at the first positional argument, so the wrapped command's
	// own flags reach it instead of failing as unknown flags of piper's.
	flags.SetInterspersed(false)

	root.AddCommand(
		newVersionCmd(),
		newConfigCmd(),
		newSoundCmd(),
	)

	return root
}

// Execute runs the CLI. It returns nil on success and a non-nil error on
// failure; main() is responsible for the exit code and the message.
func Execute() error {
	return newRootCmd().Execute()
}

// validate refuses a flag value that would make playback misbehave, before the
// wrapped command is started. A duration that will not parse is already caught
// when the flag is set; this is for the values the flag package accepts happily.
func (f runFlags) validate() error {
	if math.IsNaN(f.volume) || math.IsInf(f.volume, 0) {
		return fmt.Errorf("--volume: %v is not a number between 0 and 1", f.volume)
	}
	if f.volume < 0 || f.volume > 1 {
		return fmt.Errorf("--volume: %v is outside 0 to 1", f.volume)
	}
	return nil
}

// resolve layers the config file, then the flags that were actually set, into
// the settings for this run. A config piper cannot read or understand is
// reported and skipped rather than stopping the command the user asked for.
func (f runFlags) resolve(cmd *cobra.Command) runSettings {
	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "piper: using built-in settings: %v\n", err)
	}

	s, err := cfg.Settings()
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "piper: config: %v (using built-in settings)\n", err)
		s, _ = config.Default().Settings() // the defaults are valid; see the config tests
	}

	flags := cmd.Flags()
	if flags.Changed("delay") {
		s.Delay = f.delay
	}
	if flags.Changed("sound") {
		s.Sound = f.sound
	}
	if flags.Changed("volume") {
		s.Volume = f.volume
	}
	if flags.Changed("fade-in") {
		s.FadeIn = f.fadeIn
	}
	if flags.Changed("fade-out") {
		s.FadeOut = f.fadeOut
	}

	return runSettings{Settings: s, music: !f.silent}
}

func loadConfig() (config.Config, error) {
	path, err := config.Path()
	if err != nil {
		return config.Default(), err
	}
	return config.Load(path)
}

// run executes the wrapped command and owns piper's music for its lifetime.
func run(cmd *cobra.Command, args []string, s runSettings) error {
	proc := exec.Command(args[0], args[1:]...)
	proc.Stdin = os.Stdin
	proc.Stdout = os.Stdout
	proc.Stderr = os.Stderr

	if err := proc.Start(); err != nil {
		var missing *exec.Error
		if errors.As(err, &missing) {
			return fmt.Errorf("%s: command not found", args[0])
		}
		return err
	}

	var waitErr error
	finished := make(chan struct{})
	go func() {
		waitErr = proc.Wait()
		close(finished)
	}()

	interrupted, stopWatching := watchInterrupt(proc, finished)
	defer stopWatching()

	music := startIfStillRunning(cmd, finished, s)
	<-finished
	if music != nil {
		music.Stop()
	}

	if interrupted.Load() {
		return &ExitError{Code: 130}
	}
	return exitStatus(waitErr)
}

// startIfStillRunning waits out the delay and starts the music only if the
// command is still going. It returns nil when the music is off, when the
// command finished in time, or when playback was not possible.
func startIfStillRunning(cmd *cobra.Command, finished <-chan struct{}, s runSettings) *audio.Player {
	if !s.music {
		return nil
	}

	if s.Delay > 0 {
		timer := time.NewTimer(s.Delay)
		defer timer.Stop()

		select {
		case <-finished:
			return nil
		case <-timer.C:
		}
	}

	player, err := audio.Start(audio.Options{
		Sound:   s.Sound,
		Volume:  s.Volume,
		FadeIn:  s.FadeIn,
		FadeOut: s.FadeOut,
	})
	if err != nil {
		// The music is a nicety: a missing file or a busy audio device must
		// not fail the command the user actually asked for.
		fmt.Fprintf(cmd.ErrOrStderr(), "piper: no music: %v\n", err)
		return nil
	}
	return player
}

// exitStatus turns the child's error into piper's: nil for success, a matching
// exit code for a command that failed.
func exitStatus(err error) error {
	if err == nil {
		return nil
	}

	var exit *exec.ExitError
	if errors.As(err, &exit) {
		code := exit.ExitCode()
		if code <= 0 { // killed, or a start failure that never reached a status
			code = 1
		}
		return &ExitError{Code: code}
	}
	return err
}

// watchInterrupt gives the command a chance to handle Ctrl+C before it is
// killed, so piper is still around to fade the music out and to repeat the
// child's exit status. The returned func stops watching.
func watchInterrupt(proc *exec.Cmd, finished <-chan struct{}) (*atomic.Bool, func()) {
	var interrupted atomic.Bool

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	go func() {
		for range signals {
			interrupted.Store(true)
			select {
			case <-finished:
				// The child took the signal and is on its way out.
			case <-time.After(interruptGrace):
				_ = proc.Process.Kill()
			}
		}
	}()

	return &interrupted, func() {
		signal.Stop(signals)
		close(signals)
	}
}
