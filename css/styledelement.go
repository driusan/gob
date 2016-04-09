package css

import (
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"image/color"
	"io/ioutil"
)

const (
	DefaultFontSize = 16
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

// SortStyles will sort the rules on this element according to the CSS spec, which state:s

// 1. Find all declarations that apply too element/property (already done when this is called)
// 2. Sort according to importance (normal or important) and origin (author, user, or user agent). In ascending order of precedence:
//	1. user agent declarations (defaults)
//	2. user normal declrations (don't exist)
//	3. author normal declarations
//	4. author important declarations
//	5. user important declarations (don't exist)
// 3. Sort rules with the same importance and origin by specificity of selector: more specific selectors will override more general ones. Pseudo-elements and pseudo-classes are counted as normal elements and classes, respectively.
// 4. Finally, sort by order specified: if two declarations have the same weight, origin, and specificity, the latter specified wins. Declarations in imported stylesheets are considered to be before any declaration in the style sheet itself
// BUG(driusan): SortStyles is not implemented
func (e *StyledElement) SortStyles() error {
	return nil
}

func (e StyledElement) FollowCascadeToPx(attr string, val int) int {
	// sort according to CSS cascading rules
	e.SortStyles()

	// apply each rule
	for _, rule := range e.rules {
		// the rule has this attribute, so convert it and apply
		// it to the value calculated so far
		if cssval, ok := rule.Values[StyleAttribute(attr)]; ok {
			val, _ = ConvertUnitToPx(val, cssval)
		}
	}
	return val
}

// Follows the cascade to get the colour for the attribute named attr.
// deflt is the default to return if there is nothing found for the attribute in
// the cascade. It should be the parent's colour if it's an inherited property,
// and nil otherwise.
// error will be NoStyles if deflt is returned
func (e StyledElement) FollowCascadeToColor(attr string, deflt *color.RGBA) (*color.RGBA, error) {
	var ret *color.RGBA
	// sort according to CSS cascading rules
	e.SortStyles()

	// apply each rule
	for _, rule := range e.rules {
		// the rule has this attribute, so convert it and apply
		// it to the value calculated so far
		if cssval, ok := rule.Values[StyleAttribute(attr)]; ok {
			ret, _ = ConvertColorToRGBA(cssval)

		}
	}
	if ret == nil {
		return deflt, NoStyles
	}
	return ret, nil
}
func (e StyledElement) GetBackgroundColor(parentColour *color.RGBA) *color.RGBA {
	val, err := e.FollowCascadeToColor("background", parentColour)
	if err == NoStyles {
		return parentColour
	}

	return val
}
func (e StyledElement) GetColor(parentColour *color.RGBA) *color.RGBA {
	val, err := e.FollowCascadeToColor("color", parentColour)
	if err == NoStyles {
		return parentColour
	}
	return val
}
