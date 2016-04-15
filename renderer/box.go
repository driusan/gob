package renderer

import (
	"Gob/css"
	"image"
	"image/color"
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

	contentSize image.Point
	background  image.Image
}

func (b *outerBoxDrawer) ColorModel() color.Model {
	return color.RGBAModel
}
func (b *outerBoxDrawer) Bounds() image.Rectangle {
	return image.Rectangle{
		Min: image.Point{X: 0, Y: 0},
		Max: image.Point{
			X: b.contentSize.X + int(b.Border.Left.Width+b.Border.Right.Width) + int(b.Padding.Left.Width+b.Padding.Right.Width) + int(b.Margin.Left.Width+b.Margin.Right.Width),
			Y: b.contentSize.Y + int(b.Border.Top.Width+b.Border.Bottom.Width) + int(b.Padding.Top.Width+b.Padding.Bottom.Width) + int(b.Margin.Left.Width+b.Margin.Right.Width),
		},
	}
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

	// The padding is taken care of by the GetContentOrigin() call. Everything else
	// is a background
	return b.background.At(x, y)
}

var dfltBorder *color.RGBA = &color.RGBA{255, 128, 128, 0}

func (e RenderableDomElement) GetBorderBottomSize() int {
	if e.Styles == nil {
		return 0
	}
	if style := e.GetBorderBottomStyle(); style == "hidden" || style == "none" {
		return 0
	}

	val, err := e.Styles.GetBorderSizeInPx("bottom")
	switch err {
	case css.InheritValue:
		if e.Parent == nil {
			return 0
		}
		return e.Parent.GetBorderBottomSize()
	case css.NoStyles:
		return 0
	}
	return val
}
func (e RenderableDomElement) GetBorderBottomColor() *color.RGBA {
	if e.Styles == nil {
		return dfltBorder
	}
	val, err := e.Styles.GetBorderColor("bottom", dfltBorder)
	if err == css.InheritValue {
		if e.Parent == nil {
			return dfltBorder
		}
		return e.Parent.GetBorderBottomColor()
	}
	return val
}
func (e RenderableDomElement) GetBorderTopSize() int {
	if e.Styles == nil {
		return 0
	}
	if style := e.GetBorderTopStyle(); style == "hidden" || style == "none" {
		return 0
	}
	val, err := e.Styles.GetBorderSizeInPx("top")
	switch err {
	case css.InheritValue:
		if e.Parent == nil {
			return 0
		}
		return e.Parent.GetBorderTopSize()
	case css.NoStyles:
		return 0
	}
	return val
}
func (e RenderableDomElement) GetBorderTopColor() *color.RGBA {
	if e.Styles == nil {
		return dfltBorder
	}
	val, err := e.Styles.GetBorderColor("top", dfltBorder)
	if err == css.InheritValue {
		if e.Parent == nil {
			return dfltBorder
		}
		return e.Parent.GetBorderTopColor()
	}
	return val
}

func (e RenderableDomElement) GetBorderLeftSize() int {
	if e.Styles == nil {
		return 0
	}
	if style := e.GetBorderLeftStyle(); style == "hidden" || style == "none" {
		return 0
	}
	val, err := e.Styles.GetBorderSizeInPx("left")
	switch err {
	case css.InheritValue:
		if e.Parent == nil {
			return 0
		}
		return e.Parent.GetBorderLeftSize()
	case css.NoStyles:
		return 0
	}
	return val
}
func (e RenderableDomElement) GetBorderLeftColor() *color.RGBA {
	if e.Styles == nil {
		return dfltBorder
	}
	val, err := e.Styles.GetBorderColor("left", dfltBorder)
	if err == css.InheritValue {
		if e.Parent == nil {
			return dfltBorder
		}
		return e.Parent.GetBorderLeftColor()
	}
	return val
}

func (e RenderableDomElement) GetBorderRightSize() int {
	if e.Styles == nil {
		return 0
	}
	if style := e.GetBorderRightStyle(); style == "hidden" || style == "none" {
		return 0
	}
	val, err := e.Styles.GetBorderSizeInPx("right")
	switch err {
	case css.InheritValue:
		if e.Parent == nil {
			return 0
		}
		return e.Parent.GetBorderRightSize()
	case css.NoStyles:
		return 0
	}
	return val
}
func (e RenderableDomElement) GetBorderRightColor() *color.RGBA {
	if e.Styles == nil {
		return dfltBorder
	}
	val, err := e.Styles.GetBorderColor("right", dfltBorder)
	if err == css.InheritValue {
		if e.Parent == nil {
			return dfltBorder
		}
		return e.Parent.GetBorderRightColor()
	}
	return val
}
func (e RenderableDomElement) GetMarginLeftSize() int {
	if e.Styles == nil {
		return 0
	}
	val, err := e.Styles.FollowCascadeToPx("margin-left", e.GetFontSize())
	switch err {
	case css.InheritValue:
		if e.Parent == nil {
			return 0
		}
		return e.Parent.GetMarginLeftSize()
	case css.NoStyles:
		return 0
	}
	return val
}
func (e RenderableDomElement) GetMarginRightSize() int {
	if e.Styles == nil {
		return 0
	}
	val, err := e.Styles.FollowCascadeToPx("margin-right", e.GetFontSize())
	switch err {
	case css.InheritValue:
		if e.Parent == nil {
			return 0
		}
		return e.Parent.GetMarginRightSize()
	case css.NoStyles:
		return 0
	}
	return val
}
func (e RenderableDomElement) GetMarginTopSize() int {
	if e.Styles == nil {
		return 0
	}
	val, err := e.Styles.FollowCascadeToPx("margin-top", e.GetFontSize())
	switch err {
	case css.InheritValue:
		if e.Parent == nil {
			return 0
		}
		return e.Parent.GetMarginTopSize()
	case css.NoStyles:
		return 0
	}
	return val
}
func (e RenderableDomElement) GetMarginBottomSize() int {
	if e.Styles == nil {
		return 0
	}
	val, err := e.Styles.FollowCascadeToPx("margin-bottom", e.GetFontSize())
	switch err {
	case css.InheritValue:
		if e.Parent == nil {
			return 0
		}
		return e.Parent.GetMarginBottomSize()
	case css.NoStyles:
		return 0
	}
	return val
}
func (e RenderableDomElement) GetPaddingLeftSize() int {
	if e.Styles == nil {
		return 0
	}
	val, err := e.Styles.FollowCascadeToPx("padding-left", e.GetFontSize())
	switch err {
	case css.InheritValue:
		if e.Parent == nil {
			return 0
		}
		return e.Parent.GetPaddingLeftSize()
	case css.NoStyles:
		return 0
	}
	return val
}
func (e RenderableDomElement) GetPaddingRightSize() int {
	if e.Styles == nil {
		return 0
	}
	val, err := e.Styles.FollowCascadeToPx("padding-right", e.GetFontSize())
	switch err {
	case css.InheritValue:
		if e.Parent == nil {
			return 0
		}
		return e.Parent.GetPaddingRightSize()
	case css.NoStyles:
		return 0
	}
	return val
}
func (e RenderableDomElement) GetPaddingTopSize() int {
	if e.Styles == nil {
		return 0
	}
	val, err := e.Styles.FollowCascadeToPx("padding-top", e.GetFontSize())
	switch err {
	case css.InheritValue:
		if e.Parent == nil {
			return 0
		}
		return e.Parent.GetPaddingTopSize()
	case css.NoStyles:
		return 0
	}
	return val
}
func (e RenderableDomElement) GetPaddingBottomSize() int {
	if e.Styles == nil {
		return 0
	}
	val, err := e.Styles.FollowCascadeToPx("padding-bottom", e.GetFontSize())
	switch err {
	case css.InheritValue:
		if e.Parent == nil {
			return 0
		}
		return e.Parent.GetPaddingBottomSize()
	case css.NoStyles:
		return 0
	}
	return val
}
func (e RenderableDomElement) GetBorderTopStyle() string {
	if e.Styles == nil {
		return "none"
	}
	val, err := e.Styles.FollowCascadeToString("border-top-style")
	switch err {
	case css.InheritValue:
		if e.Parent == nil {
			return "none"
		}
		return e.Parent.GetBorderTopStyle()
	case css.NoStyles:
		return "none"
	}
	return val
}
func (e RenderableDomElement) GetBorderBottomStyle() string {
	if e.Styles == nil {
		return "none"
	}
	val, err := e.Styles.FollowCascadeToString("border-bottom-style")
	switch err {
	case css.InheritValue:
		if e.Parent == nil {
			return "none"
		}
		return e.Parent.GetBorderBottomStyle()
	case css.NoStyles:
		return "none"
	}
	return val
}
func (e RenderableDomElement) GetBorderLeftStyle() string {
	if e.Styles == nil {
		return "none"
	}
	val, err := e.Styles.FollowCascadeToString("border-left-style")
	switch err {
	case css.InheritValue:
		if e.Parent == nil {
			return "none"
		}
		return e.Parent.GetBorderLeftStyle()
	case css.NoStyles:
		return "none"
	}
	return val
}
func (e RenderableDomElement) GetBorderRightStyle() string {
	if e.Styles == nil {
		return "none"
	}
	val, err := e.Styles.FollowCascadeToString("border-right-style")
	switch err {
	case css.InheritValue:
		if e.Parent == nil {
			return "none"
		}
		return e.Parent.GetBorderRightStyle()
	case css.NoStyles:
		return "none"
	}
	return val
}

func (e RenderableDomElement) getCSSBox(img image.Image) *outerBoxDrawer {
	bg := e.GetBackgroundColor()
	if bg == nil {
		bg = dfltBorder
	}
	return &outerBoxDrawer{
		Margin: BoxMargins{
			Top:    BoxMargin{Width: e.GetMarginTopSize()},
			Left:   BoxMargin{Width: e.GetMarginLeftSize()},
			Right:  BoxMargin{Width: e.GetMarginRightSize()},
			Bottom: BoxMargin{Width: e.GetMarginBottomSize()},
		},
		Border: BoxBorders{
			Top:    BoxBorder{BoxOffset: BoxOffset{Width: e.GetBorderTopSize()}, Color: e.GetBorderTopColor(), Style: e.GetBorderTopStyle()},
			Left:   BoxBorder{BoxOffset: BoxOffset{Width: e.GetBorderLeftSize()}, Color: e.GetBorderLeftColor(), Style: e.GetBorderLeftStyle()},
			Right:  BoxBorder{BoxOffset: BoxOffset{Width: e.GetBorderRightSize()}, Color: e.GetBorderRightColor(), Style: e.GetBorderRightStyle()},
			Bottom: BoxBorder{BoxOffset: BoxOffset{Width: e.GetBorderBottomSize()}, Color: e.GetBorderBottomColor(), Style: e.GetBorderBottomStyle()},
		},
		Padding: BoxPaddings{
			Top:    BoxPadding{Width: e.GetPaddingTopSize()},
			Left:   BoxPadding{Width: e.GetPaddingLeftSize()},
			Right:  BoxPadding{Width: e.GetPaddingRightSize()},
			Bottom: BoxPadding{Width: e.GetPaddingBottomSize()},
		},
		contentSize: img.Bounds().Size(),
		background:  &image.Uniform{bg},
	}
}
