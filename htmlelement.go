package main

import (
	//	"fmt"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
	"golang.org/x/net/html"
	"image"
	"image/color"
	"image/draw"
)

// A RenderableElement is something that can be rendered to
// an image.
type RenderableElement interface {
	// Draw the element onto dst
	Render(dst *image.RGBA)

	// The final width of the element being rendered, including
	// all borders, margins and padding
	GetWidthInPx() (int, error)

	// The final height of the element being rendered, including
	// all borders, margins and padding
	GetHeightInPx() (int, error)

	GetDisplayProp() string

	GetTextContent() string
}

// A StyleElement is anything that has CSS rules applied to it. It
// can be composed into other structs
type StyledElement struct {
	// The rules that match this element.
	rules []StyleRule
}

// A TextElement is a renderable TextNode from an HTML document.
type TextElement struct {
	StyledElement
	TextContent string
}

// Rendering a TextElement just draws the string onto its parent.
func (e TextElement) Render(dst *image.RGBA) {
	fntDrawer := font.Drawer{
		dst,
		&image.Uniform{e.GetColor()},
		basicfont.Face7x13,
		fixed.P(0, 10),
	}
	//fmt.Printf("Writing: %s (%s)\n", e.TextContent, fntDrawer.MeasureString(e.TextContent))
	fntDrawer.DrawString(e.TextContent)
}
func (e TextElement) GetWidthInPx() (int, error) {
	fntDrawer := font.Drawer{
		nil,
		nil,
		basicfont.Face7x13,
		fixed.P(0, 10),
	}
	return int(fntDrawer.MeasureString(e.TextContent+" ") >> 6), nil
}
func (e TextElement) GetHeightInPx() (int, error) {
	return 15, nil
}
func (e TextElement) GetDisplayProp() string {
	return "inline"
}

func (e TextElement) GetTextContent() string {
	return e.TextContent
}

type HTMLElement struct {
	*html.Node
	StyledElement
	TextContent string

	Children []RenderableElement
}

func (e HTMLElement) GetTextContent() string {
	return e.TextContent
}
func (e *StyledElement) AddStyle(s StyleRule) {
	e.rules = append(e.rules, s)
	return
}

// SortStyles will sort the rules on this element according to the CSS spec, which state:s

// 1. Find all declarations that apply too element/property (already done when this is called)
// 2. Sort according to importance (normal or important) and origin (author, user, or user agent). In ascending order of precedence:
//	1. user agent declarations (defaults)
//	2. user normal declrations (don't exist)
//	3. author normal declarations
//	4. author important declarations
//	5. user important declarations (don't exist)
// 3. Sort rules with the same importance and origin by specificity of selector: more specific selectors will override more general ones. Pseudo-elements and pseudo-classes are counted as normal elements and classes, respectively.
// 4. Finally, sort by order specified: if two declarations have the same weight, origin, and specificity, the latter specified wins. Declarations in imported stylesheets are considered to be before any declaration in the style sheet itself
// BUG(driusan): SortStyles is not implemented
func (e *StyledElement) SortStyles() error {
	return nil
}

func (e StyledElement) followCascadeToPx(attr string, val int) int {
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

func (e StyledElement) followCascadeToColor(attr string) (*color.RGBA, error) {
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
func (e StyledElement) GetBackgroundColor() color.RGBA {
	val, err := e.followCascadeToColor("background")
	if err == NoStyles {
	}

	return *val
}
func (e StyledElement) GetColor() color.RGBA {
	val, err := e.followCascadeToColor("background")
	if err == NoStyles {
		return color.RGBA{0, 0, 0, 255}
	}
	return *val
}

func (e HTMLElement) GetWidthInPx() (int, error) {
	var calcWidth int
	for _, child := range e.Children {
		cW, _ := child.GetWidthInPx()
		if calcWidth < cW {
			calcWidth += cW
		}
	}
	if calcWidth > 0 {
		return calcWidth, nil
	}
	if e.rules == nil {
		return -1, NoStyles
	}
	return e.followCascadeToPx("width", 1000), NoStyles
}
func (e HTMLElement) GetHeightInPx() (int, error) {
	explicitHeight := e.followCascadeToPx("height", -1)
	if explicitHeight != -1 {
		return explicitHeight, nil
	}

	var calcHeight int
	for _, child := range e.Children {
		cH, _ := child.GetHeightInPx()
		calcHeight += cH
	}
	if calcHeight > 0 {
		//fmt.Printf("Calculated height of %d for %s\n", calcHeight, e)
		return calcHeight, nil
	}

	if e.rules == nil {
		return -1, NoStyles
	}
	return -2, NoStyles
	//panic("Could not calculate height of element")
}

func (e HTMLElement) ContainsBlocks() bool {
	// the CSS spec says that an element either only contains
	// blocks, or only contains inline elements.
	// If there's both inline and block children, the inline
	// ones need to implicitly have a block around them.
	for _, c := range e.Children {
		if c.GetDisplayProp() == "block" {
			return true
		}
	}
	return false
}
func (e HTMLElement) GetDisplayProp() string {
	switch e.Data {
	case "span", "a":
		return "inline"
	case "p", "div":
		fallthrough
	default:
		return "block"
	}

}

func (e HTMLElement) Render(dst *image.RGBA) {
	imageSize := dst.Bounds()

	height, err := e.GetHeightInPx()
	if err == NoStyles {
		height = imageSize.Max.Y
	}
	bg := e.GetBackgroundColor()

	width, err := e.GetWidthInPx()
	if err == NoStyles {
		width = imageSize.Max.X
	}
	if e.Type == html.ElementNode && e.Data == "body" {
		width = imageSize.Max.X
		height = imageSize.Max.Y
	}

	// Draw the background
	b := image.Rectangle{image.Point{0, 0}, image.Point{width, height}}
	draw.Draw(dst, b, &image.Uniform{bg}, image.ZP, draw.Src)

	dot := image.Point{0, 0}
	for _, c := range e.Children {

		childWidth, _ := c.GetWidthInPx()
		childHeight, _ := c.GetHeightInPx()
		childImage := image.NewRGBA(image.Rectangle{image.ZP, image.Point{width, height}})
		c.Render(childImage)

		sr := childImage.Bounds()

		r := image.Rectangle{dot, dot.Add(sr.Size())}
		draw.Draw(dst, r, childImage, sr.Min, draw.Over)
		switch c.GetDisplayProp() {
		case "inline":
			if dot.X+childWidth < imageSize.Max.X {
				dot.X += childWidth
			} else {
				dot.X = 0
				dot.Y += childHeight
			}
		case "block":
			fallthrough
		default:
			dot.X = 0
			dot.Y += childHeight
		}

	}
}
