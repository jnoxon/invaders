# Phase 4: Invader Movement, Shooting & Bullet Collision

## Goal

Verify and refine the invader grid movement, invader shooting AI, and bullet collision system. Add debug rendering for all entities so the game is visually playable (ugly rectangles are fine — proper sprites come in phase 6).

## Context

- Phases 1–3 complete: input works, player moves and fires
- `game/invaders.go` and `game/bullets.go` exist from phase 2
- This phase ensures the full combat loop works: invaders move, shoot, player shoots them, collisions resolve
- Debug rendering: draw all entities as colored rectangles

## Deliverables

### Files to modify

1. **`main.go`**
   - Expand debug rendering to draw all entities:
     - Player: white rect
     - Invaders: green rects (different shade per type: squid=lightgreen, crab=green, octopus=darkgreen)
     - Player bullets: yellow small rects (2×8)
     - Enemy bullets: red small rects (2×8)
     - Barricades: gray rects (one per barricade, filled)
     - UFO: magenta rect (when active)
   - Draw lives as small white rects in bottom-left
   - Draw level number top-right

2. **`game/invaders.go`**
   - Verify/refine `Update()`:
     - Only move when `FrameCounter >= StepInterval()`
     - On move: all alive invaders shift by `Dir * 8` pixels in X
     - Toggle `AnimFrame` on each step
     - Edge detection: if any alive invader would go past x=4 or x=256-4-InvaderW:
       - Instead of moving horizontally, move all down by 8px
       - Reverse `Dir`
     - Reset `FrameCounter` after each step
   - Verify `StepInterval()`:
     - `baseInterval = 48` (48 frames = 0.8s at 60fps for full grid)
     - `interval = baseInterval * aliveCount / (Rows*Cols)`
     - Minimum interval: 4 frames (fastest)
     - Example: 55 alive → ~48 frames, 10 alive → ~9 frames, 1 alive → ~1 frame (clamp to min 4)
   - Add `func (ig *InvaderGrid) ShouldShoot() bool`:
     - Each step, with probability proportional to speed, pick a random column
     - Return true if should fire (caller picks which invader)
   - Add `func (ig *InvaderGrid) PickShooter() *Invader`:
     - For a random column that has alive invaders, return the bottom-most alive invader in that column

3. **`game/bullets.go`**
   - Verify `Update()`:
     - Player bullets: Y -= PlayerBulletSpeed, deactivate if Y < 0
     - Enemy bullets: Y += EnemyBulletSpeed, deactivate if Y > 224
   - Add `func ActiveCount(bullets []Bullet, owner BulletOwner) int`
   - Add `func CanFire(bullets []Bullet, owner BulletOwner) bool` — check max limits

4. **`game/game.go`**
   - In `Tick()` for `StatePlaying`:
     - Update player (movement + fire)
     - Update invaders (movement)
     - Invader shooting: call `ShouldShoot()`, if true call `PickShooter()`, spawn enemy bullet (if under max)
     - Update all bullets
     - Remove inactive bullets (compact slice)
     - `CheckCollisions()`
     - Check game over conditions:
       - `Invaders.ReachedBottom()` → GameOver
       - `Lives <= 0` → GameOver
     - Check level complete:
       - `Invaders.AliveCount() == 0` → `HandleLevelComplete()`
   - UFO spawning: every ~1200 frames (20s) if not active, spawn UFO with random direction/points

5. **`game/collision.go`**
   - Verify/refine:
     - Player bullet vs invader: AABB check, if hit: invader.Alive=false, bullet.Active=false, AddScore(invader points)
     - Player bullet vs barricade: check pixel, if hit: Damage, bullet.Active=false
     - Player bullet vs UFO: AABB, if hit: UFO.Active=false, bullet.Active=false, AddScore(UFO.Points)
     - Enemy bullet vs player: AABB (skip if invulnerable), if hit: player.Hit(), bullet.Active=false, HandlePlayerDeath()
     - Enemy bullet vs barricade: check pixel, if hit: Damage, bullet.Active=false
     - Invader vs barricade: if invader rect overlaps barricade rect, destroy all overlapping pixels

### Tests to add/extend

6. **`game/invaders_test.go`**
   - Test full movement cycle: moves right N steps, hits edge, drops, moves left
   - Test step interval with different alive counts (table-driven)
   - Test ShouldShoot probability (seeded RNG, verify it fires within expected range)
   - Test PickShooter returns bottom-most invader in chosen column
   - Test animation frame toggles on each step
   - Test grid with some dead invaders (gaps in columns)

7. **`game/bullets_test.go`**
   - Test CanFire respects max limits
   - Test ActiveCount
   - Test bullet deactivates off-screen (top for player, bottom for enemy)

8. **`game/collision_test.go`**
   - Test player bullet kills specific invader (verify correct score added)
   - Test player bullet hits barricade (verify pixel damaged)
   - Test enemy bullet hits player (verify life lost)
   - Test enemy bullet blocked by barricade (player not hit)
   - Test invader overlaps barricade (pixels destroyed)
   - Test invulnerability prevents player hit
   - Test multiple bullets in same frame (order of resolution)

9. **`game/game_test.go`**
   - Test full game loop simulation (100 frames): invaders move, player can shoot
   - Test level complete triggers transition
   - Test game over when invader reaches bottom
   - Test UFO spawns after expected frames

### Acceptance Criteria

- [ ] `go build ./...` succeeds
- [ ] `go test ./game/ -v` passes all tests
- [ ] `go test ./game/ -cover` reports >80%
- [ ] In browser:
  - Invaders move in formation (step left/right, drop at edges)
  - Invaders get faster as more are killed
  - Enemy bullets appear and fall
  - Player can kill invaders (they disappear, score increases)
  - Barricades block bullets
  - Player can die (lives decrease)
  - Game over when lives = 0 or invaders reach bottom
  - UFO appears periodically at top
  - Level transition when all invaders dead
- [ ] Debug rendering shows all entities as colored rectangles
- [ ] No memory leaks (bullet slice compacts)

## Technical Notes

- Invader movement is "step" based, not continuous: all invaders jump 8px simultaneously
- The "drop at edge" logic: check if the next step would put any alive invader past the margin. If so, move down instead.
- Enemy shooting: not every frame. Use `ShouldShoot()` which returns true with some probability each step. The probability should scale with the current step interval (faster = more likely to shoot).
- Bullet cleanup: after updating, filter out inactive bullets: `bullets = append(bullets[:0], active...)` pattern
- For collision, iterate bullets and check against all potential targets. A bullet can only hit one thing (first hit wins).

## Verification

```bash
go build ./...
go test ./game/ -v -cover
./build.sh
./serve.sh &
# Browser: full gameplay loop works with rectangles
```
