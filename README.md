# Space Invaders — Go/WASM

An attempt to recreate the core mechanics of the 1978 Atari Space Invaders
arcade game in Go, compiled to WebAssembly and playable in a modern browser
from a single HTML page.

## Background

This is a simple coding test for **qwen3.8-27b**. It grew out of an impromptu
request: build a Space Invaders game in Go, compiled to WASM. The project
spec (`AGENTS.md`) and the per-phase prompts in `prompts/` were written by
Qwen, which then implemented the game phase by phase.

## Run it

```sh
make        # builds main.wasm + wasm_exec.js
make serve  # serves on http://0.0.0.0:9090
```

Open http://localhost:9090 and press Enter.

## Controls

| Key           | Action         |
|---------------|----------------|
| Left/Right or A/D | Move      |
| Space         | Fire           |
| Enter         | Start / restart|
| P             | Pause          |
| M             | Mute           |

## Tests

```sh
make test   # go test on the pure-Go packages
make vet
```

Coverage: game 99.0%, render 97.6%, audio 100%.

## Layout

- `main.go` — WASM entry point: game loop, canvas upload, audio, high score (only file using `syscall/js`)
- `game/` — pure Go game logic: state machine, entities, collisions, scoring
- `render/` — Canvas 2D renderer, pixel-art sprites, framebuffer
- `audio/` — procedurally generated Web Audio sound effects (no audio files)
- `index.html` — single-page WASM loader
- `prompts/` — the seven phase prompts the game was built from

## Status

Complete as of phase 7 (2026-08-22). See `prompts/07-checklist.md` for the
final QA results: 60 fps, browser-verified audio, persistent high score,
180 s soak with no regressions.
