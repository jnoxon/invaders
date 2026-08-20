package render

import (
	"fmt"
	"testing"

	"invaders/game"
)

type recRect struct{ x, y, w, h int }

type mockCanvas struct {
	clears int
	rects  []recRect
}

func (m *mockCanvas) Clear() { m.clears++ }

func (m *mockCanvas) FillRect(x, y, w, h int) {
	m.rects = append(m.rects, recRect{x, y, w, h})
}

func (m *mockCanvas) has(r recRect) bool {
	for _, got := range m.rects {
		if got == r {
			return true
		}
	}
	return false
}

func (m *mockCanvas) countInY(y0, y1 int) int {
	n := 0
	for _, r := range m.rects {
		if r.y >= y0 && r.y < y1 {
			n++
		}
	}
	return n
}

func spriteEqual(a, b Sprite) bool {
	return a.W == b.W && a.H == b.H && string(a.Data) == string(b.Data)
}

func TestSpritePixel(t *testing.T) {
	tests := []struct {
		name string
		s    Sprite
		x, y int
		want bool
	}{
		{"bottom row filled", playerShip, 0, playerShip.H - 1, true},
		{"bottom row filled middle", playerShip, 11, playerShip.H - 1, true},
		{"top corner transparent", playerShip, 0, 0, false},
		{"barrel tip", playerShip, 10, 0, true},
		{"x out of range", playerShip, 99, 3, false},
		{"y out of range", playerShip, 3, 99, false},
		{"negative x", playerShip, -1, 3, false},
		{"negative y", playerShip, 3, -1, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.Pixel(tt.x, tt.y); got != tt.want {
				t.Errorf("Pixel(%d,%d) = %v, want %v", tt.x, tt.y, got, tt.want)
			}
		})
	}
}

func TestSpriteIntegrity(t *testing.T) {
	tests := []struct {
		name   string
		s      Sprite
		w, h   int
		pixels int
	}{
		{"playerShip", playerShip, 24, 16, 168},
		{"playerFiring", playerFiring, 24, 16, 144},
		{"squidF0", squidF0, 20, 15, 152},
		{"squidF1", squidF1, 20, 15, 152},
		{"crabF0", crabF0, 20, 15, 120},
		{"crabF1", crabF1, 20, 15, 120},
		{"octopusF0", octopusF0, 20, 15, 168},
		{"octopusF1", octopusF1, 20, 15, 168},
		{"ufoSprite", ufoSprite, 40, 14, 376},
		{"enemyBulletF0", enemyBulletF0, 2, 8, 8},
		{"enemyBulletF1", enemyBulletF1, 2, 8, 8},
		{"lifeIcon", lifeIcon, 9, 5, 23},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.s.W != tt.w || tt.s.H != tt.h {
				t.Errorf("dims = %dx%d, want %dx%d", tt.s.W, tt.s.H, tt.w, tt.h)
			}
			if len(tt.s.Data) != tt.w*tt.h {
				t.Fatalf("data len = %d, want %d", len(tt.s.Data), tt.w*tt.h)
			}
			if n := countPixels(tt.s); n != tt.pixels {
				t.Errorf("pixels = %d, want %d", n, tt.pixels)
			}
		})
	}
}

func TestInvaderAnimationFramesDiffer(t *testing.T) {
	if spriteEqual(squidF0, squidF1) {
		t.Error("squid frames identical")
	}
	if spriteEqual(crabF0, crabF1) {
		t.Error("crab frames identical")
	}
	if spriteEqual(octopusF0, octopusF1) {
		t.Error("octopus frames identical")
	}
}

func TestInvaderSpriteSelection(t *testing.T) {
	if !spriteEqual(invaderSprite(game.InvaderSquid, 0), squidF0) {
		t.Error("squid frame 0 mismatch")
	}
	if !spriteEqual(invaderSprite(game.InvaderSquid, 1), squidF1) {
		t.Error("squid frame 1 mismatch")
	}
	if !spriteEqual(invaderSprite(game.InvaderCrab, 0), crabF0) {
		t.Error("crab frame 0 mismatch")
	}
	if !spriteEqual(invaderSprite(game.InvaderCrab, 1), crabF1) {
		t.Error("crab frame 1 mismatch")
	}
	if !spriteEqual(invaderSprite(game.InvaderOctopus, 0), octopusF0) {
		t.Error("octopus frame 0 mismatch")
	}
	if !spriteEqual(invaderSprite(game.InvaderOctopus, 1), octopusF1) {
		t.Error("octopus frame 1 mismatch")
	}
	// Frame wraps on the low bit.
	if !spriteEqual(invaderSprite(game.InvaderSquid, 2), squidF0) {
		t.Error("squid frame 2 should select frame 0")
	}
}

func TestSpriteMalformedPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("ragged rows did not panic")
		}
	}()
	sprite([]string{"##", "###"})
}

func TestScale2(t *testing.T) {
	in := sprite([]string{"#.", "##"})
	got := scale2(in)
	want := Sprite{W: 4, H: 4, Data: []byte{
		1, 1, 0, 0,
		1, 1, 0, 0,
		1, 1, 1, 1,
		1, 1, 1, 1,
	}}
	if !spriteEqual(got, want) {
		t.Errorf("scale2 data = %v, want %v", got.Data, want.Data)
	}
}

func TestPlace(t *testing.T) {
	p := place(10, 10, 2, 3, sprite([]string{"#.", ".#"}))
	if p.W != 10 || p.H != 10 {
		t.Fatalf("dims = %dx%d, want 10x10", p.W, p.H)
	}
	if !p.Pixel(2, 3) || !p.Pixel(3, 4) {
		t.Error("placed pixels missing")
	}
	if p.Pixel(3, 3) || p.Pixel(2, 4) {
		t.Error("transparent pixels filled")
	}
	if p.Pixel(0, 0) || p.Pixel(9, 9) {
		t.Error("outside placement filled")
	}
}

func TestDrawSpriteRuns(t *testing.T) {
	s := sprite([]string{"####", "##..", ".#.#", "...."})
	m := &mockCanvas{}
	r := NewRenderer(m)
	r.drawSprite(s, 10, 20)
	want := []recRect{
		{10, 20, 4, 1},
		{10, 21, 2, 1},
		{11, 22, 1, 1},
		{13, 22, 1, 1},
	}
	if len(m.rects) != len(want) {
		t.Fatalf("rects = %d, want %d: %v", len(m.rects), len(want), m.rects)
	}
	for i := range want {
		if m.rects[i] != want[i] {
			t.Errorf("rect %d = %+v, want %+v", i, m.rects[i], want[i])
		}
	}
}

func TestDrawTextGlyphs(t *testing.T) {
	m := &mockCanvas{}
	r := NewRenderer(m)
	r.drawText("AB", 0, 0)

	// A = {7,5,7,5,5}: 12 px; B = {6,5,6,5,6}: 10 px.
	if len(m.rects) != 22 {
		t.Errorf("rects = %d, want 22: %v", len(m.rects), m.rects)
	}
	// A row 0 (111) and row 1 (101).
	for _, want := range []recRect{{0, 0, 1, 1}, {1, 0, 1, 1}, {2, 0, 1, 1}, {0, 1, 1, 1}, {2, 1, 1, 1}, {2, 4, 1, 1}} {
		if !m.has(want) {
			t.Errorf("missing A px %+v", want)
		}
	}
	// B starts at x=4: row 0 (110), row 1 (101), row 4 (110).
	for _, want := range []recRect{{4, 0, 1, 1}, {5, 0, 1, 1}, {4, 1, 1, 1}, {6, 1, 1, 1}, {4, 4, 1, 1}, {5, 4, 1, 1}} {
		if !m.has(want) {
			t.Errorf("missing B px %+v", want)
		}
	}
}

func TestDrawTextUnknownCharSkipped(t *testing.T) {
	m := &mockCanvas{}
	r := NewRenderer(m)
	r.drawText("A?B", 0, 0)
	// '?' is undefined: A at x=0, B at x=8.
	for _, want := range []recRect{{8, 0, 1, 1}, {9, 0, 1, 1}, {8, 4, 1, 1}, {9, 4, 1, 1}} {
		if !m.has(want) {
			t.Errorf("missing B px %+v after skipped char", want)
		}
	}
}

func TestTextWidth(t *testing.T) {
	tests := []struct {
		s    string
		want int
	}{
		{"", 0},
		{"A", 3},
		{"AB", 7},
		{"PRESS ENTER TO START", 79},
	}
	for _, tt := range tests {
		if got := PixelFont.TextWidth(tt.s); got != tt.want {
			t.Errorf("TextWidth(%q) = %d, want %d", tt.s, got, tt.want)
		}
	}
}

func TestPixelFontCoverage(t *testing.T) {
	for _, c := range "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ :-=?" {
		if !PixelFont.Has(byte(c)) {
			t.Errorf("missing glyph %q", c)
		}
	}
}

func TestRenderClearsEveryFrame(t *testing.T) {
	g := game.NewGame()
	for st, name := range map[game.GameState]string{
		game.StateStart:           "start",
		game.StatePlaying:         "playing",
		game.StateGameOver:        "gameover",
		game.StateLevelTransition: "transition",
		game.StatePaused:          "paused",
	} {
		t.Run(name, func(t *testing.T) {
			g.State = st
			m := &mockCanvas{}
			r := NewRenderer(m)
			r.Render(g)
			if m.clears != 1 {
				t.Errorf("clears = %d, want 1", m.clears)
			}
		})
	}
}

func TestRenderStartScreen(t *testing.T) {
	g := game.NewGame()
	g.HighScore = 12345
	g.Frame = 0
	m := &mockCanvas{}
	r := NewRenderer(m)
	r.Render(g)

	// Title centered at y=16: 14 chars -> x=(256-55)/2=100.
	if !m.has(recRect{100, 16, 1, 1}) {
		t.Error("missing title 'S'")
	}
	// High score centered at y=40: 15 chars -> x=(256-59)/2=98.
	if !m.has(recRect{98, 40, 1, 1}) {
		t.Error("missing high score 'H'")
	}
	// UFO art row 0 (10 cells) -> 20px run at x=72+10.
	if !m.has(recRect{82, 64, 20, 1}) {
		t.Error("missing UFO dome top")
	}
	// Squid antennae at (88+6, 88).
	if !m.has(recRect{94, 88, 2, 1}) {
		t.Error("missing squid antennae")
	}
	// Crab art row 0 is "#....#" (6-wide): pincers at +2 and +12.
	if !m.has(recRect{90, 108, 2, 1}) || !m.has(recRect{100, 108, 2, 1}) {
		t.Error("missing crab pincers")
	}
	// Octopus head is 4 art cells -> 8px run at (88+6, 128).
	if !m.has(recRect{94, 128, 8, 1}) {
		t.Error("missing octopus head")
	}
	// Blinking prompt visible at frame 0: 20 chars -> x=(256-79)/2=88.
	if !m.has(recRect{88, 168, 1, 1}) {
		t.Error("missing start prompt")
	}
}

func gameplayGame() *game.Game {
	g := game.NewGame()
	g.State = game.StatePlaying
	g.Score = 100
	return g
}

func TestRenderGameplay(t *testing.T) {
	g := gameplayGame()
	m := &mockCanvas{}
	r := NewRenderer(m)
	r.Render(g)

	// Score text at (4,2): 'S' row 0 is 111.
	for _, want := range []recRect{{4, 2, 1, 1}, {5, 2, 1, 1}, {6, 2, 1, 1}} {
		if !m.has(want) {
			t.Errorf("missing score px %+v", want)
		}
	}
	// Level text right-aligned: "LV: 1" -> x=256-4-19=233.
	if !m.has(recRect{233, 2, 1, 1}) {
		t.Error("missing level 'L'")
	}

	// Player idle frame: barrel tip, then the 6px ### row.
	px, py := g.Player.X, g.Player.Y
	if !m.has(recRect{px + 10, py, 2, 1}) {
		t.Error("missing player barrel")
	}
	if !m.has(recRect{px + 8, py + 2, 6, 1}) {
		t.Error("missing player body row")
	}
	if m.has(recRect{px + 10, py + 2, 2, 1}) {
		t.Error("idle frame should not have flame at row 2")
	}

	// Top-left invader is a squid: antennae at +6 and +12.
	iv := &g.Invaders.Invaders[0][0]
	if !m.has(recRect{iv.X + 6, iv.Y, 2, 1}) {
		t.Error("missing invader antennae")
	}
	if !m.has(recRect{iv.X + 12, iv.Y, 2, 1}) {
		t.Error("missing invader second antenna")
	}

	// Ground line and life icons.
	if !m.has(recRect{0, groundY, game.ScreenW, 1}) {
		t.Error("missing ground line")
	}
	if got := m.countInY(lifeY, lifeY+5); got != 3*7 {
		t.Errorf("life icon rects = %d, want 21", got)
	}
	if !m.has(recRect{4, lifeY + 4, 9, 1}) {
		t.Error("missing first life icon base")
	}
}

func TestBarricadeRendering(t *testing.T) {
	g := gameplayGame()
	b := &g.Barricades[0]
	for r := range b.Pixels {
		for c := range b.Pixels[r] {
			b.Pixels[r][c] = false
		}
	}
	b.Pixels[0][5] = true
	m := &mockCanvas{}
	r := NewRenderer(m)
	r.Render(g)
	if !m.has(recRect{b.X + 10, b.Y, 2, 2}) {
		t.Errorf("missing barricade pixel at (%d, %d)", b.X+10, b.Y)
	}

	// Intact barricades: one 2x2 rect per filled shape pixel, all four
	// sharing the same Y band.
	g2 := gameplayGame()
	m2 := &mockCanvas{}
	NewRenderer(m2).Render(g2)
	want := 0
	for i := range g2.Barricades {
		for _, row := range g2.Barricades[i].Pixels {
			for _, p := range row {
				if p {
					want++
				}
			}
		}
	}
	if got := m2.countInY(g2.Barricades[0].Y, g2.Barricades[0].Y+16); got != want {
		t.Errorf("intact barricade rects = %d, want %d", got, want)
	}
}

func TestRenderGameplayFiringFrame(t *testing.T) {
	g := gameplayGame()
	g.Bullets = []game.Bullet{{X: 127, Y: 150, Owner: game.BulletPlayer, Active: true}}
	m := &mockCanvas{}
	r := NewRenderer(m)
	r.Render(g)

	// Player bullet.
	if !m.has(recRect{127, 150, game.BulletW, game.BulletH}) {
		t.Error("missing player bullet")
	}
	// Firing frame: flame extends to rows 2 and 3; no ### row at row 2.
	px, py := g.Player.X, g.Player.Y
	if !m.has(recRect{px + 10, py + 2, 2, 1}) || !m.has(recRect{px + 10, py + 3, 2, 1}) {
		t.Error("missing firing-frame flame")
	}
	if m.has(recRect{px + 8, py + 2, 6, 1}) {
		t.Error("firing frame should not have ### row at row 2")
	}
}

func TestRenderGameplayEnemyBullet(t *testing.T) {
	g := gameplayGame()
	g.Bullets = []game.Bullet{{X: 50, Y: 100, Owner: game.BulletEnemy, Active: true}}
	m := &mockCanvas{}
	NewRenderer(m).Render(g)
	// Even Y -> frame 0: row 0 single right px, row 1 2px run.
	if !m.has(recRect{50, 100, 1, 1}) {
		t.Error("missing enemy bullet frame 0 row 0")
	}
	if !m.has(recRect{50, 101, 2, 1}) {
		t.Error("missing enemy bullet frame 0 row 1 run")
	}

	// Odd Y -> frame 1: row 0 empty, row 1 single left px.
	g.Bullets = []game.Bullet{{X: 50, Y: 101, Owner: game.BulletEnemy, Active: true}}
	m = &mockCanvas{}
	NewRenderer(m).Render(g)
	if m.has(recRect{50, 101, 1, 1}) {
		t.Error("frame 1 row 0 should be empty")
	}
	if !m.has(recRect{51, 102, 1, 1}) {
		t.Error("missing enemy bullet frame 1 row 1")
	}
}

func TestRenderGameplayUFO(t *testing.T) {
	g := gameplayGame()
	g.UFOActive = true
	g.UFO = game.UFO{X: 60, Y: game.UFOY, Dir: 1, Active: true}
	m := &mockCanvas{}
	r := NewRenderer(m)
	r.Render(g)
	// UFO art row 2 (14 cells) -> 28px run over two scaled rows.
	if !m.has(recRect{66, game.UFOY + 4, 28, 1}) || !m.has(recRect{66, game.UFOY + 5, 28, 1}) {
		t.Error("missing UFO body run")
	}
}

func TestRenderPaused(t *testing.T) {
	g := gameplayGame()
	g.State = game.StatePaused
	m := &mockCanvas{}
	r := NewRenderer(m)
	r.Render(g)
	// Gameplay still drawn under the PAUSED text.
	if !m.has(recRect{0, groundY, game.ScreenW, 1}) {
		t.Error("missing gameplay under pause")
	}
	// "PAUSED" centered at y=104: 6 chars -> x=(256-23)/2=116.
	if !m.has(recRect{116, 104, 1, 1}) {
		t.Error("missing PAUSED text")
	}
}

func TestRenderGameOver(t *testing.T) {
	g := game.NewGame()
	g.State = game.StateGameOver
	g.Score = 42
	g.Frame = 0
	m := &mockCanvas{}
	r := NewRenderer(m)
	r.Render(g)
	// "GAME OVER" centered: 9 chars -> x=(256-35)/2=110; 'G' row 0 starts at col 1.
	if !m.has(recRect{111, 72, 1, 1}) {
		t.Error("missing GAME OVER")
	}
	// "SCORE: 00042" centered: 12 chars -> x=(256-47)/2=104.
	if !m.has(recRect{104, 96, 1, 1}) {
		t.Error("missing final score")
	}
	// Blink prompt visible at frame 0: 22 chars -> x=(256-87)/2=84.
	if !m.has(recRect{84, 120, 1, 1}) {
		t.Error("missing restart prompt")
	}
}

func TestRenderLevelTransition(t *testing.T) {
	g := game.NewGame()
	g.State = game.StateLevelTransition
	g.Level = 2
	m := &mockCanvas{}
	r := NewRenderer(m)
	r.Render(g)
	// "LEVEL 2" centered: 7 chars -> x=(256-27)/2=114.
	if !m.has(recRect{114, 104, 1, 1}) {
		t.Error("missing LEVEL text")
	}
	if len(m.rects) == 0 {
		t.Error("nothing drawn")
	}
}

func TestHUDFormatting(t *testing.T) {
	g := game.NewGame()
	g.State = game.StatePlaying
	g.Score = 42
	g.HighScore = 100000
	g.Level = 7
	m := &mockCanvas{}
	r := NewRenderer(m)
	r.drawHUD(g)

	// "SCORE: 00042" at (4,2).
	for _, want := range []recRect{{4, 2, 1, 1}, {5, 2, 1, 1}, {6, 2, 1, 1}} {
		if !m.has(want) {
			t.Errorf("missing score 'S' px %+v", want)
		}
	}
	// "HI-SCORE: 100000" centered: 16 chars -> x=(256-63)/2=96.
	if !m.has(recRect{96, 2, 1, 1}) {
		t.Error("missing high score 'H'")
	}
	// "LV: 7" right-aligned: x=256-4-19=233.
	if !m.has(recRect{233, 2, 1, 1}) {
		t.Error("missing level 'L'")
	}
}

func TestAllStatesRender(t *testing.T) {
	for st := game.StateStart; st <= game.StatePaused; st++ {
		t.Run(fmt.Sprint(int(st)), func(t *testing.T) {
			g := game.NewGame()
			g.State = st
			m := &mockCanvas{}
			r := NewRenderer(m)
			r.Render(g)
			if m.clears != 1 {
				t.Errorf("clears = %d, want 1", m.clears)
			}
			if len(m.rects) == 0 {
				t.Error("nothing drawn")
			}
		})
	}
}
