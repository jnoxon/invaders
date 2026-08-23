//go:build js && wasm

package audio

import (
	"math/rand/v2"
	"syscall/js"
	"time"
	"unsafe"
)

// jsBackend drives the Web Audio API. The AudioContext is created on the
// first resume, inside a user gesture: iOS permanently suspends contexts
// created before the page's first interaction.
type jsBackend struct {
	ctx       js.Value
	warbleOsc js.Value
	warbleLfo js.Value
	warbleOn  bool
}

func newBackend() backend {
	return &jsBackend{ctx: js.Undefined()}
}

// create instantiates the AudioContext. Called from a user-gesture path.
func (b *jsBackend) create() {
	ctor := js.Global().Get("AudioContext")
	if ctor.IsUndefined() {
		ctor = js.Global().Get("webkitAudioContext")
	}
	if !ctor.IsUndefined() {
		b.ctx = ctor.New()
	}
}

func (b *jsBackend) ok() bool {
	return !b.ctx.IsUndefined()
}

func (b *jsBackend) resume() {
	if !b.ok() {
		b.create()
	}
	if b.ok() && b.ctx.Get("state").String() == "suspended" {
		b.ctx.Call("resume")
	}
}

func (b *jsBackend) state() string {
	if !b.ok() {
		return ""
	}
	return b.ctx.Get("state").String()
}

// tone schedules a single oscillator note with a linear gain ramp to zero.
// Nodes are single-use and garbage collected after stop.
func (b *jsBackend) tone(t Tone, at time.Duration) {
	if !b.ok() {
		return
	}
	t0 := b.ctx.Get("currentTime").Float() + at.Seconds()
	t1 := t0 + t.Dur.Seconds()
	osc := b.ctx.Call("createOscillator")
	osc.Set("type", t.Wave)
	osc.Get("frequency").Set("value", t.Freq)
	g := b.ctx.Call("createGain")
	p := g.Get("gain")
	p.Set("value", t.Gain)
	p.Call("linearRampToValueAtTime", 0.0001, t1)
	osc.Call("connect", g)
	g.Call("connect", b.ctx.Get("destination"))
	osc.Call("start", t0)
	osc.Call("stop", t1)
}

// noise plays white noise through a decaying gain.
func (b *jsBackend) noise(dur time.Duration, gain float64) {
	if !b.ok() {
		return
	}
	rate := b.ctx.Get("sampleRate").Int()
	n := int(dur.Seconds() * float64(rate))
	if n <= 0 {
		n = 1
	}
	buf := b.ctx.Call("createBuffer", 1, n, rate)
	data := buf.Call("getChannelData", 0)
	samples := make([]float32, n)
	for i := range samples {
		samples[i] = float32(rand.Float64()*2 - 1)
	}
	// CopyBytesToJS requires a Uint8Array dst, so view the channel buffer as
	// bytes and write the float32 samples straight into it.
	u8 := js.Global().Get("Uint8Array").New(data.Get("buffer"))
	js.CopyBytesToJS(u8, unsafe.Slice((*byte)(unsafe.Pointer(unsafe.SliceData(samples))), n*4))

	t0 := b.ctx.Get("currentTime").Float()
	src := b.ctx.Call("createBufferSource")
	src.Set("buffer", buf)
	g := b.ctx.Call("createGain")
	p := g.Get("gain")
	p.Set("value", gain)
	p.Call("linearRampToValueAtTime", 0.0001, t0+dur.Seconds())
	src.Call("connect", g)
	g.Call("connect", b.ctx.Get("destination"))
	src.Call("start", t0)
}

// warbleStart runs a sine carrier whose frequency an LFO modulates.
func (b *jsBackend) warbleStart() {
	if !b.ok() || b.warbleOn {
		return
	}
	ctx := b.ctx
	osc := ctx.Call("createOscillator")
	osc.Set("type", WaveSine)
	osc.Get("frequency").Set("value", WarbleFreq)
	lfo := ctx.Call("createOscillator")
	lfo.Set("type", WaveSine)
	lfo.Get("frequency").Set("value", WarbleLFOFreq)
	depth := ctx.Call("createGain")
	depth.Get("gain").Set("value", WarbleDepth)
	lfo.Call("connect", depth)
	depth.Call("connect", osc.Get("frequency"))
	g := ctx.Call("createGain")
	g.Get("gain").Set("value", WarbleGain)
	osc.Call("connect", g)
	g.Call("connect", ctx.Get("destination"))
	t0 := ctx.Get("currentTime").Float()
	lfo.Call("start", t0)
	osc.Call("start", t0)
	b.warbleOsc = osc
	b.warbleLfo = lfo
	b.warbleOn = true
}

func (b *jsBackend) warbleStop() {
	if !b.ok() || !b.warbleOn {
		return
	}
	t0 := b.ctx.Get("currentTime").Float()
	b.warbleOsc.Call("stop", t0+0.05)
	b.warbleLfo.Call("stop", t0+0.05)
	b.warbleOsc = js.Undefined()
	b.warbleLfo = js.Undefined()
	b.warbleOn = false
}
