package game

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

func (b *Barricade) Damage(sx, sy int) {
	col := (sx - b.X) / 2
	row := (sy - b.Y) / 2
	b.clearPixel(col, row)
	b.clearPixel(col, row+1)
}

func (b *Barricade) clearPixel(col, row int) {
	if col < 0 || col >= BarricadePixelW || row < 0 || row >= BarricadePixelH {
		return
	}
	b.Pixels[row][col] = false
}

func (b *Barricade) ClearOverlapping(ix, iy, iw, ih int) {
	for r := 0; r < BarricadePixelH; r++ {
		for c := 0; c < BarricadePixelW; c++ {
			if !b.Pixels[r][c] {
				continue
			}
			px := b.X + c*2
			py := b.Y + r*2
			if AABB(px, py, 2, 2, ix, iy, iw, ih) {
				b.Pixels[r][c] = false
			}
		}
	}
}

func (b *Barricade) Rect() (int, int, int, int) {
	return b.X, b.Y, BarricadeW, BarricadeH
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
