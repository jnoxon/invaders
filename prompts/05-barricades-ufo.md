# Phase 5: Barricade Destruction & UFO Mechanics

## Goal

Refine and fully test the pixel-based barricade destruction system and UFO behavior. Ensure barricades properly erode from bullet impacts and that the UFO follows the original game's spawn/behavior rules.

## Context

- Phases 1–4 complete: full gameplay loop works with debug rendering
- Barricades and UFO exist in `game/` from phase 2, basic collision from phase 4
- This phase ensures the pixel-level destruction is correct and thoroughly tested
- UFO spawn timing, scoring, and edge behavior are finalized

## Deliverables

### Files to modify

1. **`game/barricades.go`**
   - Refine `Damage(bx, by int)`:
     - `bx, by` are screen-space coordinates of the impact
     - Convert to pixel grid coordinates: `px = (bx - b.X) / 2`, `py = (by - b.Y) / 2`
     - Clamp to grid bounds
     - Set the hit pixel to false
     - Also set 1 adjacent pixel to false (random from available neighbors: up, down, left, right) — this creates the crater effect
     - If no adjacent pixels are set, only remove the hit pixel
   - Refine `PixelAt(sx, sy int)`:
     - Returns false if (sx, sy) is outside barricade bounds
     - Convert screen coords to pixel grid: `px = (sx - b.X) / 2`, `py = (sy - b.Y) / 2`
     - Return `Pixels[py][px]`
   - Add `func (b *Barricade) Destroyed() bool`:
     - Returns true if no pixels are set
   - Add `func (b *Barricade) OverlapRect(ox, oy, ow, oh int)`:
     - For invader-barricade collision: destroy all barricade pixels that overlap the given rect
   - Add `func (b *Barricade) PixelCount() int` — for debugging/testing

2. **`game/ufo.go`**
   - Refine spawn logic:
     - UFO should not spawn if fewer than 10 invaders remain (original rule)
     - UFO Y position: 36 (between score line and invader top row)
     - UFO width: 40, height: 14
   - Add `func (u *UFO) CanSpawn(aliveInvaders int) bool`:
     - Returns true if `!u.Active && aliveInvaders >= 10`
   - Verify `Update()`:
     - Move X by `Dir * UFOSpeed`
     - Deactivate if `X < -40` or `X > 256` (fully off-screen)
   - Add `func (u *UFO) Points() int` — returns the points value (set at creation)

3. **`game/game.go`**
   - Refine UFO spawning in `Tick()`:
     - Track `KillCount` (total invaders killed this level)
     - Every 20 kills (i.e., when `KillCount % 20 == 0 && KillCount > 0`), attempt UFO spawn
     - Only spawn if `UFO.CanSpawn(Invaders.AliveCount())`
   - Refine barricade cleanup:
     - After all collisions, remove destroyed barricades (or keep them but skip rendering/collision for destroyed ones)
   - Add `KillCount int` field to Game, reset on level start

4. **`main.go`**
   - Update debug rendering for barricades:
     - Draw each set pixel as a 2×2 gray rect
     - Skip destroyed barricades
   - Update debug rendering for UFO:
     - Draw magenta rect at UFO position when active

### Tests to add/extend

5. **`game/barricades_test.go`** (comprehensive)
   - Table-driven test for `Damage`:
     - Hit center pixel → verify center + 1 neighbor removed
     - Hit corner pixel → verify only that pixel + 1 neighbor (edge case)
     - Hit already-empty pixel → no change
     - Hit when only 1 pixel remains → that pixel removed, Destroyed() = true
   - Test `PixelAt`:
     - Inside bounds, pixel set → true
     - Inside bounds, pixel cleared → false
     - Outside bounds → false
   - Test `OverlapRect`:
     - Full overlap → all pixels in range destroyed
     - Partial overlap → only overlapping pixels destroyed
     - No overlap → no change
   - Test `Destroyed`:
     - Fresh barricade → false
     - All pixels removed → true
   - Test initial shape matches spec (11×8, arch pattern)
   - Test `PixelCount` on fresh barricade

6. **`game/ufo_test.go`** (comprehensive)
   - Test `CanSpawn`:
     - Not active, 10+ invaders → true
     - Not active, 9 invaders → false
     - Active, 55 invaders → false
   - Test movement left and right
   - Test deactivation when off-screen (both directions)
   - Test points are from valid set {50, 100, 150, 300} (seeded RNG, multiple samples)
   - Test UFO Y is always 36

7. **`game/collision_test.go`** (extend)
   - Test player bullet hits UFO → UFO deactivated, score added, bullet gone
   - Test player bullet passes through destroyed barricade (no collision)
   - Test enemy bullet destroys barricade pixel → subsequent enemy bullet passes through
   - Test invader overlapping barricade destroys correct pixels

8. **`game/game_test.go`** (extend)
   - Test UFO spawns after exactly 20 kills (simulate kills, check UFO active)
   - Test UFO does not spawn with <10 invaders alive
   - Test KillCount resets on new level
   - Test full level: kill all invaders, verify transition
   - Test barricade erosion over multiple hits

### Acceptance Criteria

- [ ] `go build ./...` succeeds
- [ ] `go test ./game/ -v` passes all tests
- [ ] `go test ./game/ -cover` reports >85% (raised target due to thorough barricade/UFO tests)
- [ ] Barricades visibly erode pixel-by-pixel in browser
- [ ] Barricade crater pattern looks natural (not just a single pixel)
- [ ] UFO appears after 20 kills, not too early or too late
- [ ] UFO does not spawn when <10 invaders remain
- [ ] UFO scoring works (random from valid set)
- [ ] Destroyed barricades no longer block bullets
- [ ] Invader-barricade collision destroys overlapping pixels

## Technical Notes

- The barricade pixel grid is 11 wide × 8 tall (each cell = 2×2 screen pixels → 22×16 total)
- The original barricade shape is an arch/inverted-U:
  ```
  Row 0: .XXXXXXXX..   (cols 1-8 set)
  Row 1: .XXXXXXXX..
  Row 2: .XXXXXXXX..
  Row 3: .XXXXXXXX..
  Row 4: .XX....XX..   (cols 1-2, 8-9 set)
  Row 5: .XX....XX..
  Row 6: .XX....XX..
  Row 7: .XX....XX..
  ```
- Damage pattern: when a bullet hits pixel (px, py), set it to false, then randomly pick ONE of the 4 neighbors (up/down/left/right) that is still true and set it to false. This creates irregular erosion.
- For the random neighbor, use the Game's RNG (passed in or stored) for testability
- UFO spawn: the original game uses a kill counter per level. After 20 kills, the next "opportune" moment spawns a UFO. Simplify: spawn immediately when kill count hits a multiple of 20.
- Keep destroyed barricades in the slice (don't remove) to maintain 4 positions; just skip them in collision/render when `Destroyed()` is true

## Verification

```bash
go build ./...
go test ./game/ -v -cover
# Focus on: go test ./game/ -run TestBarricade -v
# Focus on: go test ./game/ -run TestUFO -v
./build.sh
./serve.sh &
# Browser: shoot at barricades, watch them erode; wait for UFO
```
