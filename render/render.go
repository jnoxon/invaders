package render

import (
	"fmt"

	"invaders/game"
)

// groundY is the 1px ground line above the lives row.
const groundY = 214

// lifeY is the top of the remaining-lives icons.
const lifeY = 217

// Canvas is the drawing surface. Implementations must render in a single
// color (white) on black.
type Canvas interface {
	// Clear fills the whole screen with the background color.
	Clear()
	// FillRect draws a filled rectangle at logical 256x224 coordinates.
	FillRect(x, y, w, h int)
}

// Renderer draws game state to a Canvas at the logical 256x224 resolution.
// Scaling to CSS pixels is done by the host, not here.
type Renderer struct {
	c Canvas
}

func NewRenderer(c Canvas) *Renderer {
	return &Renderer{c: c}
}

// Render clears the screen and draws the current game state.
func (r *Renderer) Render(g *game.Game) {
	r.c.Clear()
	switch g.State {
	case game.StateStart:
		r.renderStart(g)
	case game.StatePlaying, game.StatePaused:
		r.renderGameplay(g)
		if g.State == game.StatePaused {
			r.drawTextCenter("PAUSED", 104)
		}
	case game.StateGameOver:
		r.renderGameOver(g)
	case game.StateLevelTransition:
		r.renderTransition(g)
	}
}

func (r *Renderer) renderGameplay(g *game.Game) {
	r.drawHUD(g)

	for row := range g.Invaders.Invaders {
		for col := range g.Invaders.Invaders[row] {
			iv := &g.Invaders.Invaders[row][col]
			if !iv.Alive {
				continue
			}
			r.drawSprite(invaderSprite(iv.Type, iv.AnimFrame), iv.X, iv.Y)
		}
	}

	for i := range g.Barricades {
		b := &g.Barricades[i]
		for py := range game.BarricadePixelH {
			for px := range game.BarricadePixelW {
				if b.Pixels[py][px] {
					r.c.FillRect(b.X+px*2, b.Y+py*2, 2, 2)
				}
			}
		}
	}

	if g.UFOActive {
		r.drawSprite(ufoSprite, g.UFO.X, g.UFO.Y)
	}

	for i := range g.Bullets {
		b := &g.Bullets[i]
		if b.Owner == game.BulletPlayer {
			r.c.FillRect(b.X, b.Y, game.BulletW, game.BulletH)
		} else if b.Y&1 == 0 {
			r.drawSprite(enemyBulletF0, b.X, b.Y)
		} else {
			r.drawSprite(enemyBulletF1, b.X, b.Y)
		}
	}

	if g.Player.Alive {
		if game.ActiveCount(g.Bullets, game.BulletPlayer) > 0 {
			r.drawSprite(playerFiring, g.Player.X, g.Player.Y)
		} else {
			r.drawSprite(playerShip, g.Player.X, g.Player.Y)
		}
	}

	r.c.FillRect(0, groundY, game.ScreenW, 1)
	for i := range g.Lives {
		r.drawSprite(lifeIcon, 4+i*16, lifeY)
	}
}

func (r *Renderer) drawHUD(g *game.Game) {
	r.drawText(fmt.Sprintf("SCORE: %05d", g.Score), 4, 2)
	r.drawTextCenter(fmt.Sprintf("HI-SCORE: %05d", g.HighScore), 2)
	lv := fmt.Sprintf("LV: %d", g.Level)
	r.drawText(lv, game.ScreenW-4-PixelFont.TextWidth(lv), 2)
}

func (r *Renderer) renderStart(g *game.Game) {
	r.drawTextCenter("SPACE INVADERS", 16)
	r.drawTextCenter(fmt.Sprintf("HIGH SCORE: %05d", g.HighScore), 40)

	r.drawSprite(ufoSprite, 72, 64)
	r.drawText("???", 124, 68)
	r.drawSprite(squidF0, 88, 88)
	r.drawText("= 30", 124, 92)
	r.drawSprite(crabF0, 88, 108)
	r.drawText("= 20", 124, 112)
	r.drawSprite(octopusF0, 88, 128)
	r.drawText("= 10", 124, 132)

	if blink(g.Frame) {
		r.drawTextCenter("PRESS ENTER TO START", 168)
	}
}

func (r *Renderer) renderGameOver(g *game.Game) {
	r.drawTextCenter("GAME OVER", 72)
	r.drawTextCenter(fmt.Sprintf("SCORE: %05d", g.Score), 96)
	if blink(g.Frame) {
		r.drawTextCenter("PRESS ENTER TO RESTART", 120)
	}
}

func (r *Renderer) renderTransition(g *game.Game) {
	r.drawTextCenter(fmt.Sprintf("LEVEL %d", g.Level), 104)
}

// blink is on for the first half of each 1-second cycle.
func blink(frame int) bool {
	return frame%60 < 30
}

// drawSprite draws filled pixels as horizontal runs (one FillRect each).
func (r *Renderer) drawSprite(s Sprite, x, y int) {
	for py := range s.H {
		run, start := 0, 0
		for px := range s.W {
			if s.Data[py*s.W+px] != 0 {
				if run == 0 {
					start = px
				}
				run++
			} else if run > 0 {
				r.c.FillRect(x+start, y+py, run, 1)
				run = 0
			}
		}
		if run > 0 {
			r.c.FillRect(x+start, y+py, run, 1)
		}
	}
}

// drawText draws s with the pixel font at (x, y), 1px between characters.
func (r *Renderer) drawText(s string, x, y int) {
	for i := range len(s) {
		g, ok := PixelFont.Glyphs[s[i]]
		if !ok {
			continue
		}
		for row := range PixelFont.H {
			bits := g[row]
			if bits == 0 {
				continue
			}
			for col := range PixelFont.W {
				if bits&(1<<uint(PixelFont.W-1-col)) != 0 {
					r.c.FillRect(x+i*(PixelFont.W+1)+col, y+row, 1, 1)
				}
			}
		}
	}
}

func (r *Renderer) drawTextCenter(s string, y int) {
	r.drawText(s, (game.ScreenW-PixelFont.TextWidth(s))/2, y)
}
