//go:build js && wasm

package main

import (
	"syscall/js"

	"invaders/game"
	"invaders/render"
)

const tickMs = 1000.0 / 60.0

var (
	ctx js.Value
	g   *game.Game
	r   *render.Renderer

	lastMs  float64
	accumMs float64
)

// jsCanvas adapts the canvas 2D context to render.Canvas. fillStyle is
// managed by Clear: black while clearing, white for everything after.
type jsCanvas struct{}

func (jsCanvas) Clear() {
	ctx.Set("fillStyle", "#000000")
	ctx.Call("fillRect", 0, 0, game.ScreenW, game.ScreenH)
	ctx.Set("fillStyle", "#FFFFFF")
}

func (jsCanvas) FillRect(x, y, w, h int) {
	ctx.Call("fillRect", x, y, w, h)
}

func main() {
	g = game.NewGame()

	canvas := js.Global().Get("document").Call("getElementById", "game")
	ctx = canvas.Call("getContext", "2d")
	r = render.NewRenderer(jsCanvas{})

	obj := js.Global().Get("Object").New()
	obj.Set("tick", js.FuncOf(tick))
	obj.Set("setKey", js.FuncOf(setKey))
	js.Global().Set("main", obj)

	select {}
}

// tick advances the game with a fixed 60Hz timestep and renders one frame.
// now is performance.now() in milliseconds.
func tick(this js.Value, args []js.Value) any {
	now := args[0].Float()
	if lastMs == 0 {
		lastMs = now
	}
	accumMs += now - lastMs
	lastMs = now
	if accumMs > 250 {
		accumMs = 250
	}
	for accumMs >= tickMs {
		g.Tick()
		accumMs -= tickMs
	}
	r.Render(g)
	return nil
}

func setKey(this js.Value, args []js.Value) any {
	g.HandleInput(args[0].String(), args[1].Bool())
	return nil
}
