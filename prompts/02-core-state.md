# Phase 2: Core Game State & Entities (Pure Go)

## Goal

Implement all game entity types, the game state machine, and the tick/update loop as pure Go in the `game/` package. No rendering, no JS calls — just data structures and state transitions. All logic is unit-testable with `go test ./game/`.

## Context

- Phase 1 is complete: project builds, WASM loads, canvas renders black
- This phase adds `game/` package with all core logic
- `main.go` will be updated to create a `game.Game` and call `g.Tick()` each frame
- All constants (speeds, sizes, counts) defined as `const` in the game package

## Deliverables

### Files to create/modify

1. **`game/game.go`**
   - `type GameState int` with constants: `StateStart`, `StatePlaying`, `StateGameOver`, `StateLevelTransition`, `StatePaused`
   - `type Game struct`:
     - `State GameState`
     - `Player Player`
     - `Invaders InvaderGrid`
     - `Bullets []Bullet`
     - `Barricades []Barricade`
     - `UFO UFO`
     - `UFOActive bool`
     - `Score, HighScore, Lives int`
     - `Level int`
     - `Frame int` (frame counter for animation)
     - `TransitionTimer int` (for level transition countdown)
     - `RNG *rand.Rand` (seeded, for deterministic testing)
     - `Input InputState`
   - `func NewGame() *Game` — initialize all entities, state = StateStart
   - `func (g *Game) Tick()` — one frame of game logic:
     - Switch on State:
       - `StateStart`: wait for Enter → start game
       - `StatePlaying`: update player, invaders, bullets, barricades, UFO; check collisions; check game over
       - `StateGameOver`: wait for Enter → reset to new game
       - `StateLevelTransition`: countdown timer → reset level, state = StatePlaying
       - `StatePaused`: do nothing
   - `func (g *Game) StartGame()` — reset all, state = StatePlaying
   - `func (g *Game) ResetLevel()` — reset invaders/bullets/barricades, keep score/lives
   - `func (g *Game) HandleInput(code string, pressed bool)` — update InputState, handle state transitions (Enter, P)
   - `func (g *Game) CheckCollisions()` — all collision checks in one place
   - `func (g *Game) GameOver()` — set state, update high score

2. **`game/player.go`**
   - `type Player struct`:
     - `X, Y int`
     - `W, H int` (24×16)
     - `Alive bool`
     - `Invulnerable int` (frames remaining)
   - `const PlayerSpeed = 2`
   - `func (p *Player) Update(input *InputState)` — move left/right if alive, clamp to screen bounds [0, 256-W]
   - `func (p *Player) Fire() *Bullet` — return bullet at player position, or nil if already has one active (caller checks)
   - `func (p *Player) Hit()` — set Alive=false
   - `func (p *Player) Respawn()` — reset position to center bottom, Invulnerable=60

3. **`game/invaders.go`**
   - `type InvaderType int`: `InvaderSquid` (30pts), `InvaderCrab` (20pts), `InvaderOctopus` (10pts)
   - `type Invader struct`:
     - `X, Y int`
     - `Type InvaderType`
     - `Alive bool`
     - `AnimFrame int` (0 or 1)
   - `type InvaderGrid struct`:
     - `Invaders [5][11]Invader` (or `[][]Invader`)
     - `Dir int` (1 or -1)
     - `StepInterval int` (frames between moves, starts ~48, decreases as invaders die)
     - `FrameCounter int`
   - `const InvaderCols = 11`, `InvaderRows = 5`, `InvaderW = 20`, `InvaderH = 15`, `InvaderHGap = 24`, `InvaderVGap = 24`
   - `func NewInvaderGrid(level int) InvaderGrid` — populate grid, start position depends on level (higher level = starts higher)
   - `func (ig *InvaderGrid) Update()` — move all alive invaders, handle edge wrap + drop, toggle animation, compute step interval based on alive count
   - `func (ig *InvaderGrid) AliveCount() int`
   - `func (ig *InvaderGrid) StepInterval() int` — formula: base interval scaled by (aliveCount / totalCount)
   - `func (ig *InvaderGrid) BottomOfColumn(col int) *Invader` — for shooting AI
   - `func (ig *InvaderGrid) ReachedBottom() bool` — any invader y > threshold

4. **`game/bullets.go`**
   - `type BulletOwner int`: `BulletPlayer`, `BulletEnemy`
   - `type Bullet struct`:
     - `X, Y int`
     - `Owner BulletOwner`
     - `Active bool`
   - `const PlayerBulletSpeed = 6`, `EnemyBulletSpeed = 3`
   - `const MaxPlayerBullets = 1`, `MaxEnemyBullets = 4`
   - `func (b *Bullet) Update()` — move up or down based on owner, deactivate if off-screen
   - `func (b *Bullet) Rect() (x, y, w, h int)` — for collision (bullets are 2×8 pixels)

5. **`game/barricades.go`**
   - `type Barricade struct`:
     - `X, Y int`
     - `Pixels [][]bool` (11×8 grid, each pixel = 2×2 screen pixels)
   - `const BarricadeW = 22`, `BarricadeH = 16` (screen pixels), `BarricadePixelW = 11`, `BarricadePixelH = 8`
   - `func NewBarricade(x int) Barricade` — initial shape (inverted U / arch shape from original)
   - `func (b *Barricade) Damage(bx, by int)` — remove pixels around impact point (crater: remove the hit pixel + 1 adjacent)
   - `func (b *Barricade) PixelAt(sx, sy int) bool` — screen coord → pixel grid lookup
   - `func (b *Barricade) Rect() (x, y, w, h int)`
   - `func (b *Barricade) Destroyed() bool` — all pixels false

6. **`game/ufo.go`**
   - `type UFO struct`:
     - `X, Y int`
     - `Dir int`
     - `Active bool`
     - `Points int`
   - `const UFOSpeed = 2`, `UFOY = 36`
   - `func NewUFO() UFO` — random direction, random points from {50, 100, 150, 300}
   - `func (u *UFO) Update()` — move, deactivate if off-screen
   - `func (u *UFO) Rect() (x, y, w, h int)` (40×14)

7. **`game/collision.go`**
   - `func AABB(ax, ay, aw, ah, bx, by, bw, bh int) bool` — axis-aligned bounding box overlap
   - `func (g *Game) checkPlayerBulletHits()` — player bullet vs invaders, barricades, UFO
   - `func (g *Game) checkEnemyBulletHits()` — enemy bullet vs player, barricades
   - `func (g *Game) checkInvaderBarricadeCollision()` — invader overlap with barricade destroys overlapping pixels

8. **`game/scoring.go`**
   - `func (g *Game) AddScore(points int)` — add, update high score
   - `func (g *Game) InvaderPoints(t InvaderType) int`
   - `func (g *Game) HandlePlayerDeath()` — decrement lives, check game over
   - `func (g *Game) HandleLevelComplete()` — state = StateLevelTransition, timer = 120 (2s)
   - `func (g *Game) UpdateHighScore()` — if score > highScore, update

9. **`game/input.go`**
   - `type InputState struct`:
     - `Left, Right, Fire bool`
     - `JustPressed map[string]bool` (for edge-triggered events like Enter, P)
   - `type Input interface` (optional, for future):
     - `Left() bool`
     - `Right() bool`
     - `Fire() bool`
   - `func (s *InputState) Update(code string, pressed bool)` — map key codes to fields, track just-pressed
   - `func (s *InputState) ClearJustPressed()` — call after each frame

10. **`game/game_test.go`**
    - Test state transitions (Start → Playing → GameOver → Start)
    - Test level transition
    - Test game over conditions (lives = 0, invader reaches bottom)
    - Test tick does nothing when paused
    - Use seeded RNG for determinism

11. **`game/player_test.go`**
    - Test movement (left, right, clamping)
    - Test fire returns bullet at correct position
    - Test hit/respawn cycle
    - Test invulnerability countdown

12. **`game/invaders_test.go`**
    - Test grid initialization (correct types per row, correct positions)
    - Test movement: moves right, wraps at edge, drops, reverses
    - Test step interval decreases as invaders die
    - Test animation toggle
    - Test BottomOfColumn
    - Test ReachedBottom

13. **`game/bullets_test.go`**
    - Test player bullet moves up, deactivates off-screen
    - Test enemy bullet moves down, deactivates off-screen
    - Test max bullet limits

14. **`game/barricades_test.go`**
    - Test initial shape (correct pixels set)
    - Test damage removes correct pixels
    - Test PixelAt lookup
    - Test Destroyed when all pixels gone

15. **`game/ufo_test.go`**
    - Test movement (left and right)
    - Test deactivation off-screen
    - Test random points are from valid set

16. **`game/collision_test.go`**
    - Test AABB (overlap, no overlap, edge cases, zero-size)
    - Test player bullet kills invader
    - Test enemy bullet hits player
    - Test bullet hits barricade (pixel damage)
    - Test bullet hits UFO

17. **`game/scoring_test.go`**
    - Test score addition
    - Test high score update
    - Test level progression
    - Test player death / game over

### Files to modify

- **`main.go`**: Create `game.NewGame()`, call `g.Tick()` in the tick function, pass key events to `g.HandleInput()`

### Acceptance Criteria

- [ ] `go build ./...` succeeds
- [ ] `go test ./game/ -v` passes all tests
- [ ] `go test ./game/ -cover` reports >80% coverage
- [ ] No JS/syscall imports in `game/` package (pure Go)
- [ ] All magic numbers are named constants
- [ ] `gofmt` clean
- [ ] Game loop in main.go drives `g.Tick()` at 60fps
- [ ] State machine works: press Enter starts game, P pauses, etc. (verifiable via console or future rendering)

## Technical Notes

- Use `math/rand/v2` (Go 1.22) or `math/rand` with `rand.New(rand.NewSource(seed))` for deterministic tests
- Keep `Tick()` as a single entry point; it orchestrates all sub-updates
- Collision order matters: check player bullets first, then enemy bullets, then invader-barricade overlap
- The `RNG` field in Game allows tests to seed it: `g.RNG = rand.New(rand.NewSource(42))`
- Invader step interval formula: `interval = baseInterval * aliveCount / totalInitial` (so 55 alive = slow, 1 alive = fast)
- Barricade initial shape (11×8 grid):
  ```
  .XXXXXXXX..
  .XXXXXXXX..
  .XXXXXXXX..
  .XXXXXXXX..
  .XX....XX..
  .XX....XX..
  .XX....XX..
  .XX....XX..
  ```

## Verification

```bash
go build ./...
go test ./game/ -v -cover
# Expect: all tests pass, coverage >80%
```
