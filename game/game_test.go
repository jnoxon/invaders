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
	if u1.X != u2.X || u1.Dir != u2.Dir || u1.Points() != u2.Points() {
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

func TestPauseResumeMethods(t *testing.T) {
	g := playingGame()
	g.Pause()
	if g.State != StatePaused {
		t.Fatalf("state = %v, want StatePaused", g.State)
	}
	g.Pause()
	if g.State != StatePaused {
		t.Fatal("pause while paused should be a no-op")
	}
	g.Resume()
	if g.State != StatePlaying {
		t.Fatalf("state = %v, want StatePlaying", g.State)
	}
	g.Resume()
	if g.State != StatePlaying {
		t.Fatal("resume while playing should be a no-op")
	}
	s := newTestGame()
	s.Pause()
	if s.State != StateStart {
		t.Fatalf("pause in start state = %v, want StateStart", s.State)
	}
}

func TestPauseFreezesPlayer(t *testing.T) {
	g := playingGame()
	g.HandleInput("ArrowLeft", true)
	g.Pause()
	x := g.Player.X
	g.Tick()
	if g.Player.X != x {
		t.Fatalf("player moved while paused: %d -> %d", x, g.Player.X)
	}
	g.Resume()
	g.Tick()
	if g.Player.X != x-PlayerSpeed {
		t.Fatalf("player X = %d, want %d after resume", g.Player.X, x-PlayerSpeed)
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

// killOne kills one alive invader with a player bullet, exercising the
// real hit path (score, KillCount, spawn check).
func killOne(t *testing.T, g *Game) {
	t.Helper()
	if g.Invaders.AliveCount() == 0 {
		t.Fatal("no invaders left to kill")
	}
	for r := range g.Invaders.Invaders {
		for c := range g.Invaders.Invaders[r] {
			iv := &g.Invaders.Invaders[r][c]
			if !iv.Alive {
				continue
			}
			g.Bullets = []Bullet{{X: iv.X + 2, Y: iv.Y + 2, Owner: BulletPlayer, Active: true}}
			g.CheckCollisions()
			if iv.Alive {
				t.Fatalf("invader (%d,%d) still alive after bullet", r, c)
			}
			return
		}
	}
}

func TestUFOSpawnsAtTwentyKills(t *testing.T) {
	g := newTestGame()
	g.State = StatePlaying
	for i := 1; i < UFOSpawnKills; i++ {
		killOne(t, g)
		if g.UFOActive {
			t.Fatalf("UFO active after %d kills, want %d", i, UFOSpawnKills)
		}
	}
	killOne(t, g)
	if g.KillCount != UFOSpawnKills {
		t.Fatalf("kill count = %d, want %d", g.KillCount, UFOSpawnKills)
	}
	g.trySpawnUFO()
	if !g.UFOActive {
		t.Fatal("UFO should be active after 20 kills")
	}
	if !validUFOPoints[g.UFO.Points()] {
		t.Fatalf("ufo points = %d not valid", g.UFO.Points())
	}
}

func TestUFODoesNotSpawnBelowTenAlive(t *testing.T) {
	g := newTestGame()
	g.State = StatePlaying
	// Silently kill 46 invaders, leaving UFOMinAlive-1 alive.
	n := 0
	for r := range g.Invaders.Invaders {
		for c := range g.Invaders.Invaders[r] {
			if n < InvaderRows*InvaderCols-UFOMinAlive {
				g.Invaders.Invaders[r][c].Alive = false
				n++
			}
		}
	}
	g.KillCount = UFOSpawnKills - 1
	killOne(t, g)
	if g.KillCount != UFOSpawnKills {
		t.Fatalf("kill count = %d, want %d", g.KillCount, UFOSpawnKills)
	}
	g.trySpawnUFO()
	if g.UFOActive {
		t.Fatal("UFO should not spawn with fewer than 10 invaders alive")
	}
}

func TestKillCountResetsOnLevelStart(t *testing.T) {
	g := newTestGame()
	g.State = StatePlaying
	g.Player.Invulnerable = 300
	for range 25 {
		killOne(t, g)
	}
	if g.KillCount != 25 {
		t.Fatalf("kill count = %d, want 25", g.KillCount)
	}
	for r := range g.Invaders.Invaders {
		for c := range g.Invaders.Invaders[r] {
			g.Invaders.Invaders[r][c].Alive = false
		}
	}
	g.Tick()
	if g.State != StateLevelTransition {
		t.Fatalf("state = %v, want StateLevelTransition", g.State)
	}
	g.TransitionTimer = 1
	g.Tick()
	if g.State != StatePlaying {
		t.Fatalf("state = %v, want StatePlaying", g.State)
	}
	if g.KillCount != 0 {
		t.Fatalf("kill count = %d, want 0 after level start", g.KillCount)
	}
	if g.ufoNextKill != UFOSpawnKills {
		t.Fatalf("ufo threshold = %d, want %d", g.ufoNextKill, UFOSpawnKills)
	}
}

func TestFullLevelAllInvadersKilledTransitions(t *testing.T) {
	g := newTestGame()
	g.State = StatePlaying
	g.Player.Invulnerable = 1000
	for range InvaderRows * InvaderCols {
		killOne(t, g)
	}
	if g.KillCount != InvaderRows*InvaderCols {
		t.Fatalf("kill count = %d, want %d", g.KillCount, InvaderRows*InvaderCols)
	}
	g.Tick()
	if g.State != StateLevelTransition {
		t.Fatalf("state = %v, want StateLevelTransition", g.State)
	}
	if g.TransitionTimer != TransitionFrames {
		t.Fatalf("transition timer = %d, want %d", g.TransitionTimer, TransitionFrames)
	}
}

func TestBarricadeErosionOverMultipleHits(t *testing.T) {
	g := newTestGame()
	g.State = StatePlaying
	bar := &g.Barricades[0]
	count := bar.PixelCount()
	// Five filled pixels, pairwise non-adjacent, so each hit clears exactly
	// the hit pixel plus one neighbor and never erases a later target.
	hits := [][2]int{{1, 0}, {4, 0}, {7, 0}, {1, 2}, {4, 2}}
	for i, h := range hits {
		sx := bar.X + 2*h[0]
		sy := bar.Y + 2*h[1]
		g.Bullets = []Bullet{{X: sx - 1, Y: sy, Owner: BulletPlayer, Active: true}}
		g.CheckCollisions()
		if g.Bullets[0].Active {
			t.Fatalf("hit %d: bullet should be consumed", i)
		}
		count -= 2
		if got := bar.PixelCount(); got != count {
			t.Fatalf("hit %d: pixel count = %d, want %d", i, got, count)
		}
	}
}
