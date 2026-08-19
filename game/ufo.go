package game

import "math/rand"

const (
	UFOSpeed         = 2
	UFOY             = 36
	UFOW             = 40
	UFOH             = 14
	UFOKillThreshold = 20
)

var ufoPoints = []int{50, 100, 150, 300}

type UFO struct {
	X, Y   int
	Dir    int
	Active bool
	Points int
}

func NewUFO(rng *rand.Rand) UFO {
	dir := 1
	if rng.Intn(2) == 0 {
		dir = -1
	}
	x := ScreenW
	if dir > 0 {
		x = -UFOW
	}
	return UFO{
		X:      x,
		Y:      UFOY,
		Dir:    dir,
		Active: true,
		Points: ufoPoints[rng.Intn(len(ufoPoints))],
	}
}

func (u *UFO) Update() {
	if !u.Active {
		return
	}
	u.X += u.Dir * UFOSpeed
	if u.Dir > 0 && u.X > ScreenW {
		u.Active = false
	}
	if u.Dir < 0 && u.X+UFOW < 0 {
		u.Active = false
	}
}

func (u *UFO) Rect() (int, int, int, int) {
	return u.X, u.Y, UFOW, UFOH
}
