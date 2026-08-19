package game

import "testing"

func TestAddScore(t *testing.T) {
	g := newTestGame()
	g.AddScore(50)
	if g.Score != 50 {
		t.Fatalf("score = %d, want 50", g.Score)
	}
	g.AddScore(30)
	if g.Score != 80 {
		t.Fatalf("score = %d, want 80", g.Score)
	}
}

func TestHighScoreUpdate(t *testing.T) {
	g := newTestGame()
	g.AddScore(100)
	if g.HighScore != 100 {
		t.Fatalf("high = %d, want 100", g.HighScore)
	}
	g.Score = 0
	g.AddScore(10)
	if g.HighScore != 100 {
		t.Fatalf("high should stay 100, got %d", g.HighScore)
	}
	g.AddScore(5)
	if g.HighScore != 100 {
		t.Fatal("high should still be 100")
	}
}

func TestUpdateHighScoreDirect(t *testing.T) {
	g := newTestGame()
	g.Score = 500
	g.HighScore = 100
	g.UpdateHighScore()
	if g.HighScore != 500 {
		t.Fatal("high score should update")
	}
}

func TestInvaderPointsViaGame(t *testing.T) {
	g := newTestGame()
	if g.InvaderPoints(InvaderSquid) != 30 {
		t.Fatal("squid points")
	}
	if g.InvaderPoints(InvaderCrab) != 20 {
		t.Fatal("crab points")
	}
	if g.InvaderPoints(InvaderOctopus) != 10 {
		t.Fatal("octopus points")
	}
}

func TestHandlePlayerDeathDecrement(t *testing.T) {
	g := newTestGame()
	g.Lives = StartLives
	g.HandlePlayerDeath()
	if g.Lives != StartLives-1 {
		t.Fatalf("lives = %d, want %d", g.Lives, StartLives-1)
	}
	if !g.Player.Alive {
		t.Fatal("player should respawn")
	}
}

func TestHandlePlayerDeathGameOver(t *testing.T) {
	g := newTestGame()
	g.Lives = 1
	g.HandlePlayerDeath()
	if g.Lives != 0 {
		t.Fatalf("lives = %d, want 0", g.Lives)
	}
	if g.State != StateGameOver {
		t.Fatalf("state = %v, want StateGameOver", g.State)
	}
}

func TestHandleLevelComplete(t *testing.T) {
	g := newTestGame()
	g.HandleLevelComplete()
	if g.State != StateLevelTransition {
		t.Fatalf("state = %v, want StateLevelTransition", g.State)
	}
	if g.TransitionTimer != TransitionFrames {
		t.Fatalf("timer = %d, want %d", g.TransitionTimer, TransitionFrames)
	}
}
