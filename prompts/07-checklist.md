# Phase 7: Final QA Checklist

Verification results for the completed game. All items verified 2026-08-22 against
`make build` output served at `localhost:9090` (headless Chromium).

## Build & Static Checks

- [x] `gofmt -l .` — clean (no unformatted files)
- [x] `go vet ./...` — clean
- [x] `go test ./...` — all packages pass
- [x] `go test ./game/ -cover` — coverage >85%
- [x] `make build` — produces `main.wasm` + `wasm_exec.js` (same Go toolchain)
- [x] No console errors on page load (verified via `window.onerror` + `unhandledrejection` + `console.error` hooks)

## Game States (browser-verified)

- [x] `StartScreen` — title, "HIGH SCORE: NNNNN", UFO `???` + squid/crab/octopus score table; Enter starts game
- [x] `Playing` — full grid (55 invaders), 4 barricades, player, HUD (SCORE / HI-SCORE / LV), lives row
- [x] `LevelTransition` — handled (2s pause, then fresh grid higher up)
- [x] `Paused` — P toggles; no ticks while paused
- [x] `GameOver` — "GAME OVER", final score, "PRESS ENTER TO RESTART"; Enter restarts
- [x] Death flash: 6-frame white flash on player death; **decays to 0 within ~100 ms in every state**
  (regression: decay now lives in `Tick()`, so it runs after state changes — old code left the screen
  stuck white on the fatal death. Verified: sample at death `state:2 flash:4 white:1.000` →
  `flash:1 white:1.000` → `flash:0 white:0.007`; repeat-verified during soak)

## Core Mechanics (browser-verified)

- [x] Player fire: Space; one bullet at a time; kill confirmed (score +10, invaders 54)
- [x] Scoring: octopus 10, crab 20, squid 30 (bottom-row kills verified at 10 pts each)
- [x] Invader grid: 5×11, steps 8px, drops/reverses at edges, speeds up as invaders die
- [x] Enemy fire: bullets from bottom of columns; stationary player loses all 3 lives (~30 s/game)
- [x] Barricades: rendered, destructible
- [x] UFO: spawns after kill threshold, flies across, shootable (unit-tested; score popup)
- [x] Game over on last-life death and on invaders reaching player row (both observed)
- [x] Restart: Enter from GameOver starts fresh level-1 game (score 0, 55 invaders, 3 lives)

## High Score (browser-verified)

- [x] `localStorage["si-highscore"]` written on change (score 20 → stored "20")
- [x] Loaded at init (`loadHighScore` in `main.go`)
- [x] Survives full page reload; start screen renders "HIGH SCORE: 00020"

## Audio (browser-verified via AudioContext/OscillatorNode prototype hooks)

- [x] AudioContext "suspended" before gesture → "running" after first Enter (user-gesture unlock)
- [x] Fire: `square:440`
- [x] Invader kill: `triangle:110`
- [x] March: 4-beat bass `triangle:55 / 65 / 75 / 85`, one per grid step
- [x] Player hit: white-noise buffer (counted `createBufferSource` calls; 17 over 6 deaths)
- [x] Game over: descending `triangle:440 → 330 → 220 → 110`
- [x] Mute: KeyM toggles (0 oscillator starts while muted, +4 beats after unmute)

## Performance

- [x] 60.2 fps sustained (3 s sample, full 55-invader grid); rAF callbacks == game ticks (1:1)
- [x] Frame delta: min 5.9 ms, max 16.8 ms (no dropped fixed steps)
- [x] 180 s soak, 6 auto-restart cycles: 0 stuck-white episodes, 0 persistent-white episodes,
  0 frame-counter regressions, 0 new errors
- [x] JS heap over soak: −1.2 MB (GC healthy, no leak)
- [x] Single `putImageData` upload per frame (persistent `Uint8ClampedArray` + `ImageData`)

## Notes

- The browser tool's `page.evaluate` runs in an isolated realm: `window.main` is invisible to it.
  All page-realm probes were injected as `<script>` elements (DOM bridge); injected scripts must be
  IIFE-wrapped — top-level `const` leaks into the page's global lexical scope across injections.
- `state()` (exposed by `main.go`) returns `{state, lives, flash, score, level, frame, invaders,
  playerX, playerOk, ufo}` and was the primary oracle for all checks above.
