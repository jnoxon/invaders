package game

import (
	"math/rand"
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

	levelKills       int
	enemyFireCounter int
}

func NewGame() *Game {
	g := &Game{
		State:     StateStart,
		RNG:       rand.New(rand.NewSource(time.Now().UnixNano())),
		Score:     0,
		HighScore: 0,
		Lives:     StartLives,
		Level:     1,
		Input:     NewInputState(),
	}
	g.Player = NewPlayer()
	g.Invaders = NewInvaderGrid(1)
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
	g.Invaders.Update()
	g.updateBullets()
	g.updateUFO()
	g.enemyFire()
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
	if !g.Input.Fire || !g.Player.Alive {
		return
	}
	if countBullets(g.Bullets, BulletPlayer) >= MaxPlayerBullets {
		return
	}
	if b := g.Player.Fire(); b != nil {
		g.Bullets = append(g.Bullets, *b)
	}
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
		if g.levelKills >= UFOKillThreshold {
			g.UFO = NewUFO(g.RNG)
			g.UFOActive = true
		}
		return
	}
	g.UFO.Update()
	if !g.UFO.Active {
		g.UFOActive = false
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
	g.Invaders = NewInvaderGrid(g.Level)
	g.Barricades = newBarricades()
	g.Bullets = g.Bullets[:0]
	g.UFO = UFO{}
	g.UFOActive = false
	g.levelKills = 0
	g.enemyFireCounter = 0
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
