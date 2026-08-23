//go:build !js || !wasm

package audio

import "time"

// nullBackend discards everything; it keeps the package testable outside
// the browser where there is no Web Audio API.
type nullBackend struct{}

func (nullBackend) resume()                      {}
func (nullBackend) tone(Tone, time.Duration)     {}
func (nullBackend) noise(time.Duration, float64) {}
func (nullBackend) warbleStart()                 {}
func (nullBackend) warbleStop()                  {}
func (nullBackend) state() string                { return "" }

func newBackend() backend { return nullBackend{} }
