//go:build js && wasm

package main

import (
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
	ctx.Call("fillStyle", "#000000")
	ctx.Call("fillRect", 0, 0, logicalW, logicalH)
	return nil
}

func setKey(this js.Value, args []js.Value) any {
	code := args[0].String()
	pressed := args[1].Bool()
	g.HandleInput(code, pressed)
	return nil
}
