package renderer

import (
	"fmt"
	"github.com/driusan/Gob/css"
	"github.com/driusan/Gob/dom"
	"github.com/driusan/Gob/net"

	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
	"golang.org/x/net/html"

	"github.com/nfnt/resize"

	"image"
	"image/draw"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"net/url"
	"os"
	"strconv"
	"strings"
	"unicode"
	//"strconv"
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
	RenderAbort    chan bool
	ViewportHeight int
	contentWidth   int

	containerWidth  int
	containerHeight int

	curLine   []*lineBox
	lineBoxes []*lineBox

	resolver   net.URLReader
	layoutDone bool

	leftFloats, rightFloats FloatStack

	// The number of child bullet items that have been drawn.
	// Used for determining the counter to place next to the list item
	numBullets int
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
	Content     image.Image
	BorderImage image.Image

	origin   image.Point
	borigin  image.Point
	baseline int

	content string // for debugging, may be removed
	el      *RenderableDomElement
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

func (e RenderableDomElement) renderLineBox(remainingWidth int, textContent string, force bool, inlinesibling bool) (img *image.RGBA, consumed, unconsumed string, baseline int) {
	if remainingWidth < 0 {
		panic("No room to render text")
	}
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
	if ssize > remainingWidth {
		if force != true && len(words) > 0 {
			// If it's not being forced, check if at least one word
			// fits
			psize, _ := stringSize(fntDrawer, words[0])
			if psize < remainingWidth {
				// A word fits, so just keep going
				ssize = remainingWidth
			} else {
				// Nothing fits, so don't consume anything
				// since it's not being forced.
				consumed = ""
				unconsumed = textContent
				return
			}
		} else {
			// force is true, so just make it take up all the remaining width
			ssize = remainingWidth
		}
	}

	if ssize < 0 {
		panic("This should never happen")
	}
	img = image.NewRGBA(image.Rectangle{image.ZP, image.Point{ssize, lineheight}})

	baseline = fontFace.Metrics().Ascent.Floor()
	fntDrawer.Dot = fixed.P(start, baseline)
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
		consumed = lines[0]
		return
	}
	if whitespace == "nowrap" {
		//fmt.Printf("No wrap! %s", words)
	}

	for i, word := range words {
		wordSizeInPx := fntDrawer.MeasureString(word).Ceil()
		if dot+wordSizeInPx > remainingWidth && whitespace != "nowrap" {
			// The word doesn't fit on this line.
			if strings.Index(word, "-") >= 0 {
				// This isn't required, but including partial words that have a hyphen
				// on a line is what Chrome and Firefox do, and makes it easier to
				// compare test cases.
				pword := strings.SplitN(word, "-", 2)
				pword[0] = pword[0] + "-"
				wordSizeInPx = fntDrawer.MeasureString(pword[0]).Ceil()
				if dot+wordSizeInPx <= remainingWidth {
					fntDrawer.DrawString(pword[0])
					unconsumed = strings.Join(words[i+1:], " ")
					if len(unconsumed) > 0 {
						unconsumed = pword[1] + " " + unconsumed
					} else {
						unconsumed = pword[1]
					}
					consumed = strings.Join(words[:i], " ")
					if len(consumed) > 0 {
						consumed = consumed + " " + pword[0]
					} else {
						consumed = pword[0]
					}
					return
				}
				// falthrough to normal exit, we couldn't fit a partial word
			}
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
	e.leftFloats = make(FloatStack, 0)
	e.rightFloats = make(FloatStack, 0)
	var lh int
	e.LayoutPass(containerWidth, image.ZR, &image.Point{0, 0}, &lh)
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
	e.leftFloats = nil
	e.rightFloats = nil

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
func (e *RenderableDomElement) LayoutPass(containerWidth int, r image.Rectangle, dot *image.Point, nextline *int) (image.Image, image.Point) {
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

	fdot := image.Point{dot.X, dot.Y}
	height := e.GetMinHeight()
	if lh := e.GetLineHeight(); lh > *nextline {
		*nextline = lh
	}

	e.curLine = make([]*lineBox, 0)
	// special cases
	if e.Type == html.ElementNode {
		switch strings.ToLower(e.Data) {
		case "img":
			var loadedImage bool
			var iwidth, iheight int
			var ewidth, eheight bool

			for _, attr := range e.Attr {
				switch attr.Key {
				case "src":
					// Seeing this print way too many times.. something's wrong.
					//fmt.Printf("Should load: %s\n", attr.Val)
					u, err := url.Parse(attr.Val)
					if err != nil {
						loadedImage = true
						continue
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
						iwidth = size.X
						iheight = size.Y
					} else {
						fmt.Fprintf(os.Stderr, "Unknown image format: %s Err: %s\n", format, err)
						e.ContentOverlay = image.NewRGBA(image.ZR)
					}
					loadedImage = true
				case "width":
					if !ewidth {
						iwidth, _ = strconv.Atoi(attr.Val)
						ewidth = true
						if !eheight {
							iheight = 0
						}
					}
				case "height":
					if !eheight {
						iheight, _ = strconv.Atoi(attr.Val)
						eheight = true
						if !ewidth {
							iwidth = 0
						}
					}
				}
			}

			if css := e.Styles.Width.GetValue(); css != "" {
				if w := e.GetWidth(); w > 0 {
					ewidth = true
					iwidth = w
					if !eheight {
						iheight = 0
					}
				}
			}
			if css := e.Styles.Height.GetValue(); css != "" {
				if h := e.GetHeight(); h > 0 {
					eheight = true
					iheight = h
					if !ewidth {
						iwidth = 0
					}
				}
			}
			if iwidth != 0 || iheight != 0 {
				e.ContentOverlay = resize.Resize(uint(iwidth), uint(iheight), e.ContentOverlay, resize.NearestNeighbor)
				sz := e.ContentOverlay.Bounds().Size()
				iwidth = sz.X
				iheight = sz.Y
			}
			if e.Styles.Width.GetValue() == "" {
				e.Styles.Width = css.NewPxValue(iwidth)
			}
			if e.Styles.Height.GetValue() == "" {
				e.Styles.Height = css.NewPxValue(iheight)
			}
			//e.Styles.Overflow = css.NewValue("hidden")
			overlayed = NewDynamicMemoryDrawer(image.Rectangle{image.ZP, image.Point{iwidth, iheight}})
			if loadedImage {
				e.contentWidth = iwidth
				box, contentbox := e.calcCSSBox(e.ContentOverlay.Bounds().Size())
				e.BoxContentRectangle = contentbox
				overlayed.GrowBounds(contentbox)

				switch e.GetDisplayProp() {
				case "block":
					dot.Y += box.Bounds().Size().Y
					dot.X = 0
					e.BoxDrawRectangle = box.Bounds()
					return e.ContentOverlay, *dot
				default:
					fallthrough
				case "inline", "inline-block":
					//if e.Styles.LineHeight.GetValue() == "" {
					//	e.Styles.LineHeight = css.NewPxValue(iheight)
					//}
					lb := lineBox{e.ContentOverlay, nil, *dot, *dot, iheight, "[img]", e}
					e.Parent.lineBoxes = append(e.Parent.lineBoxes, &lb)
					e.Parent.curLine = append(e.Parent.curLine, &lb)
					dot.X += box.Bounds().Size().X
					if iheight > *nextline {
						*nextline = iheight
					}
				}
				return box, *dot
			}
			return e.ContentOverlay, *dot
		}
	}

	overlayed = NewDynamicMemoryDrawer(image.Rectangle{image.ZP, image.Point{width, height}})

	firstLine := true
	imageMap := NewImageMap()

	// The bottom margin of the last child, used for collapsing margins (where applicable)
	bottommargin := 0
	for c := e.FirstChild; c != nil; c = c.NextSibling {
		if fdot.Y < dot.Y {
			fdot = *dot
		}
		c.ViewportHeight = e.ViewportHeight
		if dot.X < e.leftFloats.MaxX(*dot) {
			dot.X = e.leftFloats.MaxX(*dot)
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

			lfWidth := e.leftFloats.MaxX(*dot)
			rfWidth := e.rightFloats.WidthAt(*dot)
			if dot.X < lfWidth {
				dot.X = lfWidth
			}

			dot.X += c.listIndent()
			if firstLine == true {
				dot.X += c.GetTextIndent(width)
				firstLine = false
			}

			remainingTextContent := c.Data
			whitespace := e.GetWhiteSpace()
		textdraw:
			for strings.TrimSpace(remainingTextContent) != "" {
				if whitespace == "normal" {
					if width-dot.X-rfWidth <= 0 {
						dot.Y += *nextline
						*nextline = c.GetLineHeight()

						lfWidth = e.leftFloats.MaxX(*dot)
						rfWidth = e.rightFloats.ClearFloats(*dot).WidthAt(*dot)

						dot.X = e.leftFloats.MaxX(*dot)
						e.advanceLine(dot)
						continue
					}
				}
				if width-dot.X-rfWidth <= 0 {
					continue
				}

				var childImage image.Image
				childImage, consumed, rt, baseline := c.renderLineBox(width-dot.X-rfWidth, remainingTextContent, false, c.NextSibling != nil && c.NextSibling.GetDisplayProp() == "inline")
				if consumed == "" {
					if dot.X == 0 {
						// Nothing was consumed, and we're at the start of the line, so just
						// force at least one word to be drawn.

						childImage, consumed, rt, baseline = c.renderLineBox(width-dot.X-rfWidth, remainingTextContent, true, c.NextSibling != nil && c.NextSibling.GetDisplayProp() == "inline")
					} else {
						// Go to the next line.
						dot.Y += *nextline
						*nextline = c.GetLineHeight()

						lfWidth = e.leftFloats.MaxX(*dot)
						rfWidth = e.rightFloats.WidthAt(*dot)

						dot.X = e.leftFloats.MaxX(*dot)
						e.advanceLine(dot)
						dot.X += c.listIndent()
						goto textdraw
					}
				}

				size := childImage.Bounds().Size()
				size.Y = c.GetLineHeight() - c.GetBorderBottomWidth() - c.GetBorderTopWidth()
				borderImage, cr := c.calcCSSBox(size)
				sr := childImage.Bounds()
				r = image.Rectangle{*dot, dot.Add(sr.Size())}

				// dot is a point, but the font drawer uses it as the baseline,
				// not the corner to draw at, so WidthAt can be a little unreliable.
				// Check if the new rectangle overlaps with any floats, and if it does then
				// increase according to the box it's overlapping and try again.
				for _, float := range e.leftFloats {
					if r.Overlaps(float.BoxDrawRectangle) {
						dot.Y += *nextline
						*nextline = c.GetLineHeight()

						lfWidth = e.leftFloats.MaxX(*dot)
						rfWidth = e.rightFloats.WidthAt(*dot)

						dot.X = e.leftFloats.MaxX(*dot)
						e.advanceLine(dot)
						continue textdraw
					}
				}
				// same for right floats
				for _, float := range e.rightFloats {
					if r.Overlaps(float.BoxDrawRectangle) {
						dot.Y += *nextline
						*nextline = c.GetLineHeight()

						lfWidth = e.leftFloats.MaxX(*dot)
						rfWidth = e.rightFloats.WidthAt(*dot)

						dot.X = e.leftFloats.MaxX(*dot)
						e.advanceLine(dot)
						continue textdraw
					}
				}

				// Nothing overlapped, so use this line box.
				remainingTextContent = rt

				overlayed.GrowBounds(r)

				lb := lineBox{childImage, borderImage, *dot, cr.Min, baseline, consumed, c}
				e.lineBoxes = append(e.lineBoxes, &lb)
				e.curLine = append(e.curLine, &lb)
				switch e.GetWhiteSpace() {
				case "pre":
					dot.Y += *nextline
					*nextline = c.GetLineHeight()
					dot.X = lfWidth
				case "nowrap":
					fallthrough
				case "normal":
					fallthrough
				default:
					if r.Max.X >= width-rfWidth {
						// there's no space left on this line, so advance dot to the next line.
						dot.Y += *nextline
						*nextline = c.GetLineHeight()

						lfWidth = e.leftFloats.MaxX(*dot)
						rfWidth = e.rightFloats.WidthAt(*dot)

						dot.X = lfWidth
						e.advanceLine(dot)

						// Leave space for the list marker. This is the same amount
						// added by the stylesheet in Firefox
						dot.X += c.listIndent()
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
				*nextline = c.GetLineHeight()
				e.advanceLine(dot)
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
				c.leftFloats = e.leftFloats
				c.rightFloats = e.rightFloats
				childContent, newDot := c.LayoutPass(width, image.ZR, &image.Point{dot.X, dot.Y}, nextline)
				c.ContentOverlay = childContent
				_, contentbox := c.calcCSSBox(childContent.Bounds().Size())
				c.BoxContentRectangle = contentbox
				overlayed.GrowBounds(contentbox)

				// Populate this image map. This is an inline, so we actually only care
				// about the line boxes that were generated by the children.
				childImageMap := c.ImageMap
				for _, area := range childImageMap {
					// translate the coordinate systems from the child's to this one
					newArea := area.Area

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
				if dot.X != e.leftFloats.MaxX(*dot) && display != "inline-block" {
					// This means the previous child was an inline item, and we should position dot
					// as if there were an implicit box around it.
					// floated elements don't affect dot, so only do this if it's not floated.
					if float == "none" {
						dot.X = 0
						if c.PrevSibling != nil {
							dot.Y += *nextline
							*nextline = c.GetLineHeight()
						}
					}
				}

				// draw the border, background, and CSS outer box.
				var lh int

				cdot := image.Point{}
				// Collapse margins before doing layout
				if display != "inline-block" {
					if tm := c.GetMarginTopSize(); tm != 0 && bottommargin != 0 && c.GetPaddingTop() == 0 && c.GetBorderTopWidth() == 0 {
						// collapse margins
						if tm > bottommargin {
							// Remove the bottom margin that was already added,
							// because this top margin is bigger
							dot.Y -= bottommargin
						} else {
							// Remove the new top margin, because the previous
							// bottom margin was bigger.
							dot.Y -= tm
						}

						// If it's the first child of a collapsed margin, also collapse the margin in the child container.
						if c.PrevSibling == nil {
							if tm > bottommargin {
								cdot.Y -= bottommargin
							} else {
								cdot.Y -= tm
							}
						}
					}
					if mt := c.GetMarginTopSize(); mt < 0 {
						dot.Y += mt
					}
				}

				var childContent image.Image
				// collapse child margins if applicable
				childContent, _ = c.LayoutPass(width, image.ZR, &cdot, &lh)
				c.ContentOverlay = childContent
				box, contentbox := c.calcCSSBox(childContent.Bounds().Size())
				c.BoxContentRectangle = contentbox
				sr := box.Bounds()

				if fdot.Y <= dot.Y {
					fdot.Y = dot.Y
					fdot.X = dot.X
				}
				r := image.Rectangle{fdot, fdot.Add(sr.Size())}
			positionFloats:
				switch float {
				case "right":
					size := sr.Size()
					rightFloat := e.rightFloats.ClearFloats(fdot)
					leftFloat := e.leftFloats.ClearFloats(fdot)
					rightFloatX := width - rightFloat.WidthAt(fdot)
					r = image.Rectangle{
						Min: image.Point{rightFloatX - size.X, fdot.Y},
						Max: image.Point{rightFloatX, size.Y + fdot.Y},
					}
					leftFloatX := leftFloat.MaxX(fdot)
					fdot.X = leftFloatX
					if r.Min.X <= leftFloatX {
						lfHeight := leftFloat.ClearFloats(fdot).NextFloatHeight()
						rfHeight := rightFloat.ClearFloats(fdot).NextFloatHeight()

						if lfHeight > 0 && (lfHeight <= rfHeight || rfHeight == 0) {
							fdot.Y += lfHeight
							fdot.X = leftFloat.MaxX(fdot)
						} else if rfHeight > 0 {
							fdot.Y += rfHeight
							fdot.X = leftFloat.MaxX(fdot)
						} else {
							panic("Clearing floats didn't make any space.")
						}

						rightFloatX = rightFloat.WidthAt(fdot)
						r = image.Rectangle{
							Min: image.Point{width - rightFloatX - size.X, fdot.Y},
							Max: image.Point{width - rightFloatX, size.Y + fdot.Y},
						}
					}

					for _, line := range e.lineBoxes {
						// Check if the box overlaps with any line. If so, just
						// move the float down, because there's definitely no room on
						// this line because we're dealing with a right float.
						lineBounds := image.Rectangle{
							line.origin,
							line.origin.Add(line.Content.Bounds().Size()),
						}
						if r.Overlaps(lineBounds) {
							fdot.Y += *nextline
							*nextline = c.GetLineHeight()
							fdot.X = 0
							// Recalculate all the rectangle offsets and floating
							// dot stuff now that we've moved the box (again)
							goto positionFloats
						}
					}
				case "left":
					size := sr.Size()
					leftFloatX := e.leftFloats.MaxX(fdot)
					fdot.X = leftFloatX
					r = image.Rectangle{
						Min: image.Point{leftFloatX, fdot.Y},
						Max: image.Point{leftFloatX + size.X, size.Y + fdot.Y},
					}
					if r.Max.X >= width {
						lfHeight := e.leftFloats.ClearFloats(fdot).NextFloatHeight()
						rfHeight := e.rightFloats.ClearFloats(fdot).NextFloatHeight()

						if lfHeight > 0 && (lfHeight <= rfHeight || rfHeight == 0) {
							fdot.Y += lfHeight
							fdot.X = e.leftFloats.MaxX(fdot)
						} else if rfHeight > 0 {
							fdot.Y += rfHeight
							fdot.X = e.leftFloats.MaxX(fdot)
						} else {
							panic("Clearing floats didn't make any space.")
						}

						leftFloatX = e.leftFloats.MaxX(fdot)
						r = image.Rectangle{
							Min: image.Point{leftFloatX, fdot.Y},
							Max: image.Point{leftFloatX + size.X, size.Y + fdot.Y},
						}
					}

					if len(e.lineBoxes) > 0 {
						// We need to do two passes for left floats: one to check if we can just move the
						// text or if we need to move the box itself, and another to actually move it/them.
						hasOverlap := false
						canMoveText := true
						for _, line := range e.lineBoxes {
							lineBounds := image.Rectangle{
								line.origin,
								line.origin.Add(line.Content.Bounds().Size()),
							}
							if r.Overlaps(lineBounds) {
								hasOverlap = true
								lsize := lineBounds.Size()
								o := line.origin
								if r.Max.X+lsize.X+e.rightFloats.ClearFloats(o).WidthAt(o) >= width {
									canMoveText = false
									break
								}
							}
						}

						// Second pass: move text or float
						if hasOverlap {
							if canMoveText {
								// Move the text
								nd := r.Max.X
								for i, line := range e.lineBoxes {
									lineBounds := image.Rectangle{
										line.origin,
										line.origin.Add(line.Content.Bounds().Size()),
									}
									if r.Overlaps(lineBounds) {
										line.origin.X = nd
										e.lineBoxes[i] = line

										lsize := lineBounds.Size()
										o := line.origin

										// Check if there's still room for more text when we begin rendering again
										// and move dot to the appropriate place.
										if r.Max.X+lsize.X+e.rightFloats.ClearFloats(o).WidthAt(o) >= width {
											// There's no more space, adjust dot
											// to the next line.
											dot.Y += *nextline
											*nextline = c.GetLineHeight()
											dot.X = e.leftFloats.MaxX(*dot)
										} else {
											// There's still space on this line,
											// so move dot over.
											dot.X = e.leftFloats.MaxX(o) + r.Size().X + lsize.X
											nd = dot.X
										}
									}
								}
							} else {
								// No room to move the text into, so move the
								// box down instead.
								fdot.Y += *nextline
								fdot.X = e.leftFloats.MaxX(fdot)
								goto positionFloats
							}
						}
					}
				default:
				}
				overlayed.GrowBounds(r)
				c.BoxDrawRectangle = r
				if sr.Size().Y > 0 {
					switch float {
					case "left":
						e.leftFloats = append(e.leftFloats, c)
					case "right":
						e.rightFloats = append(e.rightFloats, c)
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

				switch float {
				case "left", "right":
					// floated boxes don't affect dot
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

						if c.GetPaddingBottom() == 0 && c.GetBorderBottomWidth() == 0 {
							bottommargin = c.GetMarginBottomSize()
						} else {
							bottommargin = 0
						}

						if mb := c.GetMarginBottomSize(); mb < 0 {
							dot.Y += mb
							bottommargin = 0
						}
					}

				}
			}
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
			for _, box := range e.lineBoxes {
				sr := box.Content.Bounds()
				r := image.Rectangle{box.origin, box.origin.Add(sr.Size())}
				if box.BorderImage != nil {
					sr := box.BorderImage.Bounds()
					ro := box.origin.Sub(box.borigin)
					r := image.Rectangle{ro, ro.Add(sr.Size())}
					draw.Draw(
						e.OverlayedContent,
						r,
						box.BorderImage,
						sr.Min,
						draw.Over,
					)
				}
				draw.Draw(e.OverlayedContent, r, box.Content, sr.Min, draw.Over)
			}
		case html.ElementNode:
			if c.Data == "br" {
				continue
			}
			switch display := c.GetDisplayProp(); display {
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
			case "block", "table", "table-inline", "list-item":
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
					if display == "list-item" {
						e.numBullets++
						c.drawBullet(e.OverlayedContent, c.BoxDrawRectangle.Min, c.BoxContentRectangle.Min, e.numBullets)
					}
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

func (e *RenderableDomElement) drawBullet(dst draw.Image, drawRectangle, contentRectangle image.Point, bulletNum int) {
	clr := e.GetColor()
	fSize := e.GetFontSize()
	fontFace := e.GetFontFace(fSize)
	fntDrawer := font.Drawer{
		Dst:  dst,
		Src:  &image.Uniform{clr},
		Face: fontFace,
	}
	// FIXME: Add list-style-type support
	// normal bullet:
	bullet := "\u2219"
	if e.GetListStyleType() == "decimal" {
		bullet = fmt.Sprintf("%d.", bulletNum)
	}
	bulletSize := fntDrawer.MeasureString(bullet)
	fontMetrics := fontFace.Metrics()
	bulletOffset := e.listIndent()
	fntDrawer.Dot = fixed.P(
		// Center the X coordinate in the middle of the empty space between the draw rectangle
		// and the content rectangle
		drawRectangle.X+contentRectangle.X+bulletOffset-(20-bulletSize.Ceil()/2),
		// And have the height at the top of the first line.
		drawRectangle.Y+contentRectangle.Y+fontMetrics.Ascent.Floor(),
	)
	fntDrawer.DrawString(bullet)
}
func (e *RenderableDomElement) advanceLine(dot *image.Point) {
	if len(e.curLine) > 1 {
		// If there was more than 1 element, re-adjust all their positions with respect to
		// the vertical-align property.
		baseline := 0
		maxsize := 0

		// FIXME: This is a hack. The baseline/lineheight should be calculated in accordance with
		// the CSS spec.
		// Step 1. Figure out how big the line really is and where the baseline is.
		for _, l := range e.curLine {
			if height := l.Content.Bounds().Size().Y; height > maxsize {
				maxsize = height
			}
			if l.content != "[img]" && l.baseline > baseline {
				baseline = l.baseline
			}
		}

		// Step 2: Adjust the image origins with respect to the baseline.
		for _, l := range e.curLine {
			switch align := l.el.GetVerticalAlign(); align {
			case "text-bottom":
				l.origin.Y += maxsize - l.Content.Bounds().Size().Y
			case "text-top":
				l.origin.Y += maxsize - baseline
			case "middle":
				l.origin.Y += (maxsize / 2)
			default:
				l.origin.Y += maxsize - l.Content.Bounds().Size().Y
			}
			if s := l.Content.Bounds().Size().Y; l.origin.Y+s > dot.Y {
				// Nothing, it was already aligned to the top
				dot.Y = l.origin.Y + s
			}
		}

	}
	e.curLine = make([]*lineBox, 0)
}
func (e *RenderableDomElement) listIndent() int {
	if e == nil {
		return 0
	}
	i := 0
	for c := e; c != nil; c = c.Parent {
		if c.GetDisplayProp() == "list-item" {
			i += 40
		}
	}
	return i
}
