package render

// SpriteFont is a fixed-size pixel font. Each glyph is H rows of W-bit
// patterns; bit (W-1-col) selects column col.
type SpriteFont struct {
	W, H   int
	Glyphs map[byte][5]byte
}

// Has reports whether the font defines c.
func (f SpriteFont) Has(c byte) bool {
	_, ok := f.Glyphs[c]
	return ok
}

// TextWidth returns the rendered width: W pixels per character plus 1px
// spacing, so len(s)*(W+1)-1.
func (f SpriteFont) TextWidth(s string) int {
	if len(s) == 0 {
		return 0
	}
	return len(s)*(f.W+1) - 1
}

// PixelFont is a 3x5 font covering 0-9, A-Z, space, ':', '-', '=', '?'.
var PixelFont = SpriteFont{
	W: 3, H: 5,
	Glyphs: map[byte][5]byte{
		'0': {5, 7, 5, 5, 7},
		'1': {2, 6, 2, 2, 7},
		'2': {7, 1, 7, 4, 7},
		'3': {7, 1, 3, 1, 7},
		'4': {5, 5, 7, 1, 1},
		'5': {7, 4, 7, 1, 7},
		'6': {7, 4, 7, 5, 7},
		'7': {7, 1, 1, 2, 2},
		'8': {7, 5, 7, 5, 7},
		'9': {7, 5, 7, 1, 7},
		'A': {7, 5, 7, 5, 5},
		'B': {6, 5, 6, 5, 6},
		'C': {3, 4, 4, 4, 3},
		'D': {6, 5, 5, 5, 6},
		'E': {7, 4, 6, 4, 7},
		'F': {7, 4, 6, 4, 4},
		'G': {3, 4, 5, 5, 3},
		'H': {5, 5, 7, 5, 5},
		'I': {7, 2, 2, 2, 7},
		'J': {1, 1, 1, 5, 2},
		'K': {5, 5, 6, 5, 5},
		'L': {4, 4, 4, 4, 7},
		'M': {5, 7, 7, 5, 5},
		'N': {6, 5, 5, 5, 5},
		'O': {7, 5, 5, 5, 7},
		'P': {6, 5, 6, 4, 4},
		'Q': {6, 5, 5, 6, 2},
		'R': {6, 5, 6, 5, 5},
		'S': {7, 4, 7, 1, 7},
		'T': {7, 2, 2, 2, 2},
		'U': {5, 5, 5, 5, 7},
		'V': {5, 5, 5, 5, 2},
		'W': {5, 5, 7, 7, 5},
		'X': {5, 5, 2, 5, 5},
		'Y': {5, 5, 2, 2, 2},
		'Z': {7, 1, 2, 4, 7},
		' ': {0, 0, 0, 0, 0},
		':': {2, 2, 0, 2, 2},
		'-': {0, 0, 7, 0, 0},
		'=': {0, 7, 0, 7, 0},
		'?': {7, 5, 3, 0, 2},
	},
}
