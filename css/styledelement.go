package css

import (
	"fmt"
	"github.com/driusan/fonts"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"os"
	//"golang.org/x/image/font/basicfont"
	"image/color"
	//"io/ioutil"
	"sort"
	"strings"
)

var DefaultFontSize int

type fontStyle struct {
	fontFamily FontFamily
	fontWeight font.Weight
	fontStyle  font.Style
	fontSize   int
	//font-variant not supported. Need a good small-caps font to implement..
}

var parsedFontCache map[string]*truetype.Font
var fontCache map[fontStyle]font.Face

func init() {
	fontCache = make(map[fontStyle]font.Face)
	parsedFontCache = make(map[string]*truetype.Font)
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

	// New CSS Level 2 Attributes
	// Table related properties
	TableLayout    StyleValue
	BorderCollapse StyleValue
	BorderSpacing  StyleValue
	EmptyCells     StyleValue
	CaptionSide    StyleValue

	// Visual Formatting related properties
	Position StyleValue
	Top      StyleValue
	Bottom   StyleValue
	Left     StyleValue
	Right    StyleValue

	Overflow   StyleValue
	Visibility StyleValue
	Clip       StyleValue
	ZIndex     StyleValue

	MinHeight StyleValue
	MaxHeight StyleValue
	MinWidth  StyleValue
	MaxWidth  StyleValue

	// Generated content related properties
	Content          StyleValue
	CounterIncrement StyleValue
	CounterReset     StyleValue
	Quotes           StyleValue

	// UI Related properties
	Cursor StyleValue

	// Other
	Direction   StyleValue
	UnicodeBidi StyleValue

	// The rules that match this element.
	rules    []StyleRule
	fontSize int
}

func (e StyledElement) String() string {
	return fmt.Sprintf("%v", e.rules)
}
func (e *StyledElement) SetFontSize(size int) {
	if size < 0 {
		e.fontSize = DefaultFontSize
	} else {
		e.fontSize = size
	}
}

func (e StyledElement) MergeStyles(other ...*StyledElement) StyledElement {
	for _, o := range other {
		e.rules = append(e.rules, o.rules...)
	}
	return e
}

type FontFamily string

func (e StyledElement) GetFontFace(fsize int, fontFamily FontFamily, weight font.Weight, style font.Style) font.Face {
	fStyle := fontStyle{
		fontFamily: fontFamily,
		fontWeight: weight,
		fontStyle:  style,
		fontSize:   fsize,
	}
	if face, ok := fontCache[fStyle]; ok {
		return face
	}

	var ttfFile string
	switch fStyle.fontFamily {
	default:
		fallthrough
	case "serif":
		switch style {
		default:
			fallthrough
		case font.StyleNormal:
			if weight <= font.WeightNormal {
				ttfFile = "DejaVuSerif.ttf"
			} else { // weight > font.WeightNormal
				ttfFile = "DejaVuSerif-Bold.ttf"
			}
		case font.StyleItalic, font.StyleOblique:
			if weight <= font.WeightNormal {
				ttfFile = "DejaVuSerif-Italic.ttf"
			} else { //fontWeight > font.WeightNormal
				ttfFile = "DejaVuSerif-BoldItalic.ttf"
			}

		}
	case "sans-serif":
		switch style {
		default:
			fallthrough
		case font.StyleNormal:
			// TODO: Look up the weight of DejaVuSans-ExtraLight and use
			// it as appropriate
			if weight <= font.WeightNormal {
				ttfFile = "DejaVuSans.ttf"
			} else { //fontWeight > font.WeightNormal
				ttfFile = "DejaVuSans-Bold.ttf"
			}
		case font.StyleItalic, font.StyleOblique:
			if weight <= font.WeightNormal {
				ttfFile = "DejaVuSans-Oblique.ttf"
			} else { //fontWeight > font.WeightNormal {
				ttfFile = "DejaVuSans-BoldOblique.ttf"
			}

		}
	case "monospace":
		switch style {
		default:
			fallthrough
		case font.StyleNormal:
			// TODO: Look up the weight of DejaVuSans-ExtraLight and use
			// it as appropriate
			if weight <= font.WeightNormal {
				ttfFile = "DejaVuSansMono.ttf"
			} else { //if fontWeight > font.WeightNormal {
				ttfFile = "DejaVuSansMono-Bold.ttf"
			}
		case font.StyleItalic, font.StyleOblique:
			if weight <= font.WeightNormal {
				ttfFile = "DejaVuSansMono-Oblique.ttf"
			} else { //if fontWeight > font.WeightNormal {
				ttfFile = "DejaVuSansMono-BoldOblique.ttf"
			}

		}

	}

	var ft *truetype.Font
	if fb, ok := parsedFontCache[ttfFile]; ok {
		ft = fb
	} else {
		fontBytes, err := fonts.Asset(ttfFile)
		if err != nil {
			panic(err)
		}
		ft, _ = truetype.Parse(fontBytes)
	}
	face := truetype.NewFace(ft,
		&truetype.Options{
			Size:    float64(fsize) / PixelsPerPt,
			DPI:     PixelsPerPt * 72,
			Hinting: font.HintingFull})
	fontCache[fStyle] = face
	return face

}
func (e StyledElement) GetFontSize() (int, error) {
	if e.fontSize == 0 {
		return 0, InheritValue
	}
	return e.fontSize, nil
}

func (e *StyledElement) expandBoxBorderShorthand(att StyleAttribute, s StyleRule) {
	values := strings.Fields(s.Value.Value)
	switch len(values) {
	case 0:
		return
	case 1:
		s.Name = "border-top-" + att
		s.Value.Value = values[0]
		e.rules = append(e.rules, s)
		s.Name = "border-right-" + att
		e.rules = append(e.rules, s)
		s.Name = "border-bottom-" + att
		e.rules = append(e.rules, s)
		s.Name = "border-left-" + att
		e.rules = append(e.rules, s)
	case 2:
		s.Name = "border-top-" + att
		s.Value.Value = values[0]

		e.rules = append(e.rules, s)
		s.Name = "border-bottom-" + att
		e.rules = append(e.rules, s)

		s.Value.Value = values[1]
		s.Name = "border-right-" + att
		e.rules = append(e.rules, s)
		s.Name = "border-left-" + att
		e.rules = append(e.rules, s)
	case 3:
		s.Name = "border-top-" + att
		s.Value.Value = values[0]
		e.rules = append(e.rules, s)

		s.Name = "border-right-" + att
		s.Value.Value = values[1]
		e.rules = append(e.rules, s)
		s.Name = "border-left-" + att
		e.rules = append(e.rules, s)

		s.Name = "border-bottom-" + att
		s.Value.Value = values[2]
		e.rules = append(e.rules, s)
	case 4:
		fallthrough
	default:
		s.Name = "border-top-" + att
		s.Value.Value = values[0]
		e.rules = append(e.rules, s)

		s.Name = "border-right-" + att
		s.Value.Value = values[1]
		e.rules = append(e.rules, s)

		s.Name = "border-bottom-" + att
		s.Value.Value = values[2]
		e.rules = append(e.rules, s)

		s.Name = "border-left-" + att
		s.Value.Value = values[3]
		e.rules = append(e.rules, s)
	}
}

func (e *StyledElement) expandBorderShorthand(attrib StyleAttribute, s StyleRule) {
	values := strings.Fields(s.Value.Value)
	for _, v := range values {
		if IsLength(v) {
			s.Value.Value = v
			e.expandBoxBorderShorthand("width", s)
		} else if IsBorderStyle(v) {
			s.Value.Value = v
			e.expandBoxBorderShorthand("style", s)
		} else if IsColor(v) {
			s.Value.Value = v
			e.expandBoxBorderShorthand("color", s)
		} else {
			fmt.Fprintln(os.Stderr, "Didn't know what to do with border property", v)
		}
	}
}
func (e *StyledElement) expandBoxSideShorthand(attrib StyleAttribute, s StyleRule) {
	values := strings.Fields(s.Value.Value)
	switch len(values) {
	case 0:
		return
	case 1:
		s.Name = attrib + "-top"
		s.Value.Value = values[0]
		e.rules = append(e.rules, s)
		s.Name = attrib + "-right"
		e.rules = append(e.rules, s)
		s.Name = attrib + "-bottom"
		e.rules = append(e.rules, s)
		s.Name = attrib + "-left"
		e.rules = append(e.rules, s)
	case 2:
		s.Name = attrib + "-top"
		s.Value.Value = values[0]
		e.rules = append(e.rules, s)
		s.Name = attrib + "-bottom"
		e.rules = append(e.rules, s)

		s.Name = attrib + "-right"
		s.Value.Value = values[1]
		e.rules = append(e.rules, s)
		s.Name = attrib + "-left"
		e.rules = append(e.rules, s)
	case 3:
		s.Name = attrib + "-top"
		s.Value.Value = values[0]
		e.rules = append(e.rules, s)

		s.Name = attrib + "-right"
		s.Value.Value = values[1]
		e.rules = append(e.rules, s)
		s.Name = attrib + "-left"
		e.rules = append(e.rules, s)

		s.Name = attrib + "-bottom"
		s.Value.Value = values[2]
		e.rules = append(e.rules, s)
	case 4:
		fallthrough
	default:
		s.Name = attrib + "-top"
		s.Value.Value = values[0]
		e.rules = append(e.rules, s)

		s.Name = attrib + "-right"
		s.Value.Value = values[1]
		e.rules = append(e.rules, s)

		s.Name = attrib + "-bottom"
		s.Value.Value = values[2]
		e.rules = append(e.rules, s)

		s.Name = attrib + "-left"
		s.Value.Value = values[3]
		e.rules = append(e.rules, s)
	}
}

func (e *StyledElement) expandBackgroundShorthand(s StyleRule) {
	values := strings.Fields(s.Value.Value)

	// hack to build background-color from the shorthand even
	// though we split on whitespace and rgb(1, 2, 3) has whitespace
	buildColour := false
	for _, val := range values {
		if buildColour {
			e.rules[len(e.rules)-1].Value.Value += (" " + val)
			if strings.Index(val, ")") >= 0 {
				buildColour = false
			}
		}
		if val == "none" || IsURL(val) {
			s.Name = "background-image"
			s.Value.Value = val
			e.rules = append(e.rules, s)
		} else if IsColor(val) {
			s.Name = "background-color"
			s.Value.Value = val
			e.rules = append(e.rules, s)
			if string(val[0:4]) == "rgb(" {
				buildColour = true
			}
		}
		switch val {
		case "repeat", "repeat-x", "repeat-y", "no-repeat":
			s.Name = "background-repeat"
			s.Value.Value = val
			e.rules = append(e.rules, s)
		case "scroll", "fixed":
			s.Name = "background-attachment"
			s.Value.Value = val
			e.rules = append(e.rules, s)
		case "left", "right", "top", "center", "bottom":
			s.Name = "background-position"
			s.Value.Value = val
			e.rules = append(e.rules, s)
		default:
			if IsPercentage(val) || IsLength(val) {
				s.Name = "background-position"
				s.Value.Value = val
				e.rules = append(e.rules, s)
			}
		}
	}
}

func (e *StyledElement) AddStyle(s StyleRule) {
	switch s.Name {
	case "border":
		e.expandBorderShorthand(s.Name, s)
	case "padding", "margin":
		e.expandBoxSideShorthand(s.Name, s)
	case "background":
		e.expandBackgroundShorthand(s)
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

func (e *StyledElement) ClearStyles() {
	e.rules = nil
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
		case "min-width":
			e.MinWidth = rule.Value
		case "min-height":
			e.MinHeight = rule.Value
		case "max-width":
			e.MaxWidth = rule.Value
		case "max-height":
			e.MaxHeight = rule.Value

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

		case "overflow":
			e.Overflow = rule.Value
		}

	}
}

func (e *StyledElement) DisplayProp() string {
	for _, rule := range e.rules {
		if string(rule.Name) == "display" {
			return rule.Value.Value
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

// Returns the URL as a string, and an error. Retrieving the URL
// is left as an exercise for the caller, since the CSS package doesn't
// know the host/path to resolve relative URLs
func (e StyledElement) GetBackgroundImage() (string, error) {
	bgi := e.BackgroundImage.Value
	switch bgi {
	case "", "none":
		return "", NoStyles
	case "inherit":
		return "", InheritValue
	}
	//fmt.Printf("Background Image Value: %s\n", bgi)
	if strings.Count(bgi, "\"") >= 2 {
		realURL := bgi[strings.IndexRune(bgi, '"')+1 : strings.LastIndex(bgi, "\"")]
		return realURL, nil
	}
	if strings.Count(bgi, "'") >= 2 {
		realURL := bgi[strings.IndexRune(bgi, '\'')+1 : strings.LastIndex(bgi, "'")]
		return realURL, nil
	}
	realURL := bgi[strings.IndexRune(bgi, '(')+1 : strings.LastIndex(bgi, ")")]
	return realURL, nil
}

func (e StyledElement) GetColor(defaultColour color.Color) (color.Color, error) {
	switch e.Color.Value {
	case "inherit":
		return defaultColour, InheritValue
	case "transparent":
		return defaultColour, nil
	case "":
		return defaultColour, NoStyles
	default:
		return ConvertColorToRGBA(e.Color.Value)
	}
}
