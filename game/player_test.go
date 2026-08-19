package game

import "testing"

func TestPlayerMoveLeft(t *testing.T) {
	p := NewPlayer()
	in := NewInputState()
	in.Left = true
	x := p.X
	p.Update(&in)
	if p.X != x-PlayerSpeed {
		t.Fatalf("x = %d, want %d", p.X, x-PlayerSpeed)
	}
}

func TestPlayerMoveRight(t *testing.T) {
	p := NewPlayer()
	in := NewInputState()
	in.Right = true
	x := p.X
	p.Update(&in)
	if p.X != x+PlayerSpeed {
		t.Fatalf("x = %d, want %d", p.X, x+PlayerSpeed)
	}
}

func TestPlayerClamp(t *testing.T) {
	p := NewPlayer()
	p.X = 0
	in := NewInputState()
	in.Left = true
	p.Update(&in)
	if p.X != 0 {
		t.Fatalf("left clamp failed: x=%d", p.X)
	}

	p.X = ScreenW - PlayerW
	in.Left = false
	in.Right = true
	p.Update(&in)
	if p.X != ScreenW-PlayerW {
		t.Fatalf("right clamp failed: x=%d", p.X)
	}
}

func TestPlayerDoesNotMoveWhenDead(t *testing.T) {
	p := NewPlayer()
	p.Alive = false
	x := p.X
	in := NewInputState()
	in.Right = true
	p.Update(&in)
	if p.X != x {
		t.Fatalf("dead player moved: %d -> %d", x, p.X)
	}
}

func TestPlayerFire(t *testing.T) {
	p := NewPlayer()
	b := p.Fire()
	if b == nil {
		t.Fatal("Fire returned nil")
	}
	wantX := p.X + p.W/2 - 1
	if b.X != wantX || b.Y != p.Y {
		t.Fatalf("bullet pos = (%d,%d), want (%d,%d)", b.X, b.Y, wantX, p.Y)
	}
	if b.Owner != BulletPlayer || !b.Active {
		t.Fatalf("bullet owner/active wrong: %+v", b)
	}
}

func TestPlayerFireWhenDead(t *testing.T) {
	p := NewPlayer()
	p.Alive = false
	if b := p.Fire(); b != nil {
		t.Fatalf("dead player fired: %+v", b)
	}
}

func TestPlayerHitAndRespawn(t *testing.T) {
	p := NewPlayer()
	p.X = 5
	p.Hit()
	if p.Alive {
		t.Fatal("player still alive after Hit")
	}
	p.Respawn()
	if !p.Alive {
		t.Fatal("player not alive after Respawn")
	}
	if p.X != (ScreenW-PlayerW)/2 {
		t.Fatalf("respawn x = %d", p.X)
	}
	if p.Invulnerable != RespawnInvuln {
		t.Fatalf("invulnerable = %d, want %d", p.Invulnerable, RespawnInvuln)
	}
}

func TestPlayerInvulnerabilityCountdown(t *testing.T) {
	g := newTestGame()
	g.State = StatePlaying
	g.Player.Invulnerable = 5
	g.Tick()
	if g.Player.Invulnerable != 4 {
		t.Fatalf("invulnerable = %d, want 4", g.Player.Invulnerable)
	}
}

func TestPlayerRect(t *testing.T) {
	p := NewPlayer()
	x, y, w, h := p.Rect()
	if x != p.X || y != p.Y || w != p.W || h != p.H {
		t.Fatalf("rect = (%d,%d,%d,%d)", x, y, w, h)
	}
}
