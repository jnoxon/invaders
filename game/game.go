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

	ufoTimer int
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
	if g.UFOActive {
		g.UFO.Update()
		if !g.UFO.Active {
			g.UFOActive = false
			g.ufoTimer = 0
		}
		return
	}
	g.ufoTimer++
	if g.ufoTimer >= UFOSpawnFrames {
		g.ufoTimer = 0
		g.UFO = NewUFO(g.RNG)
		g.UFOActive = true
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
	g.ufoTimer = 0
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
