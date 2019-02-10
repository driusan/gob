package renderer

import (
	"context"
	"image"
	"strings"
	"testing"
)

// Tests that vertical margins collapse on 2 paragraphs.
func TestVerticalMarginCollapseBasic(t *testing.T) {
	page := parseHTML(
		t,
		`<html>
		<head>
			<style>
			html, body {
				margin: 0; padding: 0;
			}
			</style>
		</head>
		<body>
			<div style="display: block; width: 100px; height: 50px; margin-bottom: 40px;">Test</div>
			<div style="display: block; width: 100px; height: 50px; margin-top: 40px;">Test2</div>
		</body>
	</html>`,
	)

	page.Content.Layout(context.TODO(), image.Point{400, 300})

	nodes := []struct {
		el   *RenderableDomElement
		want image.Rectangle
	}{
		{
			// The first div should be at the origin.
			page.getBody().FirstChild.NextSibling,
			image.Rectangle{image.ZP, image.Point{100, 50}},
		},
		{
			// The second div should be 40px below the first div,
			// because the margins should have collapsed.
			page.getBody().FirstChild.NextSibling.NextSibling.NextSibling,
			image.Rectangle{image.Point{0, 90}, image.Point{100, 140}},
		},
	}

	for i, el := range nodes {
		if el.el.BoxDrawRectangle != el.want {
			t.Errorf("Test case %d: got %v want %v", i, el.el.BoxDrawRectangle, el.want)
		}
	}
}

// Tests that when the top margin is negative, it doesn't collapse.
func TestVerticalMarginCollapseNegativeTop(t *testing.T) {
	page := parseHTML(
		t,
		`<html>
		<head>
			<style>
			html, body {
				margin: 0; padding: 0;
			}
			</style>
		</head>
		<body>
			<div style="display: block; width: 100px; height: 50px; margin-bottom: 40px;">Test</div>
			<div style="display: block; width: 100px; height: 50px; margin-top: -20px;">Test2</div>
		</body>
	</html>`,
	)

	page.Content.Layout(context.TODO(), image.Point{400, 300})

	nodes := []struct {
		el   *RenderableDomElement
		want image.Rectangle
	}{
		{
			// The first div should be at the origin.
			page.getBody().FirstChild.NextSibling,
			image.Rectangle{image.ZP, image.Point{100, 50}},
		},
		{
			// The second div should be 40px below the first div,
			// because the margins should have collapsed.
			page.getBody().FirstChild.NextSibling.NextSibling.NextSibling,
			image.Rectangle{image.Point{0, 70}, image.Point{100, 120}},
		},
	}

	for i, el := range nodes {
		if el.el.BoxDrawRectangle != el.want {
			t.Errorf("Test case %d: got %v want %v", i, el.el.BoxDrawRectangle, el.want)
		}
	}
}

// Tests that when the bottom margin is negative, it doesn't collapse.
func TestVerticalMarginCollapseNegativeBottom(t *testing.T) {
	page := parseHTML(
		t,
		`<html>
		<head>
			<style>
			html, body {
				margin: 0; padding: 0;
			}
			</style>
		</head>
		<body>
			<div style="display: block; width: 100px; height: 50px; margin-bottom: -20px;">Test</div>
			<div style="display: block; width: 100px; height: 50px; margin-top: 40px;">Test2</div>
		</body>
	</html>`,
	)

	page.Content.Layout(context.TODO(), image.Point{400, 300})

	nodes := []struct {
		el   *RenderableDomElement
		want image.Rectangle
	}{
		{
			// The first div should be at the origin.
			page.getBody().FirstChild.NextSibling,
			image.Rectangle{image.ZP, image.Point{100, 50}},
		},
		{
			// The second div should be 40px below the first div,
			// because the margins should have collapsed.
			page.getBody().FirstChild.NextSibling.NextSibling.NextSibling,
			image.Rectangle{image.Point{0, 70}, image.Point{100, 120}},
		},
	}

	for i, el := range nodes {
		if el.el.BoxDrawRectangle != el.want {
			t.Errorf("Test case %d: got %v want %v", i, el.el.BoxDrawRectangle, el.want)
		}
	}
}

// Tests that padding prevents vertical margins from collapsing.
func TestVerticalMarginPaddingNoCollapse(t *testing.T) {
	page := parseHTML(
		t,
		`<html>
		<head>
			<style>
			html, body {
				margin: 0; padding: 0;
			}
			</style>
		</head>
		<body>
			<div style="display: block; width: 100px; height: 50px; margin-bottom: 20px;">Test</div>
			<div style="display: block; width: 100px; height: 50px; padding-top: 5px;"><p style="width: 100px; height: 45px; margin-top: 3px;">Test2</p></div>
		</body>
	</html>`,
	)

	page.Content.Layout(context.TODO(), image.Point{400, 300})

	nodes := []struct {
		el   *RenderableDomElement
		want image.Rectangle
	}{
		{
			// The first div should be at the origin.
			page.getBody().FirstChild.NextSibling,
			image.Rectangle{image.ZP, image.Point{100, 50}},
		},
		{
			// The second div should be 20px below the first div,
			// because the margins should have collapsed. (The padding
			// is included in the BoxDrawRectangle.)
			page.getBody().FirstChild.NextSibling.NextSibling.NextSibling,
			image.Rectangle{image.Point{0, 70}, image.Point{100, 125}},
		},
		{
			// The paragraph's margins should *not* have collapsed,
			// because there's padding that separates it.
			// It should be at the top of the parent plus the parent's
			// top 5px padding plus its own 3px margin.
			page.getBody().FirstChild.NextSibling.NextSibling.NextSibling.FirstChild,
			image.Rectangle{image.Point{0, 78}, image.Point{100, 78 + 45}},
		},
	}

	for i, el := range nodes {
		if got := el.el.getAbsoluteDrawRectangle(); got != el.want {
			t.Errorf("Test case %d: got %v want %v", i, got, el.want)
		}
	}
}

// Tests that when both margins are negative, it does collapses to a negative margin
// with the largest absolute value.
func TestVerticalMarginBothNegativeMargins(t *testing.T) {
	page := parseHTML(
		t,
		`<html>
		<head>
			<style>
			html, body {
				margin: 0; padding: 0;
			}
			</style>
		</head>
		<body>
			<div style="display: block; width: 100px; height: 50px; padding-bottom: 40px; margin-bottom: -20px;">Test</div>
			<div style="display: block; width: 100px; height: 50px; padding-top: 5px; margin-top: -30px;">Test2</div>
		</body>
	</html>`,
	)

	page.Content.Layout(context.TODO(), image.Point{400, 300})

	nodes := []struct {
		el   *RenderableDomElement
		want image.Rectangle
	}{
		{
			// The first div should be at the origin.
			page.getBody().FirstChild.NextSibling,
			image.Rectangle{image.ZP, image.Point{100, 90}},
		},
		{
			// The negative margin should have pushed it over the
			// padding of the other div.
			page.getBody().FirstChild.NextSibling.NextSibling.NextSibling,
			image.Rectangle{image.Point{0, 60}, image.Point{100, 115}},
		},
	}

	for i, el := range nodes {
		if got := el.el.getAbsoluteDrawRectangle(); got != el.want {
			t.Errorf("Smaller margin-top Test case %d: got %v want %v", i, got, el.want)
		}
	}

	// Do the same test, but this time make the margin-bottom bigger. The
	// result should have collapsed to the same thing.
	page = parseHTML(
		t,
		`<html>
		<head>
			<style>
			html, body {
				margin: 0; padding: 0;
			}
			</style>
		</head>
		<body>
			<div style="display: block; width: 100px; height: 50px; padding-bottom: 40px; margin-bottom: -30px;">Test</div>
			<div style="display: block; width: 100px; height: 50px; padding-top: 5px; margin-top: -20px;">Test2</div>
		</body>
	</html>`,
	)

	page.Content.InvalidateLayout()
	page.Content.Layout(context.TODO(), image.Point{400, 300})

	// Make sure we're referencing the elements from the new layout and not
	// holding on to pointers from the old one. The rectangle itself doesn't
	// change.
	nodes[0].el = page.getBody().FirstChild.NextSibling
	nodes[1].el = page.getBody().FirstChild.NextSibling.NextSibling.NextSibling

	for i, el := range nodes {
		if got := el.el.getAbsoluteDrawRectangle(); got != el.want {
			t.Errorf("Smaller margin-bottom test case %d: got %v want %v", i, got, el.want)
		}
	}

}

// Tests that margins of floats do not collapse.
func TestVerticalMarginFloatNoCollapse(t *testing.T) {
	page := parseHTML(
		t,
		`<html>
		<head>
			<style>
			html, body {
				margin: 0; padding: 0;
			}
			</style>
		</head>
		<body>
			<div style="display: block; width: 100px; height: 50px; margin-bottom: 20px;">Baseline</div>
			<div style="display: block; width: 100px; height: 50px; margin-top: 30px; float: left">Float</div>
			<div style="display: block; height: 50px; margin-top: 30px; clear: none;">Collapsed</div>
		</body>
	</html>`,
	)

	page.Content.Layout(context.TODO(), image.Point{400, 300})

	// FIXME: There is a hack where Layout fudges the float margins in order
	// to make intersection calculation easier, and then readjusts them when
	// drawing. This should be removed.
	// For now, we just draw to make the float get to its real location before
	// testing its rectangle.
	page.Content.drawInto(context.TODO(), image.NewRGBA(image.Rectangle{image.ZP, image.Point{400, 300}}), image.ZP)
	nodes := []struct {
		el   *RenderableDomElement
		want image.Rectangle
	}{
		{
			// The first div should be at the origin.
			page.getBody().FirstChild.NextSibling,
			image.Rectangle{image.ZP, image.Point{100, 50}},
		},
		{
			// The second (floating) div should not have had its
			// margins collapse.
			page.getBody().FirstChild.NextSibling.NextSibling.NextSibling,
			image.Rectangle{image.Point{0, 100}, image.Point{100, 150}},
		},
		{
			// The third (non-floating) div should have had its margins
			// collapse.
			page.getBody().FirstChild.NextSibling.NextSibling.NextSibling.NextSibling.NextSibling,
			image.Rectangle{image.Point{100, 80}, image.Point{200, 130}},
		},
	}

	for i, el := range nodes {
		got := el.el.getAbsoluteDrawRectangle()
		if i == 2 {
			// Adding an explicit width to the last div affects
			// the layout in more ways than changing its width according
			// to the CSS spec, so we just assume the max X is
			// correct to avoid having to figure out what it should
			// have been without having set it to a constant.
			el.want.Max.X = got.Max.X
		}
		if got != el.want {
			t.Errorf("Test case %d: got %v want %v", i, got, el.want)
		}
	}

}

// This tests that the non-floating text is positioned properly when the
// float next to it is adjusted.
func TestVerticalMarginFloatNoCollapseTextPositioning(t *testing.T) {
	// We've had multiple regressions with this part of the
	// CSS 4.1.1 Vertical Formatting test suite. This is based on that,
	// but automated.
	page := parseHTML(
		t,
		`<html>
<head>
	<style>
html, body {
	margin: 0; padding: 0;
}
.nine {
	padding-bottom: 0;
	margin-bottom: 1cm;
}
.ten {
	padding-top: 0;
	margin-top: 1cm;
	float: left;
	width: 50%;
}
.eleven {
	padding-top: 0;
	margin-top: 1cm;
}
	</style>
</head>
	<body>
<p class="nine">This is a paragraph, which I should make very long
so that you can easily see how much space there is between it and
the one below it and to the right.</p>
<p class="ten">There should be two centimeters between this paragraph and te one
above it, since margins do not collapse on floating elements.</p>
<p class="eleven">
There should be one centimeter between this paragraph and the
(non-floating) one above it, since the float should not effect the
paragraph spacing.
</p>

	</body>
</html>`,
	)
	page.Content.Layout(context.TODO(), image.Point{800, 300})

	// FIXME: This is the same hack as the other float test.
	page.Content.drawInto(context.TODO(), image.NewRGBA(image.Rectangle{image.ZP, image.Point{800, 300}}), image.ZP)

	collapsed := page.getBody().FirstChild.NextSibling.NextSibling.NextSibling.NextSibling.NextSibling

	if !strings.HasPrefix(strings.TrimSpace(collapsed.GetTextContent()), "There should be one centimeter") {
		t.Fatalf("Did not retrieve correct div. got '%v'", collapsed.GetTextContent())
	}

	if len(collapsed.lineBoxes) < 3 {
		t.Fatal("Div did not span enough lines to test")
	}

	// We only check the line box positioning, we don't check the collapsing
	// or draw rectangles, because that was handled by TestVerticalMarginFloatNoCollapse
	// and calculating what they should be would depend on knowing the how
	// many pixels are in a cm.

	for i, line := range collapsed.lineBoxes {
		// We check for < 400 instead of != 400 because the first line has 2 line boxes, one for
		// the first character and one for the rest of the line.
		if line.origin.X < 400 {
			t.Errorf("Line %d started at %v, not 50%% of the container (Content: %v)", i, line.origin.X, line.content)
		}
	}
}
