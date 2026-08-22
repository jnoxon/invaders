package render

import (
	"invaders/game"
)

// Buffer is an in-memory Canvas: a 256x224 RGBA framebuffer the host can
// upload to the canvas with a single putImageData call per frame.
type Buffer struct {
	pix []byte
}

// NewBuffer returns a black framebuffer at the logical screen size.
func NewBuffer() *Buffer {
	return &Buffer{pix: make([]byte, game.ScreenW*game.ScreenH*4)}
}

// Pix returns the raw RGBA framebuffer for upload.
func (b *Buffer) Pix() []byte {
	return b.pix
}

// Clear fills the framebuffer with black.
func (b *Buffer) Clear() {
	for i := range b.pix {
		b.pix[i] = 0
	}
}

// FillRect sets a rectangle to white, clamped to the screen.
func (b *Buffer) FillRect(x, y, w, h int) {
	x0, y0 := x, y
	x1, y1 := x+w, y+h
	if x0 < 0 {
		x0 = 0
	}
	if y0 < 0 {
		y0 = 0
	}
	if x1 > game.ScreenW {
		x1 = game.ScreenW
	}
	if y1 > game.ScreenH {
		y1 = game.ScreenH
	}
	if x0 >= x1 || y0 >= y1 {
		return
	}
	for yy := y0; yy < y1; yy++ {
		row := b.pix[yy*game.ScreenW*4+x0*4 : yy*game.ScreenW*4+x1*4]
		for i := range row {
			row[i] = 255
		}
	}
}
