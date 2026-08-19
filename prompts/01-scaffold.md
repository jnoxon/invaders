# Phase 1: Scaffold & Build System

## Goal

Set up the Go module, project structure, WASM build pipeline, and a minimal browser page that loads WASM and renders a black canvas at 256×224. By the end of this phase, opening `index.html` in a browser shows a black canvas and the console prints "invaders: ready".

## Context

- Working directory: `/Users/jeff/Code/invaders`
- Go 1.22+ installed
- No external dependencies allowed (stdlib only)
- Canvas logical resolution: 256×224 pixels
- Integer scaling (2x or 3x) to fit window

## Deliverables

### Files to create

1. **`go.mod`**
   - Module name: `invaders`
   - Go version: 1.22

2. **`build.sh`** (executable)
   - Compiles `main.go` to `invaders.wasm` + `invaders.js` in project root
   - Uses `GOOS=js GOARCH=wasm go build -o invaders.js .`
   - Copies/generates `invaders.wasm` alongside
   - Prints success message with file sizes

3. **`serve.sh`** (executable)
   - Serves project root on `http://localhost:8080`
   - Must serve `.wasm` files with `Content-Type: application/wasm` header
   - Use a small Go HTTP server or Python with proper MIME types
   - Ctrl+C to stop

4. **`index.html`**
   - Single self-contained file (inline CSS + JS, no external files)
   - Dark background, centered canvas
   - Canvas element: `id="game"`, width=256, height=224
   - Integer scaling via CSS `image-rendering: pixelated` and CSS width/height
   - Resize handler: computes best integer scale (2x, 3x, 4x) that fits window, updates canvas CSS dimensions
   - WASM loading:
     - Fetch `invaders.js` (Go-generated glue), import it
     - Instantiate `invaders.wasm`
     - Call `main.startWasm(goInstance)` (or equivalent)
     - Show "Loading..." text before ready, remove on success
   - Input wiring (stub for now):
     - Listen for `keydown`/`keyup` on window
     - Call `main.setKey(code, pressed)` exported from Go (stub in phase 1, real in phase 3)
   - Game loop driven by `requestAnimationFrame` calling `main.tick()` exported from Go

5. **`main.go`**
   - `package main`
   - Imports: `syscall/js`, `image/color`
   - `func main()`:
     - Block on JS promise (standard Go/WASM pattern: create a promise, wait on its channel)
     - Receive `go.Wasm` instance from JS
     - Get canvas 2D context from JS
     - Start `requestAnimationFrame` loop via JS callback
   - `func startWasm(go go.Wasm)`:
     - Called by JS glue after WASM init
     - Gets canvas context
     - Sets up the frame loop: JS calls `tick()` via requestAnimationFrame
   - `func tick()`:
     - Clear canvas to black (fillRect 0,0,256,224)
     - (Phase 1: that's all)
   - `func setKey(code string, pressed bool)`:
     - Stub: store in a map, print to console
     - Will be replaced in phase 3
   - Export `tick`, `setKey`, `startWasm` via `js.Global().Get("main").Set(...)`

### Acceptance Criteria

- [ ] `go build ./...` succeeds with no errors
- [ ] `./build.sh` produces `invaders.wasm` and `invaders.js` in project root
- [ ] `./serve.sh` starts a server on :8080 with correct MIME for .wasm
- [ ] Opening `http://localhost:8080/index.html` in Chrome/Firefox/Safari:
  - Shows "Loading..." briefly
  - Then shows a black 256×224 canvas, centered, pixelated
  - Console logs "invaders: ready"
  - Pressing arrow keys logs key events to console
  - Resizing window rescales canvas (integer multiples only)
- [ ] `gofmt` clean
- [ ] No external imports beyond stdlib

## Technical Notes

- Go WASM entry pattern: `main()` blocks on a channel that JS resolves after calling an exported init function
- `syscall/js` is the standard way to interop; `github.com/golang/js/wasm` is the newer API — use whichever is stable on Go 1.22
- The Go-generated `invaders.js` handles WASM instantiation; your inline JS in index.html just needs to import it and call the exported functions
- Canvas 2D context: get via `js.Global().Get("document").Call("getElementById", "game").Call("getContext", "2d")`
- For fillRect: `ctx.Call("fillStyle", "#000").Call("fillRect", 0, 0, 256, 224)`

## Verification

```bash
go build ./...
./build.sh
./serve.sh &
# open browser to http://localhost:8080
# verify black canvas, console log, key events
```
