package audio

import (
	"testing"
	"time"
)

func TestNewAudio(t *testing.T) {
	a := NewAudio()
	if a == nil {
		t.Fatal("NewAudio returned nil")
	}
	if a.Enabled {
		t.Error("new Audio should start disabled")
	}
	if a.Muted {
		t.Error("new Audio should start unmuted")
	}
}

func TestMethodsSafeWhenDisabled(t *testing.T) {
	a := NewAudio()
	a.PlayFire()
	a.PlayInvaderKill()
	a.PlayPlayerHit()
	a.PlayUFOStart()
	a.PlayUFOEnd()
	a.PlayUFOHit()
	a.PlayMarch()
	a.PlayGameOver()
}

func TestNilAudioSafe(t *testing.T) {
	var a *Audio
	a.Enable()
	a.ToggleMute()
	a.PlayFire()
	a.PlayInvaderKill()
	a.PlayPlayerHit()
	a.PlayUFOStart()
	a.PlayUFOEnd()
	a.PlayUFOHit()
	a.PlayMarch()
	a.PlayGameOver()
}

func TestEnableIsIdempotent(t *testing.T) {
	a := NewAudio()
	a.Enable()
	if !a.Enabled {
		t.Fatal("Enable did not set Enabled")
	}
	a.Enable()
	if !a.Enabled {
		t.Fatal("second Enable cleared Enabled")
	}
}

func TestToggleMute(t *testing.T) {
	a := NewAudio()
	a.Enable()
	a.ToggleMute()
	if !a.Muted {
		t.Error("ToggleMute did not mute")
	}
	a.ToggleMute()
	if a.Muted {
		t.Error("second ToggleMute did not unmute")
	}
}

func TestMuteStopsMarchCycling(t *testing.T) {
	a := NewAudio()
	a.Enable()
	a.PlayMarch()
	if a.beat != 1 {
		t.Fatalf("beat = %d, want 1", a.beat)
	}
	a.ToggleMute()
	a.PlayMarch()
	if a.beat != 1 {
		t.Fatalf("muted PlayMarch advanced beat to %d, want 1", a.beat)
	}
}

func TestSoundParameters(t *testing.T) {
	if FireTone.Freq != 440 || FireTone.Dur != 50*time.Millisecond || FireTone.Wave != WaveSquare {
		t.Errorf("FireTone = %+v, want 440Hz 50ms square", FireTone)
	}
	if KillTone.Freq != 110 || KillTone.Dur != 100*time.Millisecond || KillTone.Wave != WaveTri {
		t.Errorf("KillTone = %+v, want 110Hz 100ms triangle", KillTone)
	}
	if HitNoiseDur != 200*time.Millisecond {
		t.Errorf("HitNoiseDur = %v, want 200ms", HitNoiseDur)
	}
	if DingTone.Freq <= 1000 || DingTone.Dur <= 0 {
		t.Errorf("DingTone = %+v, want a high-pitched ding", DingTone)
	}

	wantMarch := []float64{55, 65, 75, 85}
	if len(MarchTones) != 4 {
		t.Fatalf("MarchTones has %d beats, want 4", len(MarchTones))
	}
	for i, want := range wantMarch {
		if MarchTones[i].Freq != want || MarchTones[i].Dur != 80*time.Millisecond || MarchTones[i].Wave != WaveTri {
			t.Errorf("MarchTones[%d] = %+v, want %gHz 80ms triangle", i, MarchTones[i], want)
		}
	}

	if len(GameOverTones) < 2 {
		t.Fatal("GameOverTones too short")
	}
	for i := 1; i < len(GameOverTones); i++ {
		if GameOverTones[i].Freq >= GameOverTones[i-1].Freq {
			t.Errorf("GameOverTones[%d].Freq = %g, want < %g (descending)", i, GameOverTones[i].Freq, GameOverTones[i-1].Freq)
		}
	}
}

func TestMarchCyclesBeats(t *testing.T) {
	a := NewAudio()
	a.Enable()
	for i := 0; i < len(MarchTones); i++ {
		a.PlayMarch()
	}
	if a.beat != 0 {
		t.Errorf("beat after full cycle = %d, want 0", a.beat)
	}
	a.PlayMarch()
	if a.beat != 1 {
		t.Errorf("beat after cycle+1 = %d, want 1", a.beat)
	}
}

// fakeBackend records every call so tests can verify routing and params.
type fakeBackend struct {
	tones     []Tone
	offsets   []time.Duration
	noiseN    int
	noiseDur  time.Duration
	noiseGain float64
	warbleOn  int
	warbleOff int
	resumes   int
}

func (f *fakeBackend) resume() { f.resumes++ }
func (f *fakeBackend) tone(t Tone, at time.Duration) {
	f.tones = append(f.tones, t)
	f.offsets = append(f.offsets, at)
}
func (f *fakeBackend) noise(d time.Duration, g float64) { f.noiseN++; f.noiseDur, f.noiseGain = d, g }
func (f *fakeBackend) warbleStart()                     { f.warbleOn++ }
func (f *fakeBackend) warbleStop()                      { f.warbleOff++ }
func (f *fakeBackend) state() string                    { return "running" }

func TestPlayMethodsRouteToBackend(t *testing.T) {
	b := &fakeBackend{}
	a := NewAudio()
	a.b = b
	a.Enable()
	if b.resumes != 1 {
		t.Fatalf("resumes = %d, want 1", b.resumes)
	}

	a.PlayFire()
	a.PlayInvaderKill()
	a.PlayPlayerHit()
	a.PlayUFOStart()
	a.PlayUFOHit()
	a.PlayMarch()
	a.PlayGameOver()

	wantTones := []Tone{FireTone, KillTone, DingTone, MarchTones[0],
		GameOverTones[0], GameOverTones[1], GameOverTones[2], GameOverTones[3]}
	if len(b.tones) != len(wantTones) {
		t.Fatalf("tones = %d, want %d", len(b.tones), len(wantTones))
	}
	for i, want := range wantTones {
		if b.tones[i] != want {
			t.Errorf("tone[%d] = %+v, want %+v", i, b.tones[i], want)
		}
	}
	if b.noiseN != 1 || b.noiseDur != HitNoiseDur || b.noiseGain != HitNoiseGain {
		t.Errorf("noise = %d x (%v, %v), want 1 x (%v, %v)",
			b.noiseN, b.noiseDur, b.noiseGain, HitNoiseDur, HitNoiseGain)
	}
	if b.warbleOn != 1 || b.warbleOff != 0 {
		t.Errorf("warble = %d/%d, want 1/0", b.warbleOn, b.warbleOff)
	}
	// GameOver notes are staggered by cumulative duration.
	wantOff := []time.Duration{0, 150 * time.Millisecond, 300 * time.Millisecond, 450 * time.Millisecond}
	for i, w := range wantOff {
		if b.offsets[4+i] != w {
			t.Errorf("gameOver offset[%d] = %v, want %v", i, b.offsets[4+i], w)
		}
	}
}

func TestMuteSilencesButUFOEndStillStops(t *testing.T) {
	b := &fakeBackend{}
	a := NewAudio()
	a.b = b
	a.Enable()
	a.PlayUFOStart()
	a.ToggleMute()
	a.PlayFire()
	a.PlayMarch()
	a.PlayUFOEnd()
	if len(b.tones) != 0 {
		t.Fatalf("muted tones = %v, want none", b.tones)
	}
	if b.warbleOn != 1 || b.warbleOff != 1 {
		t.Fatalf("warble = %d/%d, want 1/1 (end runs while muted)", b.warbleOn, b.warbleOff)
	}
	if b.noiseN != 0 {
		t.Error("muted noise should not fire")
	}
}
