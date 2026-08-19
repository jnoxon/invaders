# Phase 7: Audio, Performance & Final Polish

## Goal

Add procedural sound effects via Web Audio API, optimize rendering performance, add final game feel details (screen flash on player death, UFO score popup), and ensure the complete game is polished and production-ready.

## Context

- Phases 1–6 complete: full game renders with sprites, all mechanics work
- This is the final phase: audio, performance, game feel, and final QA
- After this phase, the game should feel like a complete, playable arcade experience

## Deliverables

### Files to create

1. **`audio/audio.go`**
   - `type Audio struct`:
     - `ctx js.Value` (AudioContext)
     - `enabled bool`
   - `func NewAudio() *Audio` — create AudioContext (must be created after user gesture)
   - `func (a *Audio) Enable()` — resume AudioContext (call on first Enter press)
   - `func (a *Audio) PlayFire()` — square wave, 440Hz, 50ms, quick decay
   - `func (a *Audio) PlayInvaderKill()` — triangle wave, 110Hz, 100ms
   - `func (a *Audio) PlayPlayerHit()` — white noise, 200ms, with decay
   - `func (a *Audio) PlayUFOStart()` — warbling tone (oscillator with LFO on frequency)
   - `func (a *Audio) PlayUFOEnd()` — stop warble
   - `func (a *Audio) PlayUFOHit()` — high-pitched ding
   - `func (a *Audio) PlayMarch(step int)` — 4-beat bass pattern:
     - Beat 0: 55Hz (lowest)
     - Beat 1: 65Hz
     - Beat 2: 75Hz
     - Beat 3: 85Hz
     - Each beat: triangle wave, 80ms
     - Called on each invader step, cycle through 4 beats
   - `func (a *Audio) PlayGameOver()` — descending tone sequence
   - All sounds use `OscillatorNode` + `GainNode` (procedural, no files)
   - Keep volumes low (0.1-0.3 gain)
   - Wrap all JS calls in `if a == nil || !a.enabled { return }` guards

2. **`audio/audio_test.go`**
   - Test that methods don't panic when AudioContext is nil (not in browser)
   - Test enable/disable state
   - Test that sound parameters are correct (frequency, duration) via struct inspection
   - Note: actual audio playback can't be tested without a browser

### Files to modify

3. **`main.go`**
   - Create `audio.NewAudio()`
   - On first Enter keypress: call `audio.Enable()` (user gesture requirement)
   - Hook audio calls into game events:
     - Player fires → `audio.PlayFire()`
     - Invader killed → `audio.PlayInvaderKill()`
     - Player hit → `audio.PlayPlayerHit()`
     - UFO appears → `audio.PlayUFOStart()`
     - UFO disappears/killed → `audio.PlayUFOEnd()` + `audio.PlayUFOHit()` if killed
     - Invader step → `audio.PlayMarch(stepIndex)`
     - Game over → `audio.PlayGameOver()`
   - To detect events for audio: add a `Events []GameEvent` slice to Game that's populated during Tick():
     - `type GameEvent int`: `EventFire`, `EventInvaderKilled`, `EventPlayerHit`, `EventUFOAppear`, `EventUFODisappear`, `EventUFOKilled`, `EventMarch`, `EventGameOver`, `EventLevelStart`
     - `main.go` reads `g.Events` after Tick, plays sounds, then `g.Events = nil`
   - Performance: if fillRect approach is too slow, switch to ImageData:
     - Create a `*image.RGBA` 256×224 buffer
     - Each frame: clear buffer, draw all sprites into buffer (Go-side pixel manipulation)
     - Upload: create `ImageData` from buffer, `ctx.Call("putImageData", imageData, 0, 0)`
     - This is ONE canvas call per frame instead of thousands

4. **`game/game.go`**
   - Add `Events []GameEvent` field
   - Add `type GameEvent int` with constants
   - Populate events during Tick():
     - When player fires: `append(g.Events, EventFire)`
     - When invader killed: `append(g.Events, EventInvaderKilled)`
     - When player hit: `append(g.Events, EventPlayerHit)`
     - When UFO spawns: `append(g.Events, EventUFOAppear)`
     - When UFO leaves/killed: `append(g.Events, EventUFODisappear)` / `EventUFOKilled`
     - On each invader step: `append(g.Events, EventMarch)`
     - On game over: `append(g.Events, EventGameOver)`
     - On level start: `append(g.Events, EventLevelStart)`
   - Add `Flash int` field (frames of white screen flash when player dies, ~6 frames)

5. **`render/render.go`**
   - If using ImageData approach:
     - `type Renderer struct { img *image.RGBA; ... }`
     - `func (r *Renderer) Render(g *game.Game)`:
       - Clear image buffer to black
       - Draw all entities by setting pixels in the buffer
       - Upload to canvas via putImageData
     - `func (r *Renderer) drawSpriteBuf(s Sprite, x, y int)` — write pixels to image buffer
   - Add screen flash: if `g.Flash > 0`, fill entire buffer white before uploading
   - Add UFO score popup: when UFO is killed, show points at UFO position for ~30 frames
     - Add `ScorePopup struct { X, Y, Points int; Timer int }` to Game
   - Add blinking "PRESS ENTER" (toggle every 30 frames using g.Frame)

6. **`index.html`**
   - Add audio unlock: on first click/keypress, the AudioContext can start
   - Add "M" key to mute/unmute (optional, nice to have)
   - Ensure `e.preventDefault()` on all game keys
   - Add a subtle CSS animation for the "Loading..." text (pulse)
   - Add meta viewport tag for mobile (even though it's a keyboard game)
   - Add page title: "Space Invaders"
   - Add favicon: inline SVG data URI (simple invader pixel art)

7. **`game/game.go`** (final tweaks)
   - Add `ScorePopup` struct and logic:
     - When UFO killed: set popup at UFO position with points, timer=30
     - Each tick: decrement timer, remove when 0
   - Verify all edge cases:
     - Player can't move off-screen
     - Bullets don't accumulate (slice compaction works)
     - Game resets completely on restart (all entities fresh)
     - High score persists across games (in session)

### Final QA Checklist (verify in browser)

8. **Create `prompts/07-checklist.md`** (or verify inline):
   - [ ] 60 FPS with full screen of entities (check with browser dev tools)
   - [ ] Audio plays on all events (fire, kill, hit, UFO, march)
   - [ ] Audio doesn't start until user gesture (Enter)
   - [ ] Screen flashes white on player death
   - [ ] UFO score popup appears and fades
   - [ ] Start screen: title, score table, blinking text, high score
   - [ ] Game over: "GAME OVER", final score, restart works
   - [ ] Level transition: 2s pause, then next level starts higher
   - [ ] Invaders get faster as you kill them
   - [ ] Barricades erode realistically
   - [ ] Player has 1 bullet at a time
   - [ ] Enemy bullets max 4 on screen
   - [ ] Player invulnerability after respawn (1s)
   - [ ] High score persists between games in session
   - [ ] Resize window: canvas maintains aspect ratio, integer scale
   - [ ] No browser console errors
   - [ ] No memory leaks (play 10 levels, check memory)
   - [ ] Works in Chrome, Firefox, Safari

### Acceptance Criteria

- [ ] `go build ./...` succeeds
- [ ] `go test ./...` passes all tests
- [ ] `go test ./game/ -cover` reports >85%
- [ ] `./build.sh && ./serve.sh` → complete playable game in browser
- [ ] All sounds play procedurally (no audio files)
- [ ] 60 FPS sustained (no frame drops with full entity count)
- [ ] Game feel: responsive input, satisfying audio, visual feedback (flash, popup)
- [ ] No external dependencies
- [ ] Single `index.html` (no external CSS/JS files)
- [ ] Code is gofmt-clean
- [ ] All magic numbers are constants
- [ ] Total project size < 1MB (wasm binary + html)

## Technical Notes

- Web Audio API: `new AudioContext()` must be created/resumed after a user gesture (click or keypress). Gate all audio behind an `enabled` flag set on first Enter.
- OscillatorNode for tones: create oscillator → connect to gain → connect to destination → start → stop after duration. For each sound, create fresh nodes (they're single-use).
- White noise: create an `AudioBuffer` filled with random values, play through a `BufferSourceNode` → `GainNode` (with decay).
- March sound: keep a `marchBeat int` counter (0-3), increment on each invader step, call `PlayMarch(marchBeat % 4)`.
- ImageData approach for rendering:
  ```go
  img := image.NewRGBA(image.Rect(0, 0, 256, 224))
  // ... draw pixels into img ...
  // Upload:
  buf := &img.Pix // []uint8, RGBA format
  // Create JS Uint8ClampedArray from buf, create ImageData, putImageData
  ```
  - Go → JS memory transfer: use `js.CopyBytesToJS` or `syscall/js` array creation
  - This is the key performance optimization: one putImageData call vs thousands of fillRect calls
- Screen flash: set `g.Flash = 6` on player death. In render, if `g.Flash > 0`, fill buffer white. Decrement flash in Tick().
- Score popup: `g.ScorePopup = ScorePopup{X: ufo.X, Y: ufo.Y, Points: ufo.Points, Timer: 30}`. Render as text at that position. Decrement timer in Tick().

## Verification

```bash
go build ./...
go test ./... -v -cover
./build.sh
./serve.sh &
# Browser: play a full game through multiple levels
# Check: audio, performance (dev tools), all game states, edge cases
# Compare to original: https://en.wikipedia.org/wiki/Space_Invaders
```
