package game

import "testing"

func TestNewBarricadeShape(t *testing.T) {
	b := NewBarricade(0)
	if b.Pixels[0][0] {
		t.Fatal("col 0 should be empty")
	}
	if !b.Pixels[0][1] || !b.Pixels[0][8] {
		t.Fatal("top row cols 1-8 should be filled")
	}
	if b.Pixels[0][9] {
		t.Fatal("col 9 should be empty")
	}
	if !b.Pixels[4][1] || !b.Pixels[4][2] {
		t.Fatal("lower-left leg should be filled")
	}
	if b.Pixels[4][3] {
		t.Fatal("lower gap should be empty")
	}
	if !b.Pixels[4][7] || !b.Pixels[4][8] {
		t.Fatal("lower-right leg should be filled")
	}
	if !b.Destroyed() == false {
		t.Fatal("fresh barricade should not be destroyed")
	}
}

func TestBarricadePixelAt(t *testing.T) {
	b := NewBarricade(0)
	b.Y = 0
	if !b.PixelAt(2, 0) {
		t.Fatal("should hit filled pixel at (2,0)")
	}
	if b.PixelAt(1, 0) {
		t.Fatal("should miss empty pixel at (1,0)")
	}
	if b.PixelAt(100, 100) {
		t.Fatal("out of bounds should be false")
	}
}

func TestBarricadeDamage(t *testing.T) {
	b := NewBarricade(0)
	b.Y = 0
	if !b.Pixels[0][1] {
		t.Fatal("precondition: pixel should be filled")
	}
	b.Damage(2, 0)
	if b.Pixels[0][1] {
		t.Fatal("hit pixel should be cleared")
	}
	if b.Pixels[1][1] {
		t.Fatal("adjacent pixel should be cleared")
	}
	if !b.Pixels[0][2] {
		t.Fatal("unrelated pixel should remain")
	}
}

func TestBarricadeDamageOutOfBounds(t *testing.T) {
	b := NewBarricade(0)
	b.Y = 0
	b.Damage(-10, -10)
	if b.Destroyed() {
		t.Fatal("out-of-bounds damage should not destroy")
	}
}

func TestBarricadeDestroyed(t *testing.T) {
	b := NewBarricade(0)
	for r := range b.Pixels {
		for c := range b.Pixels[r] {
			b.Pixels[r][c] = false
		}
	}
	if !b.Destroyed() {
		t.Fatal("should be destroyed when all pixels cleared")
	}
}

func TestBarricadeRect(t *testing.T) {
	b := NewBarricade(30)
	x, y, w, h := b.Rect()
	if x != 30 || y != BarricadeY || w != BarricadeW || h != BarricadeH {
		t.Fatalf("rect = (%d,%d,%d,%d)", x, y, w, h)
	}
}

func TestNewBarricadesLayout(t *testing.T) {
	bars := newBarricades()
	if len(bars) != barricadeCount {
		t.Fatalf("count = %d", len(bars))
	}
	for i := 1; i < len(bars); i++ {
		if bars[i].X <= bars[i-1].X {
			t.Fatal("barricades should be ordered left to right")
		}
	}
}

func TestBarricadeClearOverlapping(t *testing.T) {
	b := NewBarricade(0)
	b.Y = 0
	b.ClearOverlapping(0, 0, 22, 4)
	if b.Pixels[0][1] {
		t.Fatal("overlapping pixel should be cleared")
	}
	if !b.Pixels[7][1] {
		t.Fatal("non-overlapping pixel should remain")
	}
}
