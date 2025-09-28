package main

import (
	"bytes"
	"image"
	"image/color"
	"log"
	"math"

	"github.com/fabiomsouto/dfndr/internal/assets"
	"github.com/hajimehoshi/ebiten/v2"
)

const (
	shipWidth   = 50
	shipHeight  = 30
	thrustForce = 1
	dragFactor  = 0.95
	maxSpeed    = 12
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
		x:        WorldWidth / 2, // Start in the middle of the world
		y:        ScreenHeight / 2,
		vx:       0,
		vy:       0,
		image:    img,
		bullets:  bullets,
		viewport: viewport,
	}
}

// shipImgTriangle creates a simple triangular ship image
func shipImgTriangle() *ebiten.Image {
	img := ebiten.NewImage(shipWidth, shipHeight)
	// simple triangle shape for the ship for the moment
	vertices := []ebiten.Vertex{
		{DstX: 0, DstY: 0, SrcX: 1, SrcY: 1, ColorR: 1, ColorG: 255, ColorB: 1, ColorA: 1},   //tip
		{DstX: 50, DstY: 15, SrcX: 1, SrcY: 1, ColorR: 1, ColorG: 255, ColorB: 1, ColorA: 1}, //top left
		{DstX: 0, DstY: 30, SrcX: 1, SrcY: 1, ColorR: 1, ColorG: 255, ColorB: 1, ColorA: 1},  //bottom left
	}
	indices := []uint16{0, 1, 2}
	sourceImg := ebiten.NewImage(shipWidth, shipHeight)
	sourceImg.Fill(color.White)
	opts := &ebiten.DrawTrianglesOptions{}
	img.DrawTriangles(vertices, indices, sourceImg, opts)
	return img
}

func shipImg() *ebiten.Image {
	// load image from file
	img, _, err := image.Decode(bytes.NewReader(assets.ShipPNG))
	shipImg := ebiten.NewImageFromImage(img)
	if err != nil {
		log.Fatalf("failed to load ship image: %v", err)
	}
	return shipImg
}

func (p *Player) Update() {

	// Apply thrust
	if ebiten.IsKeyPressed((ebiten.KeyArrowRight)) {
		p.vx += thrustForce
	}
	if ebiten.IsKeyPressed((ebiten.KeyArrowLeft)) {
		p.vx -= thrustForce
	}
	if ebiten.IsKeyPressed((ebiten.KeyArrowUp)) {
		p.vy -= thrustForce
	}
	if ebiten.IsKeyPressed((ebiten.KeyArrowDown)) {
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
