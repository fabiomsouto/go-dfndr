package main

import (
	"bytes"
	"image"
	"log"

	"github.com/fabiomsouto/dfndr/internal/assets"
	"github.com/hajimehoshi/ebiten/v2"
)

const (
	enemyWidth  = 50
	enemyHeight = 40
)

type Enemy struct {
	x, y     float64
	vx, vy   float64
	image    *ebiten.Image
	viewport *Viewport
}

func NewEnemy(x, y, vx, vy float64, viewport *Viewport) *Enemy {
	return &Enemy{
		x:        x,
		y:        y,
		vx:       vx,
		vy:       vy,
		image:    enemyImg(),
		viewport: viewport,
	}
}

func (e *Enemy) Update() {
	e.x += e.vx
	e.y += e.vy
}

func (e *Enemy) Draw(screen *ebiten.Image, viewport *Viewport) {
	// Convert world coordinates to screen coordinates
	screenX, screenY := e.viewport.WorldToScreen(e.x, e.y)

	op := &ebiten.DrawImageOptions{}

	op.GeoM.Translate(screenX, screenY)
	screen.DrawImage(e.image, op)
}

func enemyImg() *ebiten.Image {
	// load image from embedded filesystem
	data, err := assets.Assets.ReadFile("memleak.png")
	if err != nil {
		log.Fatalf("failed to read ship.png from embedded assets: %v", err)
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		log.Fatalf("failed to decode ship image: %v", err)
	}

	return ebiten.NewImageFromImage(img)
}
