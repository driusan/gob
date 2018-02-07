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
	"image/draw"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"net/url"
	"os"
	"strings"
	"unicode"
	//"strconv"
)

const (
	DefaultFontSize = 12
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
	BoxDrawRectangle    image.Rectangle
	BoxContentRectangle image.Rectangle

	ImageMap       ImageMap
	PageLocation   *url.URL
	FirstPageOnly  bool
	RenderAbort    chan bool
	ViewportHeight int
	contentWidth   int

	containerWidth  int
	containerHeight int

	lineBoxes []lineBox

	resolver   net.URLReader
	layoutDone bool
}

func (e RenderableDomElement) String() string {
	var ret string
	if e.Element != nil {
		ret += fmt.Sprintf("%v", e.Element)
	}
	if e.Styles != nil {
		ret += fmt.Sprintf("%v", e.Styles)
	}
	return ret
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

func (e RenderableDomElement) renderLineBox(remainingWidth int, textContent string, force bool, inlinesibling bool) (img *image.RGBA, consumed, unconsumed string) {
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
		ssize, _ = stringSize(fntDrawer, strings.TrimSpace(textContent))
		if ssize > remainingWidth {
			ssize = remainingWidth
		}
	}
	lineheight := e.GetLineHeight()
	start := 0
	if unicode.IsSpace(rune(textContent[0])) {
		start = (fSize / 3)
		ssize += start
	}
	if unicode.IsSpace(rune(textContent[len(textContent)-1])) {
		ssize += (fSize / 3)
	}
	img = image.NewRGBA(image.Rectangle{image.ZP, image.Point{ssize, lineheight}})

	fntDrawer.Dot = fixed.P(start, fontFace.Metrics().Ascent.Floor())
	fntDrawer.Dst = img

	defer func() {
		if decoration := e.GetTextDecoration(); decoration != "" && decoration != "none" && decoration != "blink" {
			if strings.Contains(decoration, "underline") {
				y := fntDrawer.Dot.Y.Floor() + 1
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
	}()

	if whitespace == "pre" {
		fntDrawer.DrawString(lines[0])
		return
	}
	if whitespace == "nowrap" {
		//fmt.Printf("No wrap! %s", words)
	}

	for i, word := range words {
		wordSizeInPx := fntDrawer.MeasureString(word).Ceil()
		if dot+wordSizeInPx > remainingWidth && whitespace != "nowrap" {
			if i == 0 && force {
				// make sure at least one word gets consumed to avoid an infinite loop.
				// this isn't ideal, since some words will disappear, but if we reach this
				// point we're already in a pretty bad state..
				unconsumed = strings.Join(words[i+1:], " ")
				consumed = words[0]
			} else {
				unconsumed = strings.Join(words[i:], " ")
				consumed = strings.Join(words[:i], " ")
			}
			return
		}
		fntDrawer.DrawString(word)

		if i == len(words)-1 {
			if !inlinesibling {
				ssize = int(fntDrawer.Dot.X) >> 6
			}
			break
		}
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
		if !inlinesibling {
			ssize = int(fntDrawer.Dot.X) >> 6
		}
		fntDrawer.Dot.X = fixed.Int26_6(dot << 6)
	}
	unconsumed = ""
	consumed = textContent
	return
}

func (e *RenderableDomElement) Render(containerWidth int) image.Image {
	leftFloatStack, rightFloatStack := make(FloatStack, 0), make(FloatStack, 0)
	var lh int
	e.LayoutPass(containerWidth, image.ZR, &image.Point{0, 0}, leftFloatStack, rightFloatStack, &lh)
	return e.DrawPass()
}

func (e *RenderableDomElement) InvalidateLayout() {
	if e == nil {
		return
	}
	e.layoutDone = false
	e.OverlayedContent = nil
	e.CSSOuterBox = nil
	e.ContentOverlay = nil
	e.lineBoxes = nil

	if e.FirstChild != nil {
		e.FirstChild.InvalidateLayout()
	}
	if e.NextSibling != nil {
		e.NextSibling.InvalidateLayout()
	}
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
func (e *RenderableDomElement) LayoutPass(containerWidth int, r image.Rectangle, dot *image.Point, leftFloatStack, rightFloatStack FloatStack, nextline *int) (image.Image, image.Point) {
	var overlayed *DynamicMemoryDrawer
	if e.layoutDone {
		return e.OverlayedContent, image.Point{}
	}
	defer func() {
		if overlayed != nil {
			e.OverlayedContent = image.NewRGBA(overlayed.Bounds())
		} else {
			e.OverlayedContent = image.NewRGBA(image.ZR)
		}
		e.layoutDone = true
	}()
	e.RenderAbort = make(chan bool)

	width := e.GetContainerWidth(containerWidth)
	e.contentWidth = width
	e.containerWidth = containerWidth

	height := e.GetMinHeight()
	if lh := e.GetLineHeight(); lh > *nextline {
		*nextline = lh
	}

	// special cases
	if e.Type == html.ElementNode {
		switch strings.ToLower(e.Data) {
		case "img":
			var loadedImage bool
			for _, attr := range e.Attr {
				if loadedImage {
					_, contentbox := e.calcCSSBox(e.ContentOverlay, 0)
					e.BoxContentRectangle = contentbox
					//cr := image.Rectangle{contentStart, contentStart.Add(contentBounds.Size())}
					overlayed.GrowBounds(contentbox)
					fmt.Println(e.ContentOverlay.Bounds())
					fmt.Println(contentbox.Bounds())
					return e.ContentOverlay, *dot
				}
				switch attr.Key {
				case "src":
					// Seeing this print way too many times.. something's wrong.
					fmt.Printf("Should load: %s\n", attr.Val)
					u, err := url.Parse(attr.Val)
					if err != nil {
						loadedImage = true
						break
					}
					newURL := e.PageLocation.ResolveReference(u)
					r, code, err := e.resolver.GetURL(newURL)
					if err != nil {
						panic(err)
					}
					if code < 200 || code >= 300 {
						fmt.Println("Error", code)
						continue
					}
					content, format, err := image.Decode(r)
					if err == nil {
						e.ContentOverlay = content
						size := content.Bounds().Size()
						if e.Styles.Width.GetValue() == "" {
							e.Styles.Width = css.NewPxValue(size.X)
						}
						if e.Styles.Height.GetValue() == "" {
							e.Styles.Height = css.NewPxValue(size.Y)
						}
						width = size.X
						height = size.Y
						overlayed = NewDynamicMemoryDrawer(image.Rectangle{image.ZP, image.Point{width, height}})
						fmt.Println("Loaded", attr.Val, width, height)
					} else {
						fmt.Fprintf(os.Stderr, "Unknown image format: %s Err: %s\n", format, err)
						e.ContentOverlay = image.NewRGBA(image.ZR)
					}
					loadedImage = true
				}
			}
		}
	}

	overlayed = NewDynamicMemoryDrawer(image.Rectangle{image.ZP, image.Point{width, height}})

	firstLine := true
	imageMap := NewImageMap()
	collapsablemargin := 0
	for c := e.FirstChild; c != nil; c = c.NextSibling {
		c.ViewportHeight = e.ViewportHeight
		if dot.X < leftFloatStack.Width() {
			dot.X = leftFloatStack.Width()
		}
		float := c.GetFloat()
		switch c.Type {

		case html.TextNode:
			if strings.TrimSpace(c.Data) == "" {
				continue
			}
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

			remainingTextContent := c.Data
			whitespace := e.GetWhiteSpace()
			for remainingTextContent != "" {
				if whitespace == "normal" {
					if width-dot.X-rfWidth <= 0 {
						lfHeight := leftFloatStack.NextFloatHeight()
						rfHeight := rightFloatStack.NextFloatHeight()
						if len(leftFloatStack) == 0 && len(rightFloatStack) == 0 {
							dot.X = 0
							dot.Y += *nextline
						}

						if lfHeight <= 0 && rfHeight <= 0 {
							remainingTextContent = ""
							continue
						}
						if lfHeight > 0 && lfHeight < rfHeight {
							dot.Y += lfHeight + 1
							leftFloatStack = leftFloatStack.ClearFloats(*dot)
							dot.X = leftFloatStack.Width()
						} else if rfHeight > 0 {
							dot.Y += rfHeight + 1
							rightFloatStack = rightFloatStack.ClearFloats(*dot)
						} else {
							panic("Clearing floats didn't make any space.")
						}
						lfWidth = leftFloatStack.Width()
						rfWidth = rightFloatStack.Width()
					}
				}
				childImage, consumed, rt := c.renderLineBox(width-dot.X-rfWidth, remainingTextContent, false, c.NextSibling != nil && c.NextSibling.GetDisplayProp() == "inline")
				if consumed == "" && dot.X == 0 {
					childImage, consumed, rt = c.renderLineBox(width-dot.X-rfWidth, remainingTextContent, true, c.NextSibling != nil && c.NextSibling.GetDisplayProp() == "inline")
				}
				remainingTextContent = rt
				sr := childImage.Bounds()
				r := image.Rectangle{*dot, dot.Add(sr.Size())}
				overlayed.GrowBounds(r)

				c.lineBoxes = append(c.lineBoxes, lineBox{childImage, *dot})
				switch e.GetWhiteSpace() {
				case "pre":
					dot.Y += *nextline
					dot.X = lfWidth
				case "nowrap":
					fallthrough // dot.X = r.Max.X
				case "normal":
					fallthrough
				default:
					if r.Max.X >= width-rfWidth {
						// there's no space left on this line, so advance dot to the next line.
						dot.Y += *nextline
						// clear the floats that have been passed, and then move dot to the edge
						// of the left float.
						leftFloatStack = leftFloatStack.ClearFloats(*dot)
						rightFloatStack = rightFloatStack.ClearFloats(*dot)
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
				dot.Y += *nextline
				continue
			}
			switch display := c.GetDisplayProp(); display {
			case "none", "table-column", "table-column-group":
				// the spec says column and column-group are
				// treated the same as none.
				continue
			default:
				fallthrough
			case "inline":
				if firstLine == true {
					dot.X += c.GetTextIndent(width)
					firstLine = false
				}
				childContent, newDot := c.LayoutPass(width, image.ZR, dot, leftFloatStack, rightFloatStack, nextline)
				c.ContentOverlay = childContent
				_, contentbox := c.calcCSSBox(childContent, 0)
				c.BoxContentRectangle = contentbox
				//cr := image.Rectangle{contentStart, contentStart.Add(contentBounds.Size())}
				overlayed.GrowBounds(contentbox)

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
			case "block", "inline-block", "table", "table-inline", "list-item":
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
				var lh int
				cdot := image.Point{}

				// Collapse margins before doing layout
				if display != "inline-block" {
					if collapsablemargin > 0 {
						tm := c.GetMarginTopSize()
						if tm > collapsablemargin {
							collapsablemargin = tm
						}
						dot.Y -= collapsablemargin
					} else {
						collapsablemargin = c.GetMarginBottomSize()
					}
				} else {
					collapsablemargin = 0
				}
				var childContent image.Image
				childContent, _ = c.LayoutPass(width, image.ZR, &cdot, nil, nil, &lh)
				c.ContentOverlay = childContent
				box, contentbox := c.calcCSSBox(childContent, collapsablemargin)
				c.BoxContentRectangle = contentbox
				sr := box.Bounds()

				r := image.Rectangle{*dot, dot.Add(sr.Size())}

				switch float {
				case "right":
					size := sr.Size()
					rightFloatX := width - rightFloatStack.Width()
					r = image.Rectangle{
						Min: image.Point{rightFloatX - size.X, dot.Y},
						Max: image.Point{rightFloatX, size.Y + dot.Y},
					}
				case "left":
					size := sr.Size()
					leftFloatX := leftFloatStack.Width()
					r = image.Rectangle{
						Min: image.Point{leftFloatX, dot.Y},
						Max: image.Point{leftFloatX + size.X, size.Y + dot.Y},
					}
				}
				overlayed.GrowBounds(r)
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
					newArea := area.Area.Add(*dot).Add(contentbox.Min)
					imageMap.Add(area.Content, newArea)
				}

				// now draw the content on top of the outer box
				contentStart := dot.Add(contentbox.Min)
				contentBounds := c.ContentOverlay.Bounds()
				cr := image.Rectangle{contentStart, contentStart.Add(contentBounds.Size())}
				overlayed.GrowBounds(cr)

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
						if mb := c.GetMarginBottomSize(); mb < 0 {
							println("Negative bottom margin")
							dot.Y += mb
						}
					}

				}
			}

		}
		if float == "none" {
			leftFloatStack = leftFloatStack.ClearFloats(*dot)
			rightFloatStack = rightFloatStack.ClearFloats(*dot)
		}
	}
	e.ImageMap = imageMap
	return overlayed, *dot
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
	if e.Type == html.ElementNode {
		switch strings.ToLower(e.Data) {
		case "img":
			return e.ContentOverlay
		}
	}
	for c := e.FirstChild; c != nil; c = c.NextSibling {
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
			default:
				fallthrough
			case "inline", "inline-block":
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
			case "list-item", "table", "table-inline":
				// Hacks which are displayed as blocks because they're not implemented
				fallthrough
			case "block":
				// draw the border, background, and CSS outer box.
				childContent := c.DrawPass()
				c.ContentOverlay = childContent
				//var drawMask image.Image
				//maskP := image.ZP
				if c.CSSOuterBox != nil {
					sr := c.CSSOuterBox.Bounds()
					draw.Draw(
						e.OverlayedContent,
						c.BoxDrawRectangle,
						c.CSSOuterBox,
						sr.Min,
						draw.Over,
					)
				}

				// now draw the content on top of the outer box
				contentStart := c.BoxDrawRectangle.Min.Add(c.BoxContentRectangle.Min)
				contentBounds := c.ContentOverlay.Bounds()
				cr := image.Rectangle{contentStart, contentStart.Add(contentBounds.Size())}
				switch c.GetOverflow() {
				default:
					fallthrough
				case "visible":
					draw.Draw(
						e.OverlayedContent,
						cr,
						c.ContentOverlay,
						contentBounds.Min,
						draw.Over,
					)
				case "hidden":
					draw.DrawMask(
						e.OverlayedContent,
						cr,
						c.ContentOverlay,
						contentBounds.Min,
						c.BoxContentRectangle,
						c.BoxContentRectangle.Min,
						draw.Over,
					)
				}

			}

		}
	}
	return e.OverlayedContent
}
