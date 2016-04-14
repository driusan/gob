package renderer

import (
	"Gob/css"
	"Gob/dom"
	//	"fmt"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
	"golang.org/x/net/html"
	"image"
	"image/color"
	"image/draw"
	"strings"
	//"unicode"
	//	"unicode/utf8"
)

const (
	DefaultFontSize = 16
)

// A RenderableElement is something that can be rendered to
// an image.
type Renderer interface {
	// Returns an image representing this element.
	Render(containerWidth int) *image.RGBA
}

type RenderableDomElement struct {
	*dom.Element
	Styles *css.StyledElement

	Parent      *RenderableDomElement
	FirstChild  *RenderableDomElement
	NextSibling *RenderableDomElement
	PrevSibling *RenderableDomElement
}

func stringSize(fntDrawer font.Drawer, textContent string) (int, error) {
	var size int
	words := strings.Fields(textContent)
	fSize := int(fntDrawer.Face.Metrics().Height) >> 6
	//firstRune, _ := utf8.DecodeRuneInString(textContent)

	for _, word := range words {
		wordSizeInPx := int(fntDrawer.MeasureString(word)) >> 6
		size += wordSizeInPx

		// Add a three per em space between words, an em-quad after a period,
		// and an en-quad after other punctuation
		switch word[len(word)-1] {
		case ',', ';', ':', '!', '?':
			size += (fSize / 2)
		case '.':
			size += fSize
		default:
			size += (fSize / 3)
		}
	}
	return size, nil
}

func (e *RenderableDomElement) GetLineHeight() int {
	return e.GetFontSize()
}
func (e *RenderableDomElement) GetFontSize() int {
	fromCSS, err := e.Styles.GetFontSize()
	switch err {
	case css.NoStyles, css.InheritValue:
		if e.Parent == nil {
			return DefaultFontSize
		}
		return e.Parent.GetFontSize()
	case nil:
		return fromCSS
	default:
		panic("Could not determine font size")

	}
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
			c.Walk(callback)
		}
	}
}

func (e RenderableDomElement) GetIntrinsicHeightInPx(containerWidth int) (int, error) {
	switch e.GetDisplayProp() {
	case "inline":
		panic("This shouldn't happen. This is handled in the parent block.")
		return 0, nil //e.inlineBoxImplicitContainerHeight(containerWidth), nil
	case "block":
		fallthrough
	default:
		var calcHeight int
		var dot int = 0

		width, _ := e.GetWidthInPx(containerWidth)

		for c := e.FirstChild; c != nil; c = c.NextSibling {
			switch c.GetDisplayProp() {
			case "inline":
				remainingTextContent := strings.TrimSpace(c.GetTextContent())
				if remainingTextContent == "" {
					continue
				}
				if calcHeight == 0 {
					calcHeight = e.GetLineHeight()
				}

				for remainingTextContent != "" {
					ad, rt := c.measureLineBoxAdvancement(width-dot, remainingTextContent)
					//print( "text: ", remainingTextContent, "\nContainer width: ", width, "\ndot: ", dot, "\nad: ", ad, "\nunc: ", rt, "\n\n")
					remainingTextContent = rt
					// If anything was unconsumed, it means it didn't fit on this line, so make a new
					// one
					if rt != "" {
						dot = 0
						calcHeight += e.GetLineHeight()
					} else {
						dot += ad
					}
				}
			case "block":
				fallthrough
			default:
				dot = 0
				cH, _ := c.GetIntrinsicHeightInPx(width)
				calcHeight += cH
			}

		}

		if calcHeight > 0 {
			return calcHeight, nil
		}
	}

	if e.Styles == nil {
		return 0, css.NoStyles
	}
	return 0, css.NoStyles
}
func (e RenderableDomElement) GetWidthInPx(containerWidth int) (int, error) {
	width := e.Styles.FollowCascadeToPx("width", containerWidth)
	if e.GetDisplayProp() == "block" {
		return e.Styles.FollowCascadeToPx("width", containerWidth), nil
	}
	if e.Type == html.TextNode {
		fSize := e.GetFontSize()
		fontFace := e.Styles.GetFontFace(fSize)
		fntDrawer := font.Drawer{
			Dst:  nil,
			Src:  &image.Uniform{e.GetColor()},
			Face: fontFace,
		}
		sSize, _ := stringSize(fntDrawer, e.Data)
		if sSize > width {
			return width, nil
		}
		return sSize, nil
	}

	var calcWidth int

	for child := e.FirstChild; child != nil; child = child.NextSibling {
		cW, _ := child.GetWidthInPx(width)
		if calcWidth < cW {
			calcWidth = cW
		}
	}
	if calcWidth > 0 {
		return calcWidth, nil
	}
	return width, nil
}

func (e RenderableDomElement) GetBackgroundColor() color.Color {
	deflt := &color.RGBA{0x00, 0xE0, 0xE0, 0x00}
	//bg := e.Styles.GetBackgroundColor(&color.RGBA{0xE0, 0xE0, 0xE0, 0xFF})
	switch bg, err := e.Styles.GetBackgroundColor(deflt); err {
	case css.InheritValue:
		if e.Parent == nil {
			return &color.RGBA{0xE0, 0xE0, 0xE0, 0xFF}
		}
		return e.Parent.GetBackgroundColor()
	case css.NoStyles:
		return deflt
	default:
		return bg
	}
	//background := color.RGBA{0xE0, 0xE0, 0xE0, 0xFF}
	//return background
}
func (e RenderableDomElement) GetColor() color.Color {
	var deflt *color.RGBA
	if e.Type == html.ElementNode && e.Data == "a" {
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
	if cssVal := e.Styles.DisplayProp(); cssVal != "" {
		return cssVal
	}
	return "block"
}

type borderDrawer struct {
	i image.Image
}

func (b *borderDrawer) ColorModel() color.Model {
	return color.AlphaModel
}
func (b *borderDrawer) Bounds() image.Rectangle {
	return b.i.Bounds()
}
func (b *borderDrawer) At(x, y int) color.Color {
	return color.Alpha{0}
	// draw a 4px border for debugging.
	if x < 4 || y < 4 {
		return color.Alpha{255}
	}

	if bounds := b.i.Bounds(); x > bounds.Max.X-4 || y > bounds.Max.Y-4 {
		return color.Alpha{255}
	}
	return color.Alpha{0}
}

// measures how much printing a string will advance the cursor, and returns the amount that
// it would be advanced, and the unconsumed string that didn't fit in the line and will need
// to go into the next line.
// If nothing fits onto the line, it will still consume the first word of the string in order
// to ensure that the caller doesn't get into an infinite loop.
func (e RenderableDomElement) measureLineBoxAdvancement(remainingWidth int, textContent string) (int, string) {
	words := strings.Fields(textContent)
	//firstRune, _ := utf8.DecodeRuneInString(textContent)
	fSize := e.GetFontSize()
	fontFace := e.Styles.GetFontFace(fSize)
	var dot int
	fntDrawer := font.Drawer{
		Dst:  nil,
		Src:  &image.Uniform{e.GetColor()},
		Face: fontFace,
		Dot:  fixed.P(0, int(fontFace.Metrics().Ascent)>>6),
	}

	for i, word := range words {
		wordSizeInPx := int(fntDrawer.MeasureString(word)) >> 6
		if dot+wordSizeInPx > remainingWidth {
			if i == 0 {
				// make sure at least one word gets consumed to avoid an infinite loop.
				// this isn't ideal, since some words will disappear, but if we reach this
				// point we're already in a pretty bad state..
				return 0, strings.Join(words[i+1:], " ")
			}
			return dot, strings.Join(words[i:], " ")
		}
		dot += wordSizeInPx

		// Add a three per em space between words, an em-quad after a period,
		// and an en-quad after other punctuation
		switch word[len(word)-1] {
		case ',', ';', ':', '!', '?':
			dot += (fSize / 2)
		case '.':
			dot += fSize
		default:
			dot += (fSize / 3)
		}
	}
	return dot, ""
}
func (e RenderableDomElement) renderLineBox(remainingWidth int, textContent string) (img *image.RGBA, unconsumed string) {
	words := strings.Fields(textContent)
	//firstRune, _ := utf8.DecodeRuneInString(textContent)
	fSize := e.GetFontSize()
	fontFace := e.Styles.GetFontFace(fSize)
	var dot int
	fntDrawer := font.Drawer{
		Dst:  nil,
		Src:  &image.Uniform{e.GetColor()},
		Face: fontFace,
		Dot:  fixed.P(0, int(fontFace.Metrics().Ascent)>>6),
	}

	ssize, _ := stringSize(fntDrawer, textContent)
	if ssize > remainingWidth {
		ssize = remainingWidth
	}
	img = image.NewRGBA(image.Rectangle{image.ZP, image.Point{ssize, fSize}})
	fntDrawer.Dst = img

	for i, word := range words {
		wordSizeInPx := int(fntDrawer.MeasureString(word)) >> 6
		if dot+wordSizeInPx > remainingWidth {
			if i == 0 {
				// make sure at least one word gets consumed to avoid an infinite loop.
				// this isn't ideal, since some words will disappear, but if we reach this
				// point we're already in a pretty bad state..
				unconsumed = strings.Join(words[i+1:], " ")
			} else {
				unconsumed = strings.Join(words[i:], " ")
			}
			return
		}
		fntDrawer.DrawString(word)

		// Add a three per em space between words, an em-quad after a period,
		// and an en-quad after other punctuation
		switch word[len(word)-1] {
		case ',', ';', ':', '!', '?':
			dot = (int(fntDrawer.Dot.X) >> 6) + (fSize / 2)
		case '.':
			dot = (int(fntDrawer.Dot.X) >> 6) + fSize
		default:
			dot = (int(fntDrawer.Dot.X) >> 6) + (fSize / 3)
		}
		fntDrawer.Dot.X = fixed.Int26_6(dot << 6)
	}
	unconsumed = ""
	return
}

func (e RenderableDomElement) Render(containerWidth int) *image.RGBA {
	dot := image.Point{0, 0}

	width, _ := e.GetWidthInPx(containerWidth)
	height, _ := e.GetIntrinsicHeightInPx(width)
	if height < 0 {
		height = 0
	}
	bg := e.GetBackgroundColor()
	dst := image.NewRGBA(image.Rectangle{image.ZP, image.Point{width, height}})

	displayProp := e.GetDisplayProp()
	// draw the background
	if bg != nil && dst != nil && displayProp != "inline" {
		imageSize := dst.Bounds()
		draw.Draw(dst, imageSize, &image.Uniform{bg}, image.ZP, draw.Src)
	}

	// draw the border
	draw.DrawMask(
		dst,
		dst.Bounds(),
		&image.Uniform{color.RGBA{0, 0, 255, 255}},
		image.ZP,
		&borderDrawer{dst},
		image.ZP,
		draw.Over,
	)

	for c := e.FirstChild; c != nil; c = c.NextSibling {
		switch c.Type {
		case html.TextNode:
			c.Styles = e.Styles
			// Draw the background
			//bgChild := c.GetBackgroundColor()
			//childWidth, _ := c.GetWidthInPx(width)
			remainingTextContent := strings.TrimSpace(c.Data)
			for remainingTextContent != "" {
				childImage, rt := c.renderLineBox(width-dot.X, remainingTextContent)
				remainingTextContent = rt
				sr := childImage.Bounds()
				r := image.Rectangle{dot, dot.Add(sr.Size())}
				draw.Draw(dst, r, childImage, sr.Min, draw.Over)
				if r.Max.X >= width {
					dot.X = 0
					dot.Y += e.GetLineHeight()
				} else {
					dot.X = r.Max.X
				}
			}
		case html.ElementNode:
			switch c.GetDisplayProp() {
			case "inline":
				remainingTextContent := strings.TrimSpace(c.GetTextContent())
				for remainingTextContent != "" {
					childImage, rt := c.renderLineBox(width-dot.X, remainingTextContent)
					remainingTextContent = rt
					sr := childImage.Bounds()
					r := image.Rectangle{dot, dot.Add(sr.Size())}
					draw.Draw(dst, r, childImage, sr.Min, draw.Over)
					if r.Max.X >= width {
						dot.X = 0
						dot.Y += e.GetLineHeight()
					} else {
						dot.X = r.Max.X
					}
				}
			case "block":
				fallthrough
			default:
				childWidth, _ := c.GetWidthInPx(width)
				childHeight, _ := c.GetIntrinsicHeightInPx(childWidth)
				childImage := image.NewRGBA(image.Rectangle{image.ZP, image.Point{childWidth, childHeight}})
				childImage = c.Render(childWidth)
				sr := childImage.Bounds()
				r := image.Rectangle{dot, dot.Add(sr.Size())}
				draw.Draw(dst, r, childImage, sr.Min, draw.Over)

				dot.X = 0
				dot.Y += childHeight
			}

		}
	}
	return dst
}
