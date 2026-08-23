//go:build js && wasm

package main

import (
	"strconv"
	"syscall/js"

	"invaders/audio"
	"invaders/game"
	"invaders/render"
)

const tickMs = 1000.0 / 60.0

const highScoreKey = "si-highscore"

var (
	ctx js.Value
	g   *game.Game
	r   *render.Renderer
	au  *audio.Audio
	buf *render.Buffer

	jsPix js.Value
	jsImg js.Value

	lastMs  float64
	accumMs float64

	lastHigh int
)

func main() {
	g = game.NewGame()
	loadHighScore()

	canvas := js.Global().Get("document").Call("getElementById", "game")
	ctx = canvas.Call("getContext", "2d")

	buf = render.NewBuffer()
	r = render.NewRenderer(buf)
	au = audio.NewAudio()

	// One persistent Uint8ClampedArray + ImageData per frame: the framebuffer
	// is copied into it and uploaded with a single putImageData call.
	jsPix = js.Global().Get("Uint8ClampedArray").New(len(buf.Pix()))
	jsImg = js.Global().Get("ImageData").New(jsPix, game.ScreenW, game.ScreenH)

	obj := js.Global().Get("Object").New()
	obj.Set("tick", js.FuncOf(tick))
	obj.Set("setKey", js.FuncOf(setKey))
	obj.Set("move", js.FuncOf(move))
	obj.Set("moveEnd", js.FuncOf(moveEnd))
	obj.Set("state", js.FuncOf(stateNow))
	obj.Set("unlock", js.FuncOf(func(js.Value, []js.Value) any {
		au.Enable()
		return nil
	}))
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
		playEvents()
		saveHighScore()
		accumMs -= tickMs
	}
	r.Render(g)
	upload()
	return nil
}

// playEvents sounds the effects for events emitted since the last tick and
// drains the queue.
func playEvents() {
	for _, e := range g.Events {
		switch e {
		case game.EventFire:
			au.PlayFire()
		case game.EventInvaderKilled:
			au.PlayInvaderKill()
		case game.EventPlayerHit:
			au.PlayPlayerHit()
		case game.EventUFOAppear:
			au.PlayUFOStart()
		case game.EventUFODisappear:
			au.PlayUFOEnd()
		case game.EventUFOKilled:
			au.PlayUFOEnd()
			au.PlayUFOHit()
		case game.EventMarch:
			au.PlayMarch()
		case game.EventGameOver:
			au.PlayGameOver()
		}
	}
	g.Events = nil
}

func setKey(this js.Value, args []js.Value) any {
	code := args[0].String()
	pressed := args[1].Bool()
	if pressed {
		if code == "Enter" {
			au.Enable()
		}
		if code == "KeyM" {
			au.ToggleMute()
		}
	}
	g.HandleInput(code, pressed)
	return nil
}

// stateNow exposes a game snapshot for QA debugging.
func stateNow(js.Value, []js.Value) any {
	return map[string]any{
		"state":     int(g.State),
		"lives":     g.Lives,
		"flash":     g.Flash,
		"score":     g.Score,
		"level":     g.Level,
		"frame":     g.Frame,
		"invaders":  g.Invaders.AliveCount(),
		"bullets":   len(g.Bullets),
		"playerX":   g.Player.X,
		"playerOk":  g.Player.Alive,
		"ufo":       g.UFOActive,
		"muted":     au.Muted,
		"stateName": stateName(g.State),
	}
}

// stateName maps a game state to its QA/debug name.
func stateName(s game.GameState) string {
	switch s {
	case game.StateStart:
		return "start"
	case game.StatePlaying:
		return "playing"
	case game.StatePaused:
		return "paused"
	case game.StateLevelTransition:
		return "level"
	case game.StateGameOver:
		return "gameover"
	}
	return "unknown"
}

// move forwards a touch drag delta in logical pixels to the game.
func move(this js.Value, args []js.Value) any {
	g.AddMoveDx(args[0].Float())
	return nil
}

// moveEnd drops the queued touch drag when the drag pointer lifts.
func moveEnd(js.Value, []js.Value) any {
	g.EndMoveDx()
	return nil
}

// upload pushes the framebuffer to the canvas.
func upload() {
	pix := buf.Pix()
	js.CopyBytesToJS(jsPix, pix)
	ctx.Call("putImageData", jsImg, 0, 0)
}

// loadHighScore restores the persisted high score, if any.
func loadHighScore() {
	v := js.Global().Get("localStorage").Call("getItem", highScoreKey).String()
	if n, err := strconv.Atoi(v); err == nil {
		g.HighScore = n
	}
}

// saveHighScore persists the high score when it changes.
func saveHighScore() {
	if g.HighScore != lastHigh {
		lastHigh = g.HighScore
		js.Global().Get("localStorage").Call("setItem", highScoreKey, strconv.Itoa(g.HighScore))
	}
}
