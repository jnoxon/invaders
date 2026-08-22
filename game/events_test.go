package game

import (
	"testing"
)

func hasEvent(events []GameEvent, want GameEvent) bool {
	for _, e := range events {
		if e == want {
			return true
		}
	}
	return false
}

func playingGame() *Game {
	g := newTestGame()
	g.State = StatePlaying
	return g
}

func TestFireEmitsEvent(t *testing.T) {
	g := playingGame()
	g.HandleInput("Space", true)
	g.Tick()
	if !hasEvent(g.Events, EventFire) {
		t.Fatalf("events = %v, want EventFire", g.Events)
	}
	if ActiveCount(g.Bullets, BulletPlayer) != 1 {
		t.Fatalf("player bullets = %d, want 1", ActiveCount(g.Bullets, BulletPlayer))
	}
}

func TestFireNotEmittedWhenBulletInFlight(t *testing.T) {
	g := playingGame()
	g.Bullets = []Bullet{{X: 100, Y: 100, Owner: BulletPlayer, Active: true}}
	g.HandleInput("Space", true)
	g.Tick()
	if hasEvent(g.Events, EventFire) {
		t.Error("EventFire emitted with a bullet already in flight")
	}
}

func TestInvaderKillEmitsEvent(t *testing.T) {
	g := playingGame()
	iv := &g.Invaders.Invaders[0][0]
	g.Bullets = []Bullet{{X: iv.X, Y: iv.Y, Owner: BulletPlayer, Active: true}}
	g.Tick()
	if iv.Alive {
		t.Fatal("invader should be dead")
	}
	if !hasEvent(g.Events, EventInvaderKilled) {
		t.Fatalf("events = %v, want EventInvaderKilled", g.Events)
	}
	if g.KillCount != 1 {
		t.Fatalf("KillCount = %d, want 1", g.KillCount)
	}
}

func TestMarchEmitsOnStep(t *testing.T) {
	g := playingGame()
	var seen bool
	for i := 0; i < 120 && !seen; i++ {
		g.Tick()
		seen = hasEvent(g.Events, EventMarch)
		g.Events = nil
	}
	if !seen {
		t.Fatal("no EventMarch within 120 ticks")
	}
}

func TestUFOAppearEmitsEvent(t *testing.T) {
	g := playingGame()
	g.KillCount = UFOSpawnKills
	g.Tick()
	if !g.UFOActive {
		t.Fatal("UFO should be active")
	}
	if !hasEvent(g.Events, EventUFOAppear) {
		t.Fatalf("events = %v, want EventUFOAppear", g.Events)
	}
}

func TestUFOExitEmitsEvent(t *testing.T) {
	g := playingGame()
	g.Player.Invulnerable = 10000
	g.ufoNextKill = UFOSpawnKills
	g.UFO = NewUFO(testRNG())
	g.UFO.X = 0
	g.UFO.Dir = 1
	g.UFOActive = true
	var seen bool
	for i := 0; i < 200 && !seen; i++ {
		g.Tick()
		seen = hasEvent(g.Events, EventUFODisappear)
		g.Events = nil
	}
	if !seen {
		t.Fatal("no EventUFODisappear after UFO left the screen")
	}
	if g.UFOActive {
		t.Error("UFOActive still set after exit")
	}
}

func TestUFOKilledEmitsEventAndPopup(t *testing.T) {
	g := playingGame()
	g.ufoNextKill = UFOSpawnKills
	// x=28 falls in the gap between invader columns 0 and 1, so the
	// bullet reaches the UFO without hitting an invader. The UFO moves
	// one step (to x=26) before the collision check, and the popup
	// records the UFO's position at the moment it was killed.
	ufo := NewUFO(testRNG())
	ufo.X = 28
	ufo.Dir = -1
	g.UFO = ufo
	g.UFOActive = true
	g.Bullets = []Bullet{{X: 28, Y: UFOY, Owner: BulletPlayer, Active: true}}
	g.Tick()
	if g.UFOActive {
		t.Fatal("UFO should be inactive")
	}
	if !hasEvent(g.Events, EventUFOKilled) {
		t.Fatalf("events = %v, want EventUFOKilled", g.Events)
	}
	if g.Score != ufo.Points() {
		t.Fatalf("score = %d, want %d", g.Score, ufo.Points())
	}
	want := ScorePopup{X: 26, Y: UFOY, Points: ufo.Points(), Timer: ScorePopupFrames}
	if g.ScorePopup != want {
		t.Fatalf("popup = %+v, want %+v", g.ScorePopup, want)
	}
}

func TestPlayerHitEmitsEventAndFlash(t *testing.T) {
	g := playingGame()
	g.Player.Invulnerable = 0
	g.Bullets = []Bullet{{X: g.Player.X, Y: g.Player.Y, Owner: BulletEnemy, Active: true}}
	g.Tick()
	if !hasEvent(g.Events, EventPlayerHit) {
		t.Fatalf("events = %v, want EventPlayerHit", g.Events)
	}
	if g.Flash != DeathFlashFrames {
		t.Fatalf("Flash = %d, want %d", g.Flash, DeathFlashFrames)
	}
	if g.Lives != StartLives-1 {
		t.Fatalf("lives = %d, want %d", g.Lives, StartLives-1)
	}
	g.Tick()
	if g.Flash != DeathFlashFrames-1 {
		t.Fatalf("Flash after tick = %d, want %d", g.Flash, DeathFlashFrames-1)
	}
}

func TestFlashDecaysAfterGameOver(t *testing.T) {
	g := playingGame()
	g.Lives = 1
	g.Player.Invulnerable = 0
	g.Bullets = []Bullet{{X: g.Player.X, Y: g.Player.Y, Owner: BulletEnemy, Active: true}}
	g.Tick()
	if g.State != StateGameOver {
		t.Fatalf("state = %v, want StateGameOver", g.State)
	}
	if g.Flash != DeathFlashFrames {
		t.Fatalf("Flash = %d, want %d", g.Flash, DeathFlashFrames)
	}
	for range DeathFlashFrames {
		g.Tick()
	}
	if g.Flash != 0 {
		t.Fatalf("Flash = %d, want 0 after %d ticks", g.Flash, DeathFlashFrames)
	}
	if g.State != StateGameOver {
		t.Fatalf("state = %v, want StateGameOver", g.State)
	}
}

func TestPlayerHitIgnoredWhileInvulnerable(t *testing.T) {
	g := playingGame()
	g.Player.Invulnerable = RespawnInvuln
	g.Bullets = []Bullet{{X: g.Player.X, Y: g.Player.Y, Owner: BulletEnemy, Active: true}}
	g.Tick()
	if hasEvent(g.Events, EventPlayerHit) {
		t.Error("EventPlayerHit during invulnerability")
	}
	if g.Lives != StartLives {
		t.Errorf("lives = %d, want %d", g.Lives, StartLives)
	}
}

func TestLastLifeHitEmitsGameOver(t *testing.T) {
	g := playingGame()
	g.Lives = 1
	g.Player.Invulnerable = 0
	g.Bullets = []Bullet{{X: g.Player.X, Y: g.Player.Y, Owner: BulletEnemy, Active: true}}
	g.Tick()
	if g.State != StateGameOver {
		t.Fatalf("state = %v, want StateGameOver", g.State)
	}
	if !hasEvent(g.Events, EventPlayerHit) || !hasEvent(g.Events, EventGameOver) {
		t.Fatalf("events = %v, want EventPlayerHit and EventGameOver", g.Events)
	}
}

func TestInvadersReachBottomEmitsGameOver(t *testing.T) {
	g := playingGame()
	g.Invaders.moveY(200)
	g.Tick()
	if g.State != StateGameOver {
		t.Fatalf("state = %v, want StateGameOver", g.State)
	}
	if !hasEvent(g.Events, EventGameOver) {
		t.Fatalf("events = %v, want EventGameOver", g.Events)
	}
}

func TestStartGameEmitsLevelStart(t *testing.T) {
	g := newTestGame()
	g.HandleInput("Enter", true)
	g.Tick()
	if !hasEvent(g.Events, EventLevelStart) {
		t.Fatalf("events = %v, want EventLevelStart", g.Events)
	}
}

func TestTransitionEndEmitsLevelStart(t *testing.T) {
	g := newTestGame()
	g.State = StateLevelTransition
	g.TransitionTimer = 1
	g.Tick()
	if !hasEvent(g.Events, EventLevelStart) {
		t.Fatalf("events = %v, want EventLevelStart", g.Events)
	}
}

func TestScorePopupDecays(t *testing.T) {
	g := playingGame()
	g.ScorePopup = ScorePopup{X: 50, Y: 40, Points: 100, Timer: 2}
	g.Tick()
	g.Tick()
	if g.ScorePopup.Timer != 0 {
		t.Fatalf("popup timer = %d, want 0", g.ScorePopup.Timer)
	}
	if g.ScorePopup.Points != 100 {
		t.Fatalf("popup points = %d, want 100 (timer expires, value stays)", g.ScorePopup.Points)
	}
}

func TestRestartClearsFlashPopupAndEvents(t *testing.T) {
	g := playingGame()
	g.Flash = 3
	g.ScorePopup = ScorePopup{X: 5, Y: 5, Points: 50, Timer: 5}
	g.Events = []GameEvent{EventFire}
	g.StartGame()
	if g.Flash != 0 {
		t.Errorf("Flash = %d, want 0", g.Flash)
	}
	if g.ScorePopup != (ScorePopup{}) {
		t.Errorf("popup = %+v, want zero", g.ScorePopup)
	}
	if hasEvent(g.Events, EventFire) {
		t.Error("stale events survived restart")
	}
}
