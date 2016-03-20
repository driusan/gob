package main

import (
	"fmt"
	"golang.org/x/net/html"
	"image"
	"image/color"
	"image/draw"
)

type HTMLElement struct {
	*html.Node
	rules       []StyleRule
	TextContent string

	Children []*HTMLElement
}

func (e *HTMLElement) AddStyle(s StyleRule) {
	e.rules = append(e.rules, s)
	return
}

func (e *HTMLElement) SortStyles() error {
	//TODO: Implement this
	// This should sort HTMLElement.rules according to the spec, which states:

	// 1. Find all declarations that apply too element/property (already done when this is called)
	// 2. Sort according to importance (normal or important) and origin (author, user, or user agent). In ascending order of precedence:
	//	1. user agent declarations (defaults)
	//	2. user normal declrations (don't exist)
	//	3. author normal declarations
	//	4. author important declarations
	//	5. user important declarations (don't exist)
	// 3. Sort rules with the same importance and origin by specificity of selector: more specific selectors will override more general ones. Pseudo-elements and pseudo-classes are counted as normal elements and classes, respectively.
	// 4. Finally, sort by order specified: if two declarations have the same weight, origin, and specificity, the latter specified wins. Declarations in imported stylesheets are considered to be before any declaration in the style sheet itself
	return nil
}

func (e HTMLElement) followCascadeToPx(attr string, val int) int {
	// sort according to CSS cascading rules
	e.SortStyles()

	// apply each rule
	for _, rule := range e.rules {
		// the rule has this attribute, so convert it and apply
		// it to the value calculated so far
		if cssval, ok := rule.Values[StyleAttribute(attr)]; ok {
			val = convertUnitToPx(val, cssval)
		}
	}
	return val
}

func (e HTMLElement) followCascadeToColor(attr string) (*color.RGBA, error) {
	var ret *color.RGBA
	// sort according to CSS cascading rules
	e.SortStyles()

	// apply each rule
	for _, rule := range e.rules {
		// the rule has this attribute, so convert it and apply
		// it to the value calculated so far
		if cssval, ok := rule.Values[StyleAttribute(attr)]; ok {
			ret, _ = convertUnitToColor(cssval)

		}
	}
	if ret == nil {
		return &background, NoStyles
	}
	return ret, nil
}
func (e HTMLElement) GetWidthInPx() (int, error) {
	if e.rules == nil {
		return -1, NoStyles
	}
	return e.followCascadeToPx("width", 100), nil
}
func (e HTMLElement) GetHeightInPx() (int, error) {

	if e.rules == nil {
		return -1, NoStyles
	}
	return e.followCascadeToPx("height", 100), nil
}

func (e HTMLElement) GetBackgroundColor() color.RGBA {
	val, _ := e.followCascadeToColor("background")
	return *val
}

func (e HTMLElement) Render(dst *image.RGBA) {
	imageSize := dst.Bounds()

	width, err := e.GetWidthInPx()
	if err == NoStyles {
		fmt.Printf("hello %s\n", imageSize)
		width = imageSize.Max.X
	}
	height, err := e.GetHeightInPx()
	if err == NoStyles {
		width = imageSize.Max.Y
	}
	bg := e.GetBackgroundColor()

	if e.Type == html.ElementNode && e.Data == "body" {
		width = imageSize.Max.X
		height = imageSize.Max.Y
	} else {
	}

	fmt.Printf("W, H: %d, %d\n", width, height)
	b := image.Rectangle{image.Point{0, 0}, image.Point{width, height}}

	draw.Draw(dst, b, &image.Uniform{bg}, image.ZP, draw.Src)

	dot := image.Point{0, 0}
	for _, c := range e.Children {
		childWidth, _ := c.GetWidthInPx()
		childHeight, _ := c.GetHeightInPx()
		childImage := image.NewRGBA(image.Rectangle{image.ZP, image.Point{childWidth, childHeight}})
		c.Render(childImage)

		sr := childImage.Bounds()
		r := image.Rectangle{dot, dot.Add(sr.Size())}
		draw.Draw(dst, r, childImage, sr.Min, draw.Src)
		//r := image.Rectangle{dot, dot.Add(childImage.Size()}
		//draw.Draw(dst, r, childImage, image.Point{50, 5}, draw.Over)
		fmt.Printf("Should render %s at %s cW, cH : %d, %d\n", *c, dot, childWidth, childHeight)

		// for display: inline
		//dot.X += childWidth
		// for display: block
		dot.Y += childHeight
	}
}

func (e *HTMLElement) AppendChild(c *HTMLElement) {
	e.Children = append(e.Children, c)
}
