package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	ScreenWidth  = 1024
	ScreenHeight = 768
	Level        = 1
)

type Game struct {
	player *Player
	world  *World
}

func newGame() *Game {
	viewport := NewViewport(ScreenWidth, ScreenHeight, WorldWidth)
	player := NewPlayer(viewport)
	world := NewWorld(player, viewport, Level)

	return &Game{
		player: player,
		world:  world,
	}
}

func (g *Game) Update() error {
	g.world.Update()
	g.player.Update()
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.world.Draw(screen)
	g.player.Draw(screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
}

func main() {
	ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
	ebiten.SetWindowTitle("Go Defender")
	game := newGame()
	if err := ebiten.RunGame(game); err != nil {
		log.Fatalf("something went terribly wrong: %v", err)
	}
}
