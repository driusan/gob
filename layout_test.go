package main

// This file contains various tests for basic layout tasks.
// It should probably be in the renderer package, but that would
// result in a cyclic import between parser and renderer, so for now
// it lives at the top level.
import (
	"context"
	"image"
	"net/url"
	"strings"
	"testing"

	"github.com/driusan/gob/net"
	"github.com/driusan/gob/parser"
	"github.com/driusan/gob/renderer"
)

//
func parseHTML(t *testing.T, val string) parser.Page {
	t.Helper()
	url, err := url.Parse("https://localhost")
	if err != nil {
		t.Fatal(err)
	}
	loader := net.DefaultReader{}
	r := strings.NewReader(val)
	page := parser.LoadPage(r, loader, url, nil)
	return page
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
		el   *renderer.RenderableDomElement
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

// Test that 2 inlines of a known size are placed next to each other.
func TestBasicLayoutInline(t *testing.T) {
	page := parseHTML(
		t,
		`<html>
		<body>
			<span style="display: inline; width: 100px; height: 50px">Test</span>
			<span style="display: inline; width: 100px; height: 50px">Test2</span>
		</body>
	</html>`,
	)

	page.Content.Layout(context.TODO(), image.Point{400, 300})

	nodes := []struct {
		el   *renderer.RenderableDomElement
		want image.Rectangle
	}{
		{
			// First child = whitespace text node, second = span 1
			page.Content.FirstChild.NextSibling,
			image.Rectangle{image.ZP, image.Point{100, 50}},
		},
		{
			// Another whitespace before span 2
			page.Content.FirstChild.NextSibling.NextSibling.NextSibling,
			image.Rectangle{image.Point{100, 50}, image.Point{200, 50}},
		},
	}

	for i, el := range nodes {
		if el.el.BoxContentRectangle != el.want {
			t.Errorf("Test case %d: got %v want %v", i, el.el.BoxDrawRectangle, el.want)
		}
	}
}

