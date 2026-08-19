package game

import (
	"math/rand/v2"
	"testing"
)

func newTestGame() *Game {
	g := NewGame()
	g.RNG = rand.New(rand.NewPCG(1, 42))
	return g
}

func testRNG() *rand.Rand {
	return rand.New(rand.NewPCG(1, 42))
}

func TestNewGameInitialState(t *testing.T) {
	g := newTestGame()
	if g.State != StateStart {
		t.Fatalf("state = %v, want StateStart", g.State)
	}
	if g.Lives != StartLives {
		t.Fatalf("lives = %d, want %d", g.Lives, StartLives)
	}
	if g.Level != 1 {
		t.Fatalf("level = %d, want 1", g.Level)
	}
	if g.Invaders.AliveCount() != InvaderRows*InvaderCols {
		t.Fatalf("alive invaders = %d", g.Invaders.AliveCount())
	}
	if len(g.Barricades) != barricadeCount {
		t.Fatalf("barricades = %d", len(g.Barricades))
	}
}

func TestStartTransitionsToPlaying(t *testing.T) {
	g := newTestGame()
	g.HandleInput("Enter", true)
	g.Tick()
	if g.State != StatePlaying {
		t.Fatalf("state = %v, want StatePlaying", g.State)
	}
}

func TestStartIgnoresOtherKeys(t *testing.T) {
	g := newTestGame()
	g.HandleInput("Space", true)
	g.Tick()
	if g.State != StateStart {
		t.Fatalf("state = %v, want StateStart", g.State)
	}
}

func TestGameOverRestarts(t *testing.T) {
	g := newTestGame()
	g.State = StateGameOver
	g.HandleInput("Enter", true)
	g.Tick()
	if g.State != StatePlaying {
		t.Fatalf("state = %v, want StatePlaying", g.State)
	}
	if g.Score != 0 || g.Lives != StartLives || g.Level != 1 {
		t.Fatalf("reset failed: score=%d lives=%d level=%d", g.Score, g.Lives, g.Level)
	}
}

func TestLevelTransition(t *testing.T) {
	g := newTestGame()
	g.State = StateLevelTransition
	g.TransitionTimer = 1
	g.Level = 3
	g.Tick()
	if g.State != StatePlaying {
		t.Fatalf("state = %v, want StatePlaying", g.State)
	}
	if g.Level != 4 {
		t.Fatalf("level = %d, want 4", g.Level)
	}
}

func TestTickDoesNothingWhenPaused(t *testing.T) {
	g := newTestGame()
	g.State = StatePlaying
	g.Input.Left = true
	g.State = StatePaused
	x := g.Player.X
	g.Tick()
	if g.Player.X != x {
		t.Fatalf("player moved while paused: %d -> %d", x, g.Player.X)
	}
	if g.State != StatePaused {
		t.Fatalf("state = %v, want StatePaused", g.State)
	}
}

func TestGameOverWhenLivesExhausted(t *testing.T) {
	g := newTestGame()
	g.State = StatePlaying
	g.Lives = 1
	g.HandlePlayerDeath()
	if g.State != StateGameOver {
		t.Fatalf("state = %v, want StateGameOver", g.State)
	}
}

func TestGameOverWhenInvaderReachesBottom(t *testing.T) {
	g := newTestGame()
	g.State = StatePlaying
	g.Invaders.Invaders[0][0].Y = PlayerY
	g.Tick()
	if g.State != StateGameOver {
		t.Fatalf("state = %v, want StateGameOver", g.State)
	}
}

func TestSeededRNGIsDeterministic(t *testing.T) {
	g1 := newTestGame()
	g2 := newTestGame()
	u1 := NewUFO(g1.RNG)
	u2 := NewUFO(g2.RNG)
	if u1.X != u2.X || u1.Dir != u2.Dir || u1.Points != u2.Points {
		t.Fatalf("RNG not deterministic: %+v vs %+v", u1, u2)
	}
}

func TestPauseToggle(t *testing.T) {
	g := newTestGame()
	g.State = StatePlaying
	g.HandleInput("KeyP", true)
	if g.State != StatePaused {
		t.Fatalf("state = %v, want StatePaused", g.State)
	}
	g.HandleInput("KeyP", true)
	if g.State != StatePlaying {
		t.Fatalf("state = %v, want StatePlaying", g.State)
	}
}

func TestFullGameLoop200Frames(t *testing.T) {
	g := newTestGame()
	g.HandleInput("Enter", true)
	g.Tick()
	if g.State != StatePlaying {
		t.Fatalf("state = %v, want StatePlaying", g.State)
	}
	// Keep the player invulnerable so enemy fire can't end the sim early.
	g.Player.Invulnerable = 300
	y0 := g.Invaders.Invaders[InvaderRows-1][0].Y
	for range 200 {
		g.Tick()
	}
	if g.State != StatePlaying {
		t.Fatalf("state = %v after 200 frames, want StatePlaying", g.State)
	}
	if g.Invaders.Invaders[InvaderRows-1][0].Y != y0+invaderDrop {
		t.Fatalf("invader y = %d, want %d (one edge drop in 200 frames)",
			g.Invaders.Invaders[InvaderRows-1][0].Y, y0+invaderDrop)
	}
	g.HandleInput("Space", true)
	g.Tick()
	if ActiveCount(g.Bullets, BulletPlayer) != 1 {
		t.Fatal("player bullet should be in flight after firing")
	}
}

func TestLevelCompleteTriggersTransition(t *testing.T) {
	g := newTestGame()
	g.State = StatePlaying
	for r := range g.Invaders.Invaders {
		for c := range g.Invaders.Invaders[r] {
			g.Invaders.Invaders[r][c].Alive = false
		}
	}
	g.Tick()
	if g.State != StateLevelTransition {
		t.Fatalf("state = %v, want StateLevelTransition", g.State)
	}
	if g.TransitionTimer != TransitionFrames {
		t.Fatalf("transition timer = %d, want %d", g.TransitionTimer, TransitionFrames)
	}
}

func TestUFOSpawnAfterExpectedFrames(t *testing.T) {
	g := newTestGame()
	g.State = StatePlaying
	// Invulnerable so the long sim can't end in a game over.
	g.Player.Invulnerable = UFOSpawnFrames + 100
	for i := range UFOSpawnFrames - 1 {
		g.Tick()
		if g.UFOActive {
			t.Fatalf("UFO active at frame %d, want %d", i+1, UFOSpawnFrames)
		}
	}
	g.Tick()
	if !g.UFOActive {
		t.Fatal("UFO should be active after UFOSpawnFrames")
	}
	if g.State != StatePlaying {
		t.Fatalf("state = %v, want StatePlaying", g.State)
	}
}
