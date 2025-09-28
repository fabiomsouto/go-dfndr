package main

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
	// Center the viewport on the target
	v.x = targetX - v.width/2

	// Keep the viewport within the world bounds
	if v.x < 0 {
		v.x = 0
	}
	if v.x > v.worldWidth-v.width {
		v.x = v.worldWidth - v.width
	}

	// For now, vertical position is fixed as we only scroll horizontally
	v.y = 0
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
