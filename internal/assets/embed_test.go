package assets

import "testing"

func TestEmbeddedAssets(t *testing.T) {
	data, err := Assets.ReadFile("ship.png")
	if err != nil {
		t.Fatalf("Failed to read ship.png: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("ship.png is empty")
	}
	t.Logf("Successfully read ship.png, size: %d bytes", len(data))
}
