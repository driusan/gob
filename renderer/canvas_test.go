package renderer

import (
	"context"
	"image"
	"image/color"
	"testing"
)

func colorEQ(c1, c2 color.Color) bool {
	r1, g1, b1, a1 := c1.RGBA()
	r2, g2, b2, a2 := c2.RGBA()
	return r1 == r2 && g1 == g2 && b1 == b2 && a1 == a2
}

// Test that the canvas background gets changed by the HTML tag.
func TestCanvasBackground(t *testing.T) {
	page := parseHTML(
		t,
		`<html style="background: red">
	<body style="background: #0F0; padding: 0; margin: 25px;">
		This should have a green background and a 25px margin around it on
		a red canvas.
	</body>
</html>`,
	)

	redRGBA := color.RGBA{255, 0, 0, 255}
	greenRGBA := color.RGBA{0, 255, 0, 255}

	canvas := image.NewRGBA(image.Rectangle{image.ZP, image.Point{400, 300}})
	size := canvas.Bounds().Size()
	page.Content.Layout(context.TODO(), size)
	page.Content.RenderInto(context.TODO(), canvas, image.ZP)

	if !colorEQ(page.Background, redRGBA) {
		t.Fatalf("Unexpected canvas background. got %v want %v", page.Background, redRGBA)
	}

	// The background inside the margins is transparent, because the
	// window draws page.Background

	// Left margin
	for x := 0; x < 25; x++ {
		for y := 0; y < size.Y; y++ {
			if clr := canvas.At(x, y); !colorEQ(clr, color.Transparent) {
				t.Errorf("Left margin: Pixel (%d, %d). got %v want %v", x, y, clr, color.Transparent)
			}
		}
	}

	// Top margin
	for x := 0; x < size.X; x++ {
		for y := 0; y < 25; y++ {
			if clr := canvas.At(x, y); !colorEQ(clr, color.Transparent) {
				t.Errorf("Top margin: Pixel (%d, %d). got %v want %v", x, y, clr, color.Transparent)
			}
		}
	}

	// Right margin
	for x := size.X - 25; x < size.X; x++ {
		for y := 0; y < size.Y; y++ {
			if clr := canvas.At(x, y); !colorEQ(clr, color.Transparent) {
				t.Errorf("Right margin: Pixel (%d, %d). got %v want %v", x, y, clr, color.Transparent)
			}
		}
	}

	// We don't check the bottom margin, because we don't know
	// how big the content is relative to the margin.

	// Just check a pixel right inside the margins that it's the right
	// colour, we don't check the whole thing because we don't know
	// exactly where the text is.
	if clr := canvas.At(25, 25); !colorEQ(clr, greenRGBA) {
		t.Errorf("Body background wrong colour. got %v want %v", clr, greenRGBA)
	}
}
