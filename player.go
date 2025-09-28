package main

import (
	"bytes"
	"image"
	"image/color"
	_ "image/png" // Register PNG decoder
	"log"
	"math"
	"math/rand"

	"github.com/fabiomsouto/dfndr/internal/assets"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	shipWidth  = 100
	shipHeight = 49

	shipStartPosX = 60
	shipStartPosY = 100
	shipMaxSpeed  = 15

	thrustForce = 1
	dragFactor  = 0.95

	bulletsMax = 20 // Max active bullets
)

type Player struct {
	x, y         float64 // world coordinates
	vx, vy       float64
	image        *ebiten.Image
	bullets      []*Bullet
	viewport     *Viewport
	spaceWasDown bool // Track previous state of space key
}

type TrailPoint struct {
	x, y float64
}

type Bullet struct {
	x, y     float64
	right    bool
	active   bool
	trail    []TrailPoint
	trailHue float64 // Tracks the current hue for color morphing
}

func NewPlayer(viewport *Viewport) *Player {
	img := shipImg()
	bullets := make([]*Bullet, bulletsMax)
	for i := range bullets {
		bullets[i] = &Bullet{active: false}
	}
	return &Player{
		x:            shipStartPosX,
		y:            shipStartPosY,
		vx:           0,
		vy:           0,
		image:        img,
		bullets:      bullets,
		viewport:     viewport,
		spaceWasDown: false,
	}
}

// HSVToRGB converts HSV color values to RGB
func HSVToRGB(h, s, v float64) color.RGBA {
	h = math.Mod(h, 360) // Keep hue in [0,360)
	c := v * s
	x := c * (1 - math.Abs(math.Mod(h/60, 2)-1))
	m := v - c

	var r, g, b float64
	switch {
	case h < 60:
		r, g, b = c, x, 0
	case h < 120:
		r, g, b = x, c, 0
	case h < 180:
		r, g, b = 0, c, x
	case h < 240:
		r, g, b = 0, x, c
	case h < 300:
		r, g, b = x, 0, c
	default:
		r, g, b = c, 0, x
	}

	return color.RGBA{
		R: uint8((r + m) * 255),
		G: uint8((g + m) * 255),
		B: uint8((b + m) * 255),
		A: 255,
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

	// Handle bullet firing with simple key press detection
	spaceIsDown := ebiten.IsKeyPressed(ebiten.KeySpace)
	if spaceIsDown && !p.spaceWasDown { // Only fire on the initial press
		// Find an inactive bullet to reuse
		for _, b := range p.bullets {
			if !b.active {
				b.x = p.x + shipWidth
				b.y = p.y + shipHeight - 12
				b.right = p.vx >= 0
				b.active = true
				b.trail = make([]TrailPoint, 0, 50) // Preallocate space for 50 points
				b.trailHue = rand.Float64() * 360   // Random starting hue
				break
			}
		}
	}
	p.spaceWasDown = spaceIsDown // Update previous state

	// Update bullets
	for _, b := range p.bullets {
		if b.active {
			// Store previous position
			prevX, prevY := b.x, b.y

			if b.right {
				b.x += shipMaxSpeed + 10
				// Deactivate if out of viewport bounds
				bvx, _ := p.viewport.WorldToScreen(b.x, b.y)
				if b.x < 0 || bvx > p.viewport.width-10 { // -10 to give some margin
					b.active = false
				}

				// Remove trail points that are too far behind
				if len(b.trail) > 0 {
					for i, point := range b.trail {
						if b.x-point.x > 200 { // Remove points more than 200 pixels behind
							b.trail = b.trail[i+1:]
							break
						}
					}
				}
			} else {
				b.x -= shipMaxSpeed + 10
				// Deactivate if out of viewport bounds
				bvx, _ := p.viewport.WorldToScreen(b.x, b.y)
				if b.x < 0 || bvx < -10 { // -10 to give some margin
					b.active = false
				}

				// Remove trail points that are too far behind
				if len(b.trail) > 0 {
					for i, point := range b.trail {
						if point.x-b.x > 600 { // Remove points more than 600 pixels behind
							b.trail = b.trail[i+1:]
							break
						}
					}
				}
			} // Update trail
			if len(b.trail) == 0 || math.Hypot(b.x-prevX, b.y-prevY) > 5 {
				// Add slight randomness to y position for irregular effect
				// trailY := b.y + (rand.Float64()*2-1)*2
				// TODO: dont like the effect right now, I'll revisit later
				trailY := b.y

				// Keep track of current hue but don't assign color yet
				b.trailHue = math.Mod(b.trailHue+2, 360)

				b.trail = append(b.trail, TrailPoint{
					x: b.x,
					y: trailY,
				})

				// Keep trail at a reasonable length
				if len(b.trail) > 50 {
					b.trail = b.trail[1:]
				}
			}
		}
	}

	// Apply drag
	p.vx *= dragFactor
	p.vy *= dragFactor

	// Limit speed
	speed := math.Sqrt(p.vx*p.vx + p.vy*p.vy)
	if speed > shipMaxSpeed {
		p.vx = (p.vx / speed) * shipMaxSpeed
		p.vy = (p.vy / speed) * shipMaxSpeed
	}

	// snap very small speeds to zero, preserving sign for
	// avoiding ship suddenly flipping direction visually
	if math.Abs(p.vx) < 0.1 {
		if math.Signbit(p.vx) {
			p.vx = math.Copysign(0, -1) // Keep negative sign
		} else {
			p.vx = 0.0
		}
	}
	if math.Abs(p.vy) < 0.1 {
		p.vy = 0.0 // Vertical speed sign doesn't affect visuals
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

	// Draw bullets and their trails
	for _, b := range p.bullets {
		if b.active {
			// Draw trail
			if len(b.trail) > 1 {
				totalPoints := len(b.trail)

				// Calculate current color for the entire trail
				currentColor := HSVToRGB(b.trailHue, 1, 1)

				for i := 0; i < totalPoints-1; i++ {
					p1 := b.trail[i]
					p2 := b.trail[i+1]

					// Calculate distance from current point to bullet (0 = at bullet, 1 = furthest)
					distanceRatio := float64(i) / float64(totalPoints)

					// Skip some segments based on distance (more gaps further from bullet)
					if math.Sin(distanceRatio*20) > 0.3-distanceRatio {
						continue
					}

					// Convert trail points to screen coordinates
					x1, y1 := p.viewport.WorldToScreen(p1.x, p1.y)
					x2, y2 := p.viewport.WorldToScreen(p2.x, p2.y)

					// Apply uniform color with distance-based fade
					fadeColor := currentColor
					fadeColor.A = uint8(255 * (1 - distanceRatio*0.8))

					// Draw line segment with fading and morphing color
					vector.StrokeLine(screen,
						float32(x1), float32(y1),
						float32(x2), float32(y2),
						2, fadeColor, false)
				}
			}

			// Draw bullet
			bScreenX, bScreenY := p.viewport.WorldToScreen(b.x, b.y)
			vector.DrawFilledCircle(screen, float32(bScreenX), float32(bScreenY), 3, color.White, false)
		}
	}

	screen.DrawImage(p.image, op)
}
