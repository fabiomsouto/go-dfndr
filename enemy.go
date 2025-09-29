package main

import (
	"bytes"
	"image"
	"image/color"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/fabiomsouto/dfndr/internal/assets"
	"github.com/fabiomsouto/dfndr/internal/utils"
	"github.com/hajimehoshi/ebiten/v2"
)

const (
	enemyWidth  = 50
	enemyHeight = 40

	// Enemy behavior constants
	baseSpeed     = 2.0 // Base movement speed
	wanderFactor  = 0.8 // How much random wandering (decrease for later levels)
	precisionBase = 0.3 // Base precision in tracking (increase for later levels)
	updateRate    = 30  // How often to update random movement (frames)
)

// Difficulty levels (can be adjusted as levels progress)
type DifficultyLevel struct {
	speed     float64 // Actual movement speed
	wander    float64 // Random movement factor (0-1)
	precision float64 // Tracking precision (0-1)
}

var (
	difficultyLevels = []DifficultyLevel{
		{speed: baseSpeed * 0.3, wander: wanderFactor, precision: precisionBase},             // Level 1: Slow, erratic
		{speed: baseSpeed * 0.5, wander: wanderFactor * 0.7, precision: precisionBase * 1.5}, // Level 2: Faster, less erratic
		{speed: baseSpeed * 0.9, wander: wanderFactor * 0.5, precision: precisionBase * 2.0}, // Level 3: Even faster, more precise
		{speed: baseSpeed, wander: wanderFactor * 0.3, precision: precisionBase * 2.5},       // Level 4: Full speed, very precise
		{speed: baseSpeed * 1.3, wander: wanderFactor * 0.1, precision: precisionBase * 3.0}, // Level 5: Aggressive!
	}
)

type ExplosionParticle struct {
	x, y     float64 // Position
	vx, vy   float64 // Velocity
	size     float64 // Current size
	rotation float64 // Current rotation
	hue      float64 // Color hue
	life     float64 // Remaining life (0.0 to 1.0)
}

type Enemy struct {
	x, y          float64
	vx, vy        float64
	image         *ebiten.Image
	player        *Player
	viewport      *Viewport
	diffLevel     int                 // Current difficulty level
	wanderAngle   float64             // Current random movement angle
	updateCounter int                 // Counter for movement updates
	rng           *rand.Rand          // Per-enemy random number generator
	health        int                 // Current health points
	active        bool                // Whether the enemy is alive and active
	hitTimer      int                 // Timer for hit visual feedback
	particles     []ExplosionParticle // Explosion particles
	exploding     bool                // Whether currently exploding
}

func NewEnemy(x, y, vx, vy float64, player *Player, viewport *Viewport, level int) *Enemy {
	source := rand.NewSource(time.Now().UnixNano())
	return &Enemy{
		x:             x,
		y:             y,
		vx:            vx,
		vy:            vy,
		player:        player,
		image:         enemyImg(),
		viewport:      viewport,
		diffLevel:     level,
		wanderAngle:   rand.Float64() * 2 * math.Pi,
		updateCounter: 0,
		rng:           rand.New(source),
		health:        3, // Start with 3 health points
		active:        true,
		hitTimer:      0,
		particles:     make([]ExplosionParticle, 0),
		exploding:     false,
	}
}

func (e *Enemy) Update() {
	if !e.active && !e.exploding {
		return
	}

	// Handle explosion if active
	if e.exploding {
		e.updateExplosion()
		return
	}

	// Update hit timer
	if e.hitTimer > 0 {
		e.hitTimer--
	}

	// Get current difficulty settings
	diff := difficultyLevels[e.diffLevel]

	// Update random movement angle periodically
	e.updateCounter++
	if e.updateCounter >= updateRate {
		e.wanderAngle = e.rng.Float64() * 2 * math.Pi
		e.updateCounter = 0
	}

	// Calculate direction to player
	playerX, playerY := e.player.Position()
	dirX := playerX - e.x
	dirY := playerY - e.y
	mag := math.Sqrt((dirX * dirX) + (dirY * dirY))
	if mag > 0 {
		dirX /= mag
		dirY /= mag
	}

	// Calculate wander direction
	wanderX := math.Cos(e.wanderAngle)
	wanderY := math.Sin(e.wanderAngle)

	// Combine tracking and wandering based on precision
	finalDirX := dirX*diff.precision + wanderX*diff.wander
	finalDirY := dirY*diff.precision + wanderY*diff.wander

	// Normalize final direction
	finalMag := math.Sqrt(finalDirX*finalDirX + finalDirY*finalDirY)
	if finalMag > 0 {
		finalDirX /= finalMag
		finalDirY /= finalMag
	}

	// Apply movement
	e.vx = finalDirX * diff.speed
	e.vy = finalDirY * diff.speed

	e.x += e.vx
	e.y += e.vy

	// Wrap around world edges
	if e.x < 0 {
		e.x = WorldWidth
	} else if e.x > WorldWidth {
		e.x = 0
	}
	if e.y < 0 {
		e.y = ScreenHeight
	} else if e.y > ScreenHeight {
		e.y = 0
	}
}

func (e *Enemy) Draw(screen *ebiten.Image, viewport *Viewport) {
	if e.exploding {
		e.drawExplosion(screen)
		return
	}

	if !e.active {
		return // Don't draw if not active
	}

	// Convert world coordinates to screen coordinates
	screenX, screenY := e.viewport.WorldToScreen(e.x, e.y)

	op := &ebiten.DrawImageOptions{}

	// Flash white when hit
	if e.hitTimer > 0 {
		op.ColorScale.Scale(1.5, 1.5, 1.5, 1)
	}

	op.GeoM.Translate(screenX, screenY)
	screen.DrawImage(e.image, op)
}

// Hit is called when the enemy is hit by a bullet
func (e *Enemy) Hit() {
	if !e.active || e.exploding {
		return
	}

	e.health--
	e.hitTimer = 5 // Flash for 5 frames

	if e.health <= 0 {
		e.exploding = true
		e.initExplosion()
		// TODO: Add score
	}
}

// Returns the collision box for the enemy
func (e *Enemy) Bounds() (float64, float64, float64, float64) {
	return e.x, e.y, enemyWidth, enemyHeight
}

// CheckBulletCollision checks if a bullet hits this enemy
func (e *Enemy) CheckBulletCollision(bulletX, bulletY float64) bool {
	if !e.active || e.exploding {
		return false // Let bullets pass through when not active or during explosion
	}

	// Simple rectangle collision
	if bulletX >= e.x && bulletX <= e.x+enemyWidth &&
		bulletY >= e.y && bulletY <= e.y+enemyHeight {
		e.Hit()
		return true
	}
	return false
}

func (e *Enemy) initExplosion() {
	const numParticles = 20
	e.particles = make([]ExplosionParticle, numParticles)

	// Create particles in a circular pattern
	for i := range e.particles {
		angle := e.rng.Float64() * 2 * math.Pi
		speed := 2 + e.rng.Float64()*3

		e.particles[i] = ExplosionParticle{
			x:        e.x + float64(enemyWidth)/2,
			y:        e.y + float64(enemyHeight)/2,
			vx:       math.Cos(angle) * speed,
			vy:       math.Sin(angle) * speed,
			size:     5 + e.rng.Float64()*10,
			rotation: e.rng.Float64() * math.Pi,
			hue:      e.rng.Float64() * 60, // Random hue in red-yellow range
			life:     1.0,
		}
	}
}

func (e *Enemy) updateExplosion() bool {
	allDead := true
	for i := range e.particles {
		if e.particles[i].life > 0 {
			allDead = false

			// Update position
			e.particles[i].x += e.particles[i].vx
			e.particles[i].y += e.particles[i].vy

			// Update rotation
			e.particles[i].rotation += 0.1

			// Fade out
			e.particles[i].life -= 0.02
			e.particles[i].size *= 0.98
		}
	}

	if allDead {
		e.active = false
		e.exploding = false
	}

	return !allDead
}

func (e *Enemy) drawExplosion(screen *ebiten.Image) {
	for _, p := range e.particles {
		if p.life <= 0 {
			continue
		}

		screenX, screenY := e.viewport.WorldToScreen(p.x, p.y)

		// Calculate color based on life and hue
		c := utils.HSVToRGB(p.hue, 1.0, p.life)

		// Create a rotated rectangle
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(-p.size/2, -p.size/2) // Center rotation
		op.GeoM.Rotate(p.rotation)
		op.GeoM.Translate(screenX, screenY)
		op.ColorScale.Scale(
			float32(c.R)/255.0,
			float32(c.G)/255.0,
			float32(c.B)/255.0,
			float32(p.life),
		)

		// Create and draw the rectangle
		rect := ebiten.NewImage(int(p.size), int(p.size))
		rect.Fill(color.White)
		screen.DrawImage(rect, op)
	}
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
