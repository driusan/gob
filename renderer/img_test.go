package renderer

import (
	"context"
	"image"
	"testing"
)

// The box draw rectangle and box content rectangle interact weirdly
// for block level images to position them. This should eventually be
// removed and the BoxDrawRectangle and BoxContentRectangle should
// just be set more sanely.
func testImgDrawRectangle(t *testing.T, i *RenderableDomElement, want image.Rectangle) {
	t.Helper()
	if i.BoxDrawRectangle.Min != image.ZP {
		t.Errorf("Expected draw rectangle to start at origin")
	}
	if i.BoxDrawRectangle.Max != i.BoxContentRectangle.Max {
		t.Errorf("Draw rectangle does not cover content.")
	}
	got := i.BoxContentRectangle
	if want != got {
		t.Errorf("Unexpected image location: got %v want %v", got, want)
	}
}

// Tests that an inline image is positioned appropriately.
func TestInlineImg(t *testing.T) {
	page := parseHTML(
		t,
		`<html>
		<body style="font-size: 16px; line-height: 16px; padding: 0; margin: 0;">
			<img style="display: inline;" src="15x15.png">The image at the beginning of this sentence should be a 15px square.
		</body>
	</html>`,
	)

	size := image.Point{400, 300}
	page.Content.Layout(context.TODO(), size)

	body := page.getBody()
	if len(body.lineBoxes) < 2 {
		t.Fatal("Unexpected number of lines")
	}

	img := body.lineBoxes[0]
	txt := body.lineBoxes[1]
	if !img.IsImage() {
		t.Fatal("Expected first line box to be an image")
	}
	if txt.IsImage() {
		t.Fatal("Expected second line box to be text")
	}

	if img.origin != image.ZP {
		t.Errorf("Unexpected origin for img: got %v want (0, 0)", img.origin)
	}
	if txt.origin != (image.Point{15, 0}) {
		t.Errorf("Unexpected origin for text: got %v want (16, 0)", txt.origin)
	}
}

// Tests that an block image is positioned appropriately.
func TestBlockImg(t *testing.T) {
	page := parseHTML(
		t,
		`<html>
		<body style="font-size: 15px;padding: 0; margin: 0;">
			<img style="display: block;" src="15x15.png">The image at the beginning of this sentence should be a 15px square.
		</body>
	</html>`,
	)

	size := image.Point{400, 300}
	page.Content.Layout(context.TODO(), size)

	body := page.getBody()
	if len(body.lineBoxes) < 2 {
		t.Fatal("Unexpected number of lines")
	}

	firstchar := body.lineBoxes[0]
	if firstchar.IsImage() {
		t.Fatal("Block image generated line box")
	}

	if firstchar.origin != (image.Point{0, 16}) {
		t.Errorf("Unexpected origin for text: got %v want (0, 16)", firstchar.origin)
	}

	img := body.FirstChild.NextSibling
	if img.Data != "img" {
		t.Fatal("Could not find image")
	}
	if img.BoxDrawRectangle != (image.Rectangle{
		image.ZP,
		image.Point{15, 15},
	}) {
		t.Errorf("Unexpected image location: got %v", img.BoxDrawRectangle)
	}
}

// Tests that an block image with auto margins is positioned appropriately.
func TestCenteredBlockImg(t *testing.T) {
	page := parseHTML(
		t,
		`<html>
		<body style="font-size: 15px;padding: 0; margin: 0;">
			<img style="display: block; margin-left: auto; margin-right: auto; width: auto;" src="15x15.png">
		</body>
	</html>`,
	)

	size := image.Point{400, 300}
	page.Content.Layout(context.TODO(), size)

	img := page.getBody().FirstChild.NextSibling
	if img.Data != "img" {
		t.Fatal("Could not find image")
	}
	testImgDrawRectangle(
		t,
		img,
		image.Rectangle{
			image.Point{192, 0},
			image.Point{207, 15},
		},
	)
}

// Tests that an block image with auto margins and explicit width is positioned appropriately.
func TestResizedCenteredBlockImg(t *testing.T) {
	page := parseHTML(
		t,
		`<html>
		<body style="font-size: 15px;padding: 0; margin: 0;">
			<img style="display: block; margin-left: auto; margin-right: auto; width: 50%;" src="15x15.png">The image at the beginning of this sentence should be a 15px square.
		</body>
	</html>`,
	)

	size := image.Point{400, 300}
	page.Content.Layout(context.TODO(), size)

	img := page.getBody().FirstChild.NextSibling
	if img.Data != "img" {
		t.Fatal("Could not find image")
	}
	testImgDrawRectangle(
		t,
		img,
		image.Rectangle{
			image.Point{100, 0},
			image.Point{300, 200},
		},
	)
}

// Tests that an block image with auto left margin is positioned appropriately.
func TestResizedRightAlignedBlockImg(t *testing.T) {
	page := parseHTML(
		t,
		`<html>
		<body style="font-size: 15px; padding: 0; margin: 0;">
			<img style="display: block; margin-left: auto; margin-right: 0; width: 50%;" src="15x15.png">The image at the beginning of this sentence should be a 15px square.
		</body>
	</html>`,
	)

	size := image.Point{400, 300}
	page.Content.Layout(context.TODO(), size)

	img := page.getBody().FirstChild.NextSibling
	if img.Data != "img" {
		t.Fatal("Could not find image")
	}
	testImgDrawRectangle(t, img, image.Rectangle{
		image.Point{200, 0},
		image.Point{400, 200},
	})
}
