package main

import (
	"image/color"
	"math"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	WorldWidth = 10000
	Stars      = 500
	MaxEnemies = 20
)

type World struct {
	stars    []Star
	enemies  []*Enemy
	viewport *Viewport
}

type Star struct {
	x, y           float32
	radius         int // in pixels
	color          color.RGBA
	originalColor  color.RGBA // Store original color to prevent fading
	parallaxFactor float64    // How much this star moves relative to the camera (0.0-1.0)
}

func NewWorld(viewport *Viewport) *World {
	stars := generateStars(Stars)
	enemies := make([]*Enemy, MaxEnemies)
	for i := range enemies {
		x := float64(randInt(0, WorldWidth))
		y := float64(randInt(0, ScreenHeight))
		vx := (rand.Float64() * 2) - 1
		vy := (rand.Float64() * 2) - 1
		enemies[i] = NewEnemy(x, y, vx, vy, viewport)
	}
	return &World{
		stars:    stars,
		enemies:  enemies,
		viewport: viewport,
	}
}

func generateStars(n int) []Star {
	stars := make([]Star, n)
	for i := range n {
		radius := randInt(1, 5)
		baseColor := color.RGBA{
			R: uint8(randInt(0, 255)),
			G: uint8(randInt(0, 255)),
			B: uint8(randInt(0, 255)),
			A: 255,
		}
		// Larger stars appear closer and move faster
		parallaxFactor := 0.2 + (float64(radius)/5.0)*0.8
		stars[i] = Star{
			x:              float32(randInt(0, WorldWidth)),
			y:              float32(randInt(0, ScreenHeight)),
			color:          baseColor,
			originalColor:  baseColor,
			radius:         radius,
			parallaxFactor: parallaxFactor,
		}
	}
	return stars
}

func (world *World) Update() {
	stars := world.stars
	currentTime := time.Now().UnixMilli()
	for i := range stars {
		oscillation := math.Sin(float64(i)*0.02 + float64(currentTime)*0.0005)
		brightness := math.Abs(oscillation)

		// Interpolate using the original color values
		stars[i].color.R = uint8(float64(stars[i].originalColor.R) * (0.8 + brightness*0.4))
		stars[i].color.G = uint8(float64(stars[i].originalColor.G) * (0.8 + brightness*0.4))
		stars[i].color.B = uint8(float64(stars[i].originalColor.B) * (0.8 + brightness*0.4))
	}
}

func randInt(min, max int) int {
	return min + rand.Intn(max-min)
}

func (world *World) Draw(screen *ebiten.Image) {
	for _, star := range world.stars {
		// Apply parallax effect by scaling the viewport offset
		parallaxX := world.viewport.x * star.parallaxFactor
		screenX := float64(star.x) - parallaxX
		screenY := float64(star.y) - world.viewport.y

		// Wrap stars horizontally based on their parallax speed
		if screenX < -float64(star.radius) {
			screenX += WorldWidth
		} else if screenX > world.viewport.width+float64(star.radius) {
			screenX -= WorldWidth
		}

		// Only draw stars that are within the viewport
		if screenX >= -float64(star.radius) && screenX <= world.viewport.width+float64(star.radius) &&
			screenY >= -float64(star.radius) && screenY <= world.viewport.height+float64(star.radius) {
			vector.DrawFilledRect(screen, float32(screenX), float32(screenY), float32(star.radius), float32(star.radius), star.color, false)
		}
	}
	for _, enemy := range world.enemies {
		enemy.Update()
		enemy.Draw(screen, world.viewport)
	}
}
