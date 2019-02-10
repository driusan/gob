package renderer

import (
	"context"
	"image"
	"image/color"
	"testing"
)

// Tests that text floats around a left float and the float is on the left-hand
// side of the page. This checks for pixel-perfect rendering of borders, margin,
// padding and images.
func TestLeftFloatRendering(t *testing.T) {
	// We do the test for both text floats and image floats, with inline and
	// block images. The test net.URLReader always returns a 100x50 red png
	// for 100x50.png.
	floaters := []string{
		`<div style="float: left; display: block; width: 100px; height: 50px; margin-left: 0; margin-right: 3px; margin-top: 0; padding-top: 3px; border: 3px solid black; background: blue;">This is a float</div>`,
		`<img style="float: left; display: block; width: 100px; height: 50px; margin-left: 0; margin-right: 3px; margin-top: 0; padding-top: 3px; border: 3px solid black; background: blue;" src="100x50.png">`,
		`<img style="float: left; display: inline; width: 100px; height: 50px; margin-left: 0; margin-right: 3px; margin-top: 0; padding-top: 3px; border: 3px solid black; background: blue;" src="100x50.png">`,
	}

	// Some colours we assert in the tests.
	// img colour
	redRGBA := color.RGBA{255, 0, 0, 255}
	// border colour
	blackRGBA := color.RGBA{0, 0, 0, 255}
	// background of the float (ie. padding colour)
	blueRGBA := color.RGBA{0, 0, 255, 255}
	// background of the body.
	greenRGBA := color.RGBA{0, 128, 0, 255}

	// margins are transparent.
	transparentRGBA := color.RGBA{0, 0, 0, 0}

	for i, f := range floaters {
		page := parseHTML(
			t,
			`<html>
		<body style="background: green; padding: 0; margin: 0;">
			`+f+`
			This is non-floating text within the body. We should ensure that there's
			enough text that it goes past the end of the float.
			This is non-floating text within the body. We should ensure that there's
			enough text that it goes past the end of the float.
		</body>
	</html>`,
		)

		canvas := image.NewRGBA(image.Rectangle{image.ZP, image.Point{400, 300}})
		size := canvas.Bounds().Size()
		page.Content.Layout(context.TODO(), size)
		page.Content.RenderInto(context.TODO(), canvas, image.ZP)

		body := page.getBody()
		float := body.FirstChild.NextSibling
		switch i {
		case 0:
			// Only check that we got the right float for the
			// text float, since images don't have text content
			// to verify.
			if float.GetTextContent() != "This is a float" {
				t.Fatalf("Test %d: Could not get float element", i)
			}
		case 1, 2:
			// Only check that the content is pixel-perfect for images,
			// since we know that it's a 100x500 red rectangle.
			for x := 3; x < 103; x++ {
				for y := 6; y < 56; y++ {
					if clr := canvas.At(x, y); clr != redRGBA {
						t.Errorf("Test %d: Pixel (%d, %d). got %v want %v", i, x, y, clr, redRGBA)
					}
				}
			}
		default:
			// If something gets added to floaters, make sure
			// we add something (even if just an empty case)
			// to this switch in order to make sure we didn't
			// forget about any specific assertions here.
			t.Fatal("Added test case without making assertions")
		}

		// the BoxDrawRectangle includes the left and right margins. We
		// check the exact pixels to ensure there borders and padding are
		// drawn in the right place (and content for images) and that
		// there's nothing in the right margin.
		// (There is no top or bottom margins, so we don't care if it
		// includes them.)
		wantFloatRect := image.Rectangle{
			image.ZP,
			// X = 100px + 6px border + 3px margin
			// Y = 100px + 6px border + 3px padding
			image.Point{109, 59},
		}
		if float.BoxDrawRectangle != wantFloatRect {
			t.Errorf("Test %d: Float incorrectly positioned. got %v want %v", i, float.BoxDrawRectangle, wantFloatRect)
		}

		// check the borders
		// top
		for x := 0; x < 106; x++ {
			for y := 0; y < 3; y++ {
				if clr := canvas.At(x, y); clr != blackRGBA {
					t.Errorf("Test %d: Top border incorrect at %d, %d: got %v want: %v", i, x, y, clr, blackRGBA)

				}
			}
		}
		// bottom
		for x := 0; x < 106; x++ {
			for y := 56; y < 59; y++ {
				if clr := canvas.At(x, y); clr != blackRGBA {
					t.Errorf("Test %d: Bottom border incorrect at %d, %d: got %v want: %v", i, x, y, clr, blackRGBA)

				}
			}
		}
		// left
		for x := 0; x < 3; x++ {
			for y := 0; y < 59; y++ {
				if clr := canvas.At(x, y); clr != blackRGBA {
					t.Errorf("Test %d: Left border incorrect at %d, %d: got %v want: %v", i, x, y, clr, blackRGBA)

				}
			}
		}
		// right
		for x := 103; x < 106; x++ {
			for y := 0; y < 59; y++ {
				if clr := canvas.At(x, y); clr != blackRGBA {
					t.Errorf("Test %d: Right border incorrect at %d, %d: got %v want: %v", i, x, y, clr, blackRGBA)

				}
			}
		}

		// Check that the padding at the top is blue (the float background color)
		for x := 3; x < 103; x++ {
			for y := 3; y < 6; y++ {
				if clr := canvas.At(x, y); clr != blueRGBA {
					t.Errorf("Test %d: Top padding incorrect at %d, %d: got %v want: %v", i, x, y, clr, blueRGBA)

				}
			}
		}

		// Check that the right margin of the float is the body background.
		for x := 106; x < 109; x++ {
			for y := 0; y < 59; y++ {
				if clr := canvas.At(x, y); clr != greenRGBA {
					t.Errorf("Test %d: Right margin is incorrect at %d, %d: got %v want: %v", i, x, y, clr, transparentRGBA)

				}
			}
		}

		// Check that, somewhere greater than the margin, there's a
		// non-transparent pixel for the text of the rendered text of
		// the body.
		//
		// We don't know exactly where or the alpha value that the font
		// drawer chose, so we just check for anything non-transparent.
		found := false
		for x := 109; x < size.X; x++ {
			for y := 0; y < size.Y; y++ {
				if _, _, _, a := canvas.At(x, y).RGBA(); a > 0 {
					found = true
					break
				}

			}
			if found == true {
				break
			}
		}
		if !found {
			t.Errorf("Test %d: Did not find any text", i)
		}

		// Make sure that 1. The text lineboxes were long enough to go past
		// the float (if not we need to add more ipsum lorem text) and that 2.
		// The boxes after the float are at x=0, while the ones before are at
		// x >= 109 (the first line box is the first letter and the second is
		// the rest of the first line, so we can't check for an exact match
		found = false
		for _, lb := range body.lineBoxes {
			if lb.origin.Y >= 56+lb.Baseline() {
				if lb.origin.X != 0 {
					t.Errorf("Test %d: Text did not return to origin after float. got %v", i, lb.origin)
				} else {
					found = true
				}
			} else {
				if lb.origin.X < 109 {
					t.Errorf("Test %d: Text origin overlaps float got %v", i, lb.origin)
				}
			}
		}
		if !found {
			t.Fatalf("Test %d: No wrapping text found", i)
		}

		// Check for any intersection between the lines, or between the
		// lines and the float
		for j, lb1 := range body.lineBoxes {
			// lines
			for k, lb2 := range page.Content.lineBoxes {
				if j == k {
					continue
				}
				if lb1.Bounds().Overlaps(lb2.Bounds()) {
					t.Errorf("Error: %v (%v) overlaps %v (%v)", lb1.Content.Bounds(), lb1.content, lb2.Content.Bounds(), lb2.content)
				}
			}
			// Float
			if lb1.Bounds().Overlaps(float.BoxDrawRectangle) {
				t.Errorf("Test %d: line box %d overlaps float", i, j)
			}

		}
	}

}

// Similar to TestLeftFloatRendering, but for right floats.
func TestRightFloatRendering(t *testing.T) {
	// We do the test for both text floats and image floats, with inline and
	// block images. The test net.URLReader always returns a 100x50 red png
	// for 100x50.png.
	floaters := []string{
		`<div style="float: right; display: block; width: 100px; height: 50px; margin-left: 0; margin-left: 3px; margin-top: 0; padding-top: 3px; border: 3px solid black; background: blue;">This is a float</div>`,
		`<img style="float: right; display: block; width: 100px; height: 50px; margin-left: 0; margin-left: 3px; margin-top: 0; padding-top: 3px; border: 3px solid black; background: blue;" src="100x50.png">`,
		`<img style="float: right; display: inline; width: 100px; height: 50px; margin-left: 0; margin-left: 3px; margin-top: 0; padding-top: 3px; border: 3px solid black; background: blue;" src="100x50.png">`,
	}

	// Some colours we assert in the tests.
	// img colour
	redRGBA := color.RGBA{255, 0, 0, 255}
	// border colour
	blackRGBA := color.RGBA{0, 0, 0, 255}
	// background of the float (ie. padding colour)
	blueRGBA := color.RGBA{0, 0, 255, 255}
	// body background
	greenRGBA := color.RGBA{0, 128, 0, 255}

	for i, f := range floaters {
		page := parseHTML(
			t,
			`<html>
		<body style="background: green; padding: 0; margin: 0;">
			`+f+`
			This is non-floating text within the body. We should ensure that there's
			enough text that it goes past the end of the float.
			This is non-floating text within the body. We should ensure that there's
			enough text that it goes past the end of the float.
		</body>
	</html>`,
		)

		canvas := image.NewRGBA(image.Rectangle{image.ZP, image.Point{400, 300}})
		size := canvas.Bounds().Size()
		page.Content.Layout(context.TODO(), size)
		page.Content.RenderInto(context.TODO(), canvas, image.ZP)

		body := page.getBody()
		float := body.FirstChild.NextSibling
		switch i {
		case 0:
			// Only check that we got the right float for the
			// text float, since images don't have text content
			// to verify.
			if float.GetTextContent() != "This is a float" {
				t.Fatalf("Test %d: Could not get float element", i)
			}
		case 1, 2:
			// Only check that the content is pixel-perfect for images,
			// since we know that it's a 100x500 red rectangle.
			for x := 400 - 3 - 100; x < 400-3; x++ {
				for y := 6; y < 56; y++ {
					if clr := canvas.At(x, y); clr != redRGBA {
						t.Errorf("Test %d: Pixel (%d, %d). got %v want %v", i, x, y, clr, redRGBA)
					}
				}
			}
		default:
			// If something gets added to floaters, make sure
			// we add something (even if just an empty case)
			// to this switch in order to make sure we didn't
			// forget about any specific assertions here.
			t.Fatal("Added test case without making assertions")
		}

		// the BoxDrawRectangle includes the left and right margins. We
		// check the exact pixels to ensure there borders and padding are
		// drawn in the right place (and content for images) and that
		// there's nothing in the right margin.
		// (There is no top or bottom margins, so we don't care if it
		// includes them.)
		wantFloatRect := image.Rectangle{
			image.Point{400 - 100 - 6 - 3, 0},
			// X = 100px + 6px border + 3px margin
			// Y = 100px + 6px border + 3px padding
			image.Point{400, 59},
		}
		if float.BoxDrawRectangle != wantFloatRect {
			t.Errorf("Test %d: Float incorrectly positioned. got %v want %v", i, float.BoxDrawRectangle, wantFloatRect)
		}

		// check the borders
		// top
		for x := 400 - 100 - 6; x < 400; x++ {
			for y := 0; y < 3; y++ {
				if clr := canvas.At(x, y); clr != blackRGBA {
					t.Errorf("Test %d: Top border incorrect at %d, %d: got %v want: %v", i, x, y, clr, blackRGBA)

				}
			}
		}
		// bottom
		for x := 400 - 100 - 6; x < 400; x++ {
			for y := 56; y < 59; y++ {
				if clr := canvas.At(x, y); clr != blackRGBA {
					t.Errorf("Test %d: Bottom border incorrect at %d, %d: got %v want: %v", i, x, y, clr, blackRGBA)

				}
			}
		}
		// left
		for x := 400 - 100 - 6; x < 400-100-3; x++ {
			for y := 0; y < 59; y++ {
				if clr := canvas.At(x, y); clr != blackRGBA {
					t.Errorf("Test %d: Left border incorrect at %d, %d: got %v want: %v", i, x, y, clr, blackRGBA)

				}
			}
		}
		// right
		for x := 400 - 3; x < 400; x++ {
			for y := 0; y < 59; y++ {
				if clr := canvas.At(x, y); clr != blackRGBA {
					t.Errorf("Test %d: Right border incorrect at %d, %d: got %v want: %v", i, x, y, clr, blackRGBA)

				}
			}
		}

		// Check that the padding at the top is blue (the float background color)
		for x := 400 - 100 - 3; x < 400-3; x++ {
			for y := 3; y < 6; y++ {
				if clr := canvas.At(x, y); clr != blueRGBA {
					t.Errorf("Test %d: Top padding incorrect at %d, %d: got %v want: %v", i, x, y, clr, blueRGBA)

				}
			}
		}

		// Check that the right margin of the float is the body background.
		for x := 400 - 100 - 9; x < 400-100-6; x++ {
			for y := 0; y < 59; y++ {
				if clr := canvas.At(x, y); clr != greenRGBA {
					t.Errorf("Test %d: Right margin is incorrect at %d, %d: got %v want: %v", i, x, y, clr, greenRGBA)

				}
			}
		}

		// Check that, somewhere greater than the margin, there's a
		// non-transparent pixel for the text of the rendered text of
		// the body.
		//
		// We don't know exactly where or the alpha value that the font
		// drawer chose, so we just check for anything non-transparent.
		found := false
		for x := 109; x < size.X; x++ {
			for y := 0; y < size.Y; y++ {
				if _, _, _, a := canvas.At(x, y).RGBA(); a > 0 {
					found = true
					break
				}

			}
			if found == true {
				break
			}
		}
		if !found {
			t.Errorf("Test %d: Did not find any text", i)
		}

		// Check for any intersection between the lines, or between the
		// lines and the float
		for j, lb1 := range body.lineBoxes {
			// Lines
			for k, lb2 := range body.lineBoxes {
				if j == k {
					continue
				}
				if lb1.Bounds().Overlaps(lb2.Bounds()) {
					t.Errorf("Test %d: Error: %v (%v) overlaps %v (%v)", i, lb1.Content.Bounds(), lb1.content, lb2.Content.Bounds(), lb2.content)
				}
			}
			// Float
			if lb1.Bounds().Overlaps(float.BoxDrawRectangle) {
				t.Errorf("Test %d: line box %d overlaps float", i, j)
			}
		}
	}

}

// Test that when there is text before a left float, the float still floats before
// the text already rendered and does not overlap it.
func TestTextBeforeLeftFloat(t *testing.T) {
	// We do the test for both text floats and image floats, with inline and
	// block images. The test net.URLReader always returns a 100x50 red png
	// for 100x50.png.
	floaters := []string{
		`<div style="float: left; display: block; width: 100px; height: 50px;">This is a float</div>`,
		`<img style="float: left; display: block; width: 100px; height: 50px;" src="100x50.png">`,
		`<img style="float: left; display: inline; width: 100px; height: 50px;" src="100x50.png">`,
	}
	for i, f := range floaters {
		page := parseHTML(
			t,
			`<html style="padding: 0; margin: 0">
		<body style="background: green; padding: 0; margin: 0;">
			Before `+f+`After
		</body>
	</html>`,
		)

		canvas := image.NewRGBA(image.Rectangle{image.ZP, image.Point{400, 300}})
		size := canvas.Bounds().Size()
		page.Content.Layout(context.TODO(), size)
		page.Content.RenderInto(context.TODO(), canvas, image.ZP)

		body := page.getBody()
		float := body.FirstChild.NextSibling
		wantFloatRect := image.Rectangle{
			image.ZP,
			image.Point{100, 50},
		}
		if float.BoxDrawRectangle != wantFloatRect {
			t.Fatalf("Test %d: Float not positioned in correct location", i)
		}
		// Check for any intersection between the lines, or between the
		// lines and the float
		for j, lb1 := range body.lineBoxes {
			// Lines
			for k, lb2 := range body.lineBoxes {
				if j == k {
					continue
				}
				if lb1.Bounds().Overlaps(lb2.Bounds()) {
					t.Errorf("Test %d: Error: %v (%v) overlaps %v (%v)", i, lb1.Content.Bounds(), lb1.content, lb2.Content.Bounds(), lb2.content)
				}
			}
			// Float
			if lb1.Bounds().Overlaps(float.BoxDrawRectangle) {
				t.Errorf("Test %d: line box %d overlaps float (float: %v lb: %v)", i, j, float.BoxDrawRectangle, lb1.Content.Bounds())
			}
		}

	}
}

// Tests for a regression where right floats were being moved to the left when
// there wasn't enough room on the line for them, instead of being moved down.
func TestRegressionRightFloatInsufficientSpace(t *testing.T) {
	// We do the test for both text floats and image floats, with inline and
	// block images. The test net.URLReader always returns a 100x50 red png
	// for 100x50.png.
	floaters := []string{
		`<div style="float: right; display: block; width: 100px; height: 50px;">This is a float</div>`,
		`<img style="float: right; display: block; width: 100px; height: 50px;" src="100x50.png">`,
		`<img style="float: right; display: inline; width: 100px; height: 50px;" src="100x50.png">`,
	}
	for i, f := range floaters {
		page := parseHTML(
			t,
			`<html>
		<body style="background: green; padding: 0; margin: 0; line-height: 20px;">
			<div style="display: inline-block; width: 350px; height: 20px;">This is text</div>x
			`+f+`
		</body>
	</html>`,
		)

		canvas := image.NewRGBA(image.Rectangle{image.ZP, image.Point{400, 300}})
		size := canvas.Bounds().Size()
		page.Content.Layout(context.TODO(), size)
		page.Content.RenderInto(context.TODO(), canvas, image.ZP)

		body := page.getBody()
		float := body.FirstChild.NextSibling.NextSibling.NextSibling
		wantFloatRect := image.Rectangle{
			image.Point{400 - 100, 20},
			image.Point{400, 70},
		}
		if float.BoxDrawRectangle != wantFloatRect {
			t.Errorf("Test %d: Float not positioned in correct location. got %v want %v", i, float.BoxDrawRectangle, wantFloatRect)
		}
	}
}
