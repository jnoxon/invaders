package game

import (
	"math/rand/v2"
	"testing"
)

var validUFOPoints = map[int]bool{50: true, 100: true, 150: true, 300: true}

func TestNewUFOValid(t *testing.T) {
	rng := rand.New(rand.NewPCG(1, 1))
	for i := 0; i < 50; i++ {
		u := NewUFO(rng)
		if u.Dir != 1 && u.Dir != -1 {
			t.Fatalf("dir = %d", u.Dir)
		}
		if !validUFOPoints[u.Points] {
			t.Fatalf("points = %d not valid", u.Points)
		}
		if !u.Active {
			t.Fatal("new UFO should be active")
		}
		if u.Y != UFOY {
			t.Fatalf("y = %d, want %d", u.Y, UFOY)
		}
	}
}

func TestUFOStartsOffScreen(t *testing.T) {
	rng := rand.New(rand.NewPCG(1, 2))
	for i := 0; i < 50; i++ {
		u := NewUFO(rng)
		if u.Dir > 0 && u.X >= 0 {
			t.Fatal("right-moving UFO should start off left edge")
		}
		if u.Dir < 0 && u.X < ScreenW {
			t.Fatal("left-moving UFO should start off right edge")
		}
	}
}

func TestUFOMovesRight(t *testing.T) {
	u := UFO{X: 100, Y: UFOY, Dir: 1, Active: true}
	u.Update()
	if u.X != 100+UFOSpeed {
		t.Fatalf("x = %d", u.X)
	}
}

func TestUFOMovesLeft(t *testing.T) {
	u := UFO{X: 100, Y: UFOY, Dir: -1, Active: true}
	u.Update()
	if u.X != 100-UFOSpeed {
		t.Fatalf("x = %d", u.X)
	}
}

func TestUFODeactivatesOffScreenRight(t *testing.T) {
	u := UFO{X: ScreenW - 1, Y: UFOY, Dir: 1, Active: true}
	u.Update()
	if u.Active {
		t.Fatal("should deactivate off right edge")
	}
}

func TestUFODeactivatesOffScreenLeft(t *testing.T) {
	u := UFO{X: -UFOW, Y: UFOY, Dir: -1, Active: true}
	u.Update()
	if u.Active {
		t.Fatal("should deactivate off left edge")
	}
}

func TestUFOInactiveDoesNotMove(t *testing.T) {
	u := UFO{X: 100, Y: UFOY, Dir: 1, Active: false}
	u.Update()
	if u.X != 100 {
		t.Fatal("inactive UFO should not move")
	}
}

func TestUFORect(t *testing.T) {
	u := UFO{X: 10, Y: UFOY, Dir: 1, Active: true}
	x, y, w, h := u.Rect()
	if x != 10 || y != UFOY || w != UFOW || h != UFOH {
		t.Fatalf("rect = (%d,%d,%d,%d)", x, y, w, h)
	}
}
