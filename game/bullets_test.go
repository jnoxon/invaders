package game

import "testing"

func TestPlayerBulletMovesUp(t *testing.T) {
	b := Bullet{X: 100, Y: 100, Owner: BulletPlayer, Active: true}
	b.Update()
	if b.Y != 100-PlayerBulletSpeed {
		t.Fatalf("y = %d, want %d", b.Y, 100-PlayerBulletSpeed)
	}
}

func TestPlayerBulletDeactivatesOffScreen(t *testing.T) {
	b := Bullet{X: 100, Y: -8, Owner: BulletPlayer, Active: true}
	b.Update()
	if b.Active {
		t.Fatal("player bullet should deactivate off-screen")
	}
}

func TestEnemyBulletMovesDown(t *testing.T) {
	b := Bullet{X: 100, Y: 100, Owner: BulletEnemy, Active: true}
	b.Update()
	if b.Y != 100+EnemyBulletSpeed {
		t.Fatalf("y = %d, want %d", b.Y, 100+EnemyBulletSpeed)
	}
}

func TestEnemyBulletDeactivatesOffScreen(t *testing.T) {
	b := Bullet{X: 100, Y: ScreenH - 1, Owner: BulletEnemy, Active: true}
	b.Update()
	if b.Active {
		t.Fatal("enemy bullet should deactivate off-screen")
	}
}

func TestInactiveBulletDoesNotMove(t *testing.T) {
	b := Bullet{X: 100, Y: 100, Owner: BulletPlayer, Active: false}
	b.Update()
	if b.Y != 100 {
		t.Fatal("inactive bullet should not move")
	}
}

func TestBulletRect(t *testing.T) {
	b := Bullet{X: 10, Y: 20, Owner: BulletPlayer, Active: true}
	x, y, w, h := b.Rect()
	if x != 10 || y != 20 || w != BulletW || h != BulletH {
		t.Fatalf("rect = (%d,%d,%d,%d)", x, y, w, h)
	}
}

func TestActiveCount(t *testing.T) {
	bullets := []Bullet{
		{Owner: BulletPlayer, Active: true},
		{Owner: BulletPlayer, Active: false},
		{Owner: BulletEnemy, Active: true},
		{Owner: BulletEnemy, Active: false},
	}
	if ActiveCount(bullets, BulletPlayer) != 1 {
		t.Fatal("player bullet count wrong")
	}
	if ActiveCount(bullets, BulletEnemy) != 1 {
		t.Fatal("enemy bullet count wrong (should ignore inactive)")
	}
}

func TestPlayerLimitedToOneBullet(t *testing.T) {
	g := newTestGame()
	g.State = StatePlaying
	g.HandleInput("Space", true)
	g.Tick()
	g.HandleInput("Space", true)
	g.Tick()
	if ActiveCount(g.Bullets, BulletPlayer) != 1 {
		t.Fatalf("player bullets = %d, want 1", ActiveCount(g.Bullets, BulletPlayer))
	}
}

func TestCanFireRespectsMax(t *testing.T) {
	g := newTestGame()
	if !CanFire(g.Bullets, BulletPlayer) || !CanFire(g.Bullets, BulletEnemy) {
		t.Fatal("should be able to fire with no bullets")
	}
	g.Bullets = []Bullet{{Owner: BulletPlayer, Active: false}}
	if !CanFire(g.Bullets, BulletPlayer) {
		t.Fatal("inactive bullets should not block firing")
	}
	g.Bullets = []Bullet{{Owner: BulletPlayer, Active: true}}
	if CanFire(g.Bullets, BulletPlayer) {
		t.Fatal("player at max should not fire")
	}
	if !CanFire(g.Bullets, BulletEnemy) {
		t.Fatal("enemy should still be able to fire")
	}
	g.Bullets = make([]Bullet, MaxEnemyBullets)
	for i := range g.Bullets {
		g.Bullets[i] = Bullet{Owner: BulletEnemy, Active: true}
	}
	if CanFire(g.Bullets, BulletEnemy) {
		t.Fatal("enemy at max should not fire")
	}
	if !CanFire(g.Bullets, BulletPlayer) {
		t.Fatal("player should still be able to fire")
	}
}
