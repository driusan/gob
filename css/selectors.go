package css

import (
	"Gob/dom"
	"strings"
)

type CSSSelector string

func parseSelectors(val string) []CSSSelector {
	vals := strings.Split(val, ",")
	var ret []CSSSelector
	for _, selector := range vals {
		ret = append(ret, CSSSelector(strings.TrimSpace(selector)))
	}
	return ret
}

func (s CSSSelector) Matches(el *dom.Element) bool {
	pieces := strings.Fields(string(s))

	if len(pieces) != 1 {
		return false
		panic("I am neither a well coded error handler nor CSS parser. Can't handle complex Stylesheets.")
	}
	if s[0] == '.' {
		for _, attr := range el.Attr {
			if attr.Key == "class" && attr.Val == string(s[1:]) {
				return true
			}
		}
	}
	if s[0] == '#' {
		for _, attr := range el.Attr {
			if attr.Key == "id" && attr.Val == string(s[1:]) {
				return true
			}
		}
	}

	if el.Data == pieces[0] {
		return true
	}
	return false
}

func (s CSSSelector) NumberIDs() int {
	return strings.Count(string(s), "#")
}
func (s CSSSelector) NumberAttrs() int {
	return strings.Count(string(s), "[")
}
func (s CSSSelector) NumberClasses() int {
	return strings.Count(string(s), ".")
}
func (s CSSSelector) NumberElements() int {
	pieces := strings.Fields(string(s))
	elems := len(pieces)
	for _, piece := range pieces {
		elems += strings.Count(piece, "+")
	}
	return elems
}
func (s CSSSelector) NumberPseudo() int {
	return strings.Count(string(s), ":")
}
