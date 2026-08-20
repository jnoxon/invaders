package game

import "math/rand/v2"

const (
	BarricadeW      = 22
	BarricadeH      = 16
	BarricadePixelW = 11
	BarricadePixelH = 8
	BarricadeY      = 176
	barricadeCount  = 4
)

var barricadeShape = []string{
	".XXXXXXXX..",
	".XXXXXXXX..",
	".XXXXXXXX..",
	".XXXXXXXX..",
	".XX....XX..",
	".XX....XX..",
	".XX....XX..",
	".XX....XX..",
}

type Barricade struct {
	X, Y   int
	Pixels [][]bool
}

func NewBarricade(x int) Barricade {
	b := Barricade{X: x, Y: BarricadeY, Pixels: make([][]bool, BarricadePixelH)}
	for r := 0; r < BarricadePixelH; r++ {
		b.Pixels[r] = make([]bool, BarricadePixelW)
		for c := 0; c < BarricadePixelW; c++ {
			b.Pixels[r][c] = barricadeShape[r][c] == 'X'
		}
	}
	return b
}

func newBarricades() []Barricade {
	out := make([]Barricade, barricadeCount)
	gap := (ScreenW - barricadeCount*BarricadeW) / (barricadeCount + 1)
	for i := 0; i < barricadeCount; i++ {
		out[i] = NewBarricade(gap + i*(BarricadeW+gap))
	}
	return out
}

func (b *Barricade) PixelAt(sx, sy int) bool {
	col := (sx - b.X) / 2
	row := (sy - b.Y) / 2
	if col < 0 || col >= BarricadePixelW || row < 0 || row >= BarricadePixelH {
		return false
	}
	return b.Pixels[row][col]
}

// Damage erodes the barricade at screen-space impact (sx, sy): the hit
// pixel is cleared, then one random set neighbor (up/down/left/right) is
// cleared as well, producing an irregular crater. A hit on an already
// empty pixel changes nothing.
func (b *Barricade) Damage(sx, sy int, rng *rand.Rand) {
	col := (sx - b.X) / 2
	row := (sy - b.Y) / 2
	if col < 0 {
		col = 0
	}
	if col >= BarricadePixelW {
		col = BarricadePixelW - 1
	}
	if row < 0 {
		row = 0
	}
	if row >= BarricadePixelH {
		row = BarricadePixelH - 1
	}
	if !b.Pixels[row][col] {
		return
	}
	b.Pixels[row][col] = false

	var candidates [4][2]int
	n := 0
	for _, d := range [][2]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}} {
		nr, nc := row+d[0], col+d[1]
		if nr < 0 || nr >= BarricadePixelH || nc < 0 || nc >= BarricadePixelW {
			continue
		}
		if !b.Pixels[nr][nc] {
			continue
		}
		candidates[n] = [2]int{nr, nc}
		n++
	}
	if n > 0 {
		pick := candidates[rng.IntN(n)]
		b.Pixels[pick[0]][pick[1]] = false
	}
}

// OverlapRect clears every set pixel that overlaps the given screen rect.
func (b *Barricade) OverlapRect(ox, oy, ow, oh int) {
	for r := 0; r < BarricadePixelH; r++ {
		for c := 0; c < BarricadePixelW; c++ {
			if !b.Pixels[r][c] {
				continue
			}
			px := b.X + c*2
			py := b.Y + r*2
			if AABB(px, py, 2, 2, ox, oy, ow, oh) {
				b.Pixels[r][c] = false
			}
		}
	}
}

func (b *Barricade) Rect() (int, int, int, int) {
	return b.X, b.Y, BarricadeW, BarricadeH
}

func (b *Barricade) PixelCount() int {
	n := 0
	for r := 0; r < BarricadePixelH; r++ {
		for c := 0; c < BarricadePixelW; c++ {
			if b.Pixels[r][c] {
				n++
			}
		}
	}
	return n
}

func (b *Barricade) Destroyed() bool {
	for r := 0; r < BarricadePixelH; r++ {
		for c := 0; c < BarricadePixelW; c++ {
			if b.Pixels[r][c] {
				return false
			}
		}
	}
	return true
}
