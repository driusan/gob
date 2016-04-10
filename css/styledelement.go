package css

import (
	"fmt"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"image/color"
	"io/ioutil"
	"sort"
)

const (
	DefaultFontSize = 16
)

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
	// The rules that match this element.
	rules    []StyleRule
	fontSize int
}

func (e *StyledElement) SetFontSize(size int) {
	e.fontSize = size
}
func (e StyledElement) GetFontFace(fsize int) font.Face {
	fontBytes, _ := ioutil.ReadFile("/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf")
	fnt, _ := truetype.Parse(fontBytes)
	return truetype.NewFace(fnt,
		&truetype.Options{
			Size:    float64(fsize),
			DPI:     72,
			Hinting: font.HintingFull})

}
func (e StyledElement) GetFontSize() int {
	if e.fontSize == 0 {
		return DefaultFontSize
	}
	return e.fontSize
}

func (e *StyledElement) AddStyle(s StyleRule) {
	e.rules = append(e.rules, s)
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
// BUG(driusan): Specificity is not implemented, nor is the final tie break
func (e *StyledElement) SortStyles() error {
	sort.Sort(byCSSPrecedence(e.rules))
	fmt.Printf("%s\n", e.rules)

	return nil
}

func (e *StyledElement) FollowCascadeToPx(attr string, val int) int {
	// sort according to CSS cascading rules
	e.SortStyles()

	// apply each rule
	for _, rule := range e.rules {
		if string(rule.Name) == attr {
			val, _ = ConvertUnitToPx(val, rule.Value.string)
			return val
		}
		// the rule has this attribute, so convert it and apply
		// it to the value calculated so far
		//if cssval, ok := rule.Values[StyleAttribute(attr)]; ok {
		//val, _ = ConvertUnitToPx(val, cssval.string)
		//}
	}
	return val
}

// Follows the cascade to get the colour for the attribute named attr.
// deflt is the default to return if there is nothing found for the attribute in
// the cascade. It should be the parent's colour if it's an inherited property,
// and nil otherwise.
// error will be NoStyles if deflt is returned
func (e StyledElement) FollowCascadeToColor(attr string, deflt *color.RGBA) (*color.RGBA, error) {
	// sort according to CSS cascading rules
	e.SortStyles()

	// apply each rule
	for _, rule := range e.rules {
		if string(rule.Name) == attr {
			if rule.Value.string == "inherit" {
				return nil, InheritValue
			}
			val, _ := ConvertColorToRGBA(rule.Value.string)
			return val, nil
		}
	}
	return deflt, NoStyles
}

func (e StyledElement) GetBackgroundColor(defaultColour *color.RGBA) (*color.RGBA, error) {
	val, err := e.FollowCascadeToColor("background", defaultColour)
	switch err {
	case NoStyles:
		return defaultColour, InheritValue
	case InheritValue:
		return defaultColour, InheritValue
	case nil:
		return val, nil
	default:
		return defaultColour, InheritValue
	}
}
func (e StyledElement) GetColor(defaultColour *color.RGBA) *color.RGBA {
	val, err := e.FollowCascadeToColor("color", defaultColour)
	switch err {
	case NoStyles:
		return defaultColour
	case nil:
		return val
	default:
		panic("Could not get colour and got an error")
	}
}
