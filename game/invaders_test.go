package game

import "testing"

func forceStep(ig *InvaderGrid) {
	ig.FrameCounter = ig.StepInterval() - 1
	ig.Update()
}

func placeAll(ig *InvaderGrid, x, y int) {
	for r := range ig.Invaders {
		for c := range ig.Invaders[r] {
			ig.Invaders[r][c].X = x
			ig.Invaders[r][c].Y = y
		}
	}
}

func TestNewInvaderGridInit(t *testing.T) {
	ig := NewInvaderGrid(1)
	if ig.AliveCount() != InvaderRows*InvaderCols {
		t.Fatalf("alive = %d", ig.AliveCount())
	}
	if ig.Dir != 1 {
		t.Fatalf("dir = %d", ig.Dir)
	}
	if ig.Invaders[0][0].Type != InvaderSquid {
		t.Fatal("row 0 should be squid")
	}
	if ig.Invaders[1][0].Type != InvaderCrab {
		t.Fatal("row 1 should be crab")
	}
	if ig.Invaders[InvaderRows-1][0].Type != InvaderOctopus {
		t.Fatal("bottom row should be octopus")
	}
	if ig.Invaders[0][1].X != ig.Invaders[0][0].X+InvaderHGap {
		t.Fatal("horizontal spacing wrong")
	}
	if ig.Invaders[1][0].Y != ig.Invaders[0][0].Y+InvaderVGap {
		t.Fatal("vertical spacing wrong")
	}
}

func TestHigherLevelStartsHigher(t *testing.T) {
	l1 := NewInvaderGrid(1)
	l3 := NewInvaderGrid(3)
	if l3.Invaders[0][0].Y >= l1.Invaders[0][0].Y {
		t.Fatalf("level 3 y=%d not higher than level 1 y=%d",
			l3.Invaders[0][0].Y, l1.Invaders[0][0].Y)
	}
}

func TestInvaderMovesRight(t *testing.T) {
	ig := NewInvaderGrid(1)
	placeAll(&ig, 10, 50)
	ig.Dir = 1
	before := ig.Invaders[0][0].X
	forceStep(&ig)
	if ig.Invaders[0][0].X != before+invaderStep {
		t.Fatalf("x = %d, want %d", ig.Invaders[0][0].X, before+invaderStep)
	}
	if ig.Dir != 1 {
		t.Fatal("dir should stay 1")
	}
}

func TestInvaderWrapsAtRightEdge(t *testing.T) {
	ig := NewInvaderGrid(1)
	placeAll(&ig, 240, 50)
	ig.Dir = 1
	yBefore := ig.Invaders[0][0].Y
	forceStep(&ig)
	if ig.Dir != -1 {
		t.Fatal("dir should reverse to -1")
	}
	if ig.Invaders[0][0].Y != yBefore+invaderDrop {
		t.Fatal("should drop at edge")
	}
	if ig.Invaders[0][0].X != 240-invaderStep {
		t.Fatal("should move left after reversing")
	}
}

func TestInvaderReversesAtLeftEdge(t *testing.T) {
	ig := NewInvaderGrid(1)
	placeAll(&ig, 0, 50)
	ig.Dir = -1
	forceStep(&ig)
	if ig.Dir != 1 {
		t.Fatal("dir should reverse to 1")
	}
	if ig.Invaders[0][0].X != invaderStep {
		t.Fatalf("x = %d, want %d", ig.Invaders[0][0].X, invaderStep)
	}
}

func TestInvaderStepIntervalDecreases(t *testing.T) {
	ig := NewInvaderGrid(1)
	full := ig.StepInterval()
	for r := range ig.Invaders {
		for c := range ig.Invaders[r] {
			if (r*InvaderCols+c)%2 == 0 {
				ig.Invaders[r][c].Alive = false
			}
		}
	}
	reduced := ig.StepInterval()
	if reduced >= full {
		t.Fatalf("interval should decrease: full=%d reduced=%d", full, reduced)
	}
}

func TestInvaderStepIntervalNeverBelowOne(t *testing.T) {
	ig := NewInvaderGrid(1)
	ig.Invaders[0][0].Alive = true
	for r := range ig.Invaders {
		for c := range ig.Invaders[r] {
			if r != 0 || c != 0 {
				ig.Invaders[r][c].Alive = false
			}
		}
	}
	if ig.StepInterval() < 1 {
		t.Fatal("interval must be >= 1")
	}
}

func TestInvaderAnimationToggle(t *testing.T) {
	ig := NewInvaderGrid(1)
	placeAll(&ig, 10, 50)
	before := ig.Invaders[0][0].AnimFrame
	forceStep(&ig)
	if ig.Invaders[0][0].AnimFrame != before^1 {
		t.Fatal("animation frame should toggle")
	}
}

func TestInvaderNoStepBeforeInterval(t *testing.T) {
	ig := NewInvaderGrid(1)
	placeAll(&ig, 10, 50)
	ig.Update()
	if ig.Invaders[0][0].X != 10 {
		t.Fatal("should not move before step interval")
	}
}

func TestBottomOfColumn(t *testing.T) {
	ig := NewInvaderGrid(1)
	b := ig.BottomOfColumn(0)
	if b == nil {
		t.Fatal("expected bottom invader")
	}
	if b != &ig.Invaders[InvaderRows-1][0] {
		t.Fatal("should return bottom-most alive invader")
	}
	ig.Invaders[InvaderRows-1][0].Alive = false
	b = ig.BottomOfColumn(0)
	if b != &ig.Invaders[InvaderRows-2][0] {
		t.Fatal("should return next invader up")
	}
	ig.Invaders[InvaderRows-2][0].Alive = false
	for r := InvaderRows - 3; r >= 0; r-- {
		ig.Invaders[r][0].Alive = false
	}
	if ig.BottomOfColumn(0) != nil {
		t.Fatal("empty column should return nil")
	}
	if ig.BottomOfColumn(-1) != nil || ig.BottomOfColumn(InvaderCols) != nil {
		t.Fatal("out of range column should return nil")
	}
}

func TestReachedBottom(t *testing.T) {
	ig := NewInvaderGrid(1)
	if ig.ReachedBottom() {
		t.Fatal("fresh grid should not have reached bottom")
	}
	ig.Invaders[0][0].Y = InvaderBottomThreshold + 1
	if !ig.ReachedBottom() {
		t.Fatal("should have reached bottom")
	}
}

func TestInvaderPoints(t *testing.T) {
	if InvaderSquid.Points() != 30 || InvaderCrab.Points() != 20 || InvaderOctopus.Points() != 10 {
		t.Fatal("invader points wrong")
	}
}
