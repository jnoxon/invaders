# Phase 8: Mobile / Touch Support

## Goal

Make the game playable on phones and tablets with two-thumb touch controls:
drag horizontally in a move zone to steer the ship (1:1 with the finger),
press a fire zone to shoot. Keyboard gameplay must be unchanged.

## Context

- Phases 1–7 are complete and verified (see `prompts/07-checklist.md`): full
  game, keyboard controls, 60 fps, procedural audio, persistent high score.
- The repo is deployed to GitHub Pages (`https://jnoxon.github.io/invaders/`).
  `main.wasm` and `wasm_exec.js` are **committed** (there is no `.gitignore`).
  The site only updates after `make && git add main.wasm wasm_exec.js &&
  git commit -m "…" && git push`.
- Current input path (keyboard only):
  - `index.html` keydown/keyup → `main.setKey(code, pressed)`
  - `main.go` `setKey` → `g.HandleInput(code, pressed)` (KeyM also toggles
    mute; Enter also calls `audio.Enable()`)
  - `game.Game` holds `Input` (`InputState`: held Left/Right/Fire +
    `JustPressed` map); `Player.Update` moves the ship 2px per tick while a
    direction key is held; fire is edge-triggered (`JustPressed["Space"]`),
    one player bullet at a time
  - JS API in `main.go`: `tick(now)`, `setKey(code, pressed)`, `unlock()`,
    `state()` — the QA oracle returning `{state, lives, flash, score, level,
    frame, invaders, playerX, playerOk, ufo}`
- `index.html` scales the canvas with **integer scaling only** (`fit()` takes
  `Math.floor` on both axes) and centers it. On a phone that leaves the game
  tiny; mobile needs the canvas to actually fill the screen.
- The playfield is 256×224 (nearly square) and the invader grid spans ~240 of
  the 256px, so **touch zones cannot be overlaid on the canvas** — they must
  live in the surrounding page space.
- The whole repo fits comfortably in context. Before starting, read
  `main.go`, `game/game.go`, `game/player.go`, `game/input.go`, and
  `index.html`.

## Design (decided — implement as specified)

**Controls (two-thumb asymmetric):**
- **Move:** press and drag horizontally in the move zone. The ship follows
  the finger 1:1 (anchored mapping — see Technical Notes). Hold = continuous
  movement, matching the original's hold-to-move feel.
- **Fire:** press the fire zone. Edge-triggered exactly like Space: one press
  fires one shot (if no bullet is already up); holding does not re-fire.
- **Menus:** in Start/GameOver states, a press anywhere except the top-bar
  buttons acts as Enter. The first press also unlocks audio — the existing
  window-level `pointerdown` handler already calls `main.unlock()` and touch
  generates pointer events; verify, do not add a second path.
- **Top bar:** slim full-width bar above the play area, both orientations,
  with a pause button (⏸/▶) at the left and a mute button (🔊/🔇) at the
  right.

**Layout:**
- Detect touch with `matchMedia("(pointer: coarse)")`. Coarse pointer →
  mobile layout below. Fine pointer → keep the **current** centered
  integer-scaled layout exactly (desktop must not regress; phase 7 QA covers
  it). On fine-pointer devices the left/right page margins still function as
  move/fire zones (functional, no visual hint) so touch laptops work too.
- **Mobile, landscape (row):** `[move zone][canvas][fire zone]` under the top
  bar. Zones flex to fill the side space (min ~96px each). Canvas scale is
  **non-integer** (float, no floor): `min(availW/256, availH/224)`;
  `image-rendering: pixelated` (already set) keeps the look.
- **Mobile, portrait (column):** top bar, canvas fitted to the width (float
  scale `availW/256`), then a control bar of ~110px height (thumb reach)
  split into move zone (left ~60%) and fire zone (right ~40%), then a small
  hint line: “tip: rotate for more room”.
- If the landscape viewport cannot fit two min-width zones plus a usable
  canvas, fall back to the portrait layout.
- Page height: `100dvh` (fallback `100vh`) so mobile browser chrome never
  covers the fire zone. Re-layout on `resize` and `orientationchange`.

**Movement semantics (important):**
- The ship moves at **finger speed**, not the keyboard's 2px/frame: a fast
  swipe moves the ship fast. Deliberate — 1:1 responsiveness is what makes
  touch playable; the keyboard path keeps the original speed.
- Deltas are forwarded to Go and applied on the fixed 60Hz tick (Deliverable
  1). Unapplied delta at a clamped edge is **discarded, never buffered**
  (otherwise the ship teleports once the edge frees up — the edge-pin test in
  the deliverables exists to catch exactly that).

**Rejected alternatives (do not implement):** virtual ◀▶ hold-buttons (slower
than drag for constant dodging, and three buttons is a lot of screen for a
2-DOF game); tilt controls (iOS 13+ permission prompt, noisy accelerometer);
single-thumb tap-to-fire while dragging (fiddly, tap can start a drag);
auto-fire as the default (removes the one real decision the original has:
*when* to shoot, since only one bullet may be up). Auto-fire is an optional
stretch goal at the end of this prompt.

## Deliverables

### Files to modify

1. **`game/game.go`** — touch movement channel
   - Unexported `moveDx float64` on `Game`, plus
     `func (g *Game) AddMoveDx(dx float64)`
   - In `updatePlaying()`, after `Player.Update`: consume
     `d := int(math.Round(g.moveDx)); g.moveDx -= float64(d)`; if `d != 0`,
     move the player by `d`, clamped to the playfield. Any amount the clamp
     ate is **discarded, not re-buffered**.
   - `g.moveDx = 0` on every transition into and out of `StatePlaying`
     (start, restart, level transition, pause, resume, game over).
   - Keyboard movement (2px/frame) is unchanged and composes with dx in the
     same tick. No dx movement while the player is dead or state ≠ Playing.

2. **`main.go`** — expose the channel
   - `obj.Set("move", js.FuncOf(...))` → `g.AddMoveDx(args[0].Float())`
   - Add `"stateName"` (e.g. `"start"`, `"playing"`, `"paused"`,
     `"gameover"`, `"level"`) to the `state()` snapshot so `index.html`
     never compares raw state ints.
   - Nothing else: fire, menus, pause, and mute all reuse the existing
     `setKey`.

3. **`index.html`** — layout + pointer handling (the bulk of this phase)
   - Replace `fit()` with `layout()` implementing the layouts above; the
     fine-pointer path must produce the exact current canvas sizes.
   - Zones + top bar in the same single file (no external assets, house rule).
   - **Pointer Events, not Touch Events**, tracked per `pointerId`:
     - Move zone: the first `pointerdown` wins; ignore further pointers until
       it releases. On down: `anchor = pointerXLogical - state().playerX` and
       `setPointerCapture` the zone. On `pointermove`: forward the
       **incremental** `dx = pointerXLogical - lastPointerXLogical` via
       `main.move(dx)`, update `lastPointerXLogical`. On `pointerup`/
       `pointercancel`: release. HTML never moves the ship directly — the
       fixed-timestep Go code is the single authority on ship position.
     - Fire zone: `pointerdown` → `main.setKey("Space", true)`;
       `pointerup`/`pointerleave`/`pointercancel` → `main.setKey("Space",
       false)`. No capture.
     - Menus: when `state().stateName` is `"start"`/`"gameover"`, a
       `pointerdown` anywhere except the top bar → `main.setKey("Enter",
       true)` and `false` on `pointerup`.
     - Top-bar buttons: handle `pointerdown` (not `click`) → pause:
       `setKey("KeyP", true)` + `false`; mute: `setKey("KeyM", true)`
       (existing toggle). Update button glyphs on state changes (poll in the
       rAF loop or on a 500ms interval — trivial).
   - Coordinate mapping: `pointerXLogical = (e.clientX - rect.left) /
     scale` with `rect = canvas.getBoundingClientRect()` and `scale` from the
     last `layout()`. Anchor and deltas must be in **logical** coordinates
     (the non-integer scale makes this easy to get wrong).
   - CSS/behavior: `touch-action: none` on zones and canvas;
     `user-select: none; -webkit-user-select: none; -webkit-touch-callout:
     none`; `overscroll-behavior: none` on `html, body`; viewport meta gains
     `maximum-scale=1, user-scalable=no`; prevent `contextmenu` on the zones
     (Android long-press); `-webkit-tap-highlight-color: transparent`.
   - `pointercancel` is treated as release everywhere (system gestures steal
     pointers).

### New tests

4. **`game/move_test.go`** (table-driven, pure Go)
   - Integer dx applied on the next tick
   - Fractional accumulation: `AddMoveDx(0.5)` twice → 1px total
   - Clamped at both edges
   - **Edge-pin discard:** player at the left edge; `AddMoveDx(-50)` → 0px;
     `AddMoveDx(+30)` → still 0px; `AddMoveDx(+20)` → still 0px. Net finger
     displacement is 0, so the ship must not move; an implementation that
     re-buffers the clamped amount fails this test.
   - `moveDx` cleared on pause, game over, and start (no stale delta applies
     after a state change)
   - Composes with held-key movement in the same tick
   - Ignored while the player is dead or state ≠ Playing

### Acceptance Criteria

- [ ] `gofmt -l .` clean, `go vet ./...` clean
- [ ] `go test ./...` passes; `game/` coverage stays >85%
- [ ] `make` builds; **desktop regression:** fine-pointer browser shows the
      exact current layout (integer scale, centered), keyboard play identical,
      60 fps
- [ ] Touch (emulated, see Verification), landscape:
  - [ ] Tap anywhere on the start screen starts the game; audio is unlocked
        by that same tap
  - [ ] Drag in the move zone moves the ship 1:1 (`main.state().playerX`
        tracks the dispatched deltas), clamped at edges, no teleport after an
        edge-pin
  - [ ] Fire-zone tap fires exactly one bullet; holding does not re-fire; a
        tap after the bullet expires fires again
  - [ ] Two simultaneous pointers: drag + fire press work together
  - [ ] Pause button pauses (frame counter stops); mute button silences (0
        oscillator starts)
  - [ ] Double-tap / pinch on zones: no zoom, no page scroll, no
        pull-to-refresh, no context menu
  - [ ] No console errors; 60 fps (no dropped fixed steps)
- [ ] Touch, portrait: control bar below the canvas, both zones reachable,
  hint line present, nothing overlaps or clips at 320×568 (small phone)
- [ ] `make` → commit `main.wasm` + `wasm_exec.js` → push → the GitHub Pages
  site plays with touch in an emulated mobile viewport

### Optional stretch (only if everything above is green)

- **Auto-fire toggle:** “AUTO” button in the top bar. `Game.AutoFire bool`
  (+ setter); the fire logic fires whenever no player bullet is up, no input
  needed. Fire zone dims while active. Tests: fires when free, does not
  double-fire, toggles, cleared on restart.

## Technical Notes

- **Why incremental dx forwarding instead of virtual held keys:** mapping
  drag velocity to held Left/Right needs deadzones and hysteresis (finger
  micro-jitter flips direction), and it caps the ship at 2px/frame, so a fast
  swipe lags the finger by tens of pixels and feels broken. Anchored 1:1
  avoids both.
- **Anchored mapping:** on touchdown, `anchor = pointerXLogical - shipX`;
  the ship's target is always `pointerXLogical - anchor`. Forwarding
  incremental deltas to `main.move()` realizes that target through the
  Go-owned clamp/consume, keeping game logic the sole authority on position
  (AGENTS.md: pure Go game logic).
- **The 60Hz accumulator** (250ms clamp) means 0–2 ticks may run per rAF
  callback. HTML must not divide, average, or guess — it only appends
  deltas; Go consumes `round()` per tick and keeps the fractional remainder
  in `moveDx`.
- **Pointer Events give mouse for free:** zones also work with a mouse on
  desktop (harmless; the keyboard is primary there).
- **Do not touch `render/` or `audio/`** in this phase unless QA forces it.
- House rules still apply: no external dependencies, single `index.html`,
  `any` not `interface{}`, `const` for magic numbers, gofmt-clean, small
  functions.

## Verification

Desktop regression first (fine pointer, real window):

```bash
make && make serve &
# Chrome: centered integer-scaled canvas, keyboard play, 60 fps — unchanged
```

Mobile: headless Chromium with touch emulation (the `browser` tool). Per the
phase 7 notes: page-realm probes must be IIFE-wrapped `<script>` injections
(`window.main` is invisible to `page.evaluate`'s isolated world), and
`main.state()` is the oracle.

- `Emulation.setDeviceMetricsOverride` (e.g. iPhone 12: 390×844, dpr 3,
  `mobile: true`) + `Emulation.setTouchEmulationEnabled(true)`
- Drive the zones with `Input.dispatchTouchEvent` (touchStart / touchMove /
  touchEnd); verify `main.state().playerX` tracks the dispatched deltas 1:1
- Repeat for a landscape override (844×390) and a small phone (320×568)
- Emulation is a proxy: if a real phone is available, spot-check drag feel,
  double-tap behavior, and iOS Safari separately
