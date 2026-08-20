package game

func AABB(ax, ay, aw, ah, bx, by, bw, bh int) bool {
	if aw <= 0 || ah <= 0 || bw <= 0 || bh <= 0 {
		return false
	}
	return ax < bx+bw && bx < ax+aw && ay < by+bh && by < ay+ah
}

func (g *Game) checkPlayerBulletHits() {
	for i := range g.Bullets {
		b := &g.Bullets[i]
		if !b.Active || b.Owner != BulletPlayer {
			continue
		}
		bx, by, bw, bh := b.Rect()
		if g.hitInvader(bx, by, bw, bh) {
			b.Active = false
			continue
		}
		if g.UFOActive {
			ux, uy, uw, uh := g.UFO.Rect()
			if AABB(bx, by, bw, bh, ux, uy, uw, uh) {
				g.AddScore(g.UFO.Points())
				g.UFOActive = false
				b.Active = false
				continue
			}
		}
		if g.hitBarricade(bx, by, bw, bh) {
			b.Active = false
		}
	}
}

func (g *Game) hitInvader(bx, by, bw, bh int) bool {
	for r := range g.Invaders.Invaders {
		for c := range g.Invaders.Invaders[r] {
			iv := &g.Invaders.Invaders[r][c]
			if !iv.Alive {
				continue
			}
			if AABB(bx, by, bw, bh, iv.X, iv.Y, InvaderW, InvaderH) {
				iv.Alive = false
				g.KillCount++
				g.AddScore(iv.Type.Points())
				return true
			}
		}
	}
	return false
}

func (g *Game) hitBarricade(bx, by, bw, bh int) bool {
	for i := range g.Barricades {
		bar := &g.Barricades[i]
		if bar.Destroyed() {
			continue
		}
		x, y, w, h := bar.Rect()
		if !AABB(bx, by, bw, bh, x, y, w, h) {
			continue
		}
		if bar.PixelAt(bx+bw/2, by) {
			bar.Damage(bx+bw/2, by, g.RNG)
			return true
		}
	}
	return false
}

func (g *Game) checkEnemyBulletHits() {
	for i := range g.Bullets {
		b := &g.Bullets[i]
		if !b.Active || b.Owner != BulletEnemy {
			continue
		}
		bx, by, bw, bh := b.Rect()
		if g.Player.Alive && g.Player.Invulnerable == 0 {
			px, py, pw, ph := g.Player.Rect()
			if AABB(bx, by, bw, bh, px, py, pw, ph) {
				b.Active = false
				g.Player.Hit()
				g.HandlePlayerDeath()
				continue
			}
		}
		if g.hitBarricade(bx, by, bw, bh) {
			b.Active = false
		}
	}
}

func (g *Game) checkInvaderBarricadeCollision() {
	for r := range g.Invaders.Invaders {
		for c := range g.Invaders.Invaders[r] {
			iv := &g.Invaders.Invaders[r][c]
			if !iv.Alive {
				continue
			}
			for i := range g.Barricades {
				bar := &g.Barricades[i]
				if bar.Destroyed() {
					continue
				}
				x, y, w, h := bar.Rect()
				if AABB(iv.X, iv.Y, InvaderW, InvaderH, x, y, w, h) {
					bar.OverlapRect(iv.X, iv.Y, InvaderW, InvaderH)
				}
			}
		}
	}
}
