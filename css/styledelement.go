package css

import (
	//	"fmt"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"image/color"
	"io/ioutil"
	"sort"
	"strings"
)

const (
	DefaultFontSize = 16
)

var SansSerifFont *truetype.Font
var sansSerifFontSizeCache map[int]font.Face

func init() {
	fontBytes, _ := ioutil.ReadFile("/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf")
	SansSerifFont, _ = truetype.Parse(fontBytes)
	sansSerifFontSizeCache = make(map[int]font.Face)
}

type StyleSource uint8

func (s StyleSource) String() string {
	switch s {
	case UserAgentSrc:
		return "User Agent"
	case UserSrc:
		return "User"
	case AuthorSrc:
		return "Author"
	}
	return "Unknown Source"
}

const (
	UnknownSrc StyleSource = iota
	UserAgentSrc
	UserSrc
	AuthorSrc
	InlineStyleSrc
)

// A StyleElement is anything that has CSS rules applied to it.
type StyledElement struct {
	// CSS Level 1 attributes
	// Shorthand attributes are commented out, because
	// there's not any time in code you would want them
	// instead of the expanded versions..
	FontFamily  StyleValue
	FontStyle   StyleValue
	FontVariant StyleValue
	FontWeight  StyleValue
	FontSize    StyleValue
	//Font        StyleValue

	Color                StyleValue
	BackgroundColor      StyleValue
	BackgroundImage      StyleValue
	BackgroundRepeat     StyleValue
	BackgroundAttachment StyleValue
	BackgroundPosition   StyleValue
	//Background           StyleValue

	WordSpacing   StyleValue
	LetterSpacing StyleValue

	TextDecoration StyleValue
	VerticalAlign  StyleValue
	TextTransform  StyleValue
	TextAlign      StyleValue
	TextIndent     StyleValue
	LineHeight     StyleValue

	MarginTop    StyleValue
	MarginRight  StyleValue
	MarginBottom StyleValue
	MarginLeft   StyleValue
	//Margin       StyleValue

	PaddingTop    StyleValue
	PaddingRight  StyleValue
	PaddingBottom StyleValue
	PaddingLeft   StyleValue
	//Padding       StyleValue

	BorderTopWidth StyleValue
	BorderTopColor StyleValue
	BorderTopStyle StyleValue

	BorderBottomWidth StyleValue
	BorderBottomColor StyleValue
	BorderBottomStyle StyleValue

	BorderLeftWidth StyleValue
	BorderLeftColor StyleValue
	BorderLeftStyle StyleValue

	BorderRightWidth StyleValue
	BorderRightColor StyleValue
	BorderRightStyle StyleValue

	/*
		BorderWidth StyleValue
		BorderStyle StyleValue
		BorderColor StyleValue

		Border	    StyleValue
	*/

	Width  StyleValue
	Height StyleValue

	Float StyleValue
	Clear StyleValue

	Display StyleValue

	WhiteSpace StyleValue

	ListStyleType     StyleValue
	ListStyleImage    StyleValue
	ListStylePosition StyleValue
	// ListStyle StyleValue

	// The rules that match this element.
	rules    []StyleRule
	fontSize int
}

func (e *StyledElement) SetFontSize(size int) {
	e.fontSize = size
}
func (e StyledElement) GetFontFace(fsize int) font.Face {
	if face, ok := sansSerifFontSizeCache[fsize]; ok {
		return face
	}

	face := truetype.NewFace(SansSerifFont,
		&truetype.Options{
			Size:    float64(fsize),
			DPI:     72,
			Hinting: font.HintingFull})
	sansSerifFontSizeCache[fsize] = face
	return face

}
func (e StyledElement) GetFontSize() (int, error) {
	if e.fontSize == 0 {
		return 0, InheritValue
	}
	return e.fontSize, nil
}

func (e *StyledElement) expandBoxBorderShorthand(att StyleAttribute, s StyleRule) {
	values := strings.Fields(s.Value.string)
	switch len(values) {
	case 0:
		return
	case 1:
		s.Name = "border-top-" + att
		s.Value.string = values[0]
		e.rules = append(e.rules, s)
		s.Name = "border-right-" + att
		e.rules = append(e.rules, s)
		s.Name = "border-bottom-" + att
		e.rules = append(e.rules, s)
		s.Name = "border-left-" + att
		e.rules = append(e.rules, s)
	case 2:
		s.Name = "border-top-" + att
		s.Value.string = values[0]

		e.rules = append(e.rules, s)
		s.Name = "border-bottom-" + att
		e.rules = append(e.rules, s)

		s.Value.string = values[1]
		s.Name = "border-right-" + att
		e.rules = append(e.rules, s)
		s.Name = "border-left-" + att
		e.rules = append(e.rules, s)
	case 3:
		s.Name = "border-top-" + att
		s.Value.string = values[0]
		e.rules = append(e.rules, s)

		s.Name = "border-right-" + att
		s.Value.string = values[1]
		e.rules = append(e.rules, s)
		s.Name = "border-left-" + att
		e.rules = append(e.rules, s)

		s.Name = "border-bottom-" + att
		s.Value.string = values[2]
		e.rules = append(e.rules, s)
	case 4:
		fallthrough
	default:
		s.Name = "border-top-" + att
		s.Value.string = values[0]
		e.rules = append(e.rules, s)

		s.Name = "border-right-" + att
		s.Value.string = values[1]
		e.rules = append(e.rules, s)

		s.Name = "border-bottom-" + att
		s.Value.string = values[2]
		e.rules = append(e.rules, s)

		s.Name = "border-left-" + att
		s.Value.string = values[3]
		e.rules = append(e.rules, s)
	}
}
func (e *StyledElement) expandBoxSideShorthand(attrib StyleAttribute, s StyleRule) {
	values := strings.Fields(s.Value.string)
	switch len(values) {
	case 0:
		return
	case 1:
		s.Name = attrib + "-top"
		s.Value.string = values[0]
		e.rules = append(e.rules, s)
		s.Name = attrib + "-right"
		e.rules = append(e.rules, s)
		s.Name = attrib + "-bottom"
		e.rules = append(e.rules, s)
		s.Name = attrib + "-left"
		e.rules = append(e.rules, s)
	case 2:
		s.Name = attrib + "-top"
		s.Value.string = values[0]
		e.rules = append(e.rules, s)
		s.Name = attrib + "-bottom"
		e.rules = append(e.rules, s)

		s.Name = attrib + "-right"
		s.Value.string = values[1]
		e.rules = append(e.rules, s)
		s.Name = attrib + "-left"
		e.rules = append(e.rules, s)
	case 3:
		s.Name = attrib + "-top"
		s.Value.string = values[0]
		e.rules = append(e.rules, s)

		s.Name = attrib + "-right"
		s.Value.string = values[1]
		e.rules = append(e.rules, s)
		s.Name = attrib + "-left"
		e.rules = append(e.rules, s)

		s.Name = attrib + "-bottom"
		s.Value.string = values[2]
		e.rules = append(e.rules, s)
	case 4:
		fallthrough
	default:
		s.Name = attrib + "-top"
		s.Value.string = values[0]
		e.rules = append(e.rules, s)

		s.Name = attrib + "-right"
		s.Value.string = values[1]
		e.rules = append(e.rules, s)

		s.Name = attrib + "-bottom"
		s.Value.string = values[2]
		e.rules = append(e.rules, s)

		s.Name = attrib + "-left"
		s.Value.string = values[3]
		e.rules = append(e.rules, s)
	}
}
func (e *StyledElement) AddStyle(s StyleRule) {
	switch s.Name {
	case "padding", "margin":
		e.expandBoxSideShorthand(s.Name, s)
	case "border-width":
		// border width expands to border-side-width, not
		// border-width-side, so we can't use the normal helper
		// function
		e.expandBoxBorderShorthand("width", s)
	case "border-color":
		e.expandBoxBorderShorthand("color", s)
	case "border-style":
		e.expandBoxBorderShorthand("style", s)
	default:
		e.rules = append(e.rules, s)
	}
	return
}

// SortStyles will sort the rules on this element according to the CSS spec, which states
//
// 1. Find all declarations that apply too element/property (already done when this is called)
// 2. Sort according to importance (normal or important) and origin (author, user, or user agent). In ascending order of precedence:
//	1. user agent declarations (defaults)
//	2. user normal declrations (don't exist)
//	3. author normal declarations
//	4. author important declarations
//	5. user important declarations (don't exist)
// 3. Sort rules with the same importance and origin by specificity of selector: more specific selectors will override more general ones. Pseudo-elements and pseudo-classes are counted as normal elements and classes, respectively.
// 4. Finally, sort by order specified: if two declarations have the same weight, origin, and specificity, the latter specified wins. Declarations in imported stylesheets are considered to be before any declaration in the style sheet itself
// BUG(driusan): The final tie break is not implemented
func (e *StyledElement) SortStyles() error {
	sort.Sort(byCSSPrecedence(e.rules))
	e.populateValues()
	return nil
}

func (e *StyledElement) populateValues() {
	// rules are sorted so that the lowest index has the highest precedent, so going through
	// backwords and just blindly populating will ensure that the highest precedent rule
	// is the last to update the value
	for i := len(e.rules) - 1; i >= 0; i-- {
		switch rule := e.rules[i]; rule.Name {
		case "font-family":
			e.FontFamily = rule.Value
		case "font-style":
			e.FontStyle = rule.Value
		case "font-variant":
			e.FontVariant = rule.Value
		case "font-weight":
			e.FontWeight = rule.Value
		case "font-size":
			e.FontSize = rule.Value

		case "color":
			e.Color = rule.Value
		case "background-color":
			e.BackgroundColor = rule.Value
		case "background-image":
			e.BackgroundImage = rule.Value
		case "background-repeat":
			e.BackgroundRepeat = rule.Value
		case "background-attachment":
			e.BackgroundAttachment = rule.Value
		case "background-position":
			e.BackgroundPosition = rule.Value
		case "word-spacing":
			e.WordSpacing = rule.Value
		case "letter-spacing":
			e.LetterSpacing = rule.Value

		case "text-decoration":
			e.TextDecoration = rule.Value
		case "vertical-align":
			e.VerticalAlign = rule.Value
		case "text-transform":
			e.TextTransform = rule.Value
		case "text-align":
			e.TextAlign = rule.Value
		case "text-indent":
			e.TextIndent = rule.Value
		case "line-height":
			e.LineHeight = rule.Value

		case "margin-right":
			e.MarginRight = rule.Value
		case "margin-left":
			e.MarginLeft = rule.Value
		case "margin-top":
			e.MarginTop = rule.Value
		case "margin-bottom":
			e.MarginBottom = rule.Value

		case "padding-right":
			e.PaddingRight = rule.Value
		case "padding-left":
			e.PaddingLeft = rule.Value
		case "padding-top":
			e.PaddingTop = rule.Value
		case "padding-bottom":
			e.PaddingBottom = rule.Value

		case "border-top-width":
			e.BorderTopWidth = rule.Value
		case "border-top-color":
			e.BorderTopColor = rule.Value
		case "border-top-style":
			e.BorderTopStyle = rule.Value

		case "border-bottom-width":
			e.BorderBottomWidth = rule.Value
		case "border-bottom-color":
			e.BorderBottomColor = rule.Value
		case "border-bottom-style":
			e.BorderBottomStyle = rule.Value

		case "border-left-width":
			e.BorderLeftWidth = rule.Value
		case "border-left-color":
			e.BorderLeftColor = rule.Value
		case "border-left-style":
			e.BorderLeftStyle = rule.Value

		case "border-right-width":
			e.BorderRightWidth = rule.Value
		case "border-right-color":
			e.BorderRightColor = rule.Value
		case "border-right-style":
			e.BorderRightStyle = rule.Value

		case "width":
			e.Width = rule.Value
		case "height":
			e.Height = rule.Value

		case "float":
			e.Float = rule.Value
		case "clear":
			e.Clear = rule.Value
		case "display":
			e.Display = rule.Value
		case "white-space":
			e.WhiteSpace = rule.Value

		case "list-style-type":
			e.ListStyleType = rule.Value
		case "list-style-image":
			e.ListStyleImage = rule.Value
		case "list-style-position":
			e.ListStylePosition = rule.Value
		}
	}
}

func (e *StyledElement) DisplayProp() string {
	for _, rule := range e.rules {
		if string(rule.Name) == "display" {
			return rule.Value.string
		}
	}
	return ""
}

func (e *StyledElement) GetAttribute(attr string) StyleValue {
	for _, rule := range e.rules {
		if string(rule.Name) == attr {
			return rule.Value
		}
	}
	return StyleValue{"", false}
}

func (e StyledElement) GetBackgroundColor(defaultColour color.Color) (color.Color, error) {
	switch e.BackgroundColor.string {
	case "inherit":
		return defaultColour, InheritValue
	case "transparent":
		return color.Transparent, nil
	case "":
		return color.Transparent, NoStyles
	default:
		return ConvertColorToRGBA(e.BackgroundColor.string)
	}
}

func (e StyledElement) GetColor(defaultColour color.Color) (color.Color, error) {
	switch e.Color.string {
	case "inherit":
		return defaultColour, InheritValue
	case "transparent":
		return defaultColour, nil
	case "":
		return defaultColour, NoStyles
	default:
		return ConvertColorToRGBA(e.Color.string)
	}
}
