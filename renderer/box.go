package renderer

import (
	//	"fmt"
	"github.com/driusan/Gob/css"
	//"github.com/driusan/Gob/net"
	"image"
	"image/color"
	// the standard draw package doesn't have Copy, which we need for background Repeat.
	"image/draw"
	"net/url"
	"strings"
)

type BoxOffset struct {
	Width int
}

type BoxMargin BoxOffset
type BoxPadding BoxOffset

type BoxBorder struct {
	BoxOffset
	Color color.Color
	Style string
}

type BoxMargins struct {
	Top    BoxMargin
	Bottom BoxMargin
	Left   BoxMargin
	Right  BoxMargin
}

type BoxBorders struct {
	Top    BoxBorder
	Bottom BoxBorder
	Left   BoxBorder
	Right  BoxBorder
}

type BoxPaddings struct {
	Top    BoxPadding
	Bottom BoxPadding
	Left   BoxPadding
	Right  BoxPadding
}

type outerBoxDrawer struct {
	Padding BoxPaddings
	Border  BoxBorders

	Margin BoxMargins

	contentSize     image.Point
	background      image.Image
	backgroundColor color.Color
}

func (b *outerBoxDrawer) ColorModel() color.Model {
	return color.RGBAModel
}
func (b *outerBoxDrawer) Bounds() image.Rectangle {
	return image.Rectangle{
		Min: image.Point{X: 0, Y: 0},
		Max: image.Point{
			X: b.contentSize.X + int(b.Border.Left.Width+b.Border.Right.Width) + int(b.Padding.Left.Width+b.Padding.Right.Width) + int(b.Margin.Left.Width+b.Margin.Right.Width),
			Y: b.contentSize.Y + int(b.Border.Top.Width+b.Border.Bottom.Width) + int(b.Padding.Top.Width+b.Padding.Bottom.Width) + int(b.Margin.Top.Width+b.Margin.Bottom.Width),
		},
	}
}

func (b *outerBoxDrawer) RGBA() *image.RGBA {
	bounds := b.Bounds()
	ri := image.NewRGBA(b.Bounds())
	/* the zero colour is transparent, this isn't necessary
	draw.Draw(
		ri,
		b.Bounds(),
		&image.Uniform{color.Transparent},
		image.ZP,
		draw.Src,
	)
	*/

	// draw the background first, bounded by the margins
	draw.Draw(
		ri,
		image.Rectangle{
			Min: image.Point{
				X: b.Margin.Left.Width,
				Y: b.Margin.Top.Width,
			},
			Max: image.Point{
				X: bounds.Max.X - b.Margin.Right.Width,
				Y: bounds.Max.Y - b.Margin.Bottom.Width,
			},
		},
		b.background,
		image.ZP,
		draw.Src,
	)
	// draw the top border
	draw.Draw(
		ri,
		image.Rectangle{
			Min: image.Point{
				X: b.Margin.Left.Width,
				Y: b.Margin.Top.Width,
			},
			Max: image.Point{
				X: bounds.Max.X - b.Margin.Right.Width,
				Y: b.Margin.Top.Width + b.Border.Top.Width,
			},
		},
		&image.Uniform{b.Border.Top.Color},
		image.ZP,
		draw.Src,
	)
	// draw the left border
	draw.Draw(
		ri,
		image.Rectangle{
			Min: image.Point{
				X: b.Margin.Left.Width,
				Y: b.Margin.Top.Width,
			},
			Max: image.Point{
				X: b.Margin.Left.Width + b.Border.Left.Width,
				Y: bounds.Max.Y - b.Margin.Bottom.Width,
			},
		},
		&image.Uniform{b.Border.Left.Color},
		image.ZP,
		draw.Src,
	)
	// draw the right border
	draw.Draw(
		ri,
		image.Rectangle{
			Min: image.Point{
				X: bounds.Max.X - b.Margin.Right.Width - b.Border.Left.Width,
				Y: b.Margin.Top.Width,
			},
			Max: image.Point{
				X: bounds.Max.X - b.Margin.Right.Width,
				Y: bounds.Max.Y - b.Border.Bottom.Width,
			},
		},
		&image.Uniform{b.Border.Right.Color},
		image.ZP,
		draw.Src,
	)
	// draw the bottom border
	draw.Draw(
		ri,
		image.Rectangle{
			Min: image.Point{
				X: b.Margin.Left.Width,
				Y: bounds.Max.Y - b.Margin.Bottom.Width - b.Border.Bottom.Width,
			},
			Max: image.Point{
				X: bounds.Max.X - b.Margin.Right.Width,
				Y: bounds.Max.Y - b.Margin.Bottom.Width,
			},
		},
		&image.Uniform{b.Border.Bottom.Color},
		image.ZP,
		draw.Over,
	)

	return ri
}
func (b *outerBoxDrawer) GetContentOrigin() image.Point {
	return image.Point{
		X: b.Border.Left.Width + b.Margin.Left.Width + b.Padding.Left.Width,
		Y: b.Border.Top.Width + b.Margin.Top.Width + b.Padding.Top.Width,
	}
}
func (b *outerBoxDrawer) At(x, y int) color.Color {
	bounds := b.Bounds()
	// Deal with the margin
	if y < b.Margin.Top.Width || x < b.Margin.Left.Width {
		return &color.RGBA{0, 0, 0, 0}
	}
	if y > (bounds.Max.Y-b.Margin.Bottom.Width) || x > (bounds.Max.X-b.Margin.Right.Width) {
		return &color.RGBA{0, 0, 0, 0}
	}

	// Then the borders
	if (y - b.Margin.Top.Width) < b.Border.Top.Width {
		return b.Border.Top.Color
	}

	if (x - b.Margin.Left.Width) < b.Border.Left.Width {
		return b.Border.Left.Color
	}

	if y > bounds.Max.Y-b.Border.Bottom.Width-b.Margin.Bottom.Width {
		return b.Border.Bottom.Color
	}

	if x > bounds.Max.X-(b.Margin.Right.Width+b.Border.Right.Width) {
		return b.Border.Right.Color
	}

	return b.background.At(x, y)
}

var dfltBorder *color.RGBA = &color.RGBA{255, 128, 128, 0}
var dfltBackground *color.RGBA = &color.RGBA{255, 128, 128, 255}

func (e RenderableDomElement) GetBorderBottomWidth() int {
	if e.Styles == nil {
		return 0
	}
	value := e.Styles.BorderBottomWidth.GetValue()
	if value == "" {
		// No style, use default.
		return 0
	}
	if value == "inherit" {
		if e.Parent == nil {
			return 0
		}
		return e.Parent.GetBorderBottomWidth()

	}
	fontsize := e.GetFontSize()
	val, err := css.ConvertUnitToPx(fontsize, e.containerWidth, value)
	if err != nil {
		return 0
	}
	return val
}
func (e RenderableDomElement) GetBorderBottomColor() *color.RGBA {
	if e.Styles == nil {
		return dfltBorder
	}
	value := e.Styles.BorderBottomColor.GetValue()
	if value == "" {
		return dfltBorder
	}
	if value == "inherit" {
		if e.Parent == nil {
			return dfltBorder
		}
		return e.Parent.GetBorderBottomColor()
	}
	c, err := css.ConvertColorToRGBA(value)
	if err != nil {
		return dfltBorder
	}
	return c
}
func (e RenderableDomElement) GetBorderTopWidth() int {
	if e.Styles == nil {
		return 0
	}
	value := e.Styles.BorderTopWidth.GetValue()
	if value == "" {
		// No style, use default.
		return 0
	}
	if value == "inherit" {
		if e.Parent == nil {
			return 0
		}
		return e.Parent.GetBorderTopWidth()

	}
	fontsize := e.GetFontSize()
	val, err := css.ConvertUnitToPx(fontsize, e.containerWidth, value)
	if err != nil {
		return 0
	}
	return val
}
func (e RenderableDomElement) GetBorderTopColor() *color.RGBA {
	if e.Styles == nil {
		return dfltBorder
	}
	value := e.Styles.BorderTopColor.GetValue()
	if value == "" {
		return dfltBorder
	}
	if value == "inherit" {
		if e.Parent == nil {
			return dfltBorder
		}
		return e.Parent.GetBorderTopColor()
	}
	c, err := css.ConvertColorToRGBA(value)
	if err != nil {
		return dfltBorder
	}
	return c
}

func (e RenderableDomElement) GetBorderLeftWidth() int {
	if e.Styles == nil {
		return 0
	}
	value := e.Styles.BorderLeftWidth.GetValue()
	if value == "" {
		// No style, use default.
		return 0
	}
	if value == "inherit" {
		if e.Parent == nil {
			return 0
		}
		return e.Parent.GetBorderLeftWidth()

	}
	fontsize := e.GetFontSize()
	val, err := css.ConvertUnitToPx(fontsize, e.containerWidth, value)
	if err != nil {
		return 0
	}
	return val
}
func (e RenderableDomElement) GetBorderLeftColor() *color.RGBA {
	if e.Styles == nil {
		return dfltBorder
	}
	value := e.Styles.BorderLeftColor.GetValue()
	if value == "" {
		return dfltBorder
	}
	if value == "inherit" {
		if e.Parent == nil {
			return dfltBorder
		}
		return e.Parent.GetBorderLeftColor()
	}
	c, err := css.ConvertColorToRGBA(value)
	if err != nil {
		return dfltBorder
	}
	return c
}

func (e RenderableDomElement) GetBorderRightWidth() int {
	if e.Styles == nil {
		return 0
	}
	value := e.Styles.BorderRightWidth.GetValue()
	if value == "" {
		// No style, use default.
		return 0
	}
	if value == "inherit" {
		if e.Parent == nil {
			return 0
		}
		return e.Parent.GetBorderRightWidth()

	}
	fontsize := e.GetFontSize()
	val, err := css.ConvertUnitToPx(fontsize, e.containerWidth, value)
	if err != nil {
		return 0
	}
	return val
}
func (e RenderableDomElement) GetBorderRightColor() *color.RGBA {
	if e.Styles == nil {
		return dfltBorder
	}
	value := e.Styles.BorderRightColor.GetValue()
	if value == "" {
		return dfltBorder
	}
	if value == "inherit" {
		if e.Parent == nil {
			return dfltBorder
		}
		return e.Parent.GetBorderRightColor()
	}
	c, err := css.ConvertColorToRGBA(value)
	if err != nil {
		return dfltBorder
	}
	return c
}
func (e RenderableDomElement) GetMarginLeftSize() int {
	switch value := e.Styles.MarginLeft.GetValue(); value {
	case "":
		return 0
	case "inherit":
		if e.Parent == nil {
			return 0
		}
		return e.Parent.GetMarginLeftSize()
	case "auto":
		if e.Styles.MarginRight.GetValue() == "auto" {
			// return calculate how much is needed to center

			return (e.containerWidth - e.contentWidth - e.GetBorderLeftWidth() - e.GetBorderRightWidth() - e.GetPaddingLeft() - e.GetPaddingRight()) / 2
		}
		return 0
	default:

		fontsize := e.GetFontSize()
		val, err := css.ConvertUnitToPx(fontsize, e.containerWidth, value)
		if err != nil {
			return 0
		}
		return val
	}
}
func (e RenderableDomElement) GetMarginRightSize() int {
	value := e.Styles.MarginRight.GetValue()
	if value == "" {
		// No style, use default.
		return 0
	}
	if value == "inherit" {
		if e.Parent == nil {
			return 0
		}
		return e.Parent.GetMarginRightSize()

	}
	fontsize := e.GetFontSize()
	val, err := css.ConvertUnitToPx(fontsize, e.containerWidth, value)
	if err != nil {
		return 0
	}
	return val
}
func (e RenderableDomElement) GetMarginTopSize() int {
	value := e.Styles.MarginTop.GetValue()
	if value == "" {
		// No style, use default.
		return 0
	}
	if value == "inherit" {
		if e.Parent == nil {
			return 0
		}
		return e.Parent.GetMarginTopSize()

	}
	fontsize := e.GetFontSize()
	val, err := css.ConvertUnitToPx(fontsize, e.containerWidth, value)
	if err != nil {
		return 0
	}
	return val
}

func (e RenderableDomElement) GetMarginBottomSize() int {
	value := e.Styles.MarginBottom.GetValue()
	if value == "" {
		// No style, use default.
		return 0
	}
	if value == "inherit" {
		if e.Parent == nil {
			return 0
		}
		return e.Parent.GetMarginBottomSize()

	}
	fontsize := e.GetFontSize()
	val, err := css.ConvertUnitToPx(fontsize, e.containerWidth, value)
	if err != nil {
		return 0
	}
	return val
}
func (e RenderableDomElement) GetPaddingLeft() int {
	value := e.Styles.PaddingLeft.GetValue()
	if value == "" {
		// No style, use default.
		return 0
	}
	if value == "inherit" {
		if e.Parent == nil {
			return 0
		}
		return e.Parent.GetPaddingLeft()

	}
	fontsize := e.GetFontSize()
	val, err := css.ConvertUnitToPx(fontsize, e.containerWidth, value)
	if err != nil {
		return 0
	}
	return val
}
func (e RenderableDomElement) GetPaddingRight() int {
	value := e.Styles.PaddingRight.GetValue()
	if value == "" {
		// No style, use default.
		return 0
	}
	if value == "inherit" {
		if e.Parent == nil {
			return 0
		}
		return e.Parent.GetPaddingRight()

	}
	fontsize := e.GetFontSize()
	val, err := css.ConvertUnitToPx(fontsize, e.containerWidth, value)
	if err != nil {
		return 0
	}
	return val
}
func (e RenderableDomElement) GetPaddingTop() int {
	value := e.Styles.PaddingTop.GetValue()
	if value == "" {
		// No style, use default.
		return 0
	}
	if value == "inherit" {
		if e.Parent == nil {
			return 0
		}
		return e.Parent.GetPaddingTop()

	}
	fontsize := e.GetFontSize()
	val, err := css.ConvertUnitToPx(fontsize, e.containerWidth, value)
	if err != nil {
		return 0
	}
	return val
}
func (e RenderableDomElement) GetPaddingBottom() int {
	value := e.Styles.PaddingBottom.GetValue()
	if value == "" {
		// No style, use default.
		return 0
	}
	if value == "inherit" {
		if e.Parent == nil {
			return 0
		}
		return e.Parent.GetPaddingBottom()

	}
	fontsize := e.GetFontSize()
	val, err := css.ConvertUnitToPx(fontsize, e.containerWidth, value)
	if err != nil {
		return 0
	}
	return val
}
func (e RenderableDomElement) GetBorderTopStyle() string {
	if e.Styles == nil {
		return "none"
	}
	val := e.Styles.BorderTopStyle.GetValue()
	if val == "" {
		return "none"
	}
	if val == "inherit" {
		if e.Parent == nil {
			return "none"
		}
		return e.Parent.GetBorderTopStyle()
	}
	return val
}
func (e RenderableDomElement) GetBorderBottomStyle() string {
	if e.Styles == nil {
		return "none"
	}
	val := e.Styles.BorderBottomStyle.GetValue()
	if val == "" {
		return "none"
	}
	if val == "inherit" {
		if e.Parent == nil {
			return "none"
		}
		return e.Parent.GetBorderBottomStyle()
	}
	return val
}
func (e RenderableDomElement) GetBorderLeftStyle() string {
	if e.Styles == nil {
		return "none"
	}
	val := e.Styles.BorderLeftStyle.GetValue()
	if val == "" {
		return "none"
	}
	if val == "inherit" {
		if e.Parent == nil {
			return "none"
		}
		return e.Parent.GetBorderLeftStyle()
	}
	return val
}
func (e RenderableDomElement) GetBorderRightStyle() string {
	if e.Styles == nil {
		return "none"
	}
	val := e.Styles.BorderRightStyle.GetValue()
	if val == "" {
		return "none"
	}
	if val == "inherit" {
		if e.Parent == nil {
			return "none"
		}
		return e.Parent.GetBorderRightStyle()
	}
	return val
}

func (e *RenderableDomElement) GetBackgroundRepeat() string {
	repeat := e.Styles.BackgroundRepeat.GetValue()
	switch strings.ToLower(repeat) {
	case "inherit":
		return e.Parent.GetBackgroundRepeat()
	case "repeat", "repeat-x", "repeat-y", "no-repeat":
		return repeat
	default:
		return "repeat"
	}
}
func (e *RenderableDomElement) GetBackgroundImage() image.Image {
	iURL, err := e.Styles.GetBackgroundImage()
	switch err {
	case css.InheritValue:
		return e.Parent.GetBackgroundImage()
	case css.Invalid, css.NoStyles:
		return nil
	}
	u, err := url.Parse(iURL)
	if err != nil {
		return nil
	}
	newURL := e.PageLocation.ResolveReference(u)
	r, resp, err := e.resolver.GetURL(newURL)
	if err != nil || resp < 200 || resp >= 300 {
		return nil
	}
	content, _, err := image.Decode(r)
	if err != nil {
		return nil
	}
	return content

}

// Given an image, returns an image representing the CSS Box that should
// surround that image, and a rectangle denoting the portion of that
// image which should be used to overlay content.
// The returned image does *not* have the content overlayed, it only
// has the margin/background/borders drawn on it.
func (e *RenderableDomElement) calcCSSBox(content image.Image, collapsablemargin int) (image.Image, image.Rectangle) {
	// calculate the size of the box.
	size := content.Bounds().Size()
	if width := e.GetWidth(); width >= 0 {
		size.X = width
	}
	if height := e.GetHeight(); height >= 0 {
		size.Y = height
	}

	if minheight := e.GetMinHeight(); size.Y < minheight {
		size.Y = minheight
	}
	if minwidth := e.GetMinWidth(); size.X < minwidth {
		size.X = minwidth
	}
	if maxheight := e.GetMaxHeight(); maxheight >= 0 && size.Y > maxheight {
		size.Y = maxheight
	}
	if maxwidth := e.GetMaxWidth(); maxwidth >= 0 && size.X > maxwidth {
		size.X = maxwidth
	}

	// calculate the background image for the content box.
	bgi := e.GetBackgroundImage()
	if bgi == nil {
		bg := e.GetBackgroundColor()
		bgi = &image.Uniform{bg}
	} else {
		bg := e.GetBackgroundColor()
		solidbg := &image.Uniform{bg}
		// we need to construct the background image based on the
		// repeat and make sure that the background-color shines through any transparent
		// parts of the background image

		// allocate a new image of the appropriate size
		csize := size
		bgCanvas := image.NewRGBA(image.Rectangle{
			image.ZP,
			image.Point{
				csize.X + e.GetPaddingLeft() + e.GetPaddingRight() + e.GetBorderLeftWidth() + e.GetBorderRightWidth(),
				csize.Y + e.GetPaddingTop() + e.GetPaddingBottom() + e.GetBorderTopWidth() + e.GetBorderBottomWidth(),
			}})

		// draw the background colour over the whole image, so that the transparent parts
		// are correct.
		draw.Draw(
			bgCanvas,
			bgCanvas.Bounds(),
			solidbg,
			image.ZP,
			draw.Src,
		)

		bgiSize := bgi.Bounds().Size()

		// now draw the background image based on the background-repeat.
		switch e.GetBackgroundRepeat() {
		case "no-repeat":
			draw.Draw(
				bgCanvas,
				bgCanvas.Bounds(),
				bgi,
				image.ZP,
				draw.Over,
			)
		case "repeat-x":
			for x := 0; ; x += bgiSize.X {
				draw.Draw(
					bgCanvas,
					image.Rectangle{
						image.Point{x, 0},
						image.Point{x + bgiSize.X, bgiSize.Y},
					},
					bgi,
					image.ZP,
					draw.Over,
				)
				if x > csize.X {
					break
				}
			}
		case "repeat-y":
			for y := 0; ; y += bgiSize.X {
				draw.Draw(
					bgCanvas,
					image.Rectangle{
						image.Point{0, y},
						image.Point{bgiSize.X, y + bgiSize.Y},
					},
					bgi,
					image.ZP,
					draw.Over,
				)
				if y > csize.Y {
					break
				}
			}
		case "repeat":
			fallthrough
		default:
			for x := 0; ; x += bgiSize.X {
				for y := 0; ; y += bgiSize.Y {
					draw.Draw(
						bgCanvas,
						image.Rectangle{
							image.Point{x, y},
							image.Point{x + bgiSize.X, y + bgiSize.Y},
						},
						bgi,
						image.ZP,
						draw.Over,
					)
					if y > csize.Y {
						break
					}
				}
				if x > csize.X {
					break
				}
			}

		}
		bgi = bgCanvas
	}
	var topmargin int
	if collapsablemargin == 0 {
		topmargin = e.GetMarginTopSize()
	} else {
		if ts := e.GetMarginTopSize(); ts < collapsablemargin {
			topmargin = collapsablemargin
		} else {
			topmargin = ts
		}
	}
	box := &outerBoxDrawer{
		Margin: BoxMargins{
			Top:    BoxMargin{Width: topmargin},
			Left:   BoxMargin{Width: e.GetMarginLeftSize()},
			Right:  BoxMargin{Width: e.GetMarginRightSize()},
			Bottom: BoxMargin{Width: e.GetMarginBottomSize()},
		},
		Border: BoxBorders{
			Top:    BoxBorder{BoxOffset: BoxOffset{Width: e.GetBorderTopWidth()}, Color: e.GetBorderTopColor(), Style: e.GetBorderTopStyle()},
			Left:   BoxBorder{BoxOffset: BoxOffset{Width: e.GetBorderLeftWidth()}, Color: e.GetBorderLeftColor(), Style: e.GetBorderLeftStyle()},
			Right:  BoxBorder{BoxOffset: BoxOffset{Width: e.GetBorderRightWidth()}, Color: e.GetBorderRightColor(), Style: e.GetBorderRightStyle()},
			Bottom: BoxBorder{BoxOffset: BoxOffset{Width: e.GetBorderBottomWidth()}, Color: e.GetBorderBottomColor(), Style: e.GetBorderBottomStyle()},
		},
		Padding: BoxPaddings{
			Top:    BoxPadding{Width: e.GetPaddingTop()},
			Left:   BoxPadding{Width: e.GetPaddingLeft()},
			Right:  BoxPadding{Width: e.GetPaddingRight()},
			Bottom: BoxPadding{Width: e.GetPaddingBottom()},
		},
		contentSize: size,
		background:  bgi,
	}
	e.CSSOuterBox = box.RGBA()
	corigin := box.GetContentOrigin()
	return e.CSSOuterBox, image.Rectangle{
		Min: corigin,
		Max: image.Point{X: corigin.X + size.X, Y: corigin.Y + size.Y},
	}
}
