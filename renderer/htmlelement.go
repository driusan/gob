package renderer

import (
	"Gob/css"
	"Gob/dom"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
	"golang.org/x/net/html"
	"image"
	"image/color"
	"image/draw"
	"strings"
)

const (
	DefaultFontSize = 16
)

// A RenderableElement is something that can be rendered to
// an image.
type Renderer interface {
	// Returns an image representing this element.
	Render(containerWidth int) image.Image
}

type RenderableDomElement struct {
	*dom.Element
	Styles *css.StyledElement

	Parent      *RenderableDomElement
	FirstChild  *RenderableDomElement
	NextSibling *RenderableDomElement
	PrevSibling *RenderableDomElement
}

func getFontHeight(face font.Face) int {
	metrics := face.Metrics()
	return (metrics.Ascent + metrics.Descent).Ceil()
}
func stringSize(fntDrawer font.Drawer, textContent string) (int, error) {
	var size int
	words := strings.Fields(textContent)
	fSize := getFontHeight(fntDrawer.Face)
	//firstRune, _ := utf8.DecodeRuneInString(textContent)

	for _, word := range words {
		wordSizeInPx := fntDrawer.MeasureString(word).Ceil()
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
	fSize := e.GetFontSize()
	fontFace := e.Styles.GetFontFace(fSize)
	return getFontHeight(fontFace)
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

func (e RenderableDomElement) GetBackgroundColor() color.Color {
	deflt := &color.RGBA{0x00, 0xE0, 0xE0, 0x00}
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

func (e RenderableDomElement) GetTextDecoration() string {
	if e.Styles == nil {
		return "none"
	}

	switch decoration := e.Styles.TextDecoration.GetValue(); decoration {
	case "inherit":
		return e.Parent.GetTextDecoration()
	default:
		return strings.TrimSpace(decoration)
	}
}
func (e RenderableDomElement) GetTextTransform() string {
	if e.Styles == nil {
		return "none"
	}

	switch transformation := e.Styles.TextTransform.GetValue(); transformation {
	case "inherit":
		return e.Parent.GetTextTransform()
	case "capitalize", "uppercase", "lowercase", "none":
		return transformation
	default:
		if e.Parent == nil {
			return "none"
		}
		return e.Parent.GetTextTransform()
	}
}
func (e RenderableDomElement) renderLineBox(remainingWidth int, textContent string) (img *image.RGBA, unconsumed string) {
	switch e.GetTextTransform() {
	case "capitalize":
		textContent = strings.Title(textContent)
	case "uppercase":
		textContent = strings.ToUpper(textContent)
	case "lowercase":
		textContent = strings.ToLower(textContent)
	}
	words := strings.Fields(textContent)
	fSize := e.GetFontSize()
	fontFace := e.Styles.GetFontFace(fSize)
	var dot int
	fntDrawer := font.Drawer{
		Dst:  nil,
		Src:  &image.Uniform{e.GetColor()},
		Face: fontFace,
		Dot:  fixed.P(0, fontFace.Metrics().Ascent.Floor()),
	}

	ssize, _ := stringSize(fntDrawer, textContent)
	if ssize > remainingWidth {
		ssize = remainingWidth
	}
	lineheight := e.GetLineHeight()
	img = image.NewRGBA(image.Rectangle{image.ZP, image.Point{ssize, lineheight}})
	fntDrawer.Dst = img

	if decoration := e.GetTextDecoration(); decoration != "" && decoration != "none" && decoration != "blink" {
		color := e.GetColor()
		if strings.Contains(decoration, "underline") {
			y := fntDrawer.Dot.Y.Floor()
			for px := 0; px < ssize; px++ {
				img.Set(px, y, color)
			}
		}
		if strings.Contains(decoration, "overline") {
			y := 1
			for px := 0; px < ssize; px++ {
				img.Set(px, y, color)
			}
		}
		if strings.Contains(decoration, "line-through") {
			y := fontFace.Metrics().Ascent.Floor() / 2
			for px := 0; px < ssize; px++ {
				img.Set(px, y, color)
			}
		}
	}
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

func (e RenderableDomElement) GetTextIndent(containerWidth int) int {
	// it's inherited, with the initial value of 0
	if e.Styles == nil {
		if e.Parent == nil {
			return 0
		}
		return e.Parent.GetTextIndent(containerWidth)
	}
	val := e.Styles.TextIndent.GetValue()
	if val == "" {
		if e.Parent == nil {
			return 0
		}
		return e.Parent.GetTextIndent(containerWidth)
	}
	px, err := css.ConvertUnitToPx(e.GetFontSize(), containerWidth, val)
	if err != nil {
		return 0
	}
	return px
}
func (e RenderableDomElement) Render(containerWidth int) image.Image {
	dot := image.Point{0, 0}

	width := containerWidth - (e.GetMarginLeftSize() + e.GetMarginRightSize() + e.GetBorderLeftWidth() + e.GetBorderRightWidth() + e.GetPaddingLeftSize() + e.GetPaddingRightSize())
	height := 0

	dst := NewDynamicMemoryDrawer(image.Rectangle{image.ZP, image.Point{width, height}})

	firstLine := true
	for c := e.FirstChild; c != nil; c = c.NextSibling {
		switch c.Type {
		case html.TextNode:
			// text nodes are inline elements that didn't match
			// anything when adding styles, but that's okay,
			// because their style should be identical to their
			// parent.
			c.Styles = e.Styles

			if firstLine == true {
				dot.X += c.GetTextIndent(width)
				firstLine = false
			}

			remainingTextContent := strings.TrimSpace(c.Data)
			for remainingTextContent != "" {
				childImage, rt := c.renderLineBox(width-dot.X, remainingTextContent)
				remainingTextContent = rt
				sr := childImage.Bounds()
				r := image.Rectangle{dot, dot.Add(sr.Size())}
				dst.GrowBounds(r)
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
			case "none":
				continue
			case "inline":
				if firstLine == true {
					dot.X += c.GetTextIndent(width)
					firstLine = false
				}
				remainingTextContent := strings.TrimSpace(c.GetTextContent())
				for remainingTextContent != "" {
					childImage, rt := c.renderLineBox(width-dot.X, remainingTextContent)
					remainingTextContent = rt
					sr := childImage.Bounds()
					r := image.Rectangle{dot, dot.Add(sr.Size())}
					dst.GrowBounds(r)
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
				childContent := c.Render(width)
				box := c.getCSSBox(childContent)
				sr := box.Bounds()
				r := image.Rectangle{dot, dot.Add(sr.Size())}
				dst.GrowBounds(r)

				draw.Draw(
					dst,
					r,
					box,
					sr.Min,
					draw.Over,
				)

				contentStart := dot.Add(box.GetContentOrigin())
				contentBounds := childContent.Bounds()
				cr := image.Rectangle{contentStart, contentStart.Add(contentBounds.Size())}
				dst.GrowBounds(cr)
				draw.Draw(
					dst,
					cr,
					childContent,
					contentBounds.Min,
					draw.Over,
				)

				dot.X = 0
				dot.Y = r.Max.Y
			}

		}
	}
	return dst
}
