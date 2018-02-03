package css

import (
	//	"Gob/dom"
	"golang.org/x/net/html"
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

// match the ID and class part of the selector. This assumes that the string s
// starts with a . or #. That is, the element tag must have already been stripped
// off and separatedly matched. It will recursively call itself stripping off
// one id or class part each time, until there's either left or something
// didn't match
func matchIDAndClassSelector(el *html.Node, s string) bool {
	if el == nil || el.Type != html.ElementNode || len(s) < 1 {
		return false
	}
	remainingData := ""
	classSelector := ""
	idSelector := ""
	switch s[0] {
	case '.':
		chopped := s[1:]
		if idx := strings.IndexAny(chopped, "*.#"); idx != -1 {
			classSelector = chopped[0:idx]
			remainingData = chopped[idx:]
		} else {
			classSelector = chopped
			remainingData = ""
		}
	case '#':
		chopped := s[1:]
		if idx := strings.IndexAny(s[1:], "*.#"); idx != -1 {
			idSelector = chopped[0:idx]
			remainingData = chopped[idx:]
		} else {
			idSelector = chopped
			remainingData = ""
		}
	default:
		return false
	}

	if idSelector != "" {
		matchedId := false
		for _, attrib := range el.Attr {
			if attrib.Key == "id" {
				if strings.ToLower(attrib.Val) == strings.ToLower(idSelector) {
					matchedId = true
				} else {
					matchedId = false
				}
				break
			}
		}
		if matchedId == false {
			return false
		}
	}
	if classSelector != "" {
		matchedClass := false
		for _, attrib := range el.Attr {
			if attrib.Key == "class" {
				classes := strings.Fields(attrib.Val)
				for _, class := range classes {
					if strings.ToLower(classSelector) == strings.ToLower(class) {
						matchedClass = true
					}
				}
				break
			}
		}
		if matchedClass == false {
			return false
		}
	}

	// all the classes and ids have been matched.
	if remainingData == "" {
		return true
	}
	if remainingData == s {
		// this shouldn't happen, but if nothing was consumed assume it doesn't match and prevent
		// an infinite loop
		return false
	}
	// check the unconsumed pieces
	return matchIDAndClassSelector(el, remainingData)
}
func matchBasicSelector(el *html.Node, s string) bool {
	if el == nil || len(s) < 1 || el.Type != html.ElementNode {
		return false
	}
	elementMatchTag := ""
	remainingData := ""

	if idx := strings.IndexAny(s, "*.#"); idx != -1 {
		elementMatchTag = s[0:idx]
		remainingData = s[idx:]
		if remainingData[0] == '*' {
			remainingData = remainingData[1:]
		}
	} else {
		elementMatchTag = s
	}
	if elementMatchTag != "" && strings.ToLower(el.Data) != strings.ToLower(elementMatchTag) {
		return false
	}
	if remainingData == "" {
		// matched the tag and there's nothing left
		return true
	}

	return matchIDAndClassSelector(el, remainingData)
}
func recursiveParentMatches(el *html.Node, selectorPieces []string) bool {
	switch len(selectorPieces) {
	case 0:
		return false
	case 1:
		if matchBasicSelector(el, selectorPieces[0]) {
			return true
		}
		if el == nil {
			return false
		}
		return recursiveParentMatches(el.Parent, selectorPieces)
	default:
		lastSelector := selectorPieces[len(selectorPieces)-1]
		otherSelectors := selectorPieces[0 : len(selectorPieces)-1]
		if matchBasicSelector(el, lastSelector) == false {
			return false
		}
		return recursiveParentMatches(el.Parent, otherSelectors)
	}
}
func (s CSSSelector) Matches(el *html.Node) bool {
	pieces := strings.Fields(string(s))
	if len(pieces) <= 1 {
		return matchBasicSelector(el, pieces[0])
	}
	return recursiveParentMatches(el, pieces)
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
