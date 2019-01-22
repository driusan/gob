package renderer

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"io"
	"io/ioutil"

	"context"

	"net/url"
	"strings"
	"testing"

	"github.com/driusan/gob/net"
)

type testLoader struct {
	net.DefaultReader
}

type colouredRect struct {
	image.Rectangle
	colour color.Color
}

func (r colouredRect) At(x, y int) color.Color {
	if r.Rectangle.At(x, y) == color.Opaque {
		return r.colour
	}
	return color.Transparent
}
func (t testLoader) GetURL(u *url.URL) (io.ReadCloser, int, error) {
	// Fake some URLs used by the test suite.
	switch u.Path {
	case "/100x50.png":
		img := colouredRect{
			image.Rect(0, 0, 100, 50),
			color.RGBA{255, 0, 0, 255},
		}
		b := &bytes.Buffer{}
		if err := png.Encode(b, img); err != nil {
			println(err.Error())
			return nil, 500, err
		}
		reader := bytes.NewReader(b.Bytes())
		return ioutil.NopCloser(reader), 200, nil
	default:
		panic("Unhandled test path: " + u.Path)
	}
}

func parseHTML(t *testing.T, val string) Page {
	t.Helper()
	url, err := url.Parse("https://localhost")
	if err != nil {
		t.Fatal(err)
	}
	loader := testLoader{}
	r := strings.NewReader(val)
	return LoadPage(r, loader, url)
}

// Test that 2 blocks of a known size are placed on top of each other.
func TestBasicLayoutBlock(t *testing.T) {
	page := parseHTML(
		t,
		`<html>
		<body>
			<div style="display: block; width: 100px; height: 50px">Test</div>
			<div style="display: block; width: 100px; height: 50px">Test2</div>
		</body>
	</html>`,
	)

	page.Content.Layout(context.TODO(), image.Point{250, 300})

	nodes := []struct {
		el   *RenderableDomElement
		want image.Rectangle
	}{
		{
			// First child = whitespace text node, second = div 1
			page.Content.FirstChild.NextSibling,
			image.Rectangle{image.ZP, image.Point{100, 50}},
		},
		{
			// Another whitespace before div 2
			page.Content.FirstChild.NextSibling.NextSibling.NextSibling,
			image.Rectangle{image.Point{0, 50}, image.Point{100, 100}},
		},
	}

	for i, el := range nodes {
		if el.el.BoxDrawRectangle != el.want {
			t.Errorf("Test case %d: got %v want %v", i, el.el.BoxDrawRectangle, el.want)
		}
	}
}

// Test that 2 inline-blocks of a known size are placed next to each other.
// (Unlike inlines, we can be sure that inline-blocks are the width and height
// specified.)
func TestBasicLayoutInlineBlock(t *testing.T) {
	page := parseHTML(
		t,
		`<html>
		<body>
			<span style="display: inline-block; width: 100px; height: 50px">Test</span>
			<span style="display: inline-block; width: 100px; height: 50px">Test2</span>
		</body>
	</html>`,
	)

	page.Content.Layout(context.TODO(), image.Point{400, 300})

	nodes := []struct {
		el   *RenderableDomElement
		want image.Rectangle
	}{
		{
			// First child = whitespace text node, second = div 1
			page.Content.FirstChild.NextSibling,
			image.Rectangle{image.ZP, image.Point{100, 50}},
		},
		{
			// Another whitespace before div 2
			page.Content.FirstChild.NextSibling.NextSibling.NextSibling,
			image.Rectangle{image.Point{100, 0}, image.Point{200, 50}},
		},
	}

	for i, el := range nodes {
		if el.el.BoxDrawRectangle != el.want {
			t.Errorf("Test case %d: got %v want %v", i, el.el.BoxDrawRectangle, el.want)
		}
	}
}

// Test that a left float of a known size and a block of a known size are placed
// next to each other.
func TestBasicLeftFloat(t *testing.T) {
	page := parseHTML(
		t,
		`<html>
		<head>
			<style type="text/css">html, body { margin: 0; padding: 0}</style>
		</head>
		<body>
			<div style="display: block; float: left; width: 100px; height: 50px;">Test</div>
			<div style="display: block; width: 100px; height: 50px;">Test2</span>
		</body>
	</html>`,
	)

	page.Content.Layout(context.TODO(), image.Point{400, 300})

	nodes := []struct {
		el   *RenderableDomElement
		want image.Rectangle
	}{
		{
			// First child = whitespace text node, second = div 1
			page.Content.FirstChild.NextSibling,
			image.Rectangle{image.ZP, image.Point{100, 50}},
		},
		{
			// Another whitespace before div 2
			page.Content.FirstChild.NextSibling.NextSibling.NextSibling,
			image.Rectangle{image.Point{100, 0}, image.Point{200, 50}},
		},
	}

	for i, el := range nodes {
		if el.el.BoxDrawRectangle != el.want {
			t.Errorf("Test case %d: got %v want %v", i, el.el.BoxDrawRectangle, el.want)
		}
	}
}

// Test that a right float of a known size and a block of a known size are placed
// next to each other.
func TestBasicRightFloat(t *testing.T) {
	page := parseHTML(
		t,
		`<html>
		<head>
			<style type="text/css">html, body { margin: 0; padding: 0}</style>
		</head>
		<body>
			<div style="display: block; float: right; width: 100px; height: 50px;">Test</div>
			<div style="display: block; width: 100px; height: 50px;">Test2</span>
		</body>
	</html>`,
	)

	page.Content.Layout(context.TODO(), image.Point{400, 300})

	nodes := []struct {
		el   *RenderableDomElement
		want image.Rectangle
	}{
		{
			// Draw rectangle for the float should be before the
			// right edge.
			page.Content.FirstChild.NextSibling,
			image.Rectangle{image.Point{400 - 100, 0}, image.Point{400, 50}},
		},
		{
			// Draw rectangle for the non-float should be at the
			// left edge.
			page.Content.FirstChild.NextSibling.NextSibling.NextSibling,
			image.Rectangle{image.Point{0, 0}, image.Point{100, 50}},
		},
	}

	for i, el := range nodes {
		if el.el.BoxDrawRectangle != el.want {
			t.Errorf("Test case %d: got %v want %v", i, el.el.BoxDrawRectangle, el.want)
		}
	}
}

// Test that 2 left floats are positioned next to each other.
func TestBasicLeftFloat2(t *testing.T) {
	page := parseHTML(
		t,
		`<html>
		<head>
			<style type="text/css">html, body { margin: 0; padding: 0}</style>
		</head>
		<body>
			<div style="display: block; float: left; width: 100px; height: 50px;">Test</div>
			<div style="display: block; float: left; width: 100px; height: 50px;">Test2</span>
		</body>
	</html>`,
	)

	page.Content.Layout(context.TODO(), image.Point{400, 300})

	nodes := []struct {
		el   *RenderableDomElement
		want image.Rectangle
	}{
		{
			page.Content.FirstChild.NextSibling,
			image.Rectangle{image.Point{0, 0}, image.Point{100, 50}},
		},
		{
			// Draw rectangle for the second float should touch the
			// first.
			page.Content.FirstChild.NextSibling.NextSibling.NextSibling,
			image.Rectangle{image.Point{100, 0}, image.Point{200, 50}},
		},
	}

	for i, el := range nodes {
		if el.el.BoxDrawRectangle != el.want {
			t.Errorf("Test case %d: got %v want %v", i, el.el.BoxDrawRectangle, el.want)
		}
	}
}

// Test that 2 right floats are positioned next to each other.
func TestBasicRightFloat2(t *testing.T) {
	page := parseHTML(
		t,
		`<html>
		<head>
			<style type="text/css">html, body { margin: 0; padding: 0}</style>
		</head>
		<body>
			<div style="display: block; float: right; width: 100px; height: 50px;">Test</div>
			<div style="display: block; float: right; width: 100px; height: 50px;">Test2</span>
		</body>
	</html>`,
	)

	page.Content.Layout(context.TODO(), image.Point{400, 300})

	nodes := []struct {
		el   *RenderableDomElement
		want image.Rectangle
	}{
		{
			// Draw rectangle for the float should be before the
			// right edge.
			page.Content.FirstChild.NextSibling,
			image.Rectangle{image.Point{400 - 100, 0}, image.Point{400, 50}},
		},
		{
			// Draw rectangle for the second float should touch the
			// first.
			page.Content.FirstChild.NextSibling.NextSibling.NextSibling,
			image.Rectangle{image.Point{400 - 100 - 100, 0}, image.Point{400 - 100, 50}},
		},
	}

	for i, el := range nodes {
		if el.el.BoxDrawRectangle != el.want {
			t.Errorf("Test case %d: got %v want %v", i, el.el.BoxDrawRectangle, el.want)
		}
	}
}

// Test that 2 small inlines are placed on the same line.
func TestBasicLayoutInline(t *testing.T) {
	// Note that the width and height property is meaningless for inlines
	// so we can't verify this by explicitly setting them. We need to
	// calculate a baseline for the end of a known piece of text first.
	//
	// Also note that the ContentBoxRectangle might span multiple lines,
	// so the Min is always 0. We need to check the linebox itself to verify
	// where it's drawn, not the rectangle.
	page := parseHTML(
		t,
		`<html>
		<body>
			<span style="display: inline">Test</span><span style="display: inline">Test2</span>
		</body>
	</html>`,
	)

	page.Content.Layout(context.TODO(), image.Point{400, 300})

	span1 := page.Content.FirstChild.NextSibling
	span2 := span1.NextSibling

	// Do some sanity checks.
	if span1.BoxContentRectangle.Max.X <= span2.BoxContentRectangle.Min.X {
		t.Errorf("Span 2 was not placed after span 1")
	}
	if span1.BoxContentRectangle.Min.X != 0 {
		t.Errorf("Span 1 was not placed on the first line")
	}
	if span2.BoxContentRectangle.Min.X != 0 {
		t.Errorf("Span 2 was not placed on the first line")
	}
	if content := span1.GetTextContent(); content != "Test" {
		t.Errorf("Unexpected content for first span: got %v want 'Test'", content)
	}
	if content := span2.GetTextContent(); content != "Test2" {
		t.Errorf("Unexpected content for second span: got %v want 'Test2'", content)
	}

	// The body is the nearest block, it should have the line boxes.
	if n := len(page.Content.lineBoxes); n != 2 {
		t.Errorf("Unexpected number of line boxes for body: got %v want 2", n)
	}

	testLineBox := page.Content.lineBoxes[0]
	test2LineBox := page.Content.lineBoxes[1]

	if testLineBox.content != "Test" {
		t.Fatalf("Unexpected content for first line box: got %v want 'Test'", testLineBox.content)
	}
	if test2LineBox.content != "Test2" {
		t.Fatalf("Unexpected content for first line box: got %v want 'Test2'", test2LineBox.content)
	}

	tlbW := testLineBox.width()
	if tlbW <= 0 {
		t.Fatalf("Negative or zero width for 'Test': got %v", tlbW)
	}

	if x := page.Content.lineBoxes[1].origin.X; x != tlbW {
		t.Errorf("Unexpected X origin for Test2. got %v want %v", x, tlbW)
	}

	// Second part of the test: Do the same thing for an inline embedded in another inline.
	// This shouldn't affect the layout positioning since there's nothing changing
	// the styling (size/line height/margin/etc).
	page = parseHTML(
		t,
		`<html>
		<body>
			<span style="display: inline">Test<span style="display: inline">Test2</span></span>
		</body>
	</html>`,
	)
	page.Content.InvalidateLayout()
	page.Content.Layout(context.TODO(), image.Point{400, 300})

	span1 = page.Content.FirstChild.NextSibling
	span2 = span1.FirstChild.NextSibling

	// Do some sanity checks.
	if span1.BoxContentRectangle.Max.X <= span2.BoxContentRectangle.Min.X {
		t.Errorf("Embedded inline: Span 2 was not placed after span 1")
	}
	if span1.BoxContentRectangle.Min.X != 0 {
		t.Errorf("Embedded inline: Span 1 was not placed on the first line")
	}
	if span2.BoxContentRectangle.Min.X != 0 {
		t.Errorf("Embedded inline: Span 2 was not placed on the first line")
	}

	// Since we're checking the textContent attribute it includes the children
	if content := span1.GetTextContent(); content != "TestTest2" {
		t.Errorf("Embedded inline: Unexpected content for first span: got %v want 'Test'", content)
	}
	if content := span2.GetTextContent(); content != "Test2" {
		t.Errorf("Embedded inline: Unexpected content for second span: got %v want 'Test2'", content)
	}

	// The body is the nearest block, it should have the line boxes.
	if n := len(page.Content.lineBoxes); n != 2 {
		t.Errorf("Embedded inline: Unexpected number of line boxes for body: got %v want 2", n)
	}

	testLineBox = page.Content.lineBoxes[0]
	test2LineBox = page.Content.lineBoxes[1]

	if testLineBox.content != "Test" {
		t.Fatalf("Embedded inline: Unexpected content for first line box: got %v want 'Test'", testLineBox.content)
	}
	if test2LineBox.content != "Test2" {
		t.Fatalf("Embedded inline: Unexpected content for first line box: got %v want 'Test2'", test2LineBox.content)
	}

	// we re-use the test linebox width from the first part of the test, to
	// ensure that whether it's embedded or a sibling doesn't affect its
	// positioning.
	if x := page.Content.lineBoxes[1].origin.X; x != tlbW {
		t.Errorf("Unexpected X origin for Test2. got %v want %v", x, tlbW)
	}

}

// Test that an inline that doesn't fit in the width is split across multiple
// lines into 2 line boxes.
func TestBasicLayoutMultilineInline(t *testing.T) {
	// Like the basic inline test, we need to establish a baseline first.
	// We lay out a single really long word into a large viewport to
	// see how many pixels it is. We then double it, add a few pixels
	// for a space character, and ensure that there's 2 lineboxes (the
	// first having the word twice, and the second once.
	page := parseHTML(
		t,
		`<html>
		<body>
			<span style="display: inline">Supergalifragilisticexpialidotious</span>
		</body>
	</html>`,
	)

	page.Content.Layout(context.TODO(), image.Point{40000, 300})
	if n := len(page.Content.lineBoxes); n != 1 {
		t.Fatalf("Page did not have any line boxes: got %v want 1", n)
	}
	superwidth := page.Content.lineBoxes[0].width()
	if superwidth <= 0 {
		t.Fatal("Linebox did not have width")
	}

	page.Content.InvalidateLayout()

	page = parseHTML(
		t,
		`<html>
		<body>
			<span style="display: inline">Supercalifragilisticexpialidotious
Supercalifragilisticexpialidotious Supercalifragilisticexpialidotious
	</span>
		</body>
	</html>`,
	)

	// Now do the real test.
	page.Content.Layout(context.TODO(), image.Point{superwidth*2 + 20, 300})
	if n := len(page.Content.lineBoxes); n != 2 {
		t.Fatalf("Body had incorrect number of line boxes: got %v want 2", n)
	}

	// The default (normal) whitespace value for HTML means that the first
	// 2 should be on the first line even though there's a new line in the
	// source, and the last should be on the second line since we sized
	// the window to ensure it didn't fit.
	super1 := page.Content.lineBoxes[0]
	if super1.content != "Supercalifragilisticexpialidotious Supercalifragilisticexpialidotious" {
		t.Errorf("Unexpected content for first line box: got '%v' want 'Supercalifragilisticexpialidotious Supercalifragilisticexpialidotious'", super1.content)
	}
	super2 := page.Content.lineBoxes[1]
	if super2.content != "Supercalifragilisticexpialidotious" {
		t.Errorf("Unexpected content for first line box: got '%v' want 'Supercalifragilisticexpialidotious'", super2.content)
	}

	lineheight := super1.el.GetLineHeight()
	if lineheight <= 0 {
		t.Fatalf("Element did not have valid lineheight: got %v", lineheight)
	}

	if super1.origin != image.ZP {
		t.Errorf("Unexpected origin for first line: got %v, want (0, 0)", super1.origin)
	}
	if super2.origin.X != 0 {
		t.Errorf("Second line did not start at X=0, got %v", super2.origin.X)
	}
	if super2.origin.Y != super1.origin.Y+lineheight {
		t.Errorf("Second line did not start %v (lineheight) pixels after first line. First line start: %v. Second line start: %v", lineheight, super1.origin, super2.origin)
	}

}

// Test that a selector applying to the first line which changes the line
// height advances the correct amount.
func TestLayoutFirstline(t *testing.T) {
	// This test makes similar assumptions to the multiline one. It
	// creates a baseline, and then does a layout with a viewport which
	// it sized based on that baseline.
	page := parseHTML(
		t,
		`<html>
		<head>
		<style>
		span {
			display: inline;
			line-height: 30px;
			margin: 0;
			padding: 0;
		}
		</style>
		</head>
		<body>
			<span>TestTestTest</span>
		</body>
	</html>`,
	)

	page.Content.Layout(context.TODO(), image.Point{40000, 300})
	if n := len(page.Content.lineBoxes); n != 1 {
		t.Fatalf("Page did not have any line boxes: got %v want 1", n)
	}
	testwidth := page.Content.lineBoxes[0].width()
	if testwidth <= 0 {
		t.Fatal("Linebox did not have width")
	}

	// Now perform the actual test. We make sure there's at least 3 lines
	// so that we can ensure the line-height property from first-line only
	// got applied to the first line.
	page = parseHTML(
		t,
		`<html>
		<head>
		<style>
		span {
			display: inline;
			line-height: 30px;
			margin: 0;
			padding: 0;
		}
		span:first-line {
			line-height: 40px;
			margin: 0;
			padding: 0;
		}
		</style>
		</head>
		<body>
			<span>TestTestTest TestTestTest
TestTestTest TestTestTest
TestTestTest TestTestTest
</span>
		</body>
	</html>`,
	)

	page.Content.InvalidateLayout()
	page.Content.Layout(context.TODO(), image.Point{testwidth*2 + 40, 300})
	if n := len(page.Content.lineBoxes); n != 3 {
		t.Fatalf("Body had incorrect number of line boxes: got %v want 3", n)
	}

	for i, line := range page.Content.lineBoxes {
		// FIXME: This test probably shouldn't be required to trim the
		// space.
		if strings.TrimSpace(line.content) != "TestTestTest TestTestTest" {
			t.Fatalf("Unexpected content for line %v: got %v", i, line.content)
		}
		if line.origin.X != 0 {
			t.Errorf("Unexpected X origin for line %v: got %v, want 0", i, line.origin.X)
		}
	}

	if got := page.Content.lineBoxes[0].origin.Y; got != 0 {
		t.Errorf("Unexpected Y origin for line 1: got %v, want 0", got)
	}

	// The firstline property should have set made the line advance by 40px
	// to get to the second line.
	if got := page.Content.lineBoxes[1].origin.Y; got != 40 {
		t.Errorf("Unexpected Y origin for line 2: got %v, want 40", got)
	}

	// The last line should have only advanced by 30 (the default for span),
	// not 40 (which should have only applied to the first line.)
	if got := page.Content.lineBoxes[2].origin.Y; got != 40+30 {
		t.Errorf("Unexpected Y origin for line 3: got %v, want 70", got)
	}

}
