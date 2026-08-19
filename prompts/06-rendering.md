# Phase 6: Sprite Rendering

## Goal

Replace all debug rectangle rendering with proper pixel-art sprites defined as Go byte arrays. Implement the full renderer that draws the game state to the canvas with authentic-looking Space Invaders graphics.

## Context

- Phases 1–5 complete: full gameplay works with colored rectangle debug rendering
- All game state is in `game.Game`
- This phase creates the `render/` package that translates game state → canvas draw calls
- All sprites are 1-byte-per-pixel arrays (0=transparent, 1=filled)
- The original game uses 1-color sprites (white on black)

## Deliverables

### Files to create

1. **`render/sprites.go`**
   - All sprites as `[][]byte` or flat `[]byte` with known width/height
   - Helper: `type Sprite struct { Data []byte; W, H int }`
   - `func (s Sprite) Pixel(x, y int) bool`
   - Sprites needed (all 1-color, white):
     - **Player ship**: 24×16 (2 frames: normal, firing — firing adds a 2×4 pixel flame)
     - **Squid (top row)**: 20×15 (2 animation frames)
     - **Crab (middle rows)**: 20×15 (2 animation frames)
     - **Octopus (bottom rows)**: 20×15 (2 animation frames)
     - **UFO**: 40×14 (2 animation frames, optional)
     - **Barricade**: drawn per-pixel (no sprite needed, just fillRect for each set pixel)
     - **Player bullet**: 2×8 (solid rect, no sprite needed)
     - **Enemy bullet**: 2×8 (zigzag pattern, 2 frames)
     - **Score digits**: 3×5 pixel font (0-9) for drawing score
     - **Text**: simple 3×5 or 5×5 pixel font for "SPACE INVADERS", "GAME OVER", "PRESS ENTER", etc.
   - Define the actual pixel patterns (transcribe from original game screenshots)
   - Each sprite: `var playerSprite = Sprite{W: 24, H: 16, Data: []byte{...}}`

   Approximate sprite shapes (simplified but recognizable):

   **Player (24×16)**:
   ```
   ............X.............
   ...........XXX............
   ...........XXX............
   ........X.XXX.X...........
   ........X.XXX.X...........
   ........X.XXX.X...........
   XXXXXXXXXXXXXXXXXXXXXXXX..
   XXXXXXXXXXXXXXXXXXXXXXXX..
   XXXXXXXXXXXXXXXXXXXXXXXX..
   ```
   (Refine to look like the original cannon shape)

   **Squid (20×15)** — 2 frames (antennae up/down):
   ```
   Frame 0:              Frame 1:
   ....XXXX....          ....XXXX....
   ...XXXXXX...          ...XXXXXX...
   ..XXXXXXXX..          ..XXXXXXXX..
   .XX.XX.XX.XX.         .XXXXXXXXXX.
   XXXXXXXXXXXXX         XX.XXXXXX.XX
   .X.XXXXXXX.X.         .XXXXXXXXX.
   ..X.XXXX.X..          ..XXXXXXXX..
   ```

   **Crab (20×15)** — 2 frames (legs up/down):
   ```
   Frame 0:              Frame 1:
   ..X.......X..         .........
   ...X.....X...         ..X.....X..
   ..XXXXXXX..X..        .XXXXXXXXX.
   .XXXXXXXXXXX.         .XXXXXXXXX.
   XX.XXXXXXXX.XX        XXXXXXXXXX
   X.XXXXXXXXXXX.X       .XXXXXXXXX.
   ..X..X..X..X..        X..X..X..X.
   ```

   **Octopus (20×15)** — 2 frames:
   ```
   Frame 0:              Frame 1:
   ..XXXXXXXX..          ..XXXXXXXX..
   .XXXXXXXXXX.          .XXXXXXXXXX.
   XX.XXXXXXX.XX         XXXXXXXXXXXX
   XXXXXXXXXXXXX         XX.XXXXXX.XX
   .X.XXXXXXX.X.         .XXXXXXXXX.
   ..X..X..X..X.         ..XXXXXXXX..
   X.X..X..X..X.X        X.X..X..X..X
   ```

   **UFO (40×14)**:
   ```
   ........XXXXXXXXXXXXXXXX............
   ....XXXXXXXXXXXXXXXXXXXXXXXX........
   ...XXXXXXXXXXXXXXXXXXXXXXXXXX.......
   ..XXX..XXX..XXX..XXX..XXX..XXX......
   .XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX....
   .XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX....
   ...XXXXXXXXXXXXXXXXXXXXXXXXXX.......
   ```

   NOTE: These are starting points. Refine them to be more faithful to the original. The key is they're recognizable and 2-frame animated.

2. **`render/render.go`**
   - `type Renderer struct`:
     - `ctx js.Value` (canvas 2D context)
     - `scale int` (integer scale factor)
   - `func NewRenderer(canvas js.Value, scale int) *Renderer`
   - `func (r *Renderer) Render(g *game.Game)`:
     - Clear to black
     - Switch on `g.State`:
       - `StateStart`: renderStartScreen(g)
       - `StatePlaying`: renderGameplay(g)
       - `GameOver`: renderGameOver(g)
       - `StateLevelTransition`: renderLevelTransition(g)
       - `StatePaused`: renderGameplay(g) + "PAUSED" text
   - `func (r *Renderer) renderGameplay(g *game.Game)`:
     - Draw score (top-left), high score (top-center), level (top-right) using pixel font
     - Draw player sprite (normal or firing frame)
     - Draw all alive invaders (correct sprite + animation frame based on `AnimFrame`)
     - Draw bullets (player: white 2×8, enemy: zigzag sprite)
     - Draw barricades (per-pixel fillRect for set pixels)
     - Draw UFO if active
     - Draw lives (player sprite × remaining lives, bottom-left)
     - Draw ground line (1px white line at bottom)
   - `func (r *Renderer) renderStartScreen(g *game.Game)`:
     - "SPACE INVADERS" title (large, center)
     - Invader sprites with point values (score table)
     - "PRESS ENTER TO START" (blinking, toggle every 30 frames)
     - High score display
   - `func (r *Renderer) renderGameOver(g *game.Game)`:
     - "GAME OVER" center
     - Final score
     - "PRESS ENTER TO RESTART" (blinking)
   - `func (r *Renderer) drawSprite(s Sprite, x, y int)`:
     - For each pixel where s.Pixel(px, py) is true: `ctx.Call("fillRect", x+px, y+py, 1, 1)`
     - Optimization: draw horizontal runs as single fillRect (optional)
   - `func (r *Renderer) drawText(text string, x, y int, font SpriteFont)`:
     - Look up each character in font, drawSprite at offset position
   - `func (r *Renderer) SetScale(scale int)` — update scale for resize
   - `func (r *Renderer) Resize(newScale int)` — called on window resize

3. **`render/sprite_font.go`**
   - `type SpriteFont struct { Digits map[byte]Sprite }`
   - `var PixelFont SpriteFont` — 3×5 pixel font for 0-9, A-Z, space
   - Used for score display and text rendering
   - Characters: 0-9, A-Z, space, colon (for "SCORE: 1234")

4. **`render/render_test.go`**
   - Test sprite data integrity (correct dimensions, expected pixel count)
   - Test `Sprite.Pixel()` (in bounds, out of bounds)
   - Test `drawSprite` logic (mock or verify pixel positions)
   - Test text rendering (correct x offset per character)
   - Test renderer state (scale changes)
   - Note: canvas calls can't be tested without a browser; test the logic that determines WHAT to draw

### Files to modify

5. **`main.go`**
   - Remove ALL debug rectangle rendering
   - Create `render.NewRenderer(canvas, initialScale)`
   - In `tick()`: call `r.Render(g)` instead of manual drawing
   - On resize: call `r.SetScale(newScale)` and update canvas CSS dimensions
   - Keep the game loop: `g.Tick()` then `r.Render(g)`

6. **`index.html`**
   - Ensure resize handler calls into Go to update scale
   - Export a `setCanvasScale(scale int)` function from Go (or handle purely in CSS)
   - Actually: scale is handled in CSS (canvas element is always 256×224 internal, CSS width/height changes). The renderer always draws at 256×224. So no Go-side scale change needed.
   - Remove any debug-specific code

### Acceptance Criteria

- [ ] `go build ./...` succeeds
- [ ] `go test ./...` passes (including render package tests)
- [ ] In browser:
  - Player looks like a cannon/spaceship (not a rectangle)
  - Invaders look like the original 3 types (squid, crab, octopus)
  - Invaders animate (2-frame toggle visible)
  - UFO looks like the classic disc shape
  - Score displayed in pixel font
  - Start screen shows title, score table, "PRESS ENTER"
  - Game over screen shows "GAME OVER"
  - Barricades look like the original arch shape
  - Player bullet is a thin white line
  - Enemy bullet has zigzag pattern
  - Lives displayed as small ship icons
  - Ground line at bottom
  - All rendering at 256×224, scaled via CSS
- [ ] No visible rectangles (all sprites are pixel art)
- [ ] Performance: 60fps with all entities on screen (sprite drawing is the bottleneck — optimize if needed)

## Technical Notes

- Sprite drawing: for a 20×15 sprite, that's up to 300 fillRect calls per sprite. With 55 invaders, that's 16,500 calls worst case. This should still be fast on modern browsers.
- Optimization if needed: draw sprites as `ImageData` (put a 256×224 ImageData buffer) instead of individual fillRect calls. This is a single `putImageData` call per frame.
- Alternative: pre-render sprites to offscreen canvases (one per sprite/frame), then use `drawImage` to blit. This is the fastest approach.
- For phase 6, start with direct fillRect. If performance is an issue, switch to ImageData in phase 7.
- The pixel font should be simple: 3 pixels wide, 5 pixels tall, 1px spacing between characters. Total text width = len(text) * 4 - 1.
- Blinking text: use `g.Frame % 60 < 30` to toggle visibility (1Hz blink)
- All sprites are white on transparent. The renderer sets `fillStyle` once to white, draws everything.

## Verification

```bash
go build ./...
go test ./... -v
./build.sh
./serve.sh &
# Browser: verify all sprites render correctly, animations work, screens look right
# Compare visually to original Space Invaders screenshots
```
