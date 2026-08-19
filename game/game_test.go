package game

import (
	"math/rand"
	"testing"
)

func newTestGame() *Game {
	g := NewGame()
	g.RNG = rand.New(rand.NewSource(42))
	return g
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
