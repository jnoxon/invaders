package game

import "testing"

func TestMoveDxAppliedOnNextTick(t *testing.T) {
	g := playingGame()
	x0 := g.Player.X
	g.AddMoveDx(15)
	if g.Player.X != x0 {
		t.Fatalf("X = %d before tick, want %d (no movement before the fixed tick)", g.Player.X, x0)
	}
	g.Tick()
	if g.Player.X != x0+15 {
		t.Fatalf("X = %d after tick, want %d", g.Player.X, x0+15)
	}
}

func TestMoveDxFractionalAccumulation(t *testing.T) {
	t.Run("half steps sum to one pixel", func(t *testing.T) {
		g := playingGame()
		x0 := g.Player.X
		g.AddMoveDx(0.5)
		g.AddMoveDx(0.5)
		g.Tick()
		if g.Player.X != x0+1 {
			t.Fatalf("X = %d, want %d (two 0.5 steps = 1px)", g.Player.X, x0+1)
		}
	})
	t.Run("remainder carries across ticks", func(t *testing.T) {
		g := playingGame()
		x0 := g.Player.X
		want := []int{x0, x0 + 1, x0 + 1, x0 + 2}
		for i := 0; i < 4; i++ {
			g.AddMoveDx(0.4)
			g.Tick()
			if g.Player.X != want[i] {
				t.Fatalf("tick %d: X = %d, want %d", i+1, g.Player.X, want[i])
			}
		}
	})
}

func TestMoveDxClampedAtEdges(t *testing.T) {
	for _, tc := range []struct {
		name   string
		startX int
		dx     float64
		want   int
	}{
		{"left edge", 0, -100, 0},
		{"right edge", ScreenW - PlayerW, 100, ScreenW - PlayerW},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := playingGame()
			g.Player.X = tc.startX
			g.AddMoveDx(tc.dx)
			g.Tick()
			if g.Player.X != tc.want {
				t.Fatalf("X = %d, want %d", g.Player.X, tc.want)
			}
		})
	}
}

// The spec's edge-pin test: the ship tracks the anchored finger target.
// Deltas the playfield clamp eats are kept as debt while the finger stays
// on the pinned side of the anchor, so the ship cannot lead the finger
// (teleport) when the edge frees up.
func TestMoveDxEdgePinAnchored(t *testing.T) {
	t.Run("net zero before tick", func(t *testing.T) {
		g := playingGame()
		g.Player.X = 0
		g.AddMoveDx(-50)
		g.AddMoveDx(30)
		g.AddMoveDx(20)
		g.Tick()
		if g.Player.X != 0 {
			t.Fatalf("X = %d, want 0 (net finger displacement is 0)", g.Player.X)
		}
	})
	t.Run("spec edge-pin", func(t *testing.T) {
		g := playingGame()
		g.Player.X = 0
		g.AddMoveDx(-50)
		g.Tick()
		if g.Player.X != 0 {
			t.Fatalf("X = %d, want 0 (clamped at left edge)", g.Player.X)
		}
		g.AddMoveDx(30)
		g.Tick()
		if g.Player.X != 0 {
			t.Fatalf("X = %d, want 0 (finger still left of anchor)", g.Player.X)
		}
		g.AddMoveDx(20)
		g.Tick()
		if g.Player.X != 0 {
			t.Fatalf("X = %d, want 0 (net finger displacement is 0)", g.Player.X)
		}
	})
	t.Run("left edge tracks after anchor", func(t *testing.T) {
		g := playingGame()
		g.Player.X = 0
		g.AddMoveDx(-50)
		g.Tick()
		g.AddMoveDx(60)
		g.Tick()
		if g.Player.X != 10 {
			t.Fatalf("X = %d, want 10 (finger 10px past anchor)", g.Player.X)
		}
	})
	t.Run("right edge tracks after anchor", func(t *testing.T) {
		g := playingGame()
		g.Player.X = ScreenW - PlayerW
		g.AddMoveDx(50)
		g.Tick()
		if g.Player.X != ScreenW-PlayerW {
			t.Fatalf("X = %d, want %d (clamped at right edge)", g.Player.X, ScreenW-PlayerW)
		}
		g.AddMoveDx(-60)
		g.Tick()
		if g.Player.X != ScreenW-PlayerW-10 {
			t.Fatalf("X = %d, want %d (finger 10px past anchor)", g.Player.X, ScreenW-PlayerW-10)
		}
	})
}

func TestEndMoveDxClearsDebt(t *testing.T) {
	t.Run("left edge", func(t *testing.T) {
		g := playingGame()
		g.Player.X = 0
		g.AddMoveDx(-50)
		g.Tick()
		g.EndMoveDx()
		g.AddMoveDx(10)
		g.Tick()
		if g.Player.X != 10 {
			t.Fatalf("X = %d, want 10 (stale gesture debt must not apply)", g.Player.X)
		}
	})
	t.Run("right edge", func(t *testing.T) {
		g := playingGame()
		g.Player.X = ScreenW - PlayerW
		g.AddMoveDx(50)
		g.Tick()
		g.EndMoveDx()
		g.AddMoveDx(-10)
		g.Tick()
		if g.Player.X != ScreenW-PlayerW-10 {
			t.Fatalf("X = %d, want %d (stale gesture debt must not apply)", g.Player.X, ScreenW-PlayerW-10)
		}
	})
}

func TestMoveDxClearedOnStateChange(t *testing.T) {
	t.Run("pause", func(t *testing.T) {
		g := playingGame()
		x0 := g.Player.X
		g.AddMoveDx(30)
		g.Pause()
		g.Tick()
		g.Resume()
		g.Tick()
		if g.Player.X != x0 {
			t.Fatalf("X = %d, want %d (stale delta applied after resume)", g.Player.X, x0)
		}
	})
	t.Run("game over", func(t *testing.T) {
		g := playingGame()
		g.AddMoveDx(30)
		g.GameOver()
		g.Tick()
		g.HandleInput("Enter", true)
		g.Tick()
		if g.Player.X != (ScreenW-PlayerW)/2 {
			t.Fatalf("X = %d, want %d (stale delta applied after restart)", g.Player.X, (ScreenW-PlayerW)/2)
		}
	})
	t.Run("start", func(t *testing.T) {
		g := newTestGame()
		g.AddMoveDx(30)
		g.StartGame()
		g.Tick()
		if g.Player.X != (ScreenW-PlayerW)/2 {
			t.Fatalf("X = %d, want %d (stale delta applied after start)", g.Player.X, (ScreenW-PlayerW)/2)
		}
	})
	t.Run("level transition", func(t *testing.T) {
		g := playingGame()
		g.AddMoveDx(30)
		g.State = StateLevelTransition
		g.TransitionTimer = TransitionFrames
		g.ResetLevel()
		g.State = StatePlaying
		g.Tick()
		if g.Player.X != (ScreenW-PlayerW)/2 {
			t.Fatalf("X = %d, want %d (stale delta applied after level transition)", g.Player.X, (ScreenW-PlayerW)/2)
		}
	})
}

func TestMoveDxComposesWithHeldKey(t *testing.T) {
	g := playingGame()
	x0 := g.Player.X
	g.Input.Left = true
	g.AddMoveDx(10)
	g.Tick()
	want := x0 - PlayerSpeed + 10
	if g.Player.X != want {
		t.Fatalf("X = %d, want %d (held key and dx in the same tick)", g.Player.X, want)
	}
}

func TestMoveDxIgnoredWhenNotPlayingOrDead(t *testing.T) {
	t.Run("start state", func(t *testing.T) {
		g := newTestGame()
		x0 := g.Player.X
		g.AddMoveDx(10)
		g.Tick()
		if g.Player.X != x0 {
			t.Fatalf("X = %d, want %d (dx applied in start state)", g.Player.X, x0)
		}
	})
	t.Run("game over state", func(t *testing.T) {
		g := playingGame()
		g.GameOver()
		x0 := g.Player.X
		g.AddMoveDx(10)
		g.Tick()
		if g.Player.X != x0 {
			t.Fatalf("X = %d, want %d (dx applied in game over state)", g.Player.X, x0)
		}
	})
	t.Run("dead player", func(t *testing.T) {
		g := playingGame()
		x0 := g.Player.X
		g.Player.Hit()
		g.AddMoveDx(10)
		g.Tick()
		if g.Player.X != x0 {
			t.Fatalf("X = %d, want %d (dx applied while dead)", g.Player.X, x0)
		}
	})
}
