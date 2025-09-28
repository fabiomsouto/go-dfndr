package main

import "math"

const (
	// Horizontal deadzone - how far from center the player can move before scrolling starts
	deadzoneX = 200.0
	// Vertical deadzone - how far from center the player can move before scrolling starts
	deadzoneY = 150.0
)

type Viewport struct {
	x, y          float64 // top-left corner of viewport in world coordinates
	width, height float64
	worldWidth    float64
}

func NewViewport(width, height, worldWidth float64) *Viewport {
	return &Viewport{
		x:          0,
		y:          0,
		width:      width,
		height:     height,
		worldWidth: worldWidth,
	}
}

// Follow updates the viewport position to follow a target
func (v *Viewport) Follow(targetX, targetY float64) {
	// Get the center of the viewport
	viewportCenterX := v.x + v.width/2
	viewportCenterY := v.y + v.height/2

	// Calculate how far the target is from the viewport center
	deltaX := targetX - viewportCenterX
	deltaY := targetY - viewportCenterY

	// Only move the viewport if the target is outside the deadzone
	if math.Abs(deltaX) > deadzoneX {
		// Move the viewport, keeping the target at the edge of the deadzone
		if deltaX > 0 {
			v.x += deltaX - deadzoneX
		} else {
			v.x += deltaX + deadzoneX
		}
	}

	if math.Abs(deltaY) > deadzoneY {
		// Move the viewport, keeping the target at the edge of the deadzone
		if deltaY > 0 {
			v.y += deltaY - deadzoneY
		} else {
			v.y += deltaY + deadzoneY
		}
	}

	// Keep the viewport within the world bounds
	if v.x < 0 {
		v.x = 0
	}
	if v.x > v.worldWidth-v.width {
		v.x = v.worldWidth - v.width
	}

	// Keep viewport within vertical bounds
	if v.y < 0 {
		v.y = 0
	}
	if v.y > ScreenHeight-v.height {
		v.y = ScreenHeight - v.height
	}
}

// WorldToScreen converts world coordinates to screen coordinates
func (v *Viewport) WorldToScreen(worldX, worldY float64) (float64, float64) {
	screenX := worldX - v.x
	screenY := worldY - v.y
	return screenX, screenY
}

// ScreenToWorld converts screen coordinates to world coordinates
func (v *Viewport) ScreenToWorld(screenX, screenY float64) (float64, float64) {
	worldX := screenX + v.x
	worldY := screenY + v.y
	return worldX, worldY
}
