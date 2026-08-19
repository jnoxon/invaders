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

func newTestGrid() InvaderGrid {
	return NewInvaderGrid(1, testRNG())
}

func killAllBut(ig *InvaderGrid, n int) {
	count := 0
	for r := range ig.Invaders {
		for c := range ig.Invaders[r] {
			if count < n {
				count++
				continue
			}
			ig.Invaders[r][c].Alive = false
		}
	}
}

func TestNewInvaderGridInit(t *testing.T) {
	ig := newTestGrid()
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
	l1 := newTestGrid()
	l3 := NewInvaderGrid(3, testRNG())
	if l3.Invaders[0][0].Y >= l1.Invaders[0][0].Y {
		t.Fatalf("level 3 y=%d not higher than level 1 y=%d",
			l3.Invaders[0][0].Y, l1.Invaders[0][0].Y)
	}
}

func TestInvaderMovesRight(t *testing.T) {
	ig := newTestGrid()
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

func TestInvaderDropsAtRightEdge(t *testing.T) {
	ig := newTestGrid()
	placeAll(&ig, 232, 50)
	ig.Dir = 1
	yBefore := ig.Invaders[0][0].Y
	forceStep(&ig)
	if ig.Dir != -1 {
		t.Fatal("dir should reverse to -1")
	}
	if ig.Invaders[0][0].Y != yBefore+invaderDrop {
		t.Fatal("should drop at edge")
	}
	if ig.Invaders[0][0].X != 232 {
		t.Fatalf("x = %d, want 232 (stayed at edge, dropped instead of moving)", ig.Invaders[0][0].X)
	}
}

func TestInvaderDropsAtLeftEdge(t *testing.T) {
	ig := newTestGrid()
	placeAll(&ig, 4, 50)
	ig.Dir = -1
	yBefore := ig.Invaders[0][0].Y
	forceStep(&ig)
	if ig.Dir != 1 {
		t.Fatal("dir should reverse to 1")
	}
	if ig.Invaders[0][0].Y != yBefore+invaderDrop {
		t.Fatal("should drop at edge")
	}
	if ig.Invaders[0][0].X != 4 {
		t.Fatalf("x = %d, want 4 (stayed at edge, dropped instead of moving)", ig.Invaders[0][0].X)
	}
}

func TestFullMovementCycle(t *testing.T) {
	ig := newTestGrid()
	x0 := ig.Invaders[0][0].X
	y0 := ig.Invaders[0][0].Y
	// One full sweep: right to the edge, drop, left to the edge, drop.
	for range 10 {
		forceStep(&ig)
	}
	if ig.Invaders[0][0].X != x0 {
		t.Fatalf("x = %d, want %d after full cycle", ig.Invaders[0][0].X, x0)
	}
	if ig.Invaders[0][0].Y != y0+2*invaderDrop {
		t.Fatalf("y = %d, want %d after full cycle", ig.Invaders[0][0].Y, y0+2*invaderDrop)
	}
}

func TestInvaderStepIntervalDecreases(t *testing.T) {
	ig := newTestGrid()
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

func TestStepIntervalByAliveCount(t *testing.T) {
	cases := []struct {
		alive int
		want  int
	}{
		{55, 48},
		{27, 23},
		{10, 8},
		{4, 4},
		{1, 4},
	}
	for _, tc := range cases {
		ig := newTestGrid()
		killAllBut(&ig, tc.alive)
		if got := ig.StepInterval(); got != tc.want {
			t.Errorf("alive=%d: interval = %d, want %d", tc.alive, got, tc.want)
		}
	}
}

func TestInvaderAnimationToggle(t *testing.T) {
	ig := newTestGrid()
	placeAll(&ig, 10, 50)
	before := ig.Invaders[0][0].AnimFrame
	forceStep(&ig)
	if ig.Invaders[0][0].AnimFrame != before^1 {
		t.Fatal("animation frame should toggle")
	}
}

func TestInvaderNoStepBeforeInterval(t *testing.T) {
	ig := newTestGrid()
	placeAll(&ig, 10, 50)
	ig.Update()
	if ig.Invaders[0][0].X != 10 {
		t.Fatal("should not move before step interval")
	}
}

func TestBottomOfColumn(t *testing.T) {
	ig := newTestGrid()
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
	ig := newTestGrid()
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

func TestStepIntervalClampsToMin(t *testing.T) {
	ig := newTestGrid()
	killAllBut(&ig, 1)
	if got := ig.StepInterval(); got != invaderMinInterval {
		t.Fatalf("interval = %d, want min %d", got, invaderMinInterval)
	}
}

func TestShouldShootProbability(t *testing.T) {
	cases := []struct {
		alive  int
		trials int
		min    int
		max    int
	}{
		{55, 1000, 400, 600},
		{1, 200, 200, 200},
	}
	for _, tc := range cases {
		ig := newTestGrid()
		killAllBut(&ig, tc.alive)
		n := 0
		for range tc.trials {
			if ig.ShouldShoot() {
				n++
			}
		}
		if n < tc.min || n > tc.max {
			t.Fatalf("alive=%d: %d/%d shots, want between %d and %d",
				tc.alive, n, tc.trials, tc.min, tc.max)
		}
	}
}

func TestPickShooterBottomOfColumn(t *testing.T) {
	ig := newTestGrid()
	for c := range ig.Invaders[InvaderRows-1] {
		ig.Invaders[InvaderRows-1][c].Alive = false
	}
	for range 100 {
		iv := ig.PickShooter()
		if iv == nil || !iv.Alive {
			t.Fatal("expected an alive shooter")
		}
		col := -1
		for c := range ig.Invaders[InvaderRows-2] {
			if iv == &ig.Invaders[InvaderRows-2][c] {
				col = c
				break
			}
		}
		if col < 0 {
			t.Fatal("shooter should come from the new bottom row")
		}
		if ig.BottomOfColumn(col) != iv {
			t.Fatalf("shooter is not the bottom-most invader of column %d", col)
		}
	}
}

func TestPickShooterSkipsEmptyColumns(t *testing.T) {
	ig := newTestGrid()
	for c := 0; c < InvaderCols-1; c++ {
		for r := range ig.Invaders {
			ig.Invaders[r][c].Alive = false
		}
	}
	for range 50 {
		if ig.PickShooter() != ig.BottomOfColumn(InvaderCols-1) {
			t.Fatal("should always pick from the only non-empty column")
		}
	}
}

func TestPickShooterEmptyGrid(t *testing.T) {
	ig := newTestGrid()
	killAllBut(&ig, 0)
	if ig.PickShooter() != nil {
		t.Fatal("empty grid should return nil")
	}
}
