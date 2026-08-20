package game

import (
	"math/rand/v2"
	"time"
)

const (
	ScreenW = 256
	ScreenH = 224

	StartLives             = 3
	TransitionFrames       = 120
	InvaderBottomThreshold = PlayerY - 8
)

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

	// KillCount is the number of invaders killed this level.
	KillCount int
	// ufoNextKill is the kill count at which the next UFO may spawn.
	ufoNextKill int
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
		}
	case StatePaused:
	}
	g.Frame++
	g.Input.ClearJustPressed()
}

func (g *Game) updatePlaying() {
	g.Player.Update(&g.Input)
	g.tryFire()
	if g.Invaders.Update() && g.Invaders.ShouldShoot() {
		g.enemyShoot()
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

func (g *Game) tryFire() {
	if !g.Input.JustPressed["Space"] || !g.Player.Alive {
		return
	}
	if !CanFire(g.Bullets, BulletPlayer) {
		return
	}
	if b := g.Player.Fire(); b != nil {
		g.Bullets = append(g.Bullets, *b)
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
	}
}

func (g *Game) StartGame() {
	g.Score = 0
	g.Lives = StartLives
	g.Level = 1
	g.Player = NewPlayer()
	g.resetLevelEntities()
	g.State = StatePlaying
}

func (g *Game) ResetLevel() {
	g.Level++
	g.resetLevelEntities()
	g.Player.Respawn()
}

func (g *Game) resetLevelEntities() {
	g.Invaders = NewInvaderGrid(g.Level, g.RNG)
	g.Barricades = newBarricades()
	g.Bullets = g.Bullets[:0]
	g.UFO = UFO{}
	g.UFOActive = false
	g.KillCount = 0
	g.ufoNextKill = UFOSpawnKills
}

func (g *Game) HandleInput(code string, pressed bool) {
	g.Input.Update(code, pressed)
	if code == "KeyP" && pressed {
		if g.State == StatePlaying {
			g.State = StatePaused
		} else if g.State == StatePaused {
			g.State = StatePlaying
		}
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
}
