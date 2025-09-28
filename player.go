package main

import (
	"bytes"
	"image"
	_ "image/png" // Register PNG decoder
	"log"
	"math"

	"github.com/fabiomsouto/dfndr/internal/assets"
	"github.com/hajimehoshi/ebiten/v2"
)

const (
	shipWidth     = 100
	shipHeight    = 49
	shipStartPosX = 60
	shipStartPosY = 100
	thrustForce   = 1
	dragFactor    = 0.95
	maxSpeed      = 15
)

type Player struct {
	x, y     float64 // world coordinates
	vx, vy   float64
	image    *ebiten.Image
	bullets  []*Bullet
	viewport *Viewport
}

type Bullet struct {
	x, y   float64
	vx, vy float64
	image  *ebiten.Image
}

func NewPlayer(viewport *Viewport) *Player {
	img := shipImg()
	bullets := []*Bullet{}
	return &Player{
		x:        shipStartPosX,
		y:        shipStartPosY,
		vx:       0,
		vy:       0,
		image:    img,
		bullets:  bullets,
		viewport: viewport,
	}
}

func shipImg() *ebiten.Image {
	// load image from embedded filesystem
	data, err := assets.Assets.ReadFile("ship.png")
	if err != nil {
		log.Fatalf("failed to read ship.png from embedded assets: %v", err)
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		log.Fatalf("failed to decode ship image: %v", err)
	}

	return ebiten.NewImageFromImage(img)
}

func (p *Player) Update() {
	// Apply thrust
	// Right movement
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) || ebiten.IsKeyPressed(ebiten.KeyD) {
		p.vx += thrustForce
	}
	// Left movement
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) || ebiten.IsKeyPressed(ebiten.KeyA) {
		p.vx -= thrustForce
	}
	// Up movement
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) || ebiten.IsKeyPressed(ebiten.KeyW) {
		p.vy -= thrustForce
	}
	// Down movement
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) || ebiten.IsKeyPressed(ebiten.KeyS) {
		p.vy += thrustForce
	}

	// Apply drag
	p.vx *= dragFactor
	p.vy *= dragFactor

	// Limit speed
	speed := math.Sqrt(p.vx*p.vx + p.vy*p.vy)
	if speed > maxSpeed {
		p.vx = (p.vx / speed) * maxSpeed
		p.vy = (p.vy / speed) * maxSpeed
	}

	// snap very small speeds to zero
	if math.Abs(p.vx) < 0.1 {
		p.vx = 0
	}
	if math.Abs(p.vy) < 0.1 {
		p.vy = 0
	}

	// Update position
	p.x += p.vx
	p.y += p.vy

	// keep player within screen bounds vertically
	if p.y < 0 {
		p.y = 0
		p.vy = 0
	}
	if p.y > ScreenHeight-shipHeight {
		p.y = ScreenHeight - shipHeight
		p.vy = 0
	}

	// wrap around the world horizontally
	if p.x < -shipWidth {
		p.x = WorldWidth
	}
	if p.x > WorldWidth {
		p.x = 0
	}

	// Update viewport to follow player
	p.viewport.Follow(p.x, p.y)

	log.Printf("Player position: (%.2f, %.2f), velocity: (%.2f, %.2f)", p.x, p.y, p.vx, p.vy)
}

func (p *Player) Draw(screen *ebiten.Image) {
	// Convert world coordinates to screen coordinates
	screenX, screenY := p.viewport.WorldToScreen(p.x, p.y)

	op := &ebiten.DrawImageOptions{}

	if p.vx < 0 {
		// For left movement: first scale, then translate
		op.GeoM.Scale(-1, 1)
		op.GeoM.Translate(shipWidth, 0)
		op.GeoM.Translate(screenX, screenY)
	} else {
		// For right movement or stationary: just translate
		op.GeoM.Translate(screenX, screenY)
	}

	screen.DrawImage(p.image, op)
}
