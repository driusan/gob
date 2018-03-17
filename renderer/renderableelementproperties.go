package renderer

import (
	"fmt"
	"github.com/driusan/Gob/css"
	"golang.org/x/image/font"
	"golang.org/x/net/html"
	"image/color"
	"os"
	"strings"
)

func getFontHeight(face font.Face) int {
	metrics := face.Metrics()
	return (metrics.Ascent + metrics.Descent).Ceil()
}
func stringSize(fntDrawer font.Drawer, textContent string) (int, error) {

	var size int
	words := strings.Fields(textContent)
	fSize := getFontHeight(fntDrawer.Face)
	//firstRune, _ := utf8.DecodeRuneInString(textContent)

	for i, word := range words {
		wordSizeInPx := fntDrawer.MeasureString(word).Ceil()
		size += wordSizeInPx

		if i == len(words)-1 {
			break
		}

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
	// inherited == yes
	// percentage relative to the font size of the element itself
	fSize := e.GetFontSize()
	if e.Styles == nil {
		if e.Parent == nil {
			fontFace := e.GetFontFace(fSize)
			return getFontHeight(fontFace)
		}
		return e.Parent.getLineHeight(fSize)
	}
	stringVal := e.Styles.LineHeight.GetValue()
	if stringVal == "" || stringVal == "inherit" {
		if e.Parent == nil {
			fontFace := e.GetFontFace(fSize)
			return getFontHeight(fontFace)
		}
		return e.Parent.getLineHeight(fSize)
	}
	return e.getLineHeight(fSize)
}

// calculate the line height in pixels assuming a font size of fsize
// (Used to ensure when a child inherits the line height, it's relative to
// its own font size, not the parent's.
func (e *RenderableDomElement) getLineHeight(fSize int) int {
	stringVal := e.Styles.LineHeight.GetValue()
	if stringVal == "" || stringVal == "inherit" {
		if e.Parent == nil {
			fontFace := e.GetFontFace(fSize)
			return getFontHeight(fontFace)
		}
		return e.Parent.getLineHeight(fSize)
	}
	lHeightSize, err := css.ConvertUnitToPx(fSize, fSize, stringVal)
	if err != nil {
		fontFace := e.GetFontFace(fSize)
		return getFontHeight(fontFace)
	}
	fontFace := e.GetFontFace(lHeightSize)
	return getFontHeight(fontFace)
}

func (e *RenderableDomElement) GetFontSize() int {
	fromCSS, err := e.Styles.GetFontSize()
	switch err {
	case css.NoStyles, css.InheritValue:
		if e.Parent == nil {
			return css.DefaultFontSize
		}
		return e.Parent.GetFontSize()
	case nil:
		return fromCSS
	default:
		panic("Could not determine font size")

	}
}

func (e RenderableDomElement) GetBackgroundColor() color.Color {
	switch bgc := e.Styles.BackgroundColor.GetValue(); bgc {
	case "inherit":
		if e.Parent == nil {
			return dfltBackground
		}
		return e.Parent.GetBackgroundColor()
	case "", "transparent":
		return color.Transparent
	default:
		c, err := css.ConvertColorToRGBA(bgc)
		if err != nil {
			return color.Transparent
			//panic(err)
		}
		return c

	}
}

func (e RenderableDomElement) GetColor() color.Color {
	var deflt color.RGBA
	if e.Type == html.ElementNode && e.Data == "a" {
		deflt = color.RGBA{0, 0, 0xFF, 0xFF}
	} else {
		deflt = color.RGBA{0, 0, 0, 0xFF}
	}
	switch cssColor, err := e.Styles.GetColor(deflt); err {
	case css.InheritValue:
		if e.Parent == nil {
			return deflt
		}
		return e.Parent.GetColor()
	case css.NoStyles:
		if e.Parent == nil {
			return deflt
		}
		return e.Parent.GetColor()
	default:
		return cssColor
	}
}

func (e RenderableDomElement) GetFloat() string {
	switch float := e.Styles.Float.GetValue(); float {
	case "inherit":
		return e.Parent.GetFloat()
	case "left", "right", "none":
		return float
	default:
		return "none"

	}
}
func (e RenderableDomElement) GetDisplayProp() string {
	if e.Type == html.TextNode {
		return "inline"
	}
	if cssVal := e.Styles.DisplayProp(); cssVal != "" {
		// Apply section 9.7 of CSS spec: Relationships between display, position, and float"
		if cssVal == "none" {
			return cssVal
		}
		// position not yet implemented for point 2.

		switch e.GetFloat() {
		case "none":
			return cssVal
		default:
			switch cssVal {
			case "inline-table":
				return "table"
			case "inline", "table-row-group", "table-column", "table-column-group",
				"table-footer-group", "table-row", "table-cell", "table-caption",
				"inline-block":
				return "block"
			case "list-item":
				// FIXME: This should generate a principle box and a marker box,
				// but for now just pretending it's a block simplifies things
				return "list-item"
			default:
				return cssVal

			}
		}
		return cssVal
	}
	// CSS Level 1 default is block, CSS Level 2 is inline
	return "inline"
}

func (e *RenderableDomElement) GetTextDecoration() string {
	if e.Styles == nil {
		if e.GetDisplayProp() == "inline" {
			if e.Parent != nil {
				return e.Parent.GetTextDecoration()
			}
		}
		return "none"
	}

	switch decoration := e.Styles.TextDecoration.GetValue(); decoration {
	case "inherit":
		return e.Parent.GetTextDecoration()
	default:
		trimmed := strings.TrimSpace(decoration)
		if trimmed != "" {
			return trimmed
		}
		if e.GetDisplayProp() == "inline" {
			if e.Parent != nil {
				return e.Parent.GetTextDecoration()
			}
		}
		return "none"
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

func (e RenderableDomElement) GetContainerWidth(containerWidth int) int {
	width := containerWidth - (e.GetMarginLeftSize() + e.GetMarginRightSize() + e.GetBorderLeftWidth() + e.GetBorderRightWidth() + e.GetPaddingLeft() + e.GetPaddingRight())
	if e.Styles == nil {
		return width
	}
	cssVal := e.Styles.Width.GetValue()
	switch cssVal {
	case "inherit":
		if e.Parent == nil {
			return width
		}
		return e.Parent.GetContainerWidth(containerWidth)
	case "", "auto":
		return width
	default:
		calVal, err := css.ConvertUnitToPx(e.GetFontSize(), containerWidth, cssVal)
		if err == nil {
			return calVal
		}
		return width
	}
}
func (e RenderableDomElement) GetMaxHeight() int {
	if e.Styles == nil {
		return -1
	}
	cssVal := e.Styles.MaxHeight.GetValue()
	switch cssVal {
	case "inherit":
		if e.Parent == nil {
			return -1
		}
		return e.Parent.GetMaxHeight()
	case "":
		return -1
	default:
		calVal, err := css.ConvertUnitToPx(e.GetFontSize(), 0, cssVal)
		if err == nil {
			return calVal
		}
		return -1
	}
}
func (e RenderableDomElement) GetMaxWidth() int {
	if e.Styles == nil {
		return -1
	}
	cssVal := e.Styles.MaxWidth.GetValue()
	switch cssVal {
	case "inherit":
		if e.Parent == nil {
			return -1
		}
		return e.Parent.GetMaxWidth()
	case "":
		return -1
	default:
		calVal, err := css.ConvertUnitToPx(e.GetFontSize(), 0, cssVal)
		if err == nil {
			return calVal
		}
		return -1
	}
}
func (e RenderableDomElement) GetHeight() int {
	if e.Styles == nil {
		return -1
	}
	cssVal := e.Styles.Height.GetValue()
	switch cssVal {
	case "inherit":
		if e.Parent == nil {
			return -1
		}
		return e.Parent.GetHeight()
	case "", "auto":
		return -1
	default:
		calVal, err := css.ConvertUnitToPx(e.GetFontSize(), e.containerHeight, cssVal)
		if err == nil {
			return calVal
		}
		return -1
	}
}
func (e RenderableDomElement) GetWidth() int {
	if e.Styles == nil {
		return -1
	}
	cssVal := e.Styles.Width.GetValue()
	switch cssVal {
	case "inherit":
		if e.Parent == nil {
			return -1
		}
		return e.Parent.GetWidth()
	case "", "auto":
		return -1
	default:
		calVal, err := css.ConvertUnitToPx(e.GetFontSize(), e.containerWidth, cssVal)
		if err == nil {
			return calVal
		}
		return -1
	}
}
func (e RenderableDomElement) GetMinWidth() int {
	if e.Styles == nil {
		return 0
	}
	cssVal := e.Styles.MinWidth.GetValue()
	switch cssVal {
	case "inherit":
		if e.Parent == nil {
			return 0
		}
		return e.Parent.GetMinWidth()
	case "":
		return 0
	default:
		calVal, err := css.ConvertUnitToPx(e.GetFontSize(), 0, cssVal)
		if err == nil {
			return calVal
		}
		return 0
	}
}
func (e RenderableDomElement) GetMinHeight() int {
	if e.Styles == nil {
		return 0
	}
	cssVal := e.Styles.MinHeight.GetValue()
	switch cssVal {
	case "inherit":
		if e.Parent == nil {
			return 0
		}
		return e.Parent.GetMinHeight()
	case "":
		return 0
	default:
		calVal, err := css.ConvertUnitToPx(e.GetFontSize(), 0, cssVal)
		if err == nil {
			return calVal
		}
		return 0
	}
}

func (e *RenderableDomElement) GetFontWeight() font.Weight {
	switch e.Styles.FontWeight.GetValue() {
	case "normal":
		return font.WeightNormal
	case "100":
		return font.WeightThin
	case "200":
		return font.WeightExtraLight
	case "300":
		return font.WeightLight
	case "400":
		return font.WeightNormal
	case "500":
		return font.WeightMedium
	case "600":
		return font.WeightSemiBold
	case "700", "bold":
		return font.WeightBold
	case "800":
		return font.WeightExtraBold
	case "900":
		return font.WeightBlack
	case "bolder":
		if e.Parent == nil {
			return font.WeightMedium
		}
		parent := e.Parent.GetFontWeight()
		if parent == font.WeightBlack {
			return font.WeightBlack
		}
		return parent + 1
	case "lighter":
		if e.Parent == nil {
			return font.WeightLight
		}
		parent := e.Parent.GetFontWeight()
		if parent == font.WeightThin {
			return font.WeightThin
		}
		return parent + 1
	case "inherit":
		fallthrough
	default:
		//inherit
		if e.Parent == nil {
			return font.WeightNormal
		}
		return e.Parent.GetFontWeight()

	}
}
func (e *RenderableDomElement) GetFontStyle() font.Style {
	switch s := e.Styles.FontStyle.GetValue(); s {
	case "normal":
		return font.StyleNormal
	case "italic":
		return font.StyleItalic
	case "oblique":
		return font.StyleOblique
	case "inherit":
		fallthrough
	default:
		if e.Parent == nil {
			return font.StyleNormal
		}
		return e.Parent.GetFontStyle()

	}
}

func (e *RenderableDomElement) GetFontFamily() css.FontFamily {
	switch s := strings.ToLower(e.Styles.FontStyle.GetValue()); s {
	case "sans-serif", "serif", "monospace":
		return css.FontFamily(s)
	case "fantasy", "cursive":
		//unhandled font families that are nonetheless valid.
		// fallback on sans-serif
		return css.FontFamily("sans-serif")
	}
	if e.Parent == nil {
		return css.FontFamily("sans-serif")
	}
	return e.Parent.GetFontFamily()

}
func (e *RenderableDomElement) GetFontFace(fsize int) font.Face {
	return e.Styles.GetFontFace(fsize, e.GetFontFamily(), e.GetFontWeight(), e.GetFontStyle())
}

func (e *RenderableDomElement) GetWhiteSpace() string {
	switch s := strings.ToLower(e.Styles.WhiteSpace.GetValue()); s {
	case "normal", "pre", "nowrap":
		return s
	case "pre-wrap", "pre-line":
		panic("Unimplemented WhiteSpace value: " + s)
	}
	// default is inherited, inherit will also fall through
	// to here.
	if e.Parent == nil {
		return "normal"
	}
	return e.Parent.GetWhiteSpace()

}
func (e *RenderableDomElement) GetOverflow() string {
	switch s := strings.ToLower(e.Styles.Overflow.GetValue()); s {
	case "visible", "":
		return "visible"
	case "hidden":
		return s
	case "scroll", "auto":
		fmt.Fprintf(os.Stderr, "Unimplemented WhiteSpace value: %s. Defaulting to visible.", s)
		return "visible"
	case "inherit":
		if e.Parent == nil {
			return "visible"
		}
		return e.Parent.GetOverflow()
	default:
		return "visible"
	}
}

func (e *RenderableDomElement) GetVerticalAlign() string {
	switch s := strings.ToLower(e.Styles.VerticalAlign.GetValue()); s {
	case "baseline", "sub", "super", "top", "text-top", "middle", "botom", "text-bottom":
		return s
	case "inherit":
		if e.Parent == nil {
			return "baseline"
		}
		return e.Parent.GetVerticalAlign()
	default:
		// FIXME: Handle lengths and percentage values
		return "baseline"
	}
}
