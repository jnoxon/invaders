package game

const (
	PlayerBulletSpeed = 6
	EnemyBulletSpeed  = 3
	MaxPlayerBullets  = 1
	MaxEnemyBullets   = 4
	BulletW           = 2
	BulletH           = 8
)

type BulletOwner int

const (
	BulletPlayer BulletOwner = iota
	BulletEnemy
)

type Bullet struct {
	X, Y   int
	Owner  BulletOwner
	Active bool
}

func (b *Bullet) Update() {
	if !b.Active {
		return
	}
	if b.Owner == BulletPlayer {
		b.Y -= PlayerBulletSpeed
		if b.Y+BulletH < 0 {
			b.Active = false
		}
	} else {
		b.Y += EnemyBulletSpeed
		if b.Y > ScreenH {
			b.Active = false
		}
	}
}

func (b *Bullet) Rect() (int, int, int, int) {
	return b.X, b.Y, BulletW, BulletH
}

func ActiveCount(bullets []Bullet, owner BulletOwner) int {
	n := 0
	for i := range bullets {
		if bullets[i].Active && bullets[i].Owner == owner {
			n++
		}
	}
	return n
}

func maxBullets(owner BulletOwner) int {
	if owner == BulletPlayer {
		return MaxPlayerBullets
	}
	return MaxEnemyBullets
}

func CanFire(bullets []Bullet, owner BulletOwner) bool {
	return ActiveCount(bullets, owner) < maxBullets(owner)
}
