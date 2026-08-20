package game

import (
	"math/rand/v2"
	"testing"
)

// TestSoakMultiSeed runs many full games with deterministic input and
// verifies no panics and bounded entity state across restarts,
// level transitions, and pause cycles.
func TestSoakMultiSeed(t *testing.T) {
	const seeds = 20
	const framesPerGame = 60000
	const maxGamesPerSeed = 4

	for seed := 0; seed < seeds; seed++ {
		g := NewGame()
		g.RNG = rand.New(rand.NewPCG(1, uint64(seed)))
		in := rand.New(rand.NewPCG(1, uint64(seed+1_000_000)))

		g.Input.Update("Enter", true)
		g.Tick()
		g.Input.Update("Enter", false)

		dir := 1
		turn := in.IntN(20) + 10
		restarts := 0
		ufoSeen := false
		maxKills := 0
		levelSeen := 1
		resumeAt := 0

		for frame := 0; frame < framesPerGame && restarts < maxGamesPerSeed; frame++ {
			if frame == turn {
				dir = -dir
				turn = frame + in.IntN(20) + 10
			}
			if dir > 0 {
				g.Input.Update("ArrowRight", true)
				g.Input.Update("ArrowLeft", false)
			} else {
				g.Input.Update("ArrowLeft", true)
				g.Input.Update("ArrowRight", false)
			}
			if in.IntN(20) == 0 {
				g.Input.Update("Space", true)
			} else {
				g.Input.Update("Space", false)
			}
			if resumeAt == 0 && frame%5000 == 4999 && g.State == StatePlaying {
				g.HandleInput("KeyP", true)
				resumeAt = frame + 3
			}
			g.Tick()
			if resumeAt != 0 && frame >= resumeAt && g.State == StatePaused {
				g.HandleInput("KeyP", true)
				resumeAt = 0
			}

			if len(g.Bullets) > MaxPlayerBullets+MaxEnemyBullets {
				t.Fatalf("seed %d: bullet slice grew to %d", seed, len(g.Bullets))
			}
			if g.UFOActive {
				ufoSeen = true
			}
			if g.KillCount > maxKills {
				maxKills = g.KillCount
			}
			if g.Level > levelSeen {
				levelSeen = g.Level
			}
			if g.State == StateGameOver {
				g.Input.Update("Enter", true)
				restarts++
			}
		}

		if restarts == 0 {
			t.Fatalf("seed %d: expected at least one game over in %d frames", seed, framesPerGame)
		}
		if maxKills >= UFOSpawnKills && !ufoSeen {
			t.Fatalf("seed %d: reached %d kills without UFO", seed, maxKills)
		}
		if g.Lives < 0 || g.Lives > StartLives {
			t.Fatalf("seed %d: lives out of range: %d", seed, g.Lives)
		}
		if g.Score < 0 {
			t.Fatalf("seed %d: negative score: %d", seed, g.Score)
		}
		if levelSeen < 1 {
			t.Fatalf("seed %d: level below 1", seed)
		}
	}
}

// TestSoakLongGame keeps the player invulnerable and fires frequently so
// games run to the invader bottom-reach or level completion, exercising
// UFO spawns, grid descent, and long bullet traffic.
func TestSoakLongGame(t *testing.T) {
	for seed := 0; seed < 6; seed++ {
		g := NewGame()
		g.RNG = rand.New(rand.NewPCG(1, uint64(1000+seed)))
		in := rand.New(rand.NewPCG(1, uint64(2_000_000+seed)))

		g.Input.Update("Enter", true)
		g.Tick()
		g.Input.Update("Enter", false)

		dir := 1
		turn := in.IntN(20) + 10
		ufoSeen := false
		maxKills := 0

		for frame := 0; frame < 20000; frame++ {
			g.Player.Invulnerable = 1 << 20
			if frame == turn {
				dir = -dir
				turn = frame + in.IntN(20) + 10
			}
			if dir > 0 {
				g.Input.Update("ArrowRight", true)
				g.Input.Update("ArrowLeft", false)
			} else {
				g.Input.Update("ArrowLeft", true)
				g.Input.Update("ArrowRight", false)
			}
			if in.IntN(4) == 0 {
				g.Input.Update("Space", true)
			} else {
				g.Input.Update("Space", false)
			}
			g.Tick()
			if g.State == StateGameOver {
				g.Input.Update("Enter", true)
			}

			if len(g.Bullets) > MaxPlayerBullets+MaxEnemyBullets {
				t.Fatalf("seed %d: bullet slice grew to %d", seed, len(g.Bullets))
			}
			if g.UFOActive {
				ufoSeen = true
			}
			if g.KillCount > maxKills {
				maxKills = g.KillCount
			}
		}
		if maxKills >= UFOSpawnKills && !ufoSeen {
			t.Fatalf("seed %d: reached %d kills without UFO", seed, maxKills)
		}
	}
}
