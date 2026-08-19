package game

import "testing"

func TestInputStateUpdate(t *testing.T) {
	s := NewInputState()
	s.Update("ArrowLeft", true)
	if !s.Left {
		t.Fatal("Left should be set")
	}
	s.Update("KeyD", true)
	if !s.Right {
		t.Fatal("Right should be set (KeyD)")
	}
	s.Update("Space", true)
	if !s.Fire {
		t.Fatal("Fire should be set")
	}
	s.Update("ArrowLeft", false)
	if s.Left {
		t.Fatal("Left should clear on keyup")
	}
}

func TestInputStateJustPressed(t *testing.T) {
	s := NewInputState()
	s.Update("Enter", true)
	if !s.JustPressed["Enter"] {
		t.Fatal("Enter should be just-pressed")
	}
	s.ClearJustPressed()
	if len(s.JustPressed) != 0 {
		t.Fatal("JustPressed should be empty after clear")
	}
}

func TestInputStateNilJustPressed(t *testing.T) {
	s := InputState{}
	s.Update("Enter", true)
	if !s.JustPressed["Enter"] {
		t.Fatal("nil JustPressed should be initialized")
	}
}
