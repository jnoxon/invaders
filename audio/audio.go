// Package audio generates Space Invaders sound effects procedurally.
//
// The playback backend is selected at build time: under js/wasm it drives
// the Web Audio API (audio/js.go); everywhere else it is a no-op so the
// package stays testable in plain `go test` (audio/naive.go).
package audio

import "time"

// Tone is one oscillator note: frequency in Hz, duration, wave type and
// peak gain.
type Tone struct {
	Freq float64
	Dur  time.Duration
	Wave string
	Gain float64
}

const (
	WaveSine   = "sine"
	WaveSquare = "square"
	WaveTri    = "triangle"
)

var (
	// FireTone is the player shot: short high blip.
	FireTone = Tone{Freq: 440, Dur: 50 * time.Millisecond, Wave: WaveSquare, Gain: 0.15}
	// KillTone is the invader-killed thud.
	KillTone = Tone{Freq: 110, Dur: 100 * time.Millisecond, Wave: WaveTri, Gain: 0.2}
	// DingTone is the UFO-killed ding.
	DingTone = Tone{Freq: 1318.5, Dur: 150 * time.Millisecond, Wave: WaveTri, Gain: 0.15}

	// HitNoise* is the player-hit explosion (white noise).
	HitNoiseDur  = 200 * time.Millisecond
	HitNoiseGain = 0.25

	// MarchTones is the 4-beat invader march, lowest beat first.
	MarchTones = []Tone{
		{Freq: 55, Dur: 80 * time.Millisecond, Wave: WaveTri, Gain: 0.3},
		{Freq: 65, Dur: 80 * time.Millisecond, Wave: WaveTri, Gain: 0.3},
		{Freq: 75, Dur: 80 * time.Millisecond, Wave: WaveTri, Gain: 0.3},
		{Freq: 85, Dur: 80 * time.Millisecond, Wave: WaveTri, Gain: 0.3},
	}

	// GameOverTones is the descending game-over sequence.
	GameOverTones = []Tone{
		{Freq: 440, Dur: 150 * time.Millisecond, Wave: WaveTri, Gain: 0.2},
		{Freq: 330, Dur: 150 * time.Millisecond, Wave: WaveTri, Gain: 0.2},
		{Freq: 220, Dur: 150 * time.Millisecond, Wave: WaveTri, Gain: 0.2},
		{Freq: 110, Dur: 150 * time.Millisecond, Wave: WaveTri, Gain: 0.2},
	}

	// Warble* describes the UFO warble: a carrier whose frequency is
	// modulated by an LFO.
	WarbleFreq    = 600.0
	WarbleLFOFreq = 8.0
	WarbleDepth   = 200.0
	WarbleGain    = 0.12
)

// backend is the platform playback implementation.
type backend interface {
	resume()
	tone(t Tone, at time.Duration)
	noise(dur time.Duration, gain float64)
	warbleStart()
	warbleStop()
}

// Audio plays procedural sound effects. All methods are safe on a nil
// receiver and while not enabled.
type Audio struct {
	// Enabled gates playback; set by Enable after a user gesture.
	Enabled bool
	// Muted silences everything while true.
	Muted bool
	beat  int
	b     backend
}

// NewAudio returns a ready (but disabled) Audio.
func NewAudio() *Audio {
	return &Audio{b: newBackend()}
}

// Enable turns playback on and resumes the audio context. Browsers only
// allow this after a user gesture, so call it on the first keypress.
func (a *Audio) Enable() {
	if a == nil || a.Enabled {
		return
	}
	a.Enabled = true
	a.b.resume()
}

// ToggleMute flips the mute flag.
func (a *Audio) ToggleMute() {
	if a != nil {
		a.Muted = !a.Muted
	}
}

func (a *Audio) on() bool {
	return a != nil && a.Enabled && !a.Muted
}

// PlayFire plays the player shot.
func (a *Audio) PlayFire() {
	if a.on() {
		a.b.tone(FireTone, 0)
	}
}

// PlayInvaderKill plays the invader-killed thud.
func (a *Audio) PlayInvaderKill() {
	if a.on() {
		a.b.tone(KillTone, 0)
	}
}

// PlayPlayerHit plays the explosion noise.
func (a *Audio) PlayPlayerHit() {
	if a.on() {
		a.b.noise(HitNoiseDur, HitNoiseGain)
	}
}

// PlayUFOStart begins the UFO warble.
func (a *Audio) PlayUFOStart() {
	if a.on() {
		a.b.warbleStart()
	}
}

// PlayUFOEnd stops the UFO warble. It runs even while disabled or muted
// so a started warble is never left dangling.
func (a *Audio) PlayUFOEnd() {
	if a == nil {
		return
	}
	a.b.warbleStop()
}

// PlayUFOHit plays the UFO-killed ding.
func (a *Audio) PlayUFOHit() {
	if a.on() {
		a.b.tone(DingTone, 0)
	}
}

// PlayMarch plays the next note of the 4-beat march pattern.
func (a *Audio) PlayMarch() {
	if !a.on() {
		return
	}
	t := MarchTones[a.beat]
	a.beat = (a.beat + 1) % len(MarchTones)
	a.b.tone(t, 0)
}

// PlayGameOver plays the descending tone sequence.
func (a *Audio) PlayGameOver() {
	if !a.on() {
		return
	}
	for i, t := range GameOverTones {
		a.b.tone(t, time.Duration(i)*t.Dur)
	}
}
