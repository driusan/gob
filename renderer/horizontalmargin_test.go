package renderer

// This file currently mostly duplicates the test cases from
// https://www.w3.org/Style/CSS/Test/CSS1/current/sec412.htm
// in an automated fashion.

import (
	"context"
	"image"
	"testing"
)

// Tests that a left margin indents the block.
func TestLeftMargin(t *testing.T) {
	page := parseHTML(
		t,
		`<html>
		<body style="padding: 0; margin: 0">
			<div style="display: block; width: 100px; height: 50px; margin-left: 10px;">Test</div>
		</body>
	</html>`,
	)

	page.Content.Layout(context.TODO(), image.Point{400, 300})

	// For horizontal margins we check both the draw rectangle and the
	// content rectangle, because horizontal margins are included in the
	// CSSOuterBox (unlike vertical margins, which are not to make collapsing
	// easier.)
	nodes := []struct {
		el          *RenderableDomElement
		wantDraw    image.Rectangle
		wantContent image.Rectangle
	}{
		{
			// The first div should be at the origin.
			page.getBody().FirstChild.NextSibling,
			image.Rectangle{image.Point{0, 0}, image.Point{110, 50}},
			image.Rectangle{image.Point{10, 0}, image.Point{110, 50}},
		},
	}

	for i, el := range nodes {
		if el.el.BoxDrawRectangle != el.wantDraw {
			t.Errorf("Test case %d: got %v for draw rectangle. want %v", i, el.el.BoxDrawRectangle, el.wantDraw)
		}
		if el.el.BoxContentRectangle != el.wantContent {
			t.Errorf("Test case %d: got %v for content rectangle. want %v", i, el.el.BoxContentRectangle, el.wantContent)
		}
	}
}

// Tests that embedded horizontal margins do not collapse.
func TestEmbeddedLeftMargin(t *testing.T) {
	page := parseHTML(
		t,
		`<html>
		<head>
			<style>html, body { padding: 0; margin: 0}</style>
		</head>
		<body>
			<div style="display: block; margin-left: 10px;">
				<div style="display: block; width: 100px; height: 50px; margin-left: 10px;">
					Test
				</div>
			</div>
		</body>
	</html>`,
	)

	page.Content.Layout(context.TODO(), image.Point{400, 300})

	target := page.getBody().FirstChild.NextSibling.FirstChild.NextSibling

	want := image.Rectangle{image.Point{10, 0}, image.Point{110, 50}}
	if target.BoxContentRectangle != want {
		// BoxContentRectangle is relative to the div itself.
		t.Errorf("Embedded div's content was in incorrect place.")
	}

	want = image.Rectangle{image.Point{10, 0}, image.Point{120, 50}}
	if target.getAbsoluteDrawRectangle() != want {
		t.Errorf("Box was drawn in incorrect location.")
	}
}

// Test that when the right margin is "auto" and the left is 0, the content is
// left aligned.
func TestAutoRightMargin(t *testing.T) {
	page := parseHTML(
		t,
		`<html>
		<head>
			<style>html, body { padding: 0; margin: 0;}</style>
		</head>
		<body>
			<div style="display: block; margin-left: 0; margin-right: auto; width: 50%; height: 100px;">Test</div>
		</body>
	</html>`,
	)

	page.Content.Layout(context.TODO(), image.Point{400, 300})

	target := page.getBody().FirstChild.NextSibling

	want := image.Rectangle{image.Point{0, 0}, image.Point{200, 100}}
	if target.BoxContentRectangle != want {
		t.Errorf("Div content was in incorrect place. got %v", target.BoxContentRectangle)
	}

	want = image.Rectangle{image.Point{0, 0}, image.Point{200, 100}}
	if target.getAbsoluteDrawRectangle() != want {
		t.Errorf("Box was drawn in incorrect location.")
	}
}

// Test that when the left margin is "auto" and the right is 0, the content is
// right aligned.
func TestAutoLeftMargin(t *testing.T) {
	page := parseHTML(
		t,
		`<html>
		<head>
			<style>html, body { padding: 0; margin: 0;}</style>
		</head>
		<body>
			<div style="display: block; margin-left: auto; margin-right: 0; width: 50%; height: 100px;">Test</div>
		</body>
	</html>`,
	)

	page.Content.Layout(context.TODO(), image.Point{400, 300})

	target := page.getBody().FirstChild.NextSibling

	want := image.Rectangle{image.Point{200, 0}, image.Point{400, 100}}
	if target.BoxContentRectangle != want {
		t.Errorf("Div content was in incorrect place. got %v", target.BoxContentRectangle)
	}

	want = image.Rectangle{image.Point{0, 0}, image.Point{400, 100}}
	if got := target.getAbsoluteDrawRectangle(); got != want {
		t.Errorf("Box was drawn in incorrect location. got %v", got)
	}
}

// Test that when the both left and right margin is "auto" the content is
// centered.
func TestAutoLeftAutoRightMargin(t *testing.T) {
	page := parseHTML(
		t,
		`<html>
		<head>
			<style>html, body { padding: 0; margin: 0;}</style>
		</head>
		<body>
			<div style="display: block; margin-left: auto; margin-right: auto; width: 50%; height: 100px;">Test</div>
		</body>
	</html>`,
	)

	page.Content.Layout(context.TODO(), image.Point{400, 300})

	target := page.getBody().FirstChild.NextSibling

	want := image.Rectangle{image.Point{100, 0}, image.Point{300, 100}}
	if target.BoxContentRectangle != want {
		t.Errorf("Div content was in incorrect place. got %v", target.BoxContentRectangle)
	}

	want = image.Rectangle{image.Point{0, 0}, image.Point{300, 100}}
	if got := target.getAbsoluteDrawRectangle(); got != want {
		t.Errorf("Box was drawn in incorrect location. got %v", got)
	}
}

// Test that when the width is auto, a left margin of "auto" becomes 0 and
// the width becomes 100%.
func TestAutoWidthLeft(t *testing.T) {
	page := parseHTML(
		t,
		`<html>
		<head>
			<style>html, body { padding: 0; margin: 0;}</style>
		</head>
		<body>
			<div style="display: block; margin-left: auto; margin-right: 0; width: auto; height: 100px;">Test</div>
		</body>
	</html>`,
	)

	page.Content.Layout(context.TODO(), image.Point{400, 300})

	target := page.getBody().FirstChild.NextSibling

	want := image.Rectangle{image.Point{0, 0}, image.Point{400, 100}}
	if target.BoxContentRectangle != want {
		t.Errorf("Div content was in incorrect place. got %v", target.BoxContentRectangle)
	}

	want = image.Rectangle{image.Point{0, 0}, image.Point{400, 100}}
	if got := target.getAbsoluteDrawRectangle(); got != want {
		t.Errorf("Box was drawn in incorrect location. got %v", got)
	}
}

// Test that when the width is auto, a right margin of "auto" becomes 0 and
// the width becomes 100%.
func TestAutoWidthRight(t *testing.T) {
	page := parseHTML(
		t,
		`<html>
		<head>
			<style>html, body { padding: 0; margin: 0;}</style>
		</head>
		<body>
			<div style="display: block; margin-left: 0; margin-right: auto; width: auto; height: 100px;">Test</div>
		</body>
	</html>`,
	)

	page.Content.Layout(context.TODO(), image.Point{400, 300})

	target := page.getBody().FirstChild.NextSibling

	want := image.Rectangle{image.Point{0, 0}, image.Point{400, 100}}
	if target.BoxContentRectangle != want {
		t.Errorf("Div content was in incorrect place. got %v", target.BoxContentRectangle)
	}

	want = image.Rectangle{image.Point{0, 0}, image.Point{400, 100}}
	if got := target.getAbsoluteDrawRectangle(); got != want {
		t.Errorf("Box was drawn in incorrect location. got %v", got)
	}
}

// Test that when the width is auto, a both a left/right margin of "auto"
// becomes 0 and the width becomes 100%.
func TestAutoWidthLeftRight(t *testing.T) {
	page := parseHTML(
		t,
		`<html>
		<head>
			<style>html, body { padding: 0; margin: 0;}</style>
		</head>
		<body>
			<div style="display: block; margin-left: auto; margin-right: auto; width: auto; height: 100px;">Test</div>
		</body>
	</html>`,
	)

	page.Content.Layout(context.TODO(), image.Point{400, 300})

	target := page.getBody().FirstChild.NextSibling

	want := image.Rectangle{image.Point{0, 0}, image.Point{400, 100}}
	if target.BoxContentRectangle != want {
		t.Errorf("Div content was in incorrect place. got %v", target.BoxContentRectangle)
	}

	want = image.Rectangle{image.Point{0, 0}, image.Point{400, 100}}
	if got := target.getAbsoluteDrawRectangle(); got != want {
		t.Errorf("Box was drawn in incorrect location. got %v", got)
	}
}

// Test that when the width is 100%, and both a left/right margin of "auto"
// becomes 0 and the width becomes 100%.
func TestFullWidthLeftRight(t *testing.T) {
	page := parseHTML(
		t,
		`<html>
		<head>
			<style>html, body { padding: 0; margin: 0;}</style>
		</head>
		<body>
			<div style="display: block; margin-left: auto; margin-right: auto; width: 100%; height: 100px;">Test</div>
		</body>
	</html>`,
	)

	page.Content.Layout(context.TODO(), image.Point{400, 300})

	target := page.getBody().FirstChild.NextSibling

	want := image.Rectangle{image.Point{0, 0}, image.Point{400, 100}}
	if target.BoxContentRectangle != want {
		t.Errorf("Div content was in incorrect place. got %v", target.BoxContentRectangle)
	}

	want = image.Rectangle{image.Point{0, 0}, image.Point{400, 100}}
	if got := target.getAbsoluteDrawRectangle(); got != want {
		t.Errorf("Box was drawn in incorrect location. got %v", got)
	}
}

// Duplicates the invalid value for padding test in the CSS1 4.1.2 test
// suite.
func TestInvalidPadding(t *testing.T) {
	// It's not entirely clear why this is a useful test, but it's one
	// of the tests in https://www.w3.org/Style/CSS/Test/CSS1/current/sec412.htm
	// and there's no good reason not to automate it
	page := parseHTML(
		t,
		`<html>
		<head>
			<style>html, body { padding: 0; margin: 0;}</style>
		</head>
		<body>
			<div style="display: block; padding-left: auto; padding-right: auto; margin-left: 0; margin-right: 0; width: 50%; height: 100px;">Since auto is an invalid value for padding, the right-margin of this paragraph should be reset to "auto" and thus be expanded to 50%, and it should only occupy the left half of the viewport</div>
		</body>
	</html>`,
	)

	page.Content.Layout(context.TODO(), image.Point{400, 300})

	target := page.getBody().FirstChild.NextSibling

	want := image.Rectangle{image.Point{0, 0}, image.Point{200, 100}}
	if target.BoxContentRectangle != want {
		t.Errorf("Div content was in incorrect place. got %v", target.BoxContentRectangle)
	}

	want = image.Rectangle{image.Point{0, 0}, image.Point{200, 100}}
	if got := target.getAbsoluteDrawRectangle(); got != want {
		t.Errorf("Box was drawn in incorrect location. got %v", got)
	}
}
