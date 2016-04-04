package renderer

import (
	"fmt"
	"Gob/css"
	"Gob/dom"
	"image"
	"image/color"
	"golang.org/x/net/html"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
	"image/draw"
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

	FirstChild *RenderableDomElement
	NextSibling *RenderableDomElement
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
	return 50, nil
}
func (e RenderableDomElement) GetWidthInPx(containerWidth int) (int, error) {
	return containerWidth, nil
}

func (e RenderableDomElement) GetBackgroundColor() color.Color {
	background := color.RGBA{0xE0, 0xE0, 0xE0, 0xFF}
	return background
}
func (e RenderableDomElement) GetColor() color.Color {
	background := color.RGBA{0, 0, 0, 0xFF}
	return background
}
func (e RenderableDomElement) Render(containerWidth int) *image.RGBA {

	height, _ := e.GetHeightInPx(containerWidth)
	if height < 0 {
		height = 0
	}
	width, _ := e.GetWidthInPx(containerWidth)
	bg := e.GetBackgroundColor()
print(height, "height and ", width, "width\n")
	dst := image.NewRGBA(image.Rectangle{image.ZP, image.Point{width, height}})
	imageSize := dst.Bounds()

	if e.Element.Type == html.ElementNode && e.Element.Data == "body" {
		if height < imageSize.Max.Y {
			height = imageSize.Max.Y
		}
		b := image.Rectangle{image.Point{0, 0}, image.Point{width, height}}
		draw.Draw(dst, b, &image.Uniform{bg}, image.ZP, draw.Src)
	}

	//fmt.Printf("width, height for %s: %d, %d\n", e.Data, width, height)

	dot := image.Point{0, 0}
	fSize := e.Styles.GetFontSize()
	fontFace := e.Styles.GetFontFace(fSize)
	fmt.Printf("Font metrics: %s\n", fontFace.Metrics())
	fntDrawer := font.Drawer{
		Dst:  dst,
		Src:  &image.Uniform{e.GetColor()},
		Face: fontFace,
		//basicfont.Face7x13,
		Dot: fixed.P(dot.X, int(fontFace.Metrics().Ascent)>>6),
	}
	//containsBlocks := e.ContainsBlocks()
	//for _, c := range e.Children {
	       fmt.Printf("printing node: %s", e)
	for c := e.FirstChild; c != nil; c = c.NextSibling {

		switch c.Type {
		case html.TextNode:
			// for now, pretend all text is inline
			fntDrawer.DrawString(c.Data)
		case html.ElementNode:
			// for now, pretend all elements are blocks

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
	/*
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

*/
	}
	return dst
}
