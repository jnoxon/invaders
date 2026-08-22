package game

func (g *Game) AddScore(points int) {
	g.Score += points
	g.UpdateHighScore()
}

func (g *Game) InvaderPoints(t InvaderType) int {
	return t.Points()
}

func (g *Game) HandlePlayerDeath() {
	g.Lives--
	g.Flash = DeathFlashFrames
	g.emit(EventPlayerHit)
	if g.Lives <= 0 {
		g.GameOver()
		return
	}
	g.Player.Respawn()
}

func (g *Game) HandleLevelComplete() {
	g.State = StateLevelTransition
	g.TransitionTimer = TransitionFrames
}

func (g *Game) UpdateHighScore() {
	if g.Score > g.HighScore {
		g.HighScore = g.Score
	}
}
