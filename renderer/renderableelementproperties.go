package renderer

import (
	"github.com/driusan/Gob/css"
	"golang.org/x/image/font"
	"golang.org/x/net/html"
	"image/color"
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
	// inheritd == yes
	// percentage relative to the font size of the element itself
	fSize := e.GetFontSize()
	if e.Styles == nil {
		if e.Parent == nil {
			fontFace := e.Styles.GetFontFace(fSize)
			return getFontHeight(fontFace)
		}
		//return e.Parent.GetLineHeight()
		fontFace := e.Styles.GetFontFace(fSize)
		return getFontHeight(fontFace)
	}
	stringVal := e.Styles.LineHeight.GetValue()
	if stringVal == "" {
		if e.Parent == nil {
			fontFace := e.Styles.GetFontFace(fSize)
			return getFontHeight(fontFace)
		}
		//return e.Parent.GetLineHeight()
		fontFace := e.Styles.GetFontFace(fSize)
		return getFontHeight(fontFace)

	}
	lHeightSize, err := css.ConvertUnitToPx(fSize, fSize, stringVal)
	if err != nil {
		fontFace := e.Styles.GetFontFace(fSize)
		return getFontHeight(fontFace)
	}
	fontFace := e.Styles.GetFontFace(lHeightSize)
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

func (e RenderableDomElement) GetBackgroundColor() color.Color {
	switch bg, err := e.Styles.GetBackgroundColor(dfltBackground); err {
	case css.InheritValue:
		if e.Parent == nil {
			return color.Transparent
		}
		return e.Parent.GetBackgroundColor()
	case css.NoStyles:
		return color.Transparent
	default:
		return bg
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
			default:
				return cssVal

			}
		}
		return cssVal
	}
	// CSS Level 1 default is block, CSS Level 2 is inline
	return "block"
	//return "inline"
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

func (e RenderableDomElement) GetContentWidth(containerWidth int) int {
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
		return e.Parent.GetContentWidth(containerWidth)
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
func (e RenderableDomElement) GetContentHeight() int {
	if e.Styles == nil {
		return 0
	}
	cssVal := e.Styles.Height.GetValue()
	switch cssVal {
	case "inherit":
		if e.Parent == nil {
			return 0
		}
		return e.Parent.GetContentHeight()
	case "", "auto":
		return 0
	default:
		calVal, err := css.ConvertUnitToPx(e.GetFontSize(), 0, cssVal)
		if err == nil {
			return calVal
		}
		return 0
	}
}
