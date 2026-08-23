package game

import (
	"math"
	"math/rand/v2"
	"time"
)

const (
	ScreenW = 256
	ScreenH = 224

	StartLives             = 3
	TransitionFrames       = 120
	InvaderBottomThreshold = PlayerY - 8

	DeathFlashFrames = 6
	ScorePopupFrames = 30
)

// GameEvent is a discrete game action reported after Tick so the host can
// react (sound, haptics) without the game package knowing about them.
type GameEvent int

const (
	EventFire GameEvent = iota + 1
	EventInvaderKilled
	EventPlayerHit
	EventUFOAppear
	EventUFODisappear
	EventUFOKilled
	EventMarch
	EventGameOver
	EventLevelStart
)

// ScorePopup is a points readout shown where a UFO was shot down.
type ScorePopup struct {
	X, Y, Points int
	Timer        int
}

type GameState int

const (
	StateStart GameState = iota
	StatePlaying
	StateGameOver
	StateLevelTransition
	StatePaused
)

type Game struct {
	State           GameState
	Player          Player
	Invaders        InvaderGrid
	Bullets         []Bullet
	Barricades      []Barricade
	UFO             UFO
	UFOActive       bool
	Score           int
	HighScore       int
	Lives           int
	Level           int
	Frame           int
	TransitionTimer int
	RNG             *rand.Rand
	Input           InputState

	// Events accumulates GameEvents during Tick; the host drains it.
	Events []GameEvent
	// Flash is the number of frames of white screen left after a player death.
	Flash int
	// ScorePopup shows points at a UFO kill location until Timer hits 0.
	ScorePopup ScorePopup

	// KillCount is the number of invaders killed this level.
	KillCount int
	// ufoNextKill is the kill count at which the next UFO may spawn.
	ufoNextKill int
	// moveDx is the queued touch drag delta in logical px; see applyMoveDx.
	moveDx float64
}

func NewGame() *Game {
	g := &Game{
		State:     StateStart,
		RNG:       rand.New(rand.NewPCG(1, uint64(time.Now().UnixNano()))),
		Score:     0,
		HighScore: 0,
		Lives:     StartLives,
		Level:     1,
		Input:     NewInputState(),
	}
	g.Player = NewPlayer()
	g.Invaders = NewInvaderGrid(1, g.RNG)
	g.Barricades = newBarricades()
	return g
}

func (g *Game) Tick() {
	// Decay visual timers before state updates so an effect set this tick
	// survives a full frame, in every state (a death on the last life would
	// otherwise freeze the flash over the game-over screen).
	if g.Flash > 0 {
		g.Flash--
	}
	if g.ScorePopup.Timer > 0 {
		g.ScorePopup.Timer--
	}
	switch g.State {
	case StateStart:
		if g.Input.JustPressed["Enter"] {
			g.StartGame()
		}
	case StatePlaying:
		g.updatePlaying()
	case StateGameOver:
		if g.Input.JustPressed["Enter"] {
			g.StartGame()
		}
	case StateLevelTransition:
		g.TransitionTimer--
		if g.TransitionTimer <= 0 {
			g.ResetLevel()
			g.State = StatePlaying
			g.emit(EventLevelStart)
		}
	case StatePaused:
	}
	g.Frame++
	g.Input.ClearJustPressed()
}

func (g *Game) updatePlaying() {
	g.Player.Update(&g.Input)
	g.applyMoveDx()
	g.tryFire()
	if g.Invaders.Update() {
		g.emit(EventMarch)
		if g.Invaders.ShouldShoot() {
			g.enemyShoot()
		}
	}
	g.updateBullets()
	g.updateUFO()
	g.CheckCollisions()
	g.trySpawnUFO()

	if g.Player.Invulnerable > 0 {
		g.Player.Invulnerable--
	}

	if g.Invaders.ReachedBottom() {
		g.GameOver()
		return
	}
	if g.Invaders.AliveCount() == 0 {
		g.HandleLevelComplete()
	}
}

// AddMoveDx queues a touch drag delta in logical px. The fixed tick
// consumes it; see applyMoveDx.
func (g *Game) AddMoveDx(dx float64) {
	g.moveDx += dx
}

// applyMoveDx consumes the queued touch drag. The rounded amount moves the
// ship; the fractional remainder carries to the next tick. Any amount the
// playfield clamp eats is kept as debt, so the ship tracks the anchored
// finger target (finger minus anchor) and cannot lead it — teleporting —
// when the edge frees up. While the player is dead the queue is dropped.
func (g *Game) applyMoveDx() {
	if g.moveDx == 0 {
		return
	}
	if !g.Player.Alive {
		g.moveDx = 0
		return
	}
	d := int(math.Round(g.moveDx))
	g.moveDx -= float64(d)
	if d == 0 {
		return
	}
	x := g.Player.X + d
	if x < 0 {
		x = 0
	}
	if x > ScreenW-PlayerW {
		x = ScreenW - PlayerW
	}
	if applied := x - g.Player.X; applied != d {
		g.moveDx += float64(d - applied)
	}
	g.Player.X = x
}

// EndMoveDx drops the queued touch drag, including clamped-edge debt. The
// JS calls it when the drag pointer lifts so a stale gesture cannot steer
// the next one.
func (g *Game) EndMoveDx() {
	g.moveDx = 0
}

func (g *Game) tryFire() {
	if !g.Input.JustPressed["Space"] || !g.Player.Alive {
		return
	}
	if !CanFire(g.Bullets, BulletPlayer) {
		return
	}
	if b := g.Player.Fire(); b != nil {
		g.Bullets = append(g.Bullets, *b)
		g.emit(EventFire)
	}
}

func (g *Game) enemyShoot() {
	if !CanFire(g.Bullets, BulletEnemy) {
		return
	}
	iv := g.Invaders.PickShooter()
	if iv == nil {
		return
	}
	g.Bullets = append(g.Bullets, Bullet{
		X:      iv.X + InvaderW/2 - BulletW/2,
		Y:      iv.Y + InvaderH,
		Owner:  BulletEnemy,
		Active: true,
	})
}

func (g *Game) updateBullets() {
	n := 0
	for i := range g.Bullets {
		b := &g.Bullets[i]
		if b.Active {
			b.Update()
		}
		if b.Active {
			g.Bullets[n] = *b
			n++
		}
	}
	g.Bullets = g.Bullets[:n]
}

func (g *Game) updateUFO() {
	if !g.UFOActive {
		return
	}
	g.UFO.Update()
	if !g.UFO.Active {
		g.UFOActive = false
		g.emit(EventUFODisappear)
	}
}

// trySpawnUFO spawns a UFO once the kill count reaches the next multiple of
// UFOSpawnKills. The threshold advances on each spawn so a stalled kill
// count never re-spawns the UFO right after it leaves the screen.
func (g *Game) trySpawnUFO() {
	if g.KillCount >= g.ufoNextKill && g.UFO.CanSpawn(g.Invaders.AliveCount()) {
		g.UFO = NewUFO(g.RNG)
		g.UFOActive = true
		g.ufoNextKill += UFOSpawnKills
		g.emit(EventUFOAppear)
	}
}

func (g *Game) StartGame() {
	g.Score = 0
	g.Lives = StartLives
	g.Level = 1
	g.Player = NewPlayer()
	g.resetLevelEntities()
	g.State = StatePlaying
	g.moveDx = 0
	g.emit(EventLevelStart)
}

func (g *Game) ResetLevel() {
	g.Level++
	g.resetLevelEntities()
	g.Player.Respawn()
	g.moveDx = 0
}

func (g *Game) resetLevelEntities() {
	g.Invaders = NewInvaderGrid(g.Level, g.RNG)
	g.Barricades = newBarricades()
	g.Bullets = g.Bullets[:0]
	g.UFO = UFO{}
	g.UFOActive = false
	g.KillCount = 0
	g.ufoNextKill = UFOSpawnKills
	g.Flash = 0
	g.ScorePopup = ScorePopup{}
	g.Events = nil
}

func (g *Game) HandleInput(code string, pressed bool) {
	g.Input.Update(code, pressed)
	if code == "KeyP" && pressed {
		if g.State == StatePlaying {
			g.Pause()
		} else if g.State == StatePaused {
			g.Resume()
		}
	}
}

// Pause suspends gameplay while active.
func (g *Game) Pause() {
	if g.State == StatePlaying {
		g.State = StatePaused
		g.moveDx = 0
	}
}

// Resume resumes gameplay while paused.
func (g *Game) Resume() {
	if g.State == StatePaused {
		g.State = StatePlaying
		g.moveDx = 0
	}
}

func (g *Game) CheckCollisions() {
	g.checkPlayerBulletHits()
	g.checkEnemyBulletHits()
	g.checkInvaderBarricadeCollision()
}

func (g *Game) GameOver() {
	g.State = StateGameOver
	g.UpdateHighScore()
	g.moveDx = 0
	g.emit(EventGameOver)
}

func (g *Game) emit(e GameEvent) {
	g.Events = append(g.Events, e)
}
