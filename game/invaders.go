package game

const (
	InvaderCols = 11
	InvaderRows = 5
	InvaderW    = 20
	InvaderH    = 15
	InvaderHGap = 24
	InvaderVGap = 24

	invaderStep         = 8
	invaderDrop         = 8
	invaderBaseY        = 48
	invaderLevelDrop    = 8
	invaderMinY         = 16
	invaderBaseInterval = 48
	EdgeMargin          = 4
)

type InvaderType int

const (
	InvaderSquid InvaderType = iota
	InvaderCrab
	InvaderOctopus
)

func (t InvaderType) Points() int {
	switch t {
	case InvaderSquid:
		return 30
	case InvaderCrab:
		return 20
	default:
		return 10
	}
}

type Invader struct {
	X, Y      int
	Type      InvaderType
	Alive     bool
	AnimFrame int
}

type InvaderGrid struct {
	Invaders     [InvaderRows][InvaderCols]Invader
	Dir          int
	FrameCounter int
	totalInitial int
}

func NewInvaderGrid(level int) InvaderGrid {
	var ig InvaderGrid
	ig.Dir = 1
	ig.totalInitial = InvaderRows * InvaderCols
	startX := (ScreenW - ((InvaderCols-1)*InvaderHGap + InvaderW)) / 2
	startY := invaderStartY(level)
	for r := 0; r < InvaderRows; r++ {
		for c := 0; c < InvaderCols; c++ {
			ig.Invaders[r][c] = Invader{
				X:     startX + c*InvaderHGap,
				Y:     startY + r*InvaderVGap,
				Type:  invaderTypeForRow(r),
				Alive: true,
			}
		}
	}
	return ig
}

func invaderStartY(level int) int {
	y := invaderBaseY - (level-1)*invaderLevelDrop
	if y < invaderMinY {
		y = invaderMinY
	}
	return y
}

func invaderTypeForRow(r int) InvaderType {
	switch {
	case r == 0:
		return InvaderSquid
	case r == InvaderRows-1:
		return InvaderOctopus
	default:
		return InvaderCrab
	}
}

func (ig *InvaderGrid) Update() {
	if ig.AliveCount() == 0 {
		return
	}
	ig.FrameCounter++
	if ig.FrameCounter < ig.StepInterval() {
		return
	}
	ig.FrameCounter = 0
	ig.toggleAnim()

	left, right := ig.bounds()
	drop := false
	if ig.Dir > 0 && right+InvaderW > ScreenW-EdgeMargin {
		ig.Dir = -1
		drop = true
	} else if ig.Dir < 0 && left < EdgeMargin {
		ig.Dir = 1
		drop = true
	}
	for r := range ig.Invaders {
		for c := range ig.Invaders[r] {
			iv := &ig.Invaders[r][c]
			if !iv.Alive {
				continue
			}
			iv.X += ig.Dir * invaderStep
			if drop {
				iv.Y += invaderDrop
			}
		}
	}
}

func (ig *InvaderGrid) toggleAnim() {
	for r := range ig.Invaders {
		for c := range ig.Invaders[r] {
			iv := &ig.Invaders[r][c]
			if iv.Alive {
				iv.AnimFrame ^= 1
			}
		}
	}
}

func (ig *InvaderGrid) bounds() (left, right int) {
	left, right = ScreenW, 0
	for r := range ig.Invaders {
		for c := range ig.Invaders[r] {
			iv := &ig.Invaders[r][c]
			if !iv.Alive {
				continue
			}
			if iv.X < left {
				left = iv.X
			}
			if iv.X > right {
				right = iv.X
			}
		}
	}
	if left > right {
		return 0, 0
	}
	return left, right
}

func (ig *InvaderGrid) AliveCount() int {
	n := 0
	for r := range ig.Invaders {
		for c := range ig.Invaders[r] {
			if ig.Invaders[r][c].Alive {
				n++
			}
		}
	}
	return n
}

func (ig *InvaderGrid) StepInterval() int {
	alive := ig.AliveCount()
	if alive <= 0 {
		return 1
	}
	iv := invaderBaseInterval * alive / ig.totalInitial
	if iv < 1 {
		iv = 1
	}
	return iv
}

func (ig *InvaderGrid) BottomOfColumn(col int) *Invader {
	if col < 0 || col >= InvaderCols {
		return nil
	}
	for r := InvaderRows - 1; r >= 0; r-- {
		if ig.Invaders[r][col].Alive {
			return &ig.Invaders[r][col]
		}
	}
	return nil
}

func (ig *InvaderGrid) ReachedBottom() bool {
	for r := range ig.Invaders {
		for c := range ig.Invaders[r] {
			iv := &ig.Invaders[r][c]
			if iv.Alive && iv.Y > InvaderBottomThreshold {
				return true
			}
		}
	}
	return false
}

func (g *Game) enemyFire() {
	if countBullets(g.Bullets, BulletEnemy) >= MaxEnemyBullets {
		return
	}
	alive := g.Invaders.AliveCount()
	if alive == 0 {
		return
	}
	g.enemyFireCounter++
	interval := invaderBaseInterval * alive / (InvaderRows * InvaderCols)
	if interval < 1 {
		interval = 1
	}
	if g.enemyFireCounter < interval {
		return
	}
	g.enemyFireCounter = 0
	col := g.RNG.Intn(InvaderCols)
	iv := g.Invaders.BottomOfColumn(col)
	if iv == nil {
		return
	}
	g.Bullets = append(g.Bullets, Bullet{
		X:      iv.X + InvaderW/2,
		Y:      iv.Y + InvaderH,
		Owner:  BulletEnemy,
		Active: true,
	})
}
