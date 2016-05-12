package renderer

import (
	"fmt"
	"github.com/driusan/Gob/css"
	"github.com/driusan/Gob/dom"
	"github.com/driusan/Gob/net"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
	"golang.org/x/net/html"
	"image"
	"image/color"
	"image/draw"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"net/url"
	"os"
	//"strconv"
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

	CSSOuterBox    image.Image
	ContentOverlay image.Image

	// The CSSOuterBox with the Content overlaid on top of it. Allocated
	// during the layout phase, and drawn during the draw phase.
	OverlayedContent draw.Image
	// The location within the parent to draw the OverlayedContent
	//DrawRectangle    image.Rectangle
	//BoxOrigin        image.Point
	BoxDrawRectangle image.Rectangle
	BoxContentOrigin image.Point

	ImageMap       ImageMap
	PageLocation   *url.URL
	FirstPageOnly  bool
	RenderAbort    chan bool
	ViewportHeight int
	contentWidth   int
	containerWidth int

	lineBoxes []lineBox
}

type lineBox struct {
	//image  image.Image
	image.Image
	origin image.Point
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

func (e RenderableDomElement) renderLineBox(remainingWidth int, textContent string) (img *image.RGBA, unconsumed string) {
	switch e.GetTextTransform() {
	case "capitalize":
		textContent = strings.Title(textContent)
	case "uppercase":
		textContent = strings.ToUpper(textContent)
	case "lowercase":
		textContent = strings.ToLower(textContent)
	}
	fSize := e.GetFontSize()
	fontFace := e.GetFontFace(fSize)
	var dot int
	clr := e.GetColor()
	if clr == nil {
		clr = color.RGBA{0xff, 0xff, 0xff, 0xff}
	}
	fntDrawer := font.Drawer{
		Dst:  nil,
		Src:  &image.Uniform{clr},
		Face: fontFace,
	}

	var ssize int

	// words are used for normal, nowrap,
	var words []string
	// lines are used for pre, pre-wrap, and pre-line
	var lines []string
	whitespace := e.GetWhiteSpace()
	switch whitespace {
	case "pre":
		lines = strings.Split(textContent, "\n")
		ssize = fntDrawer.MeasureString(lines[0]).Ceil()
		unconsumed = strings.Join(lines[1:], "\n")
	case "pre-wrap":
		panic("pre-wrap not yet implemented")
		lines = strings.Split(textContent, "\n")
		words = strings.Fields(lines[0])
		ssize = fntDrawer.MeasureString(lines[0]).Ceil()
	case "nowrap":
		// same as normal, but don't cap the size
		// at remaining width
		words = strings.Fields(textContent)
		ssize, _ = stringSize(fntDrawer, textContent)
	case "pre-line":
		panic("pre-line not yet implemented")
	case "normal":
		fallthrough
	default:
		words = strings.Fields(textContent)
		ssize, _ = stringSize(fntDrawer, textContent)
		if ssize > remainingWidth {
			ssize = remainingWidth
		}
	}
	lineheight := e.GetLineHeight()
	img = image.NewRGBA(image.Rectangle{image.ZP, image.Point{ssize, lineheight}})

	//BUG(driusan): This math is wrong
	fntDrawer.Dot = fixed.P(0, fontFace.Metrics().Ascent.Floor())
	fntDrawer.Dst = img

	if decoration := e.GetTextDecoration(); decoration != "" && decoration != "none" && decoration != "blink" {
		if strings.Contains(decoration, "underline") {
			y := fntDrawer.Dot.Y.Floor()
			for px := 0; px < ssize; px++ {
				img.Set(px, y, clr)
			}
		}
		if strings.Contains(decoration, "overline") {
			y := fntDrawer.Dot.Y.Floor() - fontFace.Metrics().Ascent.Floor()
			for px := 0; px < ssize; px++ {
				img.Set(px, y, clr)
			}
		}
		if strings.Contains(decoration, "line-through") {
			y := fntDrawer.Dot.Y.Floor() - (fontFace.Metrics().Ascent.Floor() / 2)
			for px := 0; px < ssize; px++ {
				img.Set(px, y, clr)
			}
		}
	}

	if whitespace == "pre" {
		fntDrawer.DrawString(lines[0])
		return
	}
	if whitespace == "nowrap" {
		//fmt.Printf("No wrap! %s", words)
	}

	for i, word := range words {
		wordSizeInPx := fntDrawer.MeasureString(word).Ceil()
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

func (e *RenderableDomElement) Render(containerWidth int) image.Image {
	e.LayoutPass(containerWidth, image.ZR, image.Point{0, 0})
	return e.DrawPass()
}

// realRender either calculates the size of elements, or draws them, depending on if it's a layoutPass
// or not. It's done in one method because the logic is largely the same.
//
// If it's a layout pass, it will return an empty image large enough to be passed to a render pass. If it's
// a render pass, the final size, r, must be passed from the layout pass so that it can allocate an appropriately
// sized image.
//
// dot is the starting location of dot (a moving cursor used for drawing elements, representing the top-left corner
// to draw at.), and it returns both an image, and the final location of dot after drawing the element. This is
// required because inline elements might render multiple line blocks, and the whole inline isn't necessarily
// square, so you can't just take the bounds of the rendered image.
func (e *RenderableDomElement) LayoutPass(containerWidth int, r image.Rectangle, dot image.Point) (image.Image, image.Point) {
	var dst draw.Image
	defer func() {
		if dst != nil {
			e.OverlayedContent = image.NewRGBA(dst.Bounds())
		}
	}()
	e.RenderAbort = make(chan bool)
	leftFloatStack, rightFloatStack := make(FloatStack, 0), make(FloatStack, 0)

	width := e.GetContentWidth(containerWidth)
	e.contentWidth = width
	e.containerWidth = containerWidth

	height := e.GetContentHeight()

	// special cases
	if e.Type == html.ElementNode {
		switch strings.ToLower(e.Data) {
		case "img":
			var loadedImage bool
			for _, attr := range e.Attr {
				if loadedImage {
					return e.ContentOverlay, dot
				}
				switch attr.Key {
				case "src":
					// Seeing this print way too many times.. something's wrong.
					//fmt.Printf("Should load: %s\n", attr.Val)
					u, err := url.Parse(attr.Val)
					if err != nil {
						loadedImage = true
						break
					}
					newURL := e.PageLocation.ResolveReference(u)
					r, err := net.GetURLReader(newURL)
					if err != nil {
						panic(err)
					}
					content, format, err := image.Decode(r)
					if err == nil {
						e.ContentOverlay = content
					} else {
						fmt.Fprintf(os.Stderr, "Unknown image format: %s Err: %s", format, err)
					}

					loadedImage = true
				}
			}
		}
	}

	var mst *DynamicMemoryDrawer
	mst = NewDynamicMemoryDrawer(image.Rectangle{image.ZP, image.Point{width, height}})
	dst = mst

	firstLine := true
	imageMap := NewImageMap()
	for c := e.FirstChild; c != nil; c = c.NextSibling {
		c.ViewportHeight = e.ViewportHeight
		if dot.X < leftFloatStack.Width() {
			dot.X = leftFloatStack.Width()
		}
		switch c.Type {

		case html.TextNode:
			// text nodes are inline elements that didn't match
			// anything when adding styles, but that's okay,
			// because their style should be identical to their
			// parent.
			c.Styles = e.Styles

			lfWidth := leftFloatStack.Width()
			rfWidth := rightFloatStack.Width()
			if dot.X < lfWidth {
				dot.X = lfWidth
			}
			if firstLine == true {
				dot.X += c.GetTextIndent(width)
				firstLine = false
			}

			remainingTextContent := strings.TrimSpace(c.Data)
			whitespace := e.GetWhiteSpace()
			for remainingTextContent != "" {
				if whitespace == "normal" {
					if width-dot.X-rfWidth <= 0 {
						lfHeight := leftFloatStack.NextFloatHeight()
						rfHeight := rightFloatStack.NextFloatHeight()
						if len(leftFloatStack) == 0 && len(rightFloatStack) == 0 {
							panic("Not enough space to render any element and no floats to remove.")
						}
						if lfHeight <= 0 && rfHeight <= 0 {
							panic("Floats don't have any height to clear")
						}
						if lfHeight > 0 && lfHeight < rfHeight {
							dot.Y += lfHeight + 1
							leftFloatStack = leftFloatStack.ClearFloats(dot)
							dot.X = leftFloatStack.Width()
						} else if rfHeight > 0 {
							dot.Y += rfHeight + 1
							rightFloatStack = rightFloatStack.ClearFloats(dot)
						} else {
							panic("Clearing floats didn't make any space.")
						}
						lfWidth = leftFloatStack.Width()
						rfWidth = rightFloatStack.Width()
					}
				}
				childImage, rt := c.renderLineBox(width-dot.X-rfWidth, remainingTextContent)
				remainingTextContent = rt
				sr := childImage.Bounds()
				r := image.Rectangle{dot, dot.Add(sr.Size())}
				mst.GrowBounds(r)

				c.lineBoxes = append(c.lineBoxes, lineBox{childImage, dot})
				switch e.GetWhiteSpace() {
				case "pre":
					dot.Y += e.GetLineHeight()
					dot.X = lfWidth
				case "nowrap":
					dot.X = r.Max.X
				case "normal":
					fallthrough
				default:
					if r.Max.X >= width-rfWidth {
						// there's no space left on this line, so advance dot to the next line.
						dot.Y += e.GetLineHeight()
						// clear the floats that have been passed, and then move dot to the edge
						// of the left float.
						leftFloatStack = leftFloatStack.ClearFloats(dot)
						rightFloatStack = rightFloatStack.ClearFloats(dot)
						lfWidth = leftFloatStack.Width()
						rfWidth = rightFloatStack.Width()

						dot.X = lfWidth
					} else {
						// there's still space on this line, so move dot to the end
						// of the rendered text.
						dot.X = r.Max.X
					}
				}
				// add this line box to the image map.
				imageMap.Add(c, r)
			}
		case html.ElementNode:
			if c.Data == "br" {
				dot.X = 0
				dot.Y += c.GetLineHeight()
				continue
			}
			switch display := c.GetDisplayProp(); display {
			case "none":
				continue
			case "inline":
				if firstLine == true {
					dot.X += c.GetTextIndent(width)
					firstLine = false
				}
				_, newDot := c.LayoutPass(width, image.ZR, dot)

				/*
					if layoutPass == false {
						c.ContentOverlay = childContent
						bounds := childContent.Bounds()
						draw.Draw(
							dst,
							image.Rectangle{image.ZP, bounds.Max},
							c.ContentOverlay,
							bounds.Min,
							draw.Over,
						)

					}
				*/
				// Populate this image map. This is an inline, so we actually only care
				// about the line boxes that were generated by the children.
				childImageMap := c.ImageMap
				for _, area := range childImageMap {
					// translate the coordinate systems from the child's to this one
					newArea := area.Area //.Add(dot)

					if area.Content.Type == html.TextNode {
						// it was a text node, so for all intents and purposes we're actually
						// hovering over this element
						imageMap.Add(c, newArea)
					} else {
						// it was a child element node, so it's more precise to say we were hovering
						// over the child
						imageMap.Add(area.Content, newArea)
					}
				}
				dot.X = newDot.X
				dot.Y = newDot.Y
			case "block", "inline-block":
				fallthrough
			default:
				float := c.GetFloat()
				if dot.X != 0 && display != "inline-block" {
					// This means the previous child was an inline item, and we should position dot
					// as if there were an implicit box around it.
					// floated elements don't affect dot, so only do this if it's not floated.
					if float == "none" {
						dot.X = 0
						if c.PrevSibling != nil {
							dot.Y += c.PrevSibling.GetLineHeight()
						}
					}
				}

				// draw the border, background, and CSS outer box.
				childContent, _ := c.LayoutPass(width, image.ZR, image.ZP)
				c.ContentOverlay = childContent
				box, contentorigin := c.getCSSBox(childContent, true)
				c.BoxContentOrigin = contentorigin
				sr := box.Bounds()

				r := image.Rectangle{dot, dot.Add(sr.Size())}

				switch float {
				case "right":
					size := sr.Size()
					rightFloatX := width - rightFloatStack.Width()
					r = image.Rectangle{
						Min: image.Point{rightFloatX - size.X, dot.Y},
						Max: image.Point{rightFloatX, size.Y + dot.Y},
					}
					//c.BoxContentOrigin = contentorigin.Add(image.Point{rightFloatX - size.X, 0})
				case "left":
					size := sr.Size()
					leftFloatX := leftFloatStack.Width()
					r = image.Rectangle{
						Min: image.Point{leftFloatX, dot.Y},
						Max: image.Point{leftFloatX + size.X, size.Y + dot.Y},
					}
					//c.BoxContentOrigin = contentorigin.Add(image.Point{leftFloatX, 0})
				}
				mst.GrowBounds(r)
				c.BoxDrawRectangle = r

				if sr.Size().Y > 0 {
					switch float {
					case "left":
						leftFloatStack = append(leftFloatStack, c)
					case "right":
						rightFloatStack = append(rightFloatStack, c)
					}
				}

				// populate the imagemap by adding the child, then adding the children's
				// children.
				// add the child
				childImageMap := c.ImageMap
				imageMap.Add(c, r)
				// add the grandchildren
				for _, area := range childImageMap {
					// translate the coordinate systems from the child's to this one
					newArea := area.Area.Add(dot).Add(contentorigin)
					imageMap.Add(area.Content, newArea)
				}

				// now draw the content on top of the outer box
				contentStart := dot.Add(contentorigin)
				contentBounds := c.ContentOverlay.Bounds()
				cr := image.Rectangle{contentStart, contentStart.Add(contentBounds.Size())}
				mst.GrowBounds(cr)

				switch float {
				case "left", "right":
					// floated boxes don't affect dot.
				case "none":
					fallthrough
				default:
					if display == "inline-block" {
						dot.X = r.Max.X
						if dot.X > width {
							dot.X = 0
							dot.Y = r.Max.Y
						}
					} else {
						dot.X = 0
						dot.Y = r.Max.Y
					}

				}
			}

		}
		leftFloatStack = leftFloatStack.ClearFloats(dot)
		rightFloatStack = rightFloatStack.ClearFloats(dot)
	}
	e.ImageMap = imageMap
	return dst, dot
}

// realRender either calculates the size of elements, or draws them, depending on if it's a layoutPass
// or not. It's done in one method because the logic is largely the same.
//
// If it's a layout pass, it will return an empty image large enough to be passed to a render pass. If it's
// a render pass, the final size, r, must be passed from the layout pass so that it can allocate an appropriately
// sized image.
//
// dot is the starting location of dot (a moving cursor used for drawing elements, representing the top-left corner
// to draw at.), and it returns both an image, and the final location of dot after drawing the element. This is
// required because inline elements might render multiple line blocks, and the whole inline isn't necessarily
// square, so you can't just take the bounds of the rendered image.
func (e *RenderableDomElement) DrawPass() image.Image {
	// special cases
	if e.Type == html.ElementNode {
		switch strings.ToLower(e.Data) {
		case "img":
			// this was retrieved and decoded in the layout pass in order to get the
			// size. There's no need to redo it.
			// FIXME: Removing this and going into the normal path causes a panic.
			//        That shouldn't be the case.
			return e.ContentOverlay
		}
	}

	for c := e.FirstChild; c != nil; c = c.NextSibling {
		if e.FirstPageOnly && c.BoxDrawRectangle.Min.Y > e.ViewportHeight {
			return e.OverlayedContent
		}
		switch c.Type {

		case html.TextNode:
			for _, box := range c.lineBoxes {
				sr := box.Image.Bounds()
				r := image.Rectangle{box.origin, box.origin.Add(sr.Size())}
				if e.FirstPageOnly && r.Min.Y > e.ViewportHeight {
					return e.OverlayedContent
				}
				draw.Draw(e.OverlayedContent, r, box.Image, sr.Min, draw.Src)
			}
		case html.ElementNode:
			if c.Data == "br" {
				continue
			}
			switch c.GetDisplayProp() {
			case "none":
				continue
			case "inline":
				childContent := c.DrawPass()

				c.ContentOverlay = childContent
				bounds := childContent.Bounds()
				draw.Draw(
					e.OverlayedContent,
					image.Rectangle{image.ZP, bounds.Max},
					c.ContentOverlay,
					bounds.Min,
					draw.Over,
				)
			case "block":
				fallthrough
			default:
				// draw the border, background, and CSS outer box.
				childContent := c.DrawPass()
				c.ContentOverlay = childContent
				box, _ := c.getCSSBox(childContent, false)
				sr := box.Bounds()

				// draw the box onto the overlayed content
				draw.Draw(
					e.OverlayedContent,
					c.BoxDrawRectangle,
					c.CSSOuterBox,
					sr.Min,
					draw.Over,
				)

				// now draw the content on top of the outer box
				contentStart := c.BoxDrawRectangle.Min.Add(c.BoxContentOrigin)
				contentBounds := c.ContentOverlay.Bounds()
				cr := image.Rectangle{contentStart, contentStart.Add(contentBounds.Size())}
				draw.Draw(
					e.OverlayedContent,
					cr,
					c.ContentOverlay,
					contentBounds.Min,
					draw.Over,
				)

			}

		}
	}
	return e.OverlayedContent
}
