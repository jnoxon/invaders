# Phase 3: Player Movement & Input Integration

## Goal

Wire the existing `game.InputState` to real keyboard input via JS interop, and verify the player moves and fires correctly. This phase bridges the pure Go game logic (phase 2) with the browser input system (phase 1 stub).

## Context

- Phases 1 and 2 are complete: WASM loads, game logic exists in `game/` package
- `main.go` currently has a stub `setKey` that just logs
- `game.Game.HandleInput()` and `game.Player.Update()` are already implemented
- This phase makes the stub real and verifies the full input → game logic pipeline

## Deliverables

### Files to modify

1. **`main.go`**
   - Remove the stub `setKey`
   - Implement `setKey(code string, pressed bool)`:
     - Calls `g.HandleInput(code, pressed)` on the game instance
   - Ensure the game loop calls `g.Tick()` every frame (should already work from phase 2)
   - Add a simple debug render: draw player position as a white rectangle on canvas (will be replaced in phase 6)
     - `ctx.Call("fillStyle", "#fff").Call("fillRect", g.Player.X, g.Player.Y, g.Player.W, g.Player.H)`
   - Add debug: draw score as text in top-left
     - `ctx.Call("fillStyle", "#fff").Call("font", "10px monospace").Call("fillText", fmt.Sprintf("Score: %d", g.Score), 4, 12)`

2. **`index.html`**
   - Update input wiring:
     - `keydown`: call `main.setKey(e.code, true)` (e.g., "ArrowLeft", "KeyA", "Space", "Enter", "KeyP")
     - `keyup`: call `main.setKey(e.code, false)`
     - `e.preventDefault()` on handled keys (arrows, space) to prevent page scroll
   - Ensure only one `keydown` event fires per key (browser auto-repeat should be ignored — check `e.repeat` and skip)

3. **`game/game.go`** (minor fixes if needed)
   - Ensure `HandleInput` correctly maps:
     - `"ArrowLeft"`, `"KeyA"` → `InputState.Left`
     - `"ArrowRight"`, `"KeyD"` → `InputState.Right`
     - `"Space"` → `InputState.Fire`
     - `"Enter"` → `JustPressed["Enter"] = true`
     - `"KeyP"` → `JustPressed["P"] = true`
   - Ensure fire is edge-triggered (only fires on press, not held) — use `JustPressed` or track previous state

### New tests

4. **`game/input_test.go`**
   - Test key code mapping (all keys)
   - Test left/right tracking (press/release)
   - Test fire edge-triggering (press fires, holding doesn't re-fire)
   - Test Enter/P just-pressed tracking
   - Test `ClearJustPressed` resets edge triggers

5. **`game/player_test.go`** (extend if needed)
   - Test player respects InputState (moves when left held, stops when released)
   - Test player cannot move when dead
   - Test player respects invulnerability (can't be hit during invuln)

### Acceptance Criteria

- [ ] `go build ./...` succeeds
- [ ] `go test ./game/ -v` passes
- [ ] `go test ./game/ -cover` still >80%
- [ ] In browser:
  - Arrow keys / A/D move the white rectangle left/right
  - Space fires (a small white rectangle appears above player moving up)
  - Enter starts the game (transitions from start screen)
  - P pauses/unpauses (movement stops)
  - Score text visible in top-left
  - Page does not scroll when pressing arrows/space
- [ ] No key auto-repeat (holding a key doesn't cause multiple fire events)

## Technical Notes

- `e.code` gives physical key codes: `"ArrowLeft"`, `"KeyA"`, `"Space"`, `"Enter"`, `"KeyP"`
- Filter `e.repeat` in JS to avoid auto-repeat triggering multiple game events
- The `JustPressed` map pattern: set to `true` on press, read and clear at end of each `Tick()`
- Player fire should check: `input.Fire && input.JustPressed["Space"]` OR track a `wasFiring` bool
- Debug rendering is temporary — phase 6 replaces it with proper sprite rendering

## Verification

```bash
go build ./...
go test ./game/ -v -cover
./build.sh
./serve.sh &
# Browser: verify player moves, fires, state transitions work
```
