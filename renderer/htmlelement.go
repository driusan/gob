package renderer

import (
	"github.com/driusan/gob/css"
	"github.com/driusan/gob/dom"
	"github.com/driusan/gob/net"

	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
	"golang.org/x/net/html"

	"github.com/nfnt/resize"

	"context"
	"fmt"
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
	"unicode/utf8"
	//"strconv"
)

// A RenderableElement is something that can be rendered to
// an image.
type Renderer interface {
	// Lays out an element in preparation for rendering
	Layout(ctx context.Context, viewportSize image.Point) error

	// Draws the element into dst, scrolled so that the top left of the
	// image is at cursor. Layout must be called before RenderInto.
	RenderInto(ctx context.Context, dst draw.Image, cursor image.Point) error
}

type RenderableDomElement struct {
	*dom.Element
	Styles            *css.StyledElement
	ConditionalStyles struct {
		Unconditional          *css.StyledElement
		FirstLine, FirstLetter *css.StyledElement
	}

	Parent      *RenderableDomElement
	FirstChild  *RenderableDomElement
	NextSibling *RenderableDomElement
	PrevSibling *RenderableDomElement

	CSSOuterBox    image.Image
	ContentOverlay image.Image

	// The location within the parent to draw the OverlayedContent
	//DrawRectangle    image.Rectangle
	//BoxOrigin        image.Point
	BoxDrawRectangle    image.Rectangle
	BoxContentRectangle image.Rectangle
	floatAdjusted       bool

	ImageMap     ImageMap
	PageLocation *url.URL
	contentWidth int

	containerWidth  int
	containerHeight int

	// Used to determine if the left border should be drawn
	// when inlines span multiple lines.
	inlineStart bool
	curLine     []*lineBox
	lineBoxes   []*lineBox

	resolver   net.URLReader
	layoutDone bool

	leftFloats, rightFloats FloatStack

	// The number of child bullet items that have been drawn.
	// Used for determining the counter to place next to the list item
	numBullets int

	State css.State
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

	styles  *css.StyledElement
	origin  image.Point
	borigin image.Point
	metrics *font.Metrics

	content string
	el      *RenderableDomElement
}

func (lb lineBox) IsImage() bool {
	return lb.metrics == nil
}

func (lb lineBox) Baseline() int {
	if lb.metrics != nil {
		return lb.metrics.Ascent.Ceil()
	}
	// It's an image, the bottom is the baseline.
	return lb.Height()
}

func (lb lineBox) LineHeight() int {
	lb.el.Styles = lb.styles
	return lb.el.GetLineHeight()
}

func (lb lineBox) Height() int {
	lb.el.Styles = lb.styles
	if lb.IsImage() {
		if lb.BorderImage != nil {
			// The border image already includes padding and border,
			// but not top/bottom margin
			return lb.BorderImage.Bounds().Size().Y + lb.el.GetMarginTopSize() + lb.el.GetMarginBottomSize()
		}
		return lb.Content.Bounds().Size().Y + lb.el.GetMarginTopSize() + lb.el.GetMarginBottomSize()
	}
	return (lb.metrics.Height).Ceil() + lb.el.GetMarginTopSize() + lb.el.GetMarginBottomSize()
}

// Returns the width of this linebox. This is primarily used for testing.
func (lb lineBox) width() int {
	if lb.IsImage() {
		return lb.BorderImage.Bounds().Size().Y
	}
	fntDrawer, fSize := lb.getFontDrawer(nil, image.ZP)
	defer fntDrawer.Face.Close()
	return lb.measureOrDraw(true, &fntDrawer, fSize).Ceil()
}

func (lb lineBox) Bounds() image.Rectangle {
	if lb.BorderImage != nil {
		return lb.BorderImage.Bounds().Add(lb.origin)
	}
	return lb.Content.Bounds().Add(lb.origin)

}

// Gets a font drawer for this linebox to draw into draw.Image at dot.
// This opens a font.Face which it's the caller's responsibility to close.
func (lb lineBox) getFontDrawer(dst draw.Image, dot image.Point) (drawer font.Drawer, fSize int) {
	// Reapply the styles that were in effect when this was being laid out.
	lb.el.Styles = lb.styles

	fSize = lb.el.GetFontSize()
	fontFace := lb.el.GetFontFace(fSize)

	clr := lb.el.GetColor()
	drawer = font.Drawer{
		Dst:  dst,
		Src:  &image.Uniform{clr},
		Face: fontFace,
		Dot:  fixed.P(dot.X, dot.Y+lb.metrics.Ascent.Ceil()),
	}
	return
}

func (lb lineBox) measureOrDraw(measure bool, fntDrawer *font.Drawer, fSize int) fixed.Int26_6 {
	fontFace := fntDrawer.Face
	defer fontFace.Close()

	smallcaps := lb.el.FontVariant() == "small-caps"

	var smallFace font.Face
	if smallcaps {
		smallFace = lb.el.GetFontFace(fSize * 8 / 10)

		defer smallFace.Close()
	}

	rv := fixed.I(0)

	switch whitespace := lb.el.GetWhiteSpace(); whitespace {
	case "pre-wrap", "no-wrap", "pre-line":
		panic(whitespace + " not implemented")
	case "pre":
		if measure {
			return fntDrawer.MeasureString(lb.content)
		} else {
			fntDrawer.DrawString(lb.content)
		}
	case "normal":
		fallthrough
	default:
		words := strings.Fields(lb.content)
		// layout ensured that it fit and did most necessary text
		// transformations, so we just need to make sure we handle
		// whitespace in the same way and don't check anything else
		for i, word := range words {
			if smallcaps {
				var wordleft string = word
				for wordleft != "" {
					r, n := utf8.DecodeRune([]byte(wordleft))
					chr := wordleft[:n]
					wordleft = wordleft[n:]
					if unicode.IsLower(r) {
						fntDrawer.Face = smallFace
					} else {
						fntDrawer.Face = fontFace
					}
					chr = strings.ToUpper(chr)
					if measure {
						rv += fntDrawer.MeasureString(word)
					} else {
						fntDrawer.DrawString(chr)
					}

				}
			} else {
				if measure {
					rv += fntDrawer.MeasureString(word)
				} else {
					fntDrawer.DrawString(word)
				}
			}
			if i == len(words)-1 {
				break
			}
			// Add a three per em between words, an em-space after a period, and
			// an en-space after any other punctuation.
			space := fixed.I(0)

			switch word[len(word)-1] {
			case ',', ';', ':', '!', '?':
				space = fixed.Int26_6(fSize/2) << 6
			case '.':
				space = fixed.Int26_6(fSize) << 6
			default:
				space = fixed.Int26_6(fSize/3) << 6
			}
			if measure {
				rv += space
			} else {
				fntDrawer.Dot.X += space
			}
		}
	}
	return rv
}

func (lb lineBox) drawAt(ctx context.Context, dst draw.Image, dot image.Point) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	lb.el.Styles = lb.styles

	clr := lb.el.GetColor()
	fntDrawer, fSize := lb.getFontDrawer(dst, dot)
	defer fntDrawer.Face.Close()

	lb.measureOrDraw(false, &fntDrawer, fSize)

	if decoration := lb.el.GetTextDecoration(); decoration != "" && decoration != "none" && decoration != "blink" {
		if strings.Contains(decoration, "underline") {
			y := fntDrawer.Dot.Y.Floor() + 1
			for px := dot.X; px < fntDrawer.Dot.X.Ceil(); px++ {
				dst.Set(px, y, clr)
			}
		}
		if strings.Contains(decoration, "overline") {
			y := dot.Y
			for px := dot.X; px < fntDrawer.Dot.X.Ceil(); px++ {
				dst.Set(px, y, clr)
			}
		}
		if strings.Contains(decoration, "line-through") {
			y := dot.Y + lb.metrics.Ascent.Floor()/2
			for px := dot.X; px < fntDrawer.Dot.X.Ceil(); px++ {
				dst.Set(px, y, clr)
			}
		}
	}

	return nil
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

func (e RenderableDomElement) layoutLineBox(remainingWidth int, textContent string, force bool, inlinesibling bool, firstletter bool) (size image.Point, consumed, unconsumed string, metrics font.Metrics, forcenewline bool) {
	if remainingWidth < 0 {
		panic("No room to render text")
	}
	if firstletter {
		e.Styles = e.ConditionalStyles.FirstLetter
	}
	switch e.GetTextTransform() {
	case "capitalize":
		textContent = strings.Title(textContent)
	case "uppercase":
		textContent = strings.ToUpper(textContent)
	case "lowercase":
		textContent = strings.ToLower(textContent)
	}

	smallcaps := e.FontVariant() == "small-caps"

	fSize := e.GetFontSize()
	fontFace := e.GetFontFace(fSize)
	defer fontFace.Close()
	metrics = fontFace.Metrics()

	var smallFace font.Face
	if smallcaps {
		smallFace = e.GetFontFace(fSize * 8 / 10)

		defer smallFace.Close()
	}

	fntDrawer := font.Drawer{
		Dst:  nil,
		Src:  nil,
		Face: fontFace,
	}

	switch whitespace := e.GetWhiteSpace(); whitespace {
	case "pre-wrap", "no-wrap", "pre-line":
		panic(whitespace + " not implemented")
	case "pre":
		lines := strings.Split(textContent, "\n")
		consumed = lines[0]
		unconsumed = strings.Join(lines[1:], "\n")
		size = image.Point{
			fntDrawer.MeasureString(consumed).Ceil(),
			(metrics.Ascent + metrics.Descent).Ceil(),
		}
		return
	case "normal":
		fallthrough
	default:
		var sz fixed.Int26_6
		words := strings.Fields(textContent)
		for i, word := range words {
			pieces := strings.Split(word, "-")
			var face font.Face = fontFace
			for pi, piece := range pieces {
				var wsize fixed.Int26_6
				if smallcaps || firstletter {
					// Handle the word size letter by letter for smallcaps
					// or firstletter
					var wordleft = piece
					for wordleft != "" {
						r, n := utf8.DecodeRune([]byte(wordleft))
						wordleft = wordleft[n:]
						if smallcaps {
							if unicode.IsLower(r) {
								face = smallFace
							} else {
								face = fontFace
							}
							r = unicode.ToUpper(r)
						}
						// FIXME: This should also take kerning into account
						lsize, ok := face.GlyphAdvance(r)
						if !ok {
							// FIXME: Have a better fallback.
							panic("No glyph for rune in font")
						}
						wsize += lsize
						if firstletter {
							consumed += fmt.Sprintf("%c", r)
							if unicode.IsLetter(r) || force {
								wl := strings.Join(append([]string{wordleft}, pieces[pi+1:]...), "-")
								unconsumed = strings.Join(append([]string{wl}, words[i+1:]...), " ")
								size = image.Point{(wsize + sz).Ceil(), (metrics.Ascent + metrics.Descent).Ceil()}

								return
							}
						}
					}
				} else {
					wsize = fntDrawer.MeasureString(piece)
				}
				if pi != len(pieces)-1 && len(pieces) > 1 {
					dsize, ok := face.GlyphAdvance('-')
					if !ok {
						panic("Font face does not have glyph for -")
					}
					wsize += dsize
				}

				if (wsize + sz).Ceil() > remainingWidth {
					unconsumed = strings.Join(pieces[pi:], "-")
					unconsumed += " " + strings.Join(words[i+1:], " ")
					if i == 0 && pi == 0 && force {
						consumed = words[0]
						unconsumed = strings.Join(words[i+1:], " ")
						return image.Point{wsize.Ceil(), (metrics.Ascent + metrics.Descent).Ceil()}, consumed, unconsumed, metrics, true
					} else {
						forcenewline = true
						size = image.Point{sz.Ceil(), (metrics.Ascent + metrics.Descent).Ceil()}
						return
					}
				}
				// FIXME: Replace with string builder
				if !firstletter {
					consumed += piece
				}
				if pi != len(pieces)-1 && len(pieces) > 1 {
					consumed += "-"
				}
				sz += wsize
			}
			if i == len(words)-1 {
				break
			}

			osz := sz
			// Add a three per em between words, an em-space after a period, and
			// an en-space after any other punctuation.
			switch word[len(word)-1] {
			case ',', ';', ':', '!', '?':
				sz += fixed.Int26_6(fSize/2) << 6
			case '.':
				sz += fixed.Int26_6(fSize) << 6
			default:
				sz += fixed.Int26_6(fSize/3) << 6
			}

			// If the whitespace is going to put us over
			// the line, don't add it because the line
			// break acts as the whitespace and we don't
			// want the whitespace to overlap with floats
			if sz.Ceil() > remainingWidth {
				sz = osz
				unconsumed = strings.Join(words[i+1:], " ")

				break
			}
			consumed += " "
		}
		forcenewline = false
		size = image.Point{sz.Ceil(), (metrics.Ascent + metrics.Descent).Ceil()}
		return
	}
	panic("Unhandled whitespace property")
}

// Lays out the element into a viewport of size viewportSize.
func (e *RenderableDomElement) Layout(ctx context.Context, viewportSize image.Point) error {
	e.leftFloats = make(FloatStack, 0)
	e.rightFloats = make(FloatStack, 0)
	e.layoutPass(ctx, viewportSize.X, image.ZR, &image.Point{0, 0})
	return nil
}

func (e *RenderableDomElement) RenderInto(ctx context.Context, dst draw.Image, cursor image.Point) error {
	if !e.layoutDone {
		return fmt.Errorf("Element not yet laid out.")
	}
	return e.drawInto(ctx, dst, cursor)
}

func (e *RenderableDomElement) InvalidateLayout() {
	if e == nil {
		return
	}
	e.layoutDone = false
	e.CSSOuterBox = nil

	e.ContentOverlay = nil
	e.lineBoxes = nil
	e.leftFloats = nil
	e.rightFloats = nil
	e.curLine = nil
	e.lineBoxes = nil

	if e.FirstChild != nil {
		e.FirstChild.InvalidateLayout()
	}
	if e.NextSibling != nil {
		e.NextSibling.InvalidateLayout()
	}
}

func (e *RenderableDomElement) layoutPass(ctx context.Context, containerWidth int, r image.Rectangle, dot *image.Point) (image.Image, image.Point) {
	var overlayed *DynamicMemoryDrawer
	if e.layoutDone {
		return overlayed, image.Point{}
	}
	defer func() {
		e.layoutDone = true
	}()

	width := e.GetContainerWidth(containerWidth)
	e.contentWidth = width
	e.containerWidth = containerWidth

	fdot := image.Point{dot.X, dot.Y}
	height := e.GetMinHeight()

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

			if css := e.Styles.Width.Value; css != "" {
				if w := e.GetWidth(); w > 0 {
					ewidth = true
					iwidth = w
					if !eheight {
						iheight = 0
					}
				}
			}
			if css := e.Styles.Height.Value; css != "" {
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
			if loadedImage {
				e.contentWidth = iwidth
				box, contentbox := e.calcCSSBox(e.ContentOverlay.Bounds().Size(), false, false)
				e.BoxContentRectangle = contentbox
				e.BoxDrawRectangle = image.Rectangle{
					*dot,
					dot.Add(box.Bounds().Size()),
				}
				switch e.GetFloat() {
				case "right":
					e.Parent.positionRightFloat(*dot, box.Bounds().Size(), e)
				case "left":
					e.Parent.positionLeftFloat(dot, &fdot, box.Bounds().Size(), e)
				default:
					sz := box.Bounds().Size()
					iheight = sz.Y + e.GetMarginTopSize() + e.GetMarginBottomSize()

					switch e.GetDisplayProp() {
					case "block":
						dot.Y += sz.Y
						dot.X = 0
					default:
						fallthrough
					case "inline", "inline-block":
						p := e.getContainingBlock()

						if len(p.curLine) > 0 && e.BoxDrawRectangle.Max.X > containerWidth {
							p.advanceLine(dot)
							dot.X = 0
							e.BoxDrawRectangle = image.Rectangle{
								*dot,
								dot.Add(box.Bounds().Size()),
							}

						}

						lb := lineBox{
							e.ContentOverlay,
							box,
							e.Styles,
							*dot,
							contentbox.Min,
							nil,
							"[img]",
							e,
						}
						// The line block goes with the
						// nearest block. The inline may
						// be a child of another inline.
						p.lineBoxes = append(p.lineBoxes, &lb)
						p.curLine = append(p.curLine, &lb)
						dot.X += sz.X

						// Negative margins were not included
						// in the CSS Box (so that they
						// get drawn), add them back.
						if lm := e.GetMarginLeftSize(); lm < 0 {
							dot.X += lm
						}
						if rm := e.GetMarginRightSize(); rm < 0 {
							dot.X += rm
						}
					}
				}
			}
			return e.ContentOverlay, *dot
		}
	}

	overlayed = NewDynamicMemoryDrawer(image.Rectangle{image.ZP, image.Point{width, height}})

	firstLine := true
	firstletter := !e.inlineStart
	imageMap := NewImageMap()

	for c := e.FirstChild; c != nil; c = c.NextSibling {
		if ctx.Err() != nil {
			return nil, image.ZP
		}
		if fdot.Y < dot.Y {
			fdot = *dot
		}
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
			// because their style should be inherited from their
			// parent.
			c.Styles = e.Styles
			c.ConditionalStyles = e.ConditionalStyles

			// Exception: Inline element borders get applied to each line, but
			// if the parent is a block we shouldn't inherit the block border properties,
			// because there's an anonymous inline with no border properties around it.
			if c.Parent.GetDisplayProp() != "inline" {
				// FIXME: Move these into the GetX functions instead of doing so much
				// memory copying.
				c.ConditionalStyles.Unconditional = new(css.StyledElement)
				c.ConditionalStyles.FirstLine = new(css.StyledElement)
				c.ConditionalStyles.FirstLetter = new(css.StyledElement)
				c.Styles = new(css.StyledElement)

				*c.ConditionalStyles.Unconditional = *e.ConditionalStyles.Unconditional
				*c.ConditionalStyles.FirstLine = *e.ConditionalStyles.FirstLine
				*c.ConditionalStyles.FirstLetter = *e.ConditionalStyles.FirstLetter
				*c.Styles = *e.Styles

				resetboxprop := func(el *css.StyledElement) {
					el.BorderLeftWidth.Value = "0"
					el.BorderRightWidth.Value = "0"
					el.BorderTopWidth.Value = "0"
					el.BorderBottomWidth.Value = "0"

					el.PaddingLeft.Value = "0"
					el.PaddingRight.Value = "0"
					el.PaddingTop.Value = "0"
					el.PaddingBottom.Value = "0"

					el.MarginTop.Value = "0"
					el.MarginLeft.Value = "0"
					el.MarginRight.Value = "0"
					el.MarginBottom.Value = "0"

					el.BackgroundColor.Value = "transparent"
					el.BackgroundImage.Value = ""
				}

				resetboxprop(c.Styles)
				resetboxprop(c.ConditionalStyles.Unconditional)
				resetboxprop(c.ConditionalStyles.FirstLine)
				resetboxprop(c.ConditionalStyles.FirstLetter)

			}

			lfWidth := e.leftFloats.WidthAt(*dot)
			rfWidth := e.rightFloats.WidthAt(*dot)
			if dot.X < lfWidth {
				dot.X = lfWidth
			}

			dot.X += c.listIndent()
			if firstLine == true {
				c.Styles = c.ConditionalStyles.FirstLine
				dot.X += c.GetTextIndent(width)
				firstLine = false
			}
			if firstletter {
				c.Styles = c.ConditionalStyles.FirstLetter
			}
			remainingTextContent := c.Data
			ws := e.GetWhiteSpace()
		textdraw:
			for strings.TrimSpace(remainingTextContent) != "" {
				for width-dot.X-rfWidth-c.GetPaddingLeft()-c.GetBorderLeftWidth() <= 0 {
					e.advanceLine(dot)

					lfWidth = e.leftFloats.WidthAt(*dot)
					rfWidth = e.rightFloats.WidthAt(*dot)
					dot.X = e.leftFloats.MaxX(*dot)
					if dot.X == 0 && lfWidth == 0 && rfWidth == 0 {
						break
					}
				}
				if width-dot.X-rfWidth-c.GetPaddingLeft()-c.GetBorderLeftWidth() <= 0 {
					panic("No room for text")
				}

				// dot works differently for line boxes than for
				// blocks. For a block, dot represents the top left
				// corner of the border box. For a linebox, it represents
				// the top left corner of the text and the border
				// is drawn at an offset back from that, so we
				// need to add the padding and border to dot.X.
				dot.X += c.GetPaddingLeft() + c.GetBorderLeftWidth()
				var childImage image.Image
				size, consumed, rt, metrics, forcenewline := c.layoutLineBox(width-dot.X-rfWidth, remainingTextContent, false, c.NextSibling != nil && c.NextSibling.GetDisplayProp() == "inline", firstletter)
				if consumed == "" {
					if ws == "pre" {
						// There was a blank pre-formatted line, so just consume it.
						e.advanceLine(dot)
						e.Styles = e.ConditionalStyles.Unconditional
						c.Styles = c.ConditionalStyles.Unconditional
						remainingTextContent = rt
						e.inlineStart = false
						continue
					}
					if dot.X == 0 {
						// Nothing was consumed and we're
						// at the start of the line, so force a word.
						size, consumed, rt, metrics, forcenewline = c.layoutLineBox(width-dot.X-rfWidth, remainingTextContent, true, c.NextSibling != nil && c.NextSibling.GetDisplayProp() == "inline", firstletter)
					} else {
						// Advance a line and try again.
						e.advanceLine(dot)
						e.Styles = e.ConditionalStyles.Unconditional
						c.Styles = c.ConditionalStyles.Unconditional

						lfWidth = e.leftFloats.WidthAt(*dot)
						rfWidth = e.rightFloats.WidthAt(*dot)

						dot.X = e.leftFloats.MaxX(*dot)

						continue textdraw
					}
				}
				childImage = image.Rectangle{*dot, dot.Add(size)}
				if consumed == "" && ws != "pre" {
					panic("This should be impossible.")
				}

				borderImage, cr := c.calcCSSBox(size, !e.inlineStart, strings.TrimSpace(rt) != "")

				// We only grow the bounds by the amount that
				// doesn't have the border drawn, so this uses
				// size and not borderImage size.
				bz := size
				bz.Y = c.GetLineHeight()
				start := dot.Add(image.Point{0, c.GetMarginTopSize()})
				r = image.Rectangle{start, start.Add(bz)}

				// dot is a point, but the font drawer uses it as the baseline,
				// not the corner to draw at, so WidthAt can be a little unreliable.
				// Check if the new rectangle overlaps with any floats, and if it does then
				// increase according to the box it's overlapping and try again.
				for _, float := range e.leftFloats {
					if r.Overlaps(float.BoxDrawRectangle) {
						e.advanceLine(dot)
						rfWidth = e.rightFloats.WidthAt(*dot)
						lfWidth = e.leftFloats.WidthAt(*dot)
						continue textdraw
					}
				}
				// same for right floats
				for _, float := range e.rightFloats {
					if r.Overlaps(float.BoxDrawRectangle) {
						e.advanceLine(dot)
						lfWidth = e.leftFloats.WidthAt(*dot)
						rfWidth = e.rightFloats.WidthAt(*dot)
						continue textdraw
					}
				}

				// Nothing overlapped, so use this line box.
				remainingTextContent = rt
				e.inlineStart = false

				//overlayed.GrowBounds(r)
				overlayed.GrowBounds(r)

				lb := lineBox{
					Content:     childImage,
					BorderImage: borderImage,
					styles:      c.Styles,
					origin:      *dot,
					borigin:     cr.Min,
					metrics:     &metrics,
					content:     consumed,
					el:          c,
				}

				pc := c.getContainingBlock()
				pc.lineBoxes = append(pc.lineBoxes, &lb)
				pc.curLine = append(pc.curLine, &lb)

				switch e.GetWhiteSpace() {
				case "pre":
					e.advanceLine(dot)
					e.Styles = e.ConditionalStyles.Unconditional
					c.Styles = c.ConditionalStyles.Unconditional
				case "nowrap":
					fallthrough
				case "normal":
					fallthrough
				default:
					if r.Max.X >= width-rfWidth || forcenewline {
						// there's no space left on this line, so advance dot to the next line.
						e.advanceLine(dot)

						lfWidth = e.leftFloats.WidthAt(*dot)
						rfWidth = e.rightFloats.WidthAt(*dot)

						e.Styles = e.ConditionalStyles.Unconditional
						c.Styles = c.ConditionalStyles.Unconditional
					} else {
						// there's still space on this line, so move dot to the end
						// of the rendered text.
						dot.X = r.Max.X
						if firstletter {
							e.Styles = e.ConditionalStyles.FirstLine
							c.Styles = c.ConditionalStyles.FirstLine
						}
					}
				}
				firstletter = false

				// add this line box to the image map.
				imageMap.Add(c, r)
			}
		case html.ElementNode:
			if c.Data == "br" {
				e.advanceLine(dot)
				dot.Y += c.GetMarginBottomSize()
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
				if c.Data != "img" {
					// for images, this is done in the child
					// layout pass.
					dot.X += c.GetMarginLeftSize()
				}

				if c.GetFloat() == "none" {
					c.leftFloats = e.leftFloats
					c.rightFloats = e.rightFloats
				}

				c.inlineStart = true

				childContent, newDot := c.layoutPass(ctx, width, image.ZR, dot)

				var contentbox image.Rectangle
				if c.Data == "img" {
					e.ContentOverlay = c.ContentOverlay
					contentbox = c.BoxDrawRectangle
					// If it was a floating image it may
					// have changed X when it moved lineBoxes
					// over, so always take the new X value.
					// FIXME: Make this more robust.
					dot.X = newDot.X
					dot.Y = newDot.Y
				} else {
					c.ContentOverlay = childContent
					var box image.Image
					box, contentbox = c.calcCSSBox(childContent.Bounds().Size(), false, false)
					c.BoxDrawRectangle = box.Bounds().Add(*dot)

				}
				overlayed.GrowBounds(r)

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

				if c.GetFloat() == "none" && c.Data != "img" {
					dot.X = newDot.X + c.GetBorderRightWidth() + c.GetPaddingRight() + c.GetMarginRightSize()
					dot.Y = newDot.Y
				}
			case "block", "inline-block", "table", "table-inline", "list-item":
				if dot.X != e.leftFloats.MaxX(*dot) && display != "inline-block" {
					// This means the previous child was an inline item, and we should position dot
					// as if there were an implicit box around it.
					// floated elements don't affect dot, so only do this if it's not floated.
					if float == "none" {
						if c.PrevSibling != nil {
							e.advanceLine(dot)
						}
					}
				}

				if float == "none" {
					dot.Y += c.GetMarginTopSize()
					dot.Y -= c.marginCollapseOffset()
				} else {
					fdot.Y += c.GetMarginTopSize()
				}

				// draw the border, background, and CSS outer box.
				cdot := image.Point{0, 0}

				if c.GetDisplayProp() == "block" && float == "none" {
					// We tell the element that it has the whole width,
					// but add new floats (adjusted to the child's coordinate
					// space) so that if it goes past the existing floats it'll
					// go back to the full width. The adjustment only applies to
					// the height, because otherwise it's the horizontal plane
					// we're trying to track intersections with.
					// Floats don't inherit the other floats, because the parent
					// will make them collide and move them appropriately, they
					// don't take up line space from each other internally.
					c.leftFloats = make(FloatStack, 0, len(e.leftFloats))
					c.rightFloats = make(FloatStack, len(e.rightFloats))
					for _, lf := range e.leftFloats {
						float := new(RenderableDomElement)
						float.BoxDrawRectangle = lf.BoxDrawRectangle.Sub(image.Point{0, dot.Y})
						if float.BoxDrawRectangle.Max.Y > 0 {
							//if float.BoxDrawRectangle.Max.X > 0 && float.BoxDrawRectangle.Max.Y > 0 {
							c.leftFloats = append(c.leftFloats, float)
						}
					}
					for i, rf := range e.rightFloats {
						c.rightFloats[i] = new(RenderableDomElement)
						c.rightFloats[i].BoxDrawRectangle = rf.BoxDrawRectangle.Sub(image.Point{dot.X, dot.Y})
					}
				}
				var childContent image.Image
				childContent, dotAdj := c.layoutPass(ctx, width, image.ZR, &cdot)
				c.ContentOverlay = childContent
				box, contentbox := c.calcCSSBox(childContent.Bounds().Size(), false, false)
				c.BoxContentRectangle = contentbox.Sub(image.Point{dot.X, 0})
				sr := box.Bounds()

				if fdot.Y <= dot.Y {
					fdot.Y = dot.Y
					fdot.X = dot.X
				}
				var r image.Rectangle
				switch float {
				case "right":
					sz := box.Bounds().Size()
					e.positionRightFloat(*dot, sz, c)
					r = c.BoxDrawRectangle
					r.Min.Y -= c.GetMarginTopSize()
					r.Max.Y += c.GetMarginBottomSize()
					c.BoxContentRectangle = contentbox
				case "left":
					sz := box.Bounds().Size()

					e.positionLeftFloat(dot, &fdot, sz, c)
					r = c.BoxDrawRectangle
					r.Min.Y -= c.GetMarginTopSize()
					r.Max.Y += c.GetMarginBottomSize()
					c.BoxContentRectangle = contentbox
					dot.X += dotAdj.X
					fdot.X += dotAdj.X
				default:
					r = image.Rectangle{*dot, dot.Add(sr.Size())}
				}
				overlayed.GrowBounds(r)
				c.BoxDrawRectangle = r

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

						if mb := c.getEffectiveMarginBottom(); mb != 0 {
							dot.Y += mb
						}
					}

				}
			}
		}
	}

	if e.GetDisplayProp() == "block" && e.GetFloat() == "none" {
		e.advanceLine(dot)
	}
	e.ImageMap = imageMap
	return overlayed, *dot
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

// positions child c, which has the size size as a right floating element.
// dot is used to get the starting height to position the float.
func (e *RenderableDomElement) positionRightFloat(dot image.Point, size image.Point, c *RenderableDomElement) {
	// First make sure this hasn't already been positioned, since images layoutPass
	// attempts to position images from both the image itself and the parent
	for _, el := range e.rightFloats {
		if el == c {
			return
		}
	}

	width := e.contentWidth

	c.BoxDrawRectangle = image.Rectangle{
		image.Point{
			width - size.X,
			dot.Y,
		},
		image.Point{
			width,
			dot.Y + size.Y,
		},
	}
	for {
		foundhome, _ := e.handleFloatOverlap(&dot, &c.BoxDrawRectangle, c, false)
		if foundhome {
			break
		}
	}
	e.rightFloats = append(e.rightFloats, c)
}

// positions child c, which has the size size as a left floating element.
// dot is used to get the starting height to position the float. If already
// layed out text needs to be moved in order to place the float, it may also
// adjust dot.
func (e *RenderableDomElement) positionLeftFloat(dot, fdot *image.Point, size image.Point, c *RenderableDomElement) error {
	// First make sure this hasn't already been positioned, since images layoutPass
	// attempts to position images from both the image itself and the parent
	for _, el := range e.leftFloats {
		if el == c {
			return nil
		}
	}
	// Default to floating at the leftmost side of the element.
	c.BoxDrawRectangle = image.Rectangle{
		image.Point{
			0,
			fdot.Y,
		},
		image.Point{
			size.X,
			fdot.Y + size.Y,
		},
	}

	for {
		foundhome, movedot := e.handleFloatOverlap(fdot, &c.BoxDrawRectangle, c, true)
		if foundhome {
			if movedot {
				dot.X += c.BoxDrawRectangle.Size().X
			}
			break
		}
	}
	e.leftFloats = append(e.leftFloats, c)
	return nil
}

// verifies that positioning a float as position r would not overlap anything.
// if it does, move it to the next possible left float position if moveleft
// is true, and right position if it's false.
// returns true if it's found a home for r, and false if we haven't.
// linesmoved returns true if line items were moved to position this float
// and dot likely needs to be adjusted as a result.
func (e *RenderableDomElement) handleFloatOverlap(dot *image.Point, r *image.Rectangle, c *RenderableDomElement, moveleft bool) (positioned, linesmoved bool) {
	width := e.contentWidth
	// Check if it overlaps anything and move it if necessary.
	// (We pretend that it has an overlap at the start, so that we can be sure that the loop happens
	// at least once to verify it.)
	size := r.Size()

	lfWidth := e.leftFloats.WidthAt(*dot)
	rfWidth := e.rightFloats.WidthAt(*dot)
	verifyAndMove := func(other *RenderableDomElement) bool {
		if other.BoxDrawRectangle.Overlaps(*r) {
			if width-rfWidth-lfWidth-size.X < 0 {
				// No room on this line, with the floats, so move down past
				// the first float and try again.
				dot.Y += e.leftFloats.ClearFloats(*dot).NextFloatHeight()
				if moveleft {
					dot.X = 0
				} else {
					dot.X = width - rfWidth - size.X
				}
			} else {
				// There's still room in width for this float, so move it just
				// past the last float.
				if moveleft {
					dot.X = lfWidth
				} else {
					dot.X = width - rfWidth - size.X
				}
			}

			r.Min = image.Point{
				dot.X,
				dot.Y,
			}
			r.Max = image.Point{
				dot.X + size.X,
				dot.Y + size.Y,
			}
			// We've moved it, but haven't verified the new location, so tell
			// the caller that it still needs to be verified
			return false
		}
		return true
	}
	// Check if it overlaps left floats
	for _, f := range e.leftFloats {
		if verifyAndMove(f) == false {
			return false, false
		}
	}
	// Check if it overlaps any right floats
	for _, f := range e.rightFloats {
		if verifyAndMove(f) == false {
			return false, false
		}
	}

	// Check if the box overlaps with any line now that it's been placed.
	// If so, either move it down one last time, or move the lines over
	// if they fit.
	for _, line := range e.lineBoxes {
		lineBounds := line.Bounds()
		if r.Overlaps(lineBounds) {
			if moveleft {
				canMoveText := true
				for _, line := range e.lineBoxes {
					lineBounds := line.Bounds()
					lsize := lineBounds.Size()
					o := line.origin
					if r.Max.X+lsize.X+e.rightFloats.WidthAt(o) >= width {
						canMoveText = false
						break
					}
				}

				if canMoveText {
					fsz := r.Size().X
					overlapped := false
					for i, line := range e.lineBoxes {
						lsz := line.Content.Bounds().Size()
						lineBounds := image.Rectangle{
							line.origin,
							line.origin.Add(lsz),
						}

						if overlapped || r.Overlaps(lineBounds) {
							overlapped = true
							line.origin.X += fsz
							e.lineBoxes[i] = line
						}
					}
					// We were able to just move all the text on the line, there's
					// no reason to verify the box again.
					return true, true
				} else {
					e.advanceLine(dot)
					r.Min = image.Point{
						dot.X,
						dot.Y,
					}
					r.Max = image.Point{
						dot.X + size.X,
						dot.Y + size.Y,
					}
				}
				return false, false
			} else {
				// There's definitely no space to move the text on this
				// line because we're dealing with a right float, so
				// just move the float down one line.
				e.advanceLine(dot)
				r.Min = image.Point{
					width - rfWidth - size.X,
					dot.Y,
				}
				r.Max = image.Point{
					width - rfWidth,
					dot.Y + size.Y,
				}
				return true, false
			}
			return false, false
		}
	}
	return true, false
}

// Get the BoxDrawRectangle for this element, translated from the parent's
// coordinate system to the absolute coordinate system.
func (e *RenderableDomElement) getAbsoluteDrawRectangle() image.Rectangle {
	var adj image.Point
	for p := e.Parent; p != nil; p = p.Parent {
		if p.GetDisplayProp() == "inline" {
			// BoxContentRectangle isn't meaningful for inline
			// parents.
			continue
		}

		adj = adj.Add(p.BoxDrawRectangle.Min).Add(p.BoxContentRectangle.Min)
	}
	return e.BoxDrawRectangle.Add(adj)
}
func (e *RenderableDomElement) drawInto(ctx context.Context, dst draw.Image, cursor image.Point) error {
	if e.Type == html.ElementNode && strings.ToLower(e.Data) == "img" {
		if e.GetDisplayProp() == "block" {
			// Inlines are drawn as part of a lineBox, while blocks expected
			// drawInto to have drawn the image (but did the border itself).
			absrect := e.getAbsoluteDrawRectangle()

			// now draw the content on top of the outer box
			contentStart := absrect.Min.Add(e.BoxContentRectangle.Min)
			contentBounds := e.ContentOverlay.Bounds()
			cr := image.Rectangle{contentStart, contentStart.Add(contentBounds.Size())}
			draw.Draw(
				dst,
				cr.Sub(cursor),
				e.ContentOverlay,
				contentBounds.Min,
				draw.Over,
			)
		} else if e.GetDisplayProp() == "inline" && e.GetFloat() != "none" {
			// Floated images also need to get drawn, even if they're inlines.
			absrect := e.getAbsoluteDrawRectangle()

			// now draw the content on top of the outer box
			contentStart := absrect.Min.Add(
				image.Point{
					e.GetMarginTopSize() + e.GetPaddingTop() + e.GetBorderTopWidth(),
					e.GetMarginLeftSize() + e.GetPaddingLeft() + e.GetBorderLeftWidth(),
				},
			)

			contentBounds := e.ContentOverlay.Bounds()
			cr := image.Rectangle{contentStart, contentStart.Add(contentBounds.Size())}
			draw.Draw(
				dst,
				cr.Sub(cursor),
				e.ContentOverlay,
				contentBounds.Min,
				draw.Over,
			)
		}
		return nil
	}
	for c := e.FirstChild; c != nil; c = c.NextSibling {
		if ctx.Err() != nil {
			return nil
		}

		absrect := c.getAbsoluteDrawRectangle()

		switch c.Type {
		case html.ElementNode:
			// Some special cases for replaced elements.
			switch strings.ToLower(c.Data) {
			case "br":
				continue
			}

			switch display := c.GetDisplayProp(); display {
			case "none":
				continue
			default:
				/*				// Cull elements that don't fit onto dst.
								if absrect.Max.X < cursor.X {
									// the box is to the left of the viewport, don't draw it.
									continue
								}
								if absrect.Min.X > dstsize.X+cursor.X {
									// to the right of the viewport
									continue
								}
								if absrect.Max.Y < cursor.Y {
									// on top of the viewport
									continue
								}
								if absrect.Min.Y > dstsize.Y+cursor.Y {
									// below the viewport.
									continue
								}
				*/
				// when doing the layout the boxDrawRectangle was fudged
				// for floats to make it easier to calculate intersections
				// when doing the layout. Now we need to adjust.
				switch c.GetFloat() {
				case "left", "right":
					if !c.floatAdjusted {
						mt := c.GetMarginTopSize()
						mb := c.GetMarginBottomSize()
						absrect.Min.Y += mt
						absrect.Max.Y -= mb
						c.BoxDrawRectangle.Min.Y += mt
						c.BoxDrawRectangle.Max.Y -= mb
						c.floatAdjusted = true
					}
				default:
				}
				if !(c.Type == html.ElementNode && c.Data == "img" && c.GetDisplayProp() == "inline" && c.GetFloat() == "none") {
					// Inline images get drawn as part of a lineBox,
					// and borders get drawn as part of a linebox
					// for all non-image inlines.
					if c.CSSOuterBox != nil && (c.GetDisplayProp() != "inline" || c.GetFloat() != "none") {
						sr := c.CSSOuterBox.Bounds()
						draw.Draw(
							dst,
							absrect.Sub(cursor),
							c.CSSOuterBox,
							sr.Min,
							draw.Over,
						)
					}

					if err := c.drawInto(ctx, dst, cursor); err != nil {
						return err
					}
				}

			}

		case html.TextNode:
			// The parent contained the line boxes for the textnode
			// (and possibly other inline things on the same line)
			// so this was rendered by the parent.
			continue
		}
	}

	absrect := e.getAbsoluteDrawRectangle()

	for _, box := range e.lineBoxes {
		var sr image.Rectangle
		if box.BorderImage != nil {
			sr = box.BorderImage.Bounds()
		} else {
			sr = box.Content.Bounds()
		}
		bo := box.origin
		r := image.Rectangle{bo, bo.Add(sr.Size())}.Add(absrect.Min)
		if e.GetDisplayProp() != "inline" {
			r = r.Add(e.BoxContentRectangle.Min)
		}
		if box.BorderImage != nil {
			sr := box.BorderImage.Bounds()
			// ro := box.origin.Add(box.borigin).Add(absrect.Min)
			ro := box.origin.Sub(box.borigin).Add(absrect.Min)
			r := image.Rectangle{ro, ro.Add(sr.Size())}
			draw.Draw(
				dst,
				r.Sub(cursor),
				box.BorderImage,
				sr.Min,
				draw.Over,
			)
		}
		if box.IsImage() {
			// It was an inline image
			draw.Draw(dst,
				r.Sub(cursor),
				box.Content,
				sr.Min,
				draw.Over,
			)
		} else {
			// It was inline text that still needs to be drawn.
			if err := box.drawAt(ctx, dst, r.Sub(cursor).Min); err != nil {
				return err
			}

		}
	}
	return nil
}

func (e *RenderableDomElement) getContainingBlock() *RenderableDomElement {
	for p := e.Parent; p != nil; p = p.Parent {
		if p.Type != html.ElementNode {
			continue
		}
		switch p.GetDisplayProp() {
		case "inline", "inline-block":
			continue
		default:
			return p
		}
	}
	return nil
}
func (e *RenderableDomElement) advanceLine(dot *image.Point) {
	// If there was more than 1 element, re-adjust all their positions with respect to
	// the vertical-align property.
	baseline := 0
	maxsize := 0
	textbottom := 0
	texttop := 0
	textheight := 0

	nextline := e.GetLineHeight()

	// Now the we've advanced a line, we can't possibly be at either
	// the first letter or the first line, so just use the unconditional
	// styles.
	e.Styles = e.ConditionalStyles.Unconditional
	// Step 1. Figure out how big the line really is and where the baseline is.
	for _, l := range e.curLine {
		height := l.Height()
		if height > maxsize {
			maxsize = height
		}

		// If there were multiple elements on this line, ensure that
		// the largest lineheight is used for the whole line.
		lh := l.LineHeight()
		if lh > nextline {
			nextline = lh
		}

		bl := l.Baseline()
		if !l.IsImage() {
			if dsc := l.metrics.Descent.Ceil(); dsc > textbottom {
				textbottom = dsc
			}
			if asc := l.metrics.Ascent.Ceil(); asc > texttop {
				texttop = asc
			}
			h := l.metrics.Height.Ceil()
			if h > textheight {
				textheight = h
			}
			if h != lh {
				l.origin.Y += (lh - h) / 2
			}
		} else {

			switch align := l.el.GetVerticalAlign(); align {
			case "text-top":
				bl = 0
			case "middle":
				bl = height / 2
			case "text-bottom":
				bl = height - textbottom
			default:
				bl = height
			}
		}
		if bl > baseline {
			baseline = bl
		}
	}

	// Step 2: Adjust the image origins with respect to the baseline.
	for _, l := range e.curLine {
		height := l.Height()

		align := l.el.GetVerticalAlign()
		switch align {
		case "text-bottom":
			l.origin.Y += baseline - height + textbottom
		case "text-top":
			l.origin.Y += baseline - texttop
		case "middle":
			l.origin.Y += baseline - (height / 2)
		default:
			l.origin.Y += baseline - l.Baseline()
		}
		if l.IsImage() {
			tm := l.el.GetMarginTopSize()
			bm := l.el.GetMarginBottomSize()

			l.origin.Y += l.el.GetPaddingTop() + l.el.GetBorderTopWidth() + tm
			l.origin.X += l.el.GetPaddingLeft() + l.el.GetBorderLeftWidth() + l.el.GetMarginLeftSize()

			// top margin was in both the height and the origin.Y, so
			// subtract it from height when checking if it needs to
			// increase the next line size.
			// whether or not the bottom margin needs to be subtracted
			// depends on whether the margin is bigger than the other
			// text on the line, which depends on the vertical alignment
			bmoff := 0
			switch align {
			case "text-bottom":
				if bm >= textheight {
					// Remove the part of the margin that's
					// absorbed by the text line height.
					// FIXME: Figure out where this -1 comes
					// from. It seems to be required to get
					// TestImgBorderTextBottomLineheight to
					// pass, but maybe we're doing something
					// wrong.
					bmoff = texttop - 1
				}
			case "text-top":
				if bm >= textheight {
					bmoff = bm
				}
			}

			if bm < 0 {
				bm = 0
			}

			// Top was already incorporated into the origin, so
			// don't double count it in the height.
			height -= l.el.GetPaddingTop() + l.el.GetBorderTopWidth() + tm
			if end := height + l.origin.Y - bmoff + bm; end > dot.Y {
				nextline = end - dot.Y
			}
		}
	}

	dot.Y += nextline
	dot.X = e.leftFloats.MaxX(*dot)

	e.curLine = make([]*lineBox, 0)
}
