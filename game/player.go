package game

const (
	PlayerSpeed   = 2
	PlayerW       = 24
	PlayerH       = 16
	PlayerY       = 200
	RespawnInvuln = 60
)

type Player struct {
	X, Y         int
	W, H         int
	Alive        bool
	Invulnerable int
}

func NewPlayer() Player {
	return Player{
		X:     (ScreenW - PlayerW) / 2,
		Y:     PlayerY,
		W:     PlayerW,
		H:     PlayerH,
		Alive: true,
	}
}

func (p *Player) Update(input *InputState) {
	if !p.Alive {
		return
	}
	if input.Left {
		p.X -= PlayerSpeed
	}
	if input.Right {
		p.X += PlayerSpeed
	}
	if p.X < 0 {
		p.X = 0
	}
	if p.X > ScreenW-PlayerW {
		p.X = ScreenW - PlayerW
	}
}

func (p *Player) Fire() *Bullet {
	if !p.Alive {
		return nil
	}
	return &Bullet{
		X:      p.X + p.W/2 - 1,
		Y:      p.Y,
		Owner:  BulletPlayer,
		Active: true,
	}
}

func (p *Player) Hit() {
	p.Alive = false
}

func (p *Player) Respawn() {
	p.X = (ScreenW - PlayerW) / 2
	p.Y = PlayerY
	p.W = PlayerW
	p.H = PlayerH
	p.Alive = true
	p.Invulnerable = RespawnInvuln
}

func (p *Player) Rect() (int, int, int, int) {
	return p.X, p.Y, p.W, p.H
}
