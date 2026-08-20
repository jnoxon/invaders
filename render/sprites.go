package render

import "invaders/game"

// Sprite is a 1-bit pixel image: Data[y*W+x] != 0 means filled.
type Sprite struct {
	W, H int
	Data []byte
}

func (s Sprite) Pixel(x, y int) bool {
	if x < 0 || y < 0 || x >= s.W || y >= s.H {
		return false
	}
	return s.Data[y*s.W+x] != 0
}

// sprite builds a Sprite from ASCII art: '#' is filled, anything else
// transparent. It panics on malformed art (init-time only).
func sprite(rows []string) Sprite {
	w := len(rows[0])
	data := make([]byte, 0, len(rows)*w)
	for _, row := range rows {
		if len(row) != w {
			panic("render: sprite row length mismatch")
		}
		for i := range w {
			if row[i] == '#' {
				data = append(data, 1)
			} else {
				data = append(data, 0)
			}
		}
	}
	return Sprite{W: w, H: len(rows), Data: data}
}

// scale2 upscales every pixel to a 2x2 block.
func scale2(s Sprite) Sprite {
	var data []byte
	for y := range s.H {
		row := make([]byte, 0, s.W*2)
		for x := range s.W {
			b := s.Data[y*s.W+x]
			row = append(row, b, b)
		}
		data = append(data, row...)
		data = append(data, row...)
	}
	return Sprite{W: s.W * 2, H: s.H * 2, Data: data}
}

// place embeds s at offset (ox, oy) in a w x h transparent canvas.
func place(w, h, ox, oy int, s Sprite) Sprite {
	data := make([]byte, w*h)
	for y := range s.H {
		for x := range s.W {
			if s.Data[y*s.W+x] != 0 {
				data[(oy+y)*w+(ox+x)] = 1
			}
		}
	}
	return Sprite{W: w, H: h, Data: data}
}

// countPixels returns the number of filled pixels in s.
func countPixels(s Sprite) int {
	n := 0
	for _, b := range s.Data {
		if b != 0 {
			n++
		}
	}
	return n
}

// Player cannon, 12x8 art scaled 2x to the 24x16 game sprite.
var playerShip = scale2(sprite([]string{
	".....#......",
	"....###.....",
	"....###.....",
	"...#.#.#....",
	"...#.#.#....",
	"..#######...",
	".##########.",
	"############",
}))

// Firing frame: the barrel top becomes a 2x4 (screen pixels) muzzle flame.
var playerFiring = scale2(sprite([]string{
	".....#......",
	".....#......",
	"....###.....",
	"....###.....",
	"...#.#.#....",
	"...#.#.#....",
	".##########.",
	"############",
}))

// Invaders: 8x7 art scaled 2x to 16x14, centered in the 20x15 game box.
const invaderPlace = 2

var squidF0 = place(game.InvaderW, game.InvaderH, invaderPlace, 0, scale2(sprite([]string{
	"..#..#..",
	"..####..",
	"########",
	"#.####.#",
	"########",
	"##.##.##",
	"#.#..#.#",
})))

var squidF1 = place(game.InvaderW, game.InvaderH, invaderPlace, 0, scale2(sprite([]string{
	"..#..#..",
	"..####..",
	"########",
	"########",
	"#.####.#",
	".######.",
	".#.##.#.",
})))

var crabF0 = place(game.InvaderW, game.InvaderH, invaderPlace, 0, scale2(sprite([]string{
	"#....#",
	"##..##",
	"######",
	"######",
	"######",
	"#.##.#",
	".#..#.",
})))

var crabF1 = place(game.InvaderW, game.InvaderH, invaderPlace, 0, scale2(sprite([]string{
	".#..#.",
	"##..##",
	"######",
	"######",
	"######",
	"##..##",
	"#....#",
})))

var octopusF0 = place(game.InvaderW, game.InvaderH, invaderPlace, 0, scale2(sprite([]string{
	"..####..",
	".######.",
	"########",
	"########",
	".######.",
	"##.##.##",
	"#.#..#.#",
})))

var octopusF1 = place(game.InvaderW, game.InvaderH, invaderPlace, 0, scale2(sprite([]string{
	"..####..",
	".######.",
	"########",
	"########",
	".######.",
	"##.##.##",
	".#.##.#.",
})))

// UFO: 20x7 art scaled 2x to the 40x14 game sprite.
var ufoSprite = scale2(sprite([]string{
	".....##########.....",
	"....############....",
	"...##############...",
	"####################",
	"##.##.##.##.##.##.##",
	"...##############...",
	".....##########.....",
}))

// Enemy bullets: 2x8 zigzag, 2 frames.
var enemyBulletF0 = sprite([]string{
	"#.",
	"##",
	".#",
	"..",
	".#",
	"##",
	"#.",
	"..",
})

var enemyBulletF1 = sprite([]string{
	"..",
	".#",
	"##",
	"#.",
	"#.",
	"##",
	".#",
	"..",
})

// Life icon: small cannon, 10x5.
var lifeIcon = sprite([]string{
	"....#....",
	"...###...",
	"..#.#.#..",
	".#######.",
	"#########",
})

func invaderSprite(t game.InvaderType, frame int) Sprite {
	switch t {
	case game.InvaderSquid:
		if frame&1 == 0 {
			return squidF0
		}
		return squidF1
	case game.InvaderCrab:
		if frame&1 == 0 {
			return crabF0
		}
		return crabF1
	default:
		if frame&1 == 0 {
			return octopusF0
		}
		return octopusF1
	}
}
