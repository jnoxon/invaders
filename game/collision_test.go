package game

import "testing"

func TestAABB(t *testing.T) {
	cases := []struct {
		name string
		a    [4]int
		b    [4]int
		want bool
	}{
		{"overlap", [4]int{0, 0, 10, 10}, [4]int{5, 5, 10, 10}, true},
		{"no overlap", [4]int{0, 0, 10, 10}, [4]int{20, 20, 10, 10}, false},
		{"touch edge", [4]int{0, 0, 10, 10}, [4]int{10, 0, 10, 10}, false},
		{"contain", [4]int{0, 0, 100, 100}, [4]int{10, 10, 10, 10}, true},
		{"zero size a", [4]int{0, 0, 0, 0}, [4]int{0, 0, 10, 10}, false},
		{"zero size b", [4]int{0, 0, 10, 10}, [4]int{5, 5, 0, 0}, false},
		{"vertical gap", [4]int{0, 0, 10, 10}, [4]int{0, 15, 10, 10}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := AABB(tc.a[0], tc.a[1], tc.a[2], tc.a[3], tc.b[0], tc.b[1], tc.b[2], tc.b[3])
			if got != tc.want {
				t.Fatalf("AABB(%v,%v) = %v, want %v", tc.a, tc.b, got, tc.want)
			}
		})
	}
}

func TestPlayerBulletKillsInvader(t *testing.T) {
	g := newTestGame()
	g.State = StatePlaying
	iv := &g.Invaders.Invaders[0][0]
	g.Bullets = []Bullet{{X: iv.X + 5, Y: iv.Y + 5, Owner: BulletPlayer, Active: true}}
	scoreBefore := g.Score
	g.checkPlayerBulletHits()
	if iv.Alive {
		t.Fatal("invader should be dead")
	}
	if g.Score != scoreBefore+iv.Type.Points() {
		t.Fatalf("score = %d, want %d", g.Score, scoreBefore+iv.Type.Points())
	}
	if g.Bullets[0].Active {
		t.Fatal("bullet should be consumed")
	}
}

func TestPlayerBulletMissesInvader(t *testing.T) {
	g := newTestGame()
	g.State = StatePlaying
	g.Bullets = []Bullet{{X: 0, Y: 0, Owner: BulletPlayer, Active: true}}
	g.checkPlayerBulletHits()
	if g.Invaders.Invaders[0][0].X != 0 && g.Score != 0 {
		t.Fatal("should not score on miss")
	}
	if !g.Bullets[0].Active {
		t.Fatal("bullet should remain active after miss")
	}
}

func TestEnemyBulletHitsPlayer(t *testing.T) {
	g := newTestGame()
	g.State = StatePlaying
	g.Player.Invulnerable = 0
	g.Lives = StartLives
	px, py, _, _ := g.Player.Rect()
	g.Bullets = []Bullet{{X: px + 5, Y: py + 5, Owner: BulletEnemy, Active: true}}
	g.checkEnemyBulletHits()
	if g.Lives != StartLives-1 {
		t.Fatalf("lives = %d, want %d", g.Lives, StartLives-1)
	}
	if !g.Player.Alive {
		t.Fatal("player should respawn with lives remaining")
	}
}

func TestEnemyBulletGameOverWhenNoLives(t *testing.T) {
	g := newTestGame()
	g.State = StatePlaying
	g.Player.Invulnerable = 0
	g.Lives = 1
	px, py, _, _ := g.Player.Rect()
	g.Bullets = []Bullet{{X: px + 5, Y: py + 5, Owner: BulletEnemy, Active: true}}
	g.checkEnemyBulletHits()
	if g.State != StateGameOver {
		t.Fatalf("state = %v, want StateGameOver", g.State)
	}
}

func TestEnemyBulletIgnoresInvulnerablePlayer(t *testing.T) {
	g := newTestGame()
	g.State = StatePlaying
	g.Player.Invulnerable = 60
	g.Lives = StartLives
	px, py, _, _ := g.Player.Rect()
	g.Bullets = []Bullet{{X: px + 5, Y: py + 5, Owner: BulletEnemy, Active: true}}
	g.checkEnemyBulletHits()
	if g.Lives != StartLives {
		t.Fatal("invulnerable player should not lose a life")
	}
}

func TestBulletHitsBarricade(t *testing.T) {
	g := newTestGame()
	g.State = StatePlaying
	bar := g.Barricades[0]
	sx := bar.X + 2
	sy := bar.Y
	g.Bullets = []Bullet{{X: sx - 1, Y: sy, Owner: BulletPlayer, Active: true}}
	g.checkPlayerBulletHits()
	if bar.Pixels[0][1] {
		t.Fatal("barricade pixel should be damaged")
	}
	if g.Bullets[0].Active {
		t.Fatal("bullet should be consumed by barricade")
	}
}

func TestBulletHitsUFO(t *testing.T) {
	g := newTestGame()
	g.State = StatePlaying
	// Clear the formation: the UFO shares the top rows with the grid, and
	// hitInvader runs before the UFO check.
	killAllBut(&g.Invaders, 0)
	g.UFO = UFO{X: 100, Y: UFOY, Dir: 1, Active: true, points: 100}
	g.UFOActive = true
	g.Bullets = []Bullet{{X: 110, Y: UFOY + 2, Owner: BulletPlayer, Active: true}}
	g.checkPlayerBulletHits()
	if g.UFOActive {
		t.Fatal("UFO should be deactivated")
	}
	if g.Score != 100 {
		t.Fatalf("score = %d, want 100", g.Score)
	}
}

func TestInvaderDestroysBarricadeOnOverlap(t *testing.T) {
	g := newTestGame()
	bar := g.Barricades[0]
	g.Invaders.Invaders[0][0] = Invader{X: bar.X, Y: bar.Y, Type: InvaderCrab, Alive: true}
	g.checkInvaderBarricadeCollision()
	if !bar.Destroyed() {
		t.Fatal("overlapping invader should destroy barricade pixels")
	}
}

func TestEnemyBulletBlockedByBarricade(t *testing.T) {
	g := newTestGame()
	g.State = StatePlaying
	lives := g.Lives
	bar := g.Barricades[0]
	sx := bar.X + 2
	g.Bullets = []Bullet{{X: sx - 1, Y: bar.Y, Owner: BulletEnemy, Active: true}}
	g.checkEnemyBulletHits()
	if g.Bullets[0].Active {
		t.Fatal("bullet should be consumed by barricade")
	}
	if g.Lives != lives {
		t.Fatal("player should not lose a life")
	}
	if bar.Pixels[0][1] {
		t.Fatal("barricade pixel should be damaged")
	}
}

func TestMultipleBulletsSameFrame(t *testing.T) {
	g := newTestGame()
	g.State = StatePlaying
	iv1 := &g.Invaders.Invaders[0][0]
	iv2 := &g.Invaders.Invaders[0][2]
	g.Bullets = []Bullet{
		{X: iv1.X + 5, Y: iv1.Y + 5, Owner: BulletPlayer, Active: true},
		{X: iv2.X + 5, Y: iv2.Y + 5, Owner: BulletPlayer, Active: true},
	}
	g.checkPlayerBulletHits()
	if iv1.Alive || iv2.Alive {
		t.Fatal("both invaders should be dead")
	}
	if ActiveCount(g.Bullets, BulletPlayer) != 0 {
		t.Fatal("both bullets should be consumed")
	}
	if g.Score != InvaderSquid.Points()*2 {
		t.Fatalf("score = %d, want %d", g.Score, InvaderSquid.Points()*2)
	}
}

func TestBulletHitsFirstTargetOnly(t *testing.T) {
	g := newTestGame()
	g.State = StatePlaying
	bar := &g.Barricades[0]
	iv := &g.Invaders.Invaders[0][0]
	iv.X = bar.X + 2
	iv.Y = bar.Y - 5
	g.Bullets = []Bullet{{X: iv.X + 5, Y: iv.Y + 10, Owner: BulletPlayer, Active: true}}
	g.checkPlayerBulletHits()
	if iv.Alive {
		t.Fatal("invader should be dead")
	}
	if g.Bullets[0].Active {
		t.Fatal("bullet should be consumed by invader")
	}
	if !bar.Pixels[2][4] {
		t.Fatal("barricade should be undamaged (invader hit first)")
	}
}

func TestPlayerBulletPassesDestroyedBarricade(t *testing.T) {
	g := newTestGame()
	g.State = StatePlaying
	bar := &g.Barricades[0]
	for r := range bar.Pixels {
		for c := range bar.Pixels[r] {
			bar.Pixels[r][c] = false
		}
	}
	if !bar.Destroyed() {
		t.Fatal("barricade should be destroyed")
	}
	g.Bullets = []Bullet{{X: bar.X, Y: bar.Y, Owner: BulletPlayer, Active: true}}
	g.checkPlayerBulletHits()
	if !g.Bullets[0].Active {
		t.Fatal("bullet should pass through a destroyed barricade")
	}
}

func TestEnemyBulletPassesErasedPixel(t *testing.T) {
	g := newTestGame()
	g.State = StatePlaying
	lives := g.Lives
	bar := g.Barricades[0]
	sx := bar.X + 2
	// First shot erases the top pixel at col 1.
	g.Bullets = []Bullet{{X: sx - 1, Y: bar.Y, Owner: BulletEnemy, Active: true}}
	g.checkEnemyBulletHits()
	if g.Bullets[0].Active {
		t.Fatal("first bullet should be consumed by the barricade")
	}
	if bar.Pixels[0][1] {
		t.Fatal("top pixel should be erased")
	}
	// Second shot through the same point passes through the crater.
	g.Bullets = []Bullet{{X: sx - 1, Y: bar.Y, Owner: BulletEnemy, Active: true}}
	g.checkEnemyBulletHits()
	if !g.Bullets[0].Active {
		t.Fatal("second bullet should pass through the erased pixel")
	}
	if g.Lives != lives {
		t.Fatal("player should not lose a life")
	}
}

func TestInvaderPartiallyErodesBarricade(t *testing.T) {
	g := newTestGame()
	bar := &g.Barricades[0]
	// Invader covers rows 4-7, cols 6-10 of the pixel grid: only the
	// right leg (cols 7-8, rows 4-7) is set there.
	g.Invaders.Invaders[0][0] = Invader{X: bar.X + 12, Y: bar.Y + 8, Type: InvaderCrab, Alive: true}
	g.checkInvaderBarricadeCollision()
	if bar.Pixels[4][7] || bar.Pixels[7][8] {
		t.Fatal("overlapping right-leg pixels should be cleared")
	}
	if !bar.Pixels[0][1] || !bar.Pixels[4][1] {
		t.Fatal("non-overlapping pixels should remain")
	}
	if got := bar.PixelCount(); got != 40 {
		t.Fatalf("pixel count = %d, want 40", got)
	}
}
