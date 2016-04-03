package main

import (
	"fmt"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
	"unicode"
	"unicode/utf8"
	//	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/net/html"
	//"os"
	"image"
	"image/color"
	"image/draw"
	"io/ioutil"
	"strings"
	"Gob/css"
)

const (
	DefaultFontSize = 16
)

// A RenderableElement is something that can be rendered to
// an image.
type RenderableElement interface {
	// Returns an image representing this element.
	Render(containerWidth int) *image.RGBA

	// The final width of the element being rendered, including
	// all borders, margins and padding
	GetWidthInPx(parentWidth int) (int, error)

	// The final height of the element being rendered, including
	// all borders, margins and padding
	GetHeightInPx(parentWidth int) (int, error)

	GetDisplayProp() string

	GetFontFace(int) font.Face
	GetFontSize() int
	SetFontSize(int)
	GetTextContent() string
	GetBackgroundColor() color.RGBA
}

// A StyleElement is anything that has CSS rules applied to it. It
// can be composed into other structs
type StyledElement struct {
	// The rules that match this element.
	rules []StyleRule

	fontSize int
}

// A TextElement is a renderable TextNode from an HTML document.
type TextElement struct {
	*StyledElement
	TextContent string
}

// Rendering a TextElement just draws the string onto its parent.
func (e TextElement) Render(containerWidth int) *image.RGBA {
	return nil
}
func (e TextElement) GetWidthInPx(parentWidth int) (int, error) {
	fntDrawer := font.Drawer{
		nil,
		nil,
		basicfont.Face7x13,
		fixed.P(0, 10),
	}
	return int(fntDrawer.MeasureString(e.TextContent) >> 6), nil
}
func (e TextElement) GetHeightInPx(parentWidth int) (int, error) {
	if parentWidth == 0 {
		return 0, NoStyles
	}
	if width, err := e.GetWidthInPx(parentWidth); width > parentWidth && err == nil {
		// crude approximation of how many lines long this is,
		// times the font size
		return e.fontSize * (width / parentWidth), nil
	}
	return e.fontSize, nil
}
func (e TextElement) GetDisplayProp() string {
	return "inline"
}

func (e TextElement) GetTextContent() string {
	return e.TextContent
}

func (e *StyledElement) SetFontSize(size int) {
	e.fontSize = size
}
func (e StyledElement) GetFontFace(fsize int) font.Face {
	fontBytes, _ := ioutil.ReadFile("/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf")
	fnt, _ := truetype.Parse(fontBytes)
	return truetype.NewFace(fnt,
		&truetype.Options{
			Size:    float64(fsize),
			DPI:     72,
			Hinting: font.HintingFull})

}
func (e StyledElement) GetFontSize() int {
	if e.fontSize == 0 {
		return DefaultFontSize
	}
	return e.fontSize
}

type HTMLElement struct {
	*html.Node
	StyledElement
	TextContent string

	Children []RenderableElement
}

func (e HTMLElement) GetFontSize() int {
	if e.Type == html.ElementNode {
		switch e.Data {
		case "h1":
			return DefaultFontSize * 2
		}
	}
	return DefaultFontSize
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
			val, _ = css.ConvertUnitToPx(val, cssval)
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
			ret, _ = css.ConvertColorToRGBA(cssval)

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
	val, err := e.followCascadeToColor("color")
	if err == NoStyles {
		return color.RGBA{0, 0, 0, 255}
	}
	return *val
}

func (e HTMLElement) GetWidthInPx(parentWidth int) (int, error) {
	var calcWidth int
	for _, child := range e.Children {
		cW, _ := child.GetWidthInPx(parentWidth)
		if calcWidth < cW {
			calcWidth = cW
		}
	}
	if calcWidth > 0 {
		return calcWidth, nil
	}
	if e.rules == nil {
		return parentWidth, NoStyles
	}
	return parentWidth, NoStyles
}
func (e HTMLElement) GetHeightInPx(parentWidth int) (int, error) {
	explicitHeight := e.followCascadeToPx("height", -1)
	if explicitHeight != -1 {
		return explicitHeight, nil
	}

	var calcHeight int
	for _, child := range e.Children {
		// Cascade the font size down to the children before
		// calculating the height

		cH, _ := child.GetHeightInPx(parentWidth)
		if cH < e.GetFontSize() {
			calcHeight += e.GetFontSize()
		} else {
			calcHeight += cH
		}
	}
	if calcHeight > 0 {
		return calcHeight, nil
	}

	if e.rules == nil {
		return -1, NoStyles
	}
	return -2, NoStyles
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
	case "p", "div", "h1":
		fallthrough
	default:
		return "block"
	}

}

func (e HTMLElement) Render(containerWidth int) *image.RGBA {

	height, _ := e.GetHeightInPx(containerWidth)
	width, _ := e.GetWidthInPx(containerWidth)
	width = containerWidth
	bg := e.GetBackgroundColor()
	dst := image.NewRGBA(image.Rectangle{image.ZP, image.Point{width, height}})
	imageSize := dst.Bounds()

	if e.Type == html.ElementNode && e.Data == "body" {
		if height < imageSize.Max.Y {
			height = imageSize.Max.Y
		} else {
			fmt.Printf("Should be able to scroll")
		}
		b := image.Rectangle{image.Point{0, 0}, image.Point{width, height}}
		draw.Draw(dst, b, &image.Uniform{bg}, image.ZP, draw.Src)
	}

	//fmt.Printf("width, height for %s: %d, %d\n", e.Data, width, height)

	dot := image.Point{0, 0}
	fSize := e.GetFontSize()
	fontFace := e.GetFontFace(fSize)
	fmt.Printf("Font metrics: %s\n", fontFace.Metrics())
	fntDrawer := font.Drawer{
		Dst:  dst,
		Src:  &image.Uniform{e.GetColor()},
		Face: fontFace,
		//basicfont.Face7x13,
		Dot: fixed.P(dot.X, int(fontFace.Metrics().Ascent)>>6),
	}
	//containsBlocks := e.ContainsBlocks()
	for _, c := range e.Children {
		c.SetFontSize(fSize)

		switch c.GetDisplayProp() {
		case "inline":
			// Draw the background
			//bgChild := c.GetBackgroundColor()

			// draw the content
			textContent := c.GetTextContent()
			words := strings.Fields(textContent)
			firstRune, _ := utf8.DecodeRuneInString(textContent)
			if unicode.IsSpace(firstRune) {
				dot.X += (fSize / 3)
				fntDrawer.Dot = fixed.P(dot.X, dot.Y+int(fontFace.Metrics().Ascent)>>6)
			}
			for _, word := range words {
				//fntDrawer.Dot = fixed.P(dot.X, dot.Y+int(fontFace.Metrics().Ascent) >> 6 )
				wordSizeInPx := int(fntDrawer.MeasureString(word) >> 6)
				if dot.X+wordSizeInPx > width {
					dot.X = 0
					dot.Y += fSize
					fntDrawer.Dot = fixed.P(dot.X, dot.Y+int(fontFace.Metrics().Ascent)>>6)
				}
				fntDrawer.DrawString(word)

				// If it's the last word in this inline chunk,
				// only add a space after if the textcontent
				// ends with space. This is so that space
				// doesn't get added in things like <span>foo</span>bar
				/*
					This doesn't work, because the string is trimmed before it gets here.
					if widx == len(words)-1 {
						lastRune, _ := utf8.DecodeLastRuneInString(textContent)
						if !unicode.IsSpace(lastRune) {
							fmt.Printf("No space after %s", word)
							continue
						}
					}*/

				// Add a three per em space between words, an em-quad after a period,
				// and an en-quad after other punctuation
				switch word[len(word)-1] {
				case ',', ';', ':', '!', '?':
					dot.X = (int(fntDrawer.Dot.X) >> 6) + (fSize / 2)
				case '.':
					dot.X = (int(fntDrawer.Dot.X) >> 6) + fSize
				default:
					dot.X = (int(fntDrawer.Dot.X) >> 6) + (fSize / 3)
				}
				fntDrawer.Dot = fixed.P(dot.X, dot.Y+int(fontFace.Metrics().Ascent)>>6)
			}
		case "block":
			fallthrough
		default:
			// Draw the block itself, and move dot.
			childHeight, _ := c.GetHeightInPx(containerWidth)
			childImage := image.NewRGBA(image.Rectangle{image.ZP, image.Point{width, height}})
			childImage = c.Render(width)

			sr := childImage.Bounds()

			r := image.Rectangle{dot, dot.Add(sr.Size())}
			draw.Draw(dst, r, childImage, sr.Min, draw.Over)
			dot.X = 0
			dot.Y += childHeight
			fntDrawer.Dot = fixed.P(dot.X, dot.Y+int(fontFace.Metrics().Ascent)>>6)
		}

	}
	return dst
}
