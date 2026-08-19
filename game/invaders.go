package game

import "math/rand/v2"

const (
	InvaderCols = 11
	InvaderRows = 5
	InvaderW    = 20
	InvaderH    = 15
	InvaderHGap = 22
	InvaderVGap = 24

	// 2px (not the spec's 8px): the 240px-wide grid leaves only 8px of
	// sweep room on the 256px screen, so an 8px step could never move
	// horizontally without crossing the 4px edge margin.
	invaderStep         = 2
	invaderDrop         = 8
	invaderBaseY        = 32
	invaderLevelDrop    = 8
	invaderMinY         = 16
	invaderBaseInterval = 48
	invaderMinInterval  = 4
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
	rng          *rand.Rand
}

func NewInvaderGrid(level int, rng *rand.Rand) InvaderGrid {
	ig := InvaderGrid{
		Dir:          1,
		totalInitial: InvaderRows * InvaderCols,
		rng:          rng,
	}
	startX := (ScreenW - ((InvaderCols-1)*InvaderHGap + InvaderW)) / 2
	startY := invaderStartY(level)
	for r := range InvaderRows {
		for c := range InvaderCols {
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

// Update advances the grid by one frame. Returns true if the grid stepped.
func (ig *InvaderGrid) Update() bool {
	if ig.AliveCount() == 0 {
		return false
	}
	ig.FrameCounter++
	if ig.FrameCounter < ig.StepInterval() {
		return false
	}
	ig.FrameCounter = 0
	ig.toggleAnim()

	left, right := ig.bounds()
	if ig.Dir > 0 && right+invaderStep > ScreenW-EdgeMargin-InvaderW {
		ig.moveY(invaderDrop)
		ig.Dir = -1
		return true
	}
	if ig.Dir < 0 && left-invaderStep < EdgeMargin {
		ig.moveY(invaderDrop)
		ig.Dir = 1
		return true
	}
	ig.moveX(ig.Dir * invaderStep)
	return true
}

func (ig *InvaderGrid) moveX(dx int) {
	for r := range ig.Invaders {
		for c := range ig.Invaders[r] {
			iv := &ig.Invaders[r][c]
			if iv.Alive {
				iv.X += dx
			}
		}
	}
}

func (ig *InvaderGrid) moveY(dy int) {
	for r := range ig.Invaders {
		for c := range ig.Invaders[r] {
			iv := &ig.Invaders[r][c]
			if iv.Alive {
				iv.Y += dy
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
		return invaderMinInterval
	}
	iv := invaderBaseInterval * alive / ig.totalInitial
	if iv < invaderMinInterval {
		iv = invaderMinInterval
	}
	return iv
}

// ShouldShoot reports whether the grid fires this step.
// Probability is inversely proportional to the step interval (faster = more likely).
func (ig *InvaderGrid) ShouldShoot() bool {
	return ig.rng.IntN(2*ig.StepInterval()) < invaderBaseInterval
}

// PickShooter returns the bottom-most alive invader of a random non-empty column.
func (ig *InvaderGrid) PickShooter() *Invader {
	if ig.AliveCount() == 0 {
		return nil
	}
	col := ig.rng.IntN(InvaderCols)
	for {
		if iv := ig.BottomOfColumn(col); iv != nil {
			return iv
		}
		col = (col + 1) % InvaderCols
	}
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
