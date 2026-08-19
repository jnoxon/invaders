package game

type InputState struct {
	Left, Right, Fire bool
	JustPressed       map[string]bool
}

func NewInputState() InputState {
	return InputState{JustPressed: map[string]bool{}}
}

func (s *InputState) Update(code string, pressed bool) {
	if s.JustPressed == nil {
		s.JustPressed = map[string]bool{}
	}
	switch code {
	case "ArrowLeft", "KeyA":
		s.Left = pressed
	case "ArrowRight", "KeyD":
		s.Right = pressed
	case "Space":
		s.Fire = pressed
	}
	if pressed {
		s.JustPressed[code] = true
	} else {
		delete(s.JustPressed, code)
	}
}

func (s *InputState) ClearJustPressed() {
	for k := range s.JustPressed {
		delete(s.JustPressed, k)
	}
}
