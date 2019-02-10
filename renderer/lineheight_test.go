package renderer

// The tests in this file are intended to mirror the tests at
// https://www.w3.org/Style/CSS/Test/CSS1/current/sec44.htm
// in a more automated fashion. Any given one of the properties
// being tested also deserves its own unit tests, but these tests
// are only verifying that the line height is being set appropriately.
import (
	"context"
	"fmt"
	"image"
	"testing"
)

// Test that when an image is in the middle of a line, the lineheight for the
// line containing the image is increased appropriately.
func TestImgTextLineheight(t *testing.T) {
	page := parseHTML(
		t,
		`<html>
		<body>
			<p style="line-height: 20px; font-size: 14px">Here is an image. <img src="15x15.png" style="vertical-align: text-top; width: 200px; height:200px;">.This<br> is inline text within the body. It has a line-height of 20px,
and a font-size of 14px. This means that each line should be exactly 20px below
the previous one, with the 6px evenly distributed above and below the line.
To test this. This is  We need to ensure that there are multiple lines. This should be
enough text.
		</body>
	</html>`,
	)
	size := image.Point{500, 300}
	page.Content.Layout(context.TODO(), size)
	body := page.getBody()
	p := body.FirstChild.NextSibling
	// There wasn't much text before the image, so there should
	// be the linebox for the firstchar, then the rest of the
	// text, then the image
	sameline := p.lineBoxes[1]
	img := p.lineBoxes[2]
	// There was a little text that should be on the same line, then
	// an explicit br, so the line box after that should be on the next line.
	nextline := p.lineBoxes[4]
	nextnextline := p.lineBoxes[5]

	if !img.IsImage() {
		t.Fatal("Could not find image linebox.")
	}
	if img.origin.Y != 0 {
		t.Errorf("Unexpected Y position for image: got %v want 0", img.origin.Y)
	}

	// The 3px half-leading still factors into the position of the
	// text. (Doesn't make much sense to me, but both Firefox and Chrome
	// seem to have this behaviour.)
	if sameline.origin.Y != 3 {
		t.Errorf("Unexpected Y position for text on image line: got %v want 3", sameline.origin.Y)
	}

	// Even though the paragraph has a line-height of 20, the image
	// should have increased the line with the image to 200.
	if nextline.origin.Y != 203 {
		t.Errorf("Unexpected Y position for text on next line: got %v want 203", nextline.origin.Y)
	}

	// Lines after that should have went back to a 20px lineheight,
	// then split the 6px leading between the top and the bottom of the text
	if nextnextline.origin.Y != 223 {
		t.Errorf("Unexpected Y position for text on next line: got %v want 223", nextnextline.origin.Y)
	}
}

// Test that when an image is in the middle of a line, the lineheight for the
// line containing the image is increased appropriately.
func TestImgTextBottomLineheight(t *testing.T) {
	page := parseHTML(
		t,
		`<html>
		<body>
			<p style="line-height: 20px; font-size: 14px">Here is an image. <img src="15x15.png" style="vertical-align: text-bottom; width: 200px; height:200px;">.This<br> is inline text within the body. It has a line-height of 20px,
and a font-size of 14px. This means that each line should be exactly 20px below
the previous one, with the 6px evenly distributed above and below the line.
To test this. This is  We need to ensure that there are multiple lines. This should be
enough text.
		</body>
	</html>`,
	)
	size := image.Point{500, 300}
	page.Content.Layout(context.TODO(), size)
	p := page.getBody().FirstChild.NextSibling
	// There wasn't much text before the image, so there should
	// be the linebox for the firstchar, then the rest of the
	// text, then the image
	sameline := p.lineBoxes[1]
	img := p.lineBoxes[2]
	// There was a little text that should be on the same line, then
	// an explicit br, so the line box after that should be on the next line.
	nextline := p.lineBoxes[4]
	nextnextline := p.lineBoxes[5]

	if !img.IsImage() {
		t.Fatal("Could not find image linebox.")
	}
	if img.origin.Y != 0 {
		t.Errorf("Unexpected Y position for image: got %v want 0", img.origin.Y)
	}

	if want := 200 - sameline.metrics.Height.Ceil(); sameline.origin.Y != want {
		t.Errorf("Unexpected Y position for text on image line: got %v want %v", sameline.origin.Y, want)
	}

	// Even though the paragraph has a line-height of 20, the image
	// should have increased the line with the image to 200.
	if nextline.origin.Y != 203 {
		t.Errorf("Unexpected Y position for text on next line: got %v want 203", nextline.origin.Y)
	}

	// Lines after that should have went back to a 20px lineheight,
	// then split the 6px leading between the top and the bottom of the text
	if nextnextline.origin.Y != 223 {
		t.Errorf("Unexpected Y position for text on next line: got %v want 223", nextnextline.origin.Y)
	}
}

// Test that when multiple images are on the same line with different vertical-align
// properties, the lines are positioned appropriately.
func TestImgMixedVerticalAlign(t *testing.T) {
	page := parseHTML(
		t,
		`<html>
		<body>
			<p style="line-height: 20px; font-size: 14px">Here is <img src="15x15.png" style="vertical-align: text-bottom; width: 200px; height:200px;">.This
<img src="15x15.png" style="vertical-align: text-top; width: 200px; height:200px;">.This
<br> is inline text within the body. It has a line-height of 20px,
		</body>
	</html>`,
	)
	size := image.Point{600, 300}
	page.Content.Layout(context.TODO(), size)
	p := page.getBody().FirstChild.NextSibling
	// There wasn't much text before the image, so there should
	// be the linebox for the firstchar, then the rest of the
	// text, then the image
	sameline := p.lineBoxes[1]
	img := p.lineBoxes[2]
	sameline2 := p.lineBoxes[3]
	img2 := p.lineBoxes[4]
	sameline3 := p.lineBoxes[5]
	nextline := p.lineBoxes[6]

	if !img.IsImage() {
		t.Fatal("Could not find image linebox.")
	}
	if !img2.IsImage() {
		t.Fatal("Could not find second image linebox.")
	}

	if want := 200 - sameline.metrics.Height.Ceil(); sameline.origin.Y != want {
		t.Errorf("Unexpected Y position for text before images: got %v want %v", sameline.origin.Y, want)
	}
	if want := 200 - sameline2.metrics.Height.Ceil(); sameline2.origin.Y != want {
		t.Errorf("Unexpected Y position for text between images: got %v want %v", sameline2.origin.Y, want)
	}
	if want := 200 - sameline3.metrics.Height.Ceil(); sameline3.origin.Y != want {
		t.Errorf("Unexpected Y position for text after images: got %v want %v", sameline3.origin.Y, want)
	}

	// One image was top aligned, one bottom aligned, for a total of 400px,
	// but the top-aligned one is offset by the text height, so we subtract
	// that from the total expected.
	if want := 400 - sameline2.metrics.Height.Ceil(); nextline.origin.Y != want {
		t.Errorf("Unexpected Y position for text on image line: got %v want %v", nextline.origin.Y, want)
	}
}

// Test that when an image is in the middle of a line with a margin, border,
// and padding, the margin/border/padding is taken into account when adjusting
// the lineheight.
func TestImgBorderTextLineheight(t *testing.T) {
	page := parseHTML(
		t,
		`<html>
		<body>
			<p style="line-height: 20px; font-size: 14px">Here is an image. <img src="15x15.png" style="vertical-align: text-top; width: 200px; height:200px; padding: 5px; border: 10px solid black; margin: 15px;">.This<br> is inline text within the body. It has a line-height of 20px,
and a font-size of 14px. This means that each line should be exactly 20px below
the previous one, with the 6px evenly distributed above and below the line.
To test this. This is  We need to ensure that there are multiple lines. This should be
enough text.
		</body>
	</html>`,
	)
	size := image.Point{500, 300}
	page.Content.Layout(context.TODO(), size)
	p := page.getBody().FirstChild.NextSibling
	// There wasn't much text before the image, so there should
	// be the linebox for the firstchar, then the rest of the
	// text, then the image
	sameline := p.lineBoxes[1]
	img := p.lineBoxes[2]
	// There was a little text that should be on the same line, then
	// an explicit br, so the line box after that should be on the next line.
	nextline := p.lineBoxes[4]
	nextnextline := p.lineBoxes[5]

	if !img.IsImage() {
		t.Fatal("Could not find image linebox.")
	}

	// Origin is the content origin, not the border origin, so it should
	// be 15(margin) + 10 (border) + 5(padding) = 30
	if want := 30; img.origin.Y != want {
		t.Errorf("Unexpected Y position for image: got %v want %v", img.origin.Y, want)
	}

	// The 3px half-leading still factors into the position of the
	// text. (Doesn't make much sense to me, but both Firefox and Chrome
	// seem to have this behaviour.)
	if sameline.origin.Y != 3 {
		t.Errorf("Unexpected Y position for text on image line: got %v want 3", sameline.origin.Y)
	}

	// Even though the paragraph has a line-height of 20, the image
	// should have increased the line with the image to 260 with the
	// margin and padding and border.
	if want := 263; nextline.origin.Y != want {
		t.Errorf("Unexpected Y position for text on next line: got %v want %v", nextline.origin.Y, want)
	}

	// Lines after that should have went back to a 20px lineheight,
	// then split the 6px leading between the top and the bottom of the text
	if want := 283; nextnextline.origin.Y != want {
		t.Errorf("Unexpected Y position for text on next line: got %v want %v", nextnextline.origin.Y, want)
	}
}

// Test that when an image is in the middle of a line with a margin, border,
// and padding, the margin/border/padding is taken into account when adjusting
// the lineheight.
func TestImgBorderTextBottomLineheight(t *testing.T) {
	page := parseHTML(
		t,
		`<html>
		<body>
			<p style="line-height: 20px; font-size: 14px">Here is an image. <img src="15x15.png" style="vertical-align: text-bottom; width: 200px; height:200px; padding: 5px; border: 10px solid black; margin: 15px;">.This<br> is inline text within the body. It has a line-height of 20px,
and a font-size of 14px. This means that each line should be exactly 20px below
the previous one, with the 6px evenly distributed above and below the line.
To test this. This is  We need to ensure that there are multiple lines. This should be
enough text.
		</body>
	</html>`,
	)
	size := image.Point{500, 300}
	page.Content.Layout(context.TODO(), size)
	p := page.getBody().FirstChild.NextSibling
	// There wasn't much text before the image, so there should
	// be the linebox for the firstchar, then the rest of the
	// text, then the image
	sameline := p.lineBoxes[1]
	img := p.lineBoxes[2]
	// There was a little text that should be on the same line, then
	// an explicit br, so the line box after that should be on the next line.
	nextline := p.lineBoxes[4]
	nextnextline := p.lineBoxes[5]

	if !img.IsImage() {
		t.Fatal("Could not find image linebox.")
	}

	// Origin is the content origin, not the border origin, so it should
	// be 15(margin) + 10 (border) + 5(padding) = 30
	if want := 30; img.origin.Y != want {
		t.Errorf("Unexpected Y position for image: got %v want %v", img.origin.Y, want)
	}

	if want := 260 - sameline.metrics.Height.Ceil(); sameline.origin.Y != want {
		t.Errorf("Unexpected Y position for text on image line: got %v want %v", sameline.origin.Y, want)
	}

	// Even though the paragraph has a line-height of 20, the image
	// should have increased the line with the image to 260 with the
	// margin and padding and border. The text is at 260-14 = 246, then
	// the next line should be at the maximum of 260 or 246+20 = 266
	if want := 266; nextline.origin.Y != want {
		t.Errorf("Unexpected Y position for text on next line: got %v want %v", nextline.origin.Y, want)
	}

	// Lines after that should have went back to a 20px lineheight,
	// then split the 6px leading between the top and the bottom of the text
	if want := 286; nextnextline.origin.Y != want {
		t.Errorf("Unexpected Y position for text on next line: got %v want %v", nextnextline.origin.Y, want)
	}
}

// Test that when multiple images are on the same line with different vertical-align
// properties, the lines are positioned appropriately in the case where the
// image has border/padding/margin.
func TestImgBorderMixedVerticalAlign(t *testing.T) {
	page := parseHTML(
		t,
		`<html>
		<body>
			<p style="line-height: 20px; font-size: 14px">Here is <img src="15x15.png" style="vertical-align: text-bottom; width: 200px; height:200px;padding: 5px; border: 10px solid black; margin: 15px;">.This
<img src="15x15.png" style="vertical-align: text-top; width: 200px; height:200px;padding: 5px; border: 10px solid black; margin: 15px;">.This
<br> is inline text within the body. It has a line-height of 20px,
		</body>
	</html>`,
	)
	size := image.Point{900, 300}
	page.Content.Layout(context.TODO(), size)
	p := page.getBody().FirstChild.NextSibling
	// There wasn't much text before the image, so there should
	// be the linebox for the firstchar, then the rest of the
	// text, then the image
	sameline := p.lineBoxes[1]
	img := p.lineBoxes[2]
	sameline2 := p.lineBoxes[3]
	img2 := p.lineBoxes[4]
	sameline3 := p.lineBoxes[5]
	nextline := p.lineBoxes[6]

	if !img.IsImage() {
		t.Fatal("Could not find image linebox.")
	}
	if !img2.IsImage() {
		t.Fatal("Could not find second image linebox.")
	}

	if want := 260 - sameline.metrics.Height.Ceil(); sameline.origin.Y != want {
		t.Errorf("Unexpected Y position for text before images: got %v want %v", sameline.origin.Y, want)
	}
	if want := 260 - sameline2.metrics.Height.Ceil(); sameline2.origin.Y != want {
		t.Errorf("Unexpected Y position for text between images: got %v want %v", sameline2.origin.Y, want)
	}
	if want := 260 - sameline3.metrics.Height.Ceil(); sameline3.origin.Y != want {
		t.Errorf("Unexpected Y position for text after images (%v): got %v want %v", sameline3.content, sameline3.origin.Y, want)
	}

	// The reasoning here is the same as in TestImgMixedVerticalAlign.
	// FIXME: Verify that the logic still applies with margins.
	if want := (260 * 2) - sameline2.metrics.Height.Ceil(); nextline.origin.Y != want {
		t.Errorf("Unexpected Y position for text after image line: got %v want %v", nextline.origin.Y, want)
	}
}

// Test that when an image is in the middle of a line, the lineheight for the
// line containing the image is increased appropriately.
func TestImgMiddleLineheight(t *testing.T) {
	page := parseHTML(
		t,
		`<html>
		<body>
			<p style="line-height: 20px; font-size: 14px">Here is an image. <img src="15x15.png" style="vertical-align: middle; width: 200px; height:200px;">.This<br> is inline text within the body. It has a line-height of 20px,
and a font-size of 14px. This means that each line should be exactly 20px below
the previous one, with the 6px evenly distributed above and below the line.
To test this. This is  We need to ensure that there are multiple lines. This should be
enough text.
		</body>
	</html>`,
	)
	size := image.Point{500, 300}
	page.Content.Layout(context.TODO(), size)
	p := page.getBody().FirstChild.NextSibling
	// There wasn't much text before the image, so there should
	// be the linebox for the firstchar, then the rest of the
	// text, then the image
	sameline := p.lineBoxes[1]
	img := p.lineBoxes[2]
	// There was a little text that should be on the same line, then
	// an explicit br, so the line box after that should be on the next line.
	nextline := p.lineBoxes[4]
	nextnextline := p.lineBoxes[5]

	if !img.IsImage() {
		t.Fatal("Could not find image linebox.")
	}
	if img.origin.Y != 0 {
		t.Errorf("Unexpected Y position for image: got %v want 0", img.origin.Y)
	}

	// FIXME: This should be tested in accordance with the CSS
	// spec, which says it should align "the vertical midpoint of
	// the element with the baseline plus half the x-height of the
	// parent", but for now truetype doesn't seem to be setting the
	// XHeight properly in the font metrics, so this probably needs
	// to wait until the font parsing is switched to sfnt, and this is
	// a good enough approximation for now.
	if want := 100 - (sameline.LineHeight() / 2); sameline.origin.Y != want {
		fmt.Printf("%v %v %v\n", sameline.metrics.Ascent, sameline.metrics.Height, sameline.metrics.Descent)
		t.Errorf("Unexpected Y position for text on image line: got %v want %v", sameline.origin.Y, want)
	}

	// Even though the paragraph has a line-height of 20, the image
	// should have increased the line with the image to 200.
	if nextline.origin.Y != 203 {
		t.Errorf("Unexpected Y position for text on next line: got %v want 203", nextline.origin.Y)
	}

	// Lines after that should have went back to a 20px lineheight,
	// then split the 6px leading between the top and the bottom of the text
	if nextnextline.origin.Y != 223 {
		t.Errorf("Unexpected Y position for text on next line: got %v want 223", nextnextline.origin.Y)
	}
}

// Tests that when right or bottom borders are negative, they suck in
// the text.
func TestImgNegativeMargin(t *testing.T) {
	page := parseHTML(
		t,
		`<html>
		<body>
			<p style="line-height: 20px; font-size: 14px">Texty<img src="15x15.png" style="vertical-align: middle; width: 200px; height:200px; margin: -15px">After<br>Next
			</p>
		</body>
	</html>`,
	)
	size := image.Point{500, 300}
	page.Content.Layout(context.TODO(), size)
	p := page.getBody().FirstChild.NextSibling
	img := p.lineBoxes[2]

	if !img.IsImage() {
		t.Fatal("Could not find image linebox.")
	}

	// first letter + rest of text = width of the first text.
	width := p.lineBoxes[0].width() + p.lineBoxes[1].width()
	if want := (image.Point{width - 15, -15}); img.origin != want {
		t.Errorf("Image origin at unexpected location. got %v want %v", img.origin, want)

	}
	after := p.lineBoxes[3]
	if want := width + 200 - 30; after.origin.X != want {
		t.Errorf("Text after origin at unexpected location. got %v want %v", after.origin.X, want)
	}

	nextline := p.lineBoxes[4]
	if want := 200 - 30 + 3; nextline.origin.Y != want {
		t.Errorf("Unexpected origin for next line. got %v want %v", nextline.origin.Y, want)
	}
}

// Tests that when right or bottom borders are negative, they suck in
// the text based on the border, not the content.
func TestImgNegativeMarginBorder(t *testing.T) {
	page := parseHTML(
		t,
		`<html>
		<body>
			<p style="line-height: 20px; font-size: 14px">Texty<img src="15x15.png" style="width: 200px; height:200px; margin: -15px; border: 10px solid black; padding: 5px; ">After<br>Next
			</p>
		</body>
	</html>`,
	)
	size := image.Point{500, 300}
	page.Content.Layout(context.TODO(), size)
	p := page.getBody().FirstChild.NextSibling
	img := p.lineBoxes[2]

	if !img.IsImage() {
		t.Fatal("Could not find image linebox.")
	}

	// first letter + rest of text = width of the first text
	width := p.lineBoxes[0].width() + p.lineBoxes[1].width()
	// The Y origin is at 0, because the padding/border compensate for the
	// negative margin
	if want := (image.Point{width - 15 + 10 + 5, 0}); img.origin != want {
		t.Errorf("Image origin at unexpected location. got %v want %v", img.origin, want)

	}
	after := p.lineBoxes[3]
	if want := width + 230 - (15 * 2); after.origin.X != want {
		t.Errorf("Text after origin at unexpected location. got %v want %v", after.origin.X, want)
	}

	nextline := p.lineBoxes[4]
	if want := 230 - 30 + 3; nextline.origin.Y != want {
		t.Errorf("Unexpected origin for next line. got %v want %v", nextline.origin.Y, want)
	}

}
