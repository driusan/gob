package renderer

import (
	"Gob/css"
	"Gob/dom"
	"fmt"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
	"golang.org/x/net/html"
	"image"
	"image/color"
	"image/draw"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	DefaultFontSize = 16
)

// A RenderableElement is something that can be rendered to
// an image.
type Renderer interface {
	// Returns an image representing this element.
	Render(containerWidth int) *image.RGBA

	/*
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
	*/
}

type RenderableDomElement struct {
	*dom.Element
	Styles *css.StyledElement

	FirstChild  *RenderableDomElement
	NextSibling *RenderableDomElement
}

func stringSize(fntDrawer font.Drawer, textContent string) (int, error) {
	var size int
	words := strings.Fields(textContent)
	fSize := int(fntDrawer.Face.Metrics().Height) >> 6
	firstRune, _ := utf8.DecodeRuneInString(textContent)

	if unicode.IsSpace(firstRune) {
		size = fSize / 3
	}
	for _, word := range words {
		wordSizeInPx := int(fntDrawer.MeasureString(word)) >> 6
		size += wordSizeInPx

		// Add a three per em space between words, an em-quad after a period,
		// and an en-quad after other punctuation
		switch word[len(word)-1] {
		case ',', ';', ':', '!', '?':
			size += (int(fntDrawer.Dot.X) >> 6) + (fSize / 2)
		case '.':
			size += (int(fntDrawer.Dot.X) >> 6) + fSize
		default:
			size += (int(fntDrawer.Dot.X) >> 6) + (fSize / 3)
		}
	}
	return size, nil
}

func (e *RenderableDomElement) Walk(callback func(*RenderableDomElement)) {
	if e == nil {
		return
	}

	if e.Type == html.ElementNode {
		callback(e)
	}

	for c := e.FirstChild; c != nil; c = c.NextSibling {
		switch c.Type {
		case html.ElementNode:
			callback(c)
		}
	}
}
func (e RenderableDomElement) GetHeightInPx(containerWidth int) (int, error) {
	explicitHeight := e.Styles.FollowCascadeToPx("height", -1)
	if explicitHeight != -1 {
		return explicitHeight, nil
	}

	var calcHeight int
	for c := e.FirstChild; c != nil; c = c.NextSibling {
		// Cascade the font size down to the children before
		// calculating the height

		cH, _ := c.GetHeightInPx(containerWidth)
		if cH < e.Styles.GetFontSize() {
			calcHeight += e.Styles.GetFontSize()
		} else {
			calcHeight += cH
		}
	}
	if calcHeight > 0 {
		return calcHeight, nil
	}

	if e.Styles == nil {
		return -1, css.NoStyles
	}
	return e.Styles.GetFontSize(), css.NoStyles
}
func (e RenderableDomElement) GetWidthInPx(containerWidth int) (int, error) {
	var calcWidth int
	if e.Type == html.TextNode {
		fSize := e.Styles.GetFontSize()
		fontFace := e.Styles.GetFontFace(fSize)
		fntDrawer := font.Drawer{
			Dst:  nil,
			Src:  &image.Uniform{e.GetColor()},
			Face: fontFace,
		}
		return stringSize(fntDrawer, e.Data)
		sizeInPx := int(fntDrawer.MeasureString(e.Data) >> 6)
		return sizeInPx, nil
	}
	if e.GetDisplayProp() == "block" {
		return containerWidth, nil
	}
	for child := e.FirstChild; child != nil; child = child.NextSibling {
		cW, _ := child.GetWidthInPx(containerWidth)
		if calcWidth < cW {
			calcWidth = cW
		}
	}
	if calcWidth > 0 {
		return calcWidth, nil
	}
	return containerWidth, nil
}

func (e RenderableDomElement) GetBackgroundColor() color.Color {
	background := color.RGBA{0xE0, 0xE0, 0xE0, 0xFF}
	return background
}
func (e RenderableDomElement) GetColor() color.Color {
	var deflt *color.RGBA
	if e.Type == html.ElementNode && e.Data == "a" {
		fmt.Printf("Default colour for link")
		deflt = &color.RGBA{0, 0, 0xFF, 0xFF}
	} else {
		deflt = &color.RGBA{0, 0, 0, 0xFF}
	}
	cssColor := e.Styles.GetColor(deflt)
	return cssColor
}

func (e RenderableDomElement) GetDisplayProp() string {
	if e.Type == html.TextNode {
		return "inline"
	}
	switch e.Data {
	case "span", "a":
		return "inline"
	case "p", "div", "h1":
		return "block"
	}
	return "block"
}
func (e RenderableDomElement) Render(containerWidth int) *image.RGBA {
	// font size is inherited, so if it's an h1 propagate it down. This is a hack until
	// the CSS package properly implements GetFontSize.
	dot := image.Point{0, 0}
	fSize := e.Styles.GetFontSize()
	fontFace := e.Styles.GetFontFace(fSize)
	fntDrawer := font.Drawer{
		Dst:  nil,
		Src:  &image.Uniform{e.GetColor()},
		Face: fontFace,
		Dot:  fixed.P(dot.X, int(fontFace.Metrics().Ascent)>>6),
	}
	if e.Element.Type == html.ElementNode && e.Element.Data == "h1" {
		e.Styles.SetFontSize(DefaultFontSize * 2)
		for c := e.FirstChild; c != nil; c = c.NextSibling {

			c.Styles.SetFontSize(DefaultFontSize * 2)
			sz, _ := stringSize(fntDrawer, c.Element.Data)
			fmt.Printf("Data: %s Size: %d\n", c.Element.Data, sz)
		}
	}

	height, _ := e.GetHeightInPx(containerWidth)
	if height < 0 {
		height = 0
	}
	width, _ := e.GetWidthInPx(containerWidth)
	bg := e.GetBackgroundColor()
	dst := image.NewRGBA(image.Rectangle{image.ZP, image.Point{width, height}})
	fntDrawer.Dst = dst
	imageSize := dst.Bounds()

	if e.Element.Type == html.ElementNode && e.Element.Data == "body" {
		if height < imageSize.Max.Y {
			height = imageSize.Max.Y
		}
		b := image.Rectangle{image.Point{0, 0}, image.Point{width, height}}
		draw.Draw(dst, b, &image.Uniform{bg}, image.ZP, draw.Src)
	}

	for c := e.FirstChild; c != nil; c = c.NextSibling {
		switch c.Type {
		case html.TextNode:
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
				wordSizeInPx := int(fntDrawer.MeasureString(word)) >> 6
				if dot.X+wordSizeInPx > width {
					dot.X = 0
					dot.Y += fSize
					fntDrawer.Dot = fixed.P(dot.X, dot.Y+int(fontFace.Metrics().Ascent>>6))
				} else {

				}
				fntDrawer.DrawString(word)

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
			// for now, pretend all text is inline
			//fntDrawer.DrawString(c.Data)
		case html.ElementNode:

			// for now, pretend all elements are blocks

			// Draw the block itself, and move dot.
			childHeight, _ := c.GetHeightInPx(width)
			childWidth, _ := c.GetWidthInPx(containerWidth)
			childImage := image.NewRGBA(image.Rectangle{image.ZP, image.Point{childWidth, childHeight}})
			childImage = c.Render(width)

			sr := childImage.Bounds()

			r := image.Rectangle{dot, dot.Add(sr.Size())}
			draw.Draw(dst, r, childImage, sr.Min, draw.Over)
			if c.GetDisplayProp() != "inline" {
				dot.X = 0
				dot.Y += childHeight
			} else {
				dot.X += childWidth
			}
			fntDrawer.Dot = fixed.P(dot.X, dot.Y+int(fontFace.Metrics().Ascent)>>6)

		}
	}
	return dst
}
