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
			return (e.containerWidth - e.contentWidth) / 2
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
func (e RenderableDomElement) GetPaddingLeftSize() int {
	value := e.Styles.PaddingLeft.GetValue()
	if value == "" {
		// No style, use default.
		return 0
	}
	if value == "inherit" {
		if e.Parent == nil {
			return 0
		}
		return e.Parent.GetPaddingLeftSize()

	}
	fontsize := e.GetFontSize()
	val, err := css.ConvertUnitToPx(fontsize, e.containerWidth, value)
	if err != nil {
		return 0
	}
	return val
}
func (e RenderableDomElement) GetPaddingRightSize() int {
	value := e.Styles.PaddingRight.GetValue()
	if value == "" {
		// No style, use default.
		return 0
	}
	if value == "inherit" {
		if e.Parent == nil {
			return 0
		}
		return e.Parent.GetPaddingRightSize()

	}
	fontsize := e.GetFontSize()
	val, err := css.ConvertUnitToPx(fontsize, e.containerWidth, value)
	if err != nil {
		return 0
	}
	return val
}
func (e RenderableDomElement) GetPaddingTopSize() int {
	value := e.Styles.PaddingTop.GetValue()
	if value == "" {
		// No style, use default.
		return 0
	}
	if value == "inherit" {
		if e.Parent == nil {
			return 0
		}
		return e.Parent.GetPaddingTopSize()

	}
	fontsize := e.GetFontSize()
	val, err := css.ConvertUnitToPx(fontsize, e.containerWidth, value)
	if err != nil {
		return 0
	}
	return val
}
func (e RenderableDomElement) GetPaddingBottomSize() int {
	value := e.Styles.PaddingBottom.GetValue()
	if value == "" {
		// No style, use default.
		return 0
	}
	if value == "inherit" {
		if e.Parent == nil {
			return 0
		}
		return e.Parent.GetPaddingBottomSize()

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
			Top:    BoxBorder{BoxOffset: BoxOffset{Width: e.GetBorderTopWidth()}, Color: e.GetBorderTopColor(), Style: e.GetBorderTopStyle()},
			Left:   BoxBorder{BoxOffset: BoxOffset{Width: e.GetBorderLeftWidth()}, Color: e.GetBorderLeftColor(), Style: e.GetBorderLeftStyle()},
			Right:  BoxBorder{BoxOffset: BoxOffset{Width: e.GetBorderRightWidth()}, Color: e.GetBorderRightColor(), Style: e.GetBorderRightStyle()},
			Bottom: BoxBorder{BoxOffset: BoxOffset{Width: e.GetBorderBottomWidth()}, Color: e.GetBorderBottomColor(), Style: e.GetBorderBottomStyle()},
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
