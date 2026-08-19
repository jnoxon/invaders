package game

import "testing"

func TestInputKeyCodeMapping(t *testing.T) {
	cases := []struct {
		code  string
		left  bool
		right bool
		fire  bool
	}{
		{"ArrowLeft", true, false, false},
		{"KeyA", true, false, false},
		{"ArrowRight", false, true, false},
		{"KeyD", false, true, false},
		{"Space", false, false, true},
		{"Enter", false, false, false},
	}
	for _, tc := range cases {
		t.Run(tc.code, func(t *testing.T) {
			s := NewInputState()
			s.Update(tc.code, true)
			if s.Left != tc.left {
				t.Fatalf("Left = %v, want %v", s.Left, tc.left)
			}
			if s.Right != tc.right {
				t.Fatalf("Right = %v, want %v", s.Right, tc.right)
			}
			if s.Fire != tc.fire {
				t.Fatalf("Fire = %v, want %v", s.Fire, tc.fire)
			}
		})
	}
}

func TestInputLeftRightRelease(t *testing.T) {
	cases := []struct {
		code string
		get  func(*InputState) bool
	}{
		{"ArrowLeft", func(s *InputState) bool { return s.Left }},
		{"KeyA", func(s *InputState) bool { return s.Left }},
		{"ArrowRight", func(s *InputState) bool { return s.Right }},
		{"KeyD", func(s *InputState) bool { return s.Right }},
	}
	for _, tc := range cases {
		s := NewInputState()
		s.Update(tc.code, true)
		if !tc.get(&s) {
			t.Fatalf("%s: not set on press", tc.code)
		}
		s.Update(tc.code, false)
		if tc.get(&s) {
			t.Fatalf("%s: still set on release", tc.code)
		}
	}
}

func TestInputFireEdgeTrigger(t *testing.T) {
	g := newTestGame()
	g.State = StatePlaying
	g.HandleInput("Space", true)
	g.Tick()
	if got := countBullets(g.Bullets, BulletPlayer); got != 1 {
		t.Fatalf("after press: player bullets = %d, want 1", got)
	}
	g.Tick()
	g.Tick()
	if got := countBullets(g.Bullets, BulletPlayer); got != 1 {
		t.Fatalf("while holding: player bullets = %d, want 1", got)
	}
}

func TestInputEnterJustPressed(t *testing.T) {
	g := newTestGame()
	g.HandleInput("Enter", true)
	if !g.Input.JustPressed["Enter"] {
		t.Fatal("Enter should be just-pressed")
	}
	g.Tick()
	if g.Input.JustPressed["Enter"] {
		t.Fatal("Enter just-pressed should clear after tick")
	}
}

func TestInputPJustPressed(t *testing.T) {
	g := newTestGame()
	g.HandleInput("KeyP", true)
	if !g.Input.JustPressed["KeyP"] {
		t.Fatal("KeyP should be tracked as just-pressed")
	}
	g.Tick()
	if g.Input.JustPressed["KeyP"] {
		t.Fatal("KeyP just-pressed should clear after tick")
	}
}

func TestInputClearJustPressed(t *testing.T) {
	s := NewInputState()
	s.Update("Enter", true)
	s.Update("Space", true)
	if len(s.JustPressed) != 2 {
		t.Fatalf("just-pressed count = %d, want 2", len(s.JustPressed))
	}
	s.ClearJustPressed()
	if len(s.JustPressed) != 0 {
		t.Fatalf("just-pressed should be empty after clear, got %d", len(s.JustPressed))
	}
}

func TestInputStateNilJustPressed(t *testing.T) {
	s := InputState{}
	s.Update("Enter", true)
	if !s.JustPressed["Enter"] {
		t.Fatal("nil JustPressed should be initialized")
	}
}
