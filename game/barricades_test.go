package game

import (
	"math/rand/v2"
	"testing"
)

func TestNewBarricadeShape(t *testing.T) {
	b := NewBarricade(0)
	want := [BarricadePixelH]string{
		".XXXXXXXX..",
		".XXXXXXXX..",
		".XXXXXXXX..",
		".XXXXXXXX..",
		".XX....XX..",
		".XX....XX..",
		".XX....XX..",
		".XX....XX..",
	}
	for r := range BarricadePixelH {
		for c := range BarricadePixelW {
			wantSet := want[r][c] == 'X'
			if got := b.Pixels[r][c]; got != wantSet {
				t.Fatalf("pixel (%d,%d) = %v, want %v (row %q)", c, r, got, wantSet, want[r])
			}
		}
	}
	if b.Destroyed() {
		t.Fatal("fresh barricade should not be destroyed")
	}
}

func TestNewBarricadePixelCount(t *testing.T) {
	b := NewBarricade(0)
	if got := b.PixelCount(); got != 48 {
		t.Fatalf("pixel count = %d, want 48", got)
	}
}

func TestBarricadePixelAt(t *testing.T) {
	b := NewBarricade(0)
	b.Y = 0
	if !b.PixelAt(2, 0) {
		t.Fatal("should hit filled pixel at (2,0)")
	}
	b.Pixels[0][1] = false
	if b.PixelAt(2, 0) {
		t.Fatal("should miss cleared pixel at (2,0)")
	}
	if b.PixelAt(100, 100) {
		t.Fatal("out of bounds should be false")
	}
	if b.PixelAt(-2, 0) {
		t.Fatal("negative x should be false")
	}
}

func TestBarricadeDamageCenter(t *testing.T) {
	b := NewBarricade(0)
	b.Y = 0
	rng := rand.New(rand.NewPCG(1, 7))
	count := b.PixelCount()
	// pixel (4,2): all four neighbors set
	b.Damage(8, 4, rng)
	if b.Pixels[2][4] {
		t.Fatal("hit pixel should be cleared")
	}
	adj := 0
	for _, p := range [][2]int{{4, 1}, {4, 3}, {3, 2}, {5, 2}} {
		if !b.Pixels[p[1]][p[0]] {
			adj++
		}
	}
	if adj != 1 {
		t.Fatalf("neighbors cleared = %d, want 1", adj)
	}
	if got := b.PixelCount(); got != count-2 {
		t.Fatalf("pixel count = %d, want %d", got, count-2)
	}
}

func TestBarricadeDamageCorner(t *testing.T) {
	b := NewBarricade(0)
	b.Y = 0
	rng := rand.New(rand.NewPCG(1, 8))
	count := b.PixelCount()
	// pixel (1,0): top-left corner of the arch; only down and right set
	b.Damage(2, 0, rng)
	if b.Pixels[0][1] {
		t.Fatal("hit pixel should be cleared")
	}
	adj := 0
	for _, p := range [][2]int{{1, 1}, {2, 0}} {
		if !b.Pixels[p[1]][p[0]] {
			adj++
		}
	}
	if adj != 1 {
		t.Fatalf("neighbors cleared = %d, want 1", adj)
	}
	if !b.Pixels[1][2] {
		t.Fatal("diagonal pixel should remain")
	}
	if got := b.PixelCount(); got != count-2 {
		t.Fatalf("pixel count = %d, want %d", got, count-2)
	}
}

func TestBarricadeDamageEmptyPixelNoChange(t *testing.T) {
	b := NewBarricade(0)
	b.Y = 0
	rng := rand.New(rand.NewPCG(1, 9))
	count := b.PixelCount()
	// pixel (4,5) is inside the arch gap
	b.Damage(8, 10, rng)
	if got := b.PixelCount(); got != count {
		t.Fatalf("pixel count = %d, want %d (no change)", got, count)
	}
}

func TestBarricadeDamageLastPixel(t *testing.T) {
	b := NewBarricade(0)
	for r := range b.Pixels {
		for c := range b.Pixels[r] {
			b.Pixels[r][c] = false
		}
	}
	b.Pixels[0][0] = true
	rng := rand.New(rand.NewPCG(1, 10))
	b.Damage(0, 0, rng)
	if !b.Destroyed() {
		t.Fatal("last pixel should be destroyed")
	}
}

func TestBarricadeDamageClampsToBounds(t *testing.T) {
	b := NewBarricade(0)
	for r := range b.Pixels {
		for c := range b.Pixels[r] {
			b.Pixels[r][c] = false
		}
	}
	b.Pixels[7][10] = true
	rng := rand.New(rand.NewPCG(1, 11))
	b.Damage(1000, 1000, rng)
	if !b.Destroyed() {
		t.Fatal("clamped hit should clear bottom-right pixel")
	}
}

func TestBarricadeDamageDeterministic(t *testing.T) {
	b1 := NewBarricade(0)
	b2 := NewBarricade(0)
	rng1 := rand.New(rand.NewPCG(1, 42))
	rng2 := rand.New(rand.NewPCG(1, 42))
	b1.Damage(8, 4, rng1)
	b2.Damage(8, 4, rng2)
	for r := range b1.Pixels {
		for c := range b1.Pixels[r] {
			if b1.Pixels[r][c] != b2.Pixels[r][c] {
				t.Fatalf("seeded damage diverged at (%d,%d)", c, r)
			}
		}
	}
}

func TestBarricadeDamageNeighborVaries(t *testing.T) {
	// With 4 candidates the random pick must not be a fixed neighbor.
	seen := map[int]bool{}
	for seed := range 64 {
		b := NewBarricade(0)
		b.Y = 0
		rng := rand.New(rand.NewPCG(1, uint64(seed)))
		b.Damage(8, 4, rng) // pixel (4,2)
		for i, p := range [][2]int{{4, 1}, {4, 3}, {3, 2}, {5, 2}} {
			if !b.Pixels[p[1]][p[0]] {
				seen[i] = true
			}
		}
	}
	if len(seen) < 2 {
		t.Fatalf("neighbor pick never varied: %v", seen)
	}
}

func TestBarricadeDestroyed(t *testing.T) {
	b := NewBarricade(0)
	if b.Destroyed() {
		t.Fatal("fresh barricade should not be destroyed")
	}
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

func TestBarricadeOverlapRectFull(t *testing.T) {
	b := NewBarricade(0)
	b.OverlapRect(b.X, b.Y, BarricadeW, BarricadeH)
	if !b.Destroyed() {
		t.Fatal("full overlap should destroy the barricade")
	}
}

func TestBarricadeOverlapRectPartial(t *testing.T) {
	b := NewBarricade(0)
	b.Y = 0
	b.OverlapRect(0, 0, 22, 4)
	if b.Pixels[0][1] || b.Pixels[1][1] {
		t.Fatal("rows 0-1 should be cleared")
	}
	if !b.Pixels[2][1] || !b.Pixels[7][1] {
		t.Fatal("rows 2-7 should remain")
	}
	if got := b.PixelCount(); got != 32 {
		t.Fatalf("pixel count = %d, want 32", got)
	}
}

func TestBarricadeOverlapRectNoOverlap(t *testing.T) {
	b := NewBarricade(0)
	b.Y = 0
	b.OverlapRect(50, 50, 10, 10)
	if got := b.PixelCount(); got != 48 {
		t.Fatalf("pixel count = %d, want 48 (unchanged)", got)
	}
}
