//go:build js && wasm

package main

import (
	"fmt"
	"syscall/js"

	"invaders/game"
)

const (
	logicalW = 256
	logicalH = 224
)

var (
	ctx js.Value
	g   *game.Game
)

func main() {
	g = game.NewGame()

	obj := js.Global().Get("Object").New()
	obj.Set("tick", js.FuncOf(tick))
	obj.Set("setKey", js.FuncOf(setKey))
	js.Global().Set("main", obj)

	canvas := js.Global().Get("document").Call("getElementById", "game")
	ctx = canvas.Call("getContext", "2d")

	select {}
}

func tick(this js.Value, args []js.Value) any {
	g.Tick()
	ctx.Set("fillStyle", "#000000")
	ctx.Call("fillRect", 0, 0, logicalW, logicalH)
	ctx.Set("font", "10px monospace")
	switch g.State {
	case game.StateStart:
		text("SPACE INVADERS", logicalW/2, 80, "center")
		text(fmt.Sprintf("HIGH SCORE: %d", g.HighScore), logicalW/2, 104, "center")
		text("PRESS ENTER TO START", logicalW/2, 128, "center")
	case game.StatePlaying:
		drawEntities()
	case game.StateLevelTransition:
		text(fmt.Sprintf("LEVEL %d", g.Level), logicalW/2, logicalH/2, "center")
	case game.StateGameOver:
		text("GAME OVER", logicalW/2, 80, "center")
		text(fmt.Sprintf("SCORE: %d", g.Score), logicalW/2, 104, "center")
		text("PRESS ENTER TO RESTART", logicalW/2, 128, "center")
	case game.StatePaused:
		drawEntities()
		text("PAUSED", logicalW/2, logicalH/2, "center")
	}
	return nil
}

func drawEntities() {
	for r := range g.Invaders.Invaders {
		for c := range g.Invaders.Invaders[r] {
			iv := &g.Invaders.Invaders[r][c]
			if !iv.Alive {
				continue
			}
			ctx.Set("fillStyle", invaderColor(iv.Type))
			ctx.Call("fillRect", iv.X, iv.Y, game.InvaderW, game.InvaderH)
		}
	}

	ctx.Set("fillStyle", "#808080")
	for i := range g.Barricades {
		bar := &g.Barricades[i]
		if bar.Destroyed() {
			continue
		}
		for r := range bar.Pixels {
			for c := range bar.Pixels[r] {
				if !bar.Pixels[r][c] {
					continue
				}
				ctx.Call("fillRect", bar.X+c*2, bar.Y+r*2, 2, 2)
			}
		}
	}

	if g.UFOActive {
		ctx.Set("fillStyle", "#FF00FF")
		x, y, w, h := g.UFO.Rect()
		ctx.Call("fillRect", x, y, w, h)
	}

	for i := range g.Bullets {
		b := &g.Bullets[i]
		if b.Owner == game.BulletPlayer {
			ctx.Set("fillStyle", "#FFFF00")
		} else {
			ctx.Set("fillStyle", "#FF0000")
		}
		ctx.Call("fillRect", b.X, b.Y, game.BulletW, game.BulletH)
	}

	if g.Player.Alive {
		ctx.Set("fillStyle", "#FFFFFF")
		ctx.Call("fillRect", g.Player.X, g.Player.Y, g.Player.W, g.Player.H)
	}

	for i := range g.Lives {
		ctx.Set("fillStyle", "#FFFFFF")
		ctx.Call("fillRect", 4+i*16, 216, 12, 6)
	}

	ctx.Set("fillStyle", "#FFFFFF")
	text(fmt.Sprintf("SCORE: %d", g.Score), 4, 12, "left")
	text(fmt.Sprintf("LV %d", g.Level), logicalW-4, 12, "right")
}

func invaderColor(t game.InvaderType) string {
	switch t {
	case game.InvaderSquid:
		return "#90EE90"
	case game.InvaderCrab:
		return "#00FF00"
	default:
		return "#006400"
	}
}

func text(s string, x, y int, align string) {
	ctx.Set("fillStyle", "#FFFFFF")
	ctx.Set("textAlign", align)
	ctx.Call("fillText", s, x, y)
	ctx.Set("textAlign", "left")
}

func setKey(this js.Value, args []js.Value) any {
	code := args[0].String()
	pressed := args[1].Bool()
	g.HandleInput(code, pressed)
	return nil
}
