//go:build js && wasm

package main

import (
	"fmt"
	"syscall/js"
)

const (
	logicalW = 256
	logicalH = 224
)

var ctx js.Value

func main() {
	obj := js.Global().Get("Object").New()
	obj.Set("tick", js.FuncOf(tick))
	obj.Set("setKey", js.FuncOf(setKey))
	js.Global().Set("main", obj)

	canvas := js.Global().Get("document").Call("getElementById", "game")
	ctx = canvas.Call("getContext", "2d")

	fmt.Println("invaders: ready")
	select {}
}

func tick(this js.Value, args []js.Value) any {
	ctx.Call("fillStyle", "#000000")
	ctx.Call("fillRect", 0, 0, logicalW, logicalH)
	return nil
}

func setKey(this js.Value, args []js.Value) any {
	code := args[0].String()
	pressed := args[1].Bool()
	fmt.Printf("key %s %v\n", code, pressed)
	return nil
}
