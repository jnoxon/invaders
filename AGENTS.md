# Space Invaders — Go/WASM

## Mission

Build a faithful reimplementation of the 1978 Atari Space Invaders arcade game in Go, compiled to WebAssembly, playable in a modern browser via a single HTML page.

## Desired Outcome

A working game that:
- Runs at 60 FPS in Chrome/Firefox/Safari via `index.html`
- Faithfully replicates core Atari Space Invaders mechanics
- Has a clean Go codebase with >80% test coverage on game logic
- Builds with a single `make` command
- No external Go dependencies (stdlib + `js/wasm` only)

## Architecture

```
invaders/
├── AGENTS.md          # This file
├── go.mod             # Module: invaders, go 1.26.6 (toolchain auto-downloads)
├── Makefile           # make, make test, make vet, make serve, make clean
├── index.html         # WASM loader + canvas + input wiring
├── main.go            # Entry: WASM init, game loop, JS interop (//go:build js && wasm)
├── main.wasm          # Built (gitignore)
├── wasm_exec.js       # Go toolchain JS glue, copied from GOROOT at build time
├── game/
│   ├── game.go        # Game struct, state machine, tick/update
│   ├── game_test.go
│   ├── player.go      # Player entity
│   ├── player_test.go
│   ├── invaders.go    # Invader grid, formation, movement
│   ├── invaders_test.go
│   ├── bullets.go     # Bullet entity (player + enemy)
│   ├── bullets_test.go
│   ├── barricades.go  # Destructible pixel barriers
│   ├── barricades_test.go
│   ├── ufo.go         # Bonus UFO
│   ├── ufo_test.go
│   ├── collision.go   # AABB + pixel collision detection
│   ├── collision_test.go
│   ├── scoring.go     # Score, lives, level progression
│   ├── scoring_test.go
│   └── input.go       # Input interface + state
├── render/
│   ├── render.go      # Canvas 2D renderer
│   ├── sprites.go     # Pixel art as Go byte arrays
│   └── render_test.go
└── audio/
    └── audio.go       # Web Audio API sound effects via JS interop
```

## Game Spec (Atari Space Invaders)

### Canvas
- Logical resolution: 256 × 224 pixels (original arcade)
- Display scaled 2x or 3x (integer scaling) to fit browser window
- Background: black

### Player
- 1 ship at bottom, 24×16 pixels
- Moves left/right with arrow keys or A/D
- Speed: ~2 pixels per frame (120 px/s at 60fps)
- Can fire 1 bullet at a time (original constraint)
- Bullet speed: ~6 pixels/frame upward
- 3 lives (displayed as ship icons below player)
- Respawn brief invulnerability (1s)

### Invaders
- 5 rows × 11 columns
- 3 types (top row: squid 30pts, middle: crab 20pts, bottom: octopus 10pts)
- Each invader is 20×15 pixels
- Grid spacing: 22px horizontal, 24px vertical (grid is 240px wide; 24px horizontal would be 260px and not fit the 256px screen)
- Movement: entire grid moves 1 step (8px) left/right
- On reaching screen edge (4px margin), grid drops 8px and reverses
- Speed increases as fewer invaders remain (fewer = faster)
- Each invader has 2 animation frames (toggle on each step)
- Invaders shoot: random invader from bottom of each column fires
- Enemy bullet speed: ~3 pixels/frame downward
- Enemy fire rate scales with invader speed
- Game over condition: any invader reaches player row (y > player.y - 8)

### Barricades (Fortifications)
- 4 barriers between player and invaders
- Each barrier: 22×16 pixels, represented as a pixel grid (1 pixel = 2×2 screen pixels)
- Destructible: each hit removes 1–2 pixels from impact point (crater pattern)
- Placed at y ≈ 176

### UFO (Mystery Ship)
- Appears at top of screen after ~20 invader kills
- Flies left-to-right or right-to-left at random
- Speed: 2 pixels/frame
- Points: random from {50, 100, 150, 300}
- Can be shot by player bullet
- Disappears off-screen

### Scoring
- Octopus (bottom row): 10 points
- Crab (middle rows): 20 points
- Squid (top row): 30 points
- UFO: 50–300 points (random)
- Level progression: all invaders cleared → next level
- Next level: invaders start higher, speed slightly increased
- High score displayed (session-only, localStorage)

### Game States
- `StartScreen`: Title, "Press ENTER to Start", high score
- `Playing`: Active gameplay
- `GameOver`: "GAME OVER", final score, "Press ENTER to Restart"
- `LevelTransition`: 2s pause between levels

### Controls
- Left/Right or A/D: Move
- Space: Fire
- Enter: Start/Restart
- P: Pause/Resume

### Sound (Web Audio API)
- Player fire: short blip (square wave, 440Hz, 50ms)
- Invader killed: low thud (triangle, 110Hz, 100ms)
- Player hit: explosion (noise, 200ms)
- UFO: warbling tone while on screen
- Invader march: 4-beat bass pattern (loops, speeds up with invaders)
- All sounds generated procedurally (no audio files)

## Rules & Conventions

### Go
- Go 1.26.6 (set in `go.mod`; the toolchain auto-downloads on first build)
- Module name: `invaders`
- No external dependencies
- Use `syscall/js` for browser interop
- `main.go` (the only file using `syscall/js`) carries `//go:build js && wasm`
- All other packages are pure Go — they build/test/vet natively via `./...`
- Game logic must be pure Go (no JS calls) — testable without a browser
- Rendering is the only layer that touches JS/Canvas
- Use `requestAnimationFrame` (via JS callback) for the game loop
- Fixed timestep: 60 updates/sec, render every frame

### Testing
- All game logic in `game/` package must be unit-tested
- Use table-driven tests
- Target >80% coverage on `game/` package
- Tests must not require a browser (pure Go)
- Run: `go test ./game/ -v`
- Integration test: verify WASM binary builds and loads

### Rendering
- Canvas 2D API (no WebGL)
- All sprites defined as byte arrays in Go (no image files)
- Pixel art: 1 byte per pixel (0=transparent, 1=filled)
- Renderer translates game state → canvas draw calls
- Integer scaling only (no fractional pixel sizes)

### HTML
- Single `index.html`, no external CSS/JS files
- Inline `<script>` for WASM loading and input wiring
- Dark theme, centered canvas, minimal chrome
- Show "Loading..." until WASM is ready
- Handle window resize (maintain aspect ratio)

### Build
- `make` (or `make build`): Compiles to `main.wasm`, then copies the toolchain's
  `wasm_exec.js` glue from `$(go env GOROOT)/lib/wasm/` into the project root.
  The glue and the wasm must come from the same Go version.
- `make test`: Runs `go test` on the pure-Go packages (`game/`, `render/`, ...).
- `make vet`: Runs `go vet` on the pure-Go packages.
- `make serve`: Serves the project root on `0.0.0.0:9090`.
- `make clean`: Removes `main.wasm` and `wasm_exec.js`.
- `wasm_exec.js` is Go's generated JS glue (not hand-written) — always copy from GOROOT, never edit.
- Build invocation is `GOOS=js GOARCH=wasm go build -o main.wasm .` (output the `.wasm`, not `.js`).

### Code Style
- No comments unless explaining non-obvious logic
- Use `any` rather than `interface{}`
- Follow standard Go conventions (gofmt, golint-clean)
- Small functions (<30 lines preferred)
- Error handling: return errors, don't panic (except in init)
- Use `const` for all magic numbers (speeds, sizes, counts)

### Phase Deliverables
Each phase (see `prompts/`) must leave the project in a buildable, testable state:
- `go build ./...` succeeds
- `go test ./game/` passes (if game code exists)
- If rendering exists: `make && make serve` shows progress
