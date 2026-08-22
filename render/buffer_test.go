package render

import (
	"testing"

	"invaders/game"
)

func pixel(b *Buffer, x, y int) (r, g, bl, a byte) {
	i := (y*game.ScreenW + x) * 4
	return b.pix[i], b.pix[i+1], b.pix[i+2], b.pix[i+3]
}

func TestBufferStartsBlack(t *testing.T) {
	b := NewBuffer()
	for _, p := range [][2]int{{0, 0}, {128, 112}, {255, 223}} {
		if r, _, _, _ := pixel(b, p[0], p[1]); r != 0 {
			t.Errorf("pixel %v not black at start", p)
		}
	}
}

func TestBufferFillRect(t *testing.T) {
	b := NewBuffer()
	b.FillRect(10, 20, 4, 3)
	if r, g, bl, a := pixel(b, 10, 20); r != 255 || g != 255 || bl != 255 || a != 255 {
		t.Errorf("pixel (10,20) = %d %d %d %d, want white", r, g, bl, a)
	}
	if r, _, _, _ := pixel(b, 13, 22); r != 255 {
		t.Error("pixel (13,22) not filled")
	}
	// Corners just outside the rect stay black.
	for _, p := range [][2]int{{9, 20}, {14, 20}, {10, 19}, {10, 23}} {
		if r, _, _, _ := pixel(b, p[0], p[1]); r != 0 {
			t.Errorf("pixel %v outside rect is white", p)
		}
	}
}

func TestBufferFillRectClamps(t *testing.T) {
	b := NewBuffer()
	b.FillRect(-5, -5, 10, 10)
	b.FillRect(250, 220, 10, 10)
	if r, _, _, _ := pixel(b, 0, 0); r != 255 {
		t.Error("negative origin should clamp into the screen")
	}
	if r, _, _, _ := pixel(b, 255, 223); r != 255 {
		t.Error("off-screen origin should clamp into the screen")
	}
	b.FillRect(300, 300, 4, 4)
	b.FillRect(0, 0, -1, -1)
}

func TestBufferClear(t *testing.T) {
	b := NewBuffer()
	b.FillRect(0, 0, game.ScreenW, game.ScreenH)
	b.Clear()
	if r, _, _, _ := pixel(b, 100, 100); r != 0 {
		t.Error("Clear did not reset the buffer")
	}
}

func TestRenderFlashFillsScreen(t *testing.T) {
	g := gameplayGame()
	g.Flash = 1
	b := NewBuffer()
	NewRenderer(b).Render(g)
	// Entire screen white, including pixels only gameplay would draw.
	for _, p := range [][2]int{{0, 0}, {128, 112}, {255, 223}, {0, 223}} {
		if r, _, _, _ := pixel(b, p[0], p[1]); r != 255 {
			t.Errorf("flash pixel %v = %d, want 255", p, r)
		}
	}
}

func TestRenderScorePopup(t *testing.T) {
	g := gameplayGame()
	// y=150 sits in the empty band between invaders and barricades.
	g.ScorePopup = game.ScorePopup{X: 10, Y: 150, Points: 100, Timer: 30}
	b := NewBuffer()
	NewRenderer(b).Render(g)
	// "100": '1' = {2,6,2,2,7}; row 0 lit column is 1 -> pixel (11, 150).
	if r, _, _, _ := pixel(b, 11, 150); r != 255 {
		t.Error("missing score popup digit")
	}
}

func TestRenderScorePopupClampedRight(t *testing.T) {
	g := gameplayGame()
	g.ScorePopup = game.ScorePopup{X: 250, Y: 150, Points: 300, Timer: 30}
	b := NewBuffer()
	NewRenderer(b).Render(g)
	// "300" is 11px wide: clamped to x=245; last '0' spans x 253-255.
	if r, _, _, _ := pixel(b, 255, 153); r != 255 {
		t.Error("clamped popup should end at the right screen edge")
	}
}

func TestRenderScorePopupHiddenWhenExpired(t *testing.T) {
	g := gameplayGame()
	g.ScorePopup = game.ScorePopup{X: 10, Y: 150, Points: 100, Timer: 0}
	b := NewBuffer()
	NewRenderer(b).Render(g)
	if r, _, _, _ := pixel(b, 11, 150); r != 0 {
		t.Error("expired popup should not draw")
	}
}
