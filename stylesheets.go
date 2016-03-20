package main

//import "fmt"
import "strings"

type Color string
type Size string
type Stylesheet []StyleRule
type CSSSelector string
type StyleRule struct {
	Selector CSSSelector
	Values   map[StyleAttribute]string
}

type StyleAttribute string

func parseSelectors(val string) []CSSSelector {
	vals := strings.Split(val, ",")
	var ret []CSSSelector
	for _, selector := range vals {
		ret = append(ret, CSSSelector(strings.TrimSpace(selector)))
	}
	return ret
}

func parseBlock(val string) map[StyleAttribute]string {
	m := make(map[StyleAttribute]string)
	pieces := strings.Split(val, ";")
	for _, attrib := range pieces {
		if strings.TrimSpace(attrib) == "" {
			continue
		}
		split := strings.Split(attrib, ":")
		if len(split) != 2 {
			panic("Got a bad selector" + attrib)
		}
		m[StyleAttribute(strings.TrimSpace(split[0]))] = strings.TrimSpace(split[1])

	}
	return m
}

func ParseStylesheet(val string) Stylesheet {
	s := make([]StyleRule, 0)
	selectorStart := 0
	blockStart := -1
	var selectors []CSSSelector
	for idx, chr := range val {
		switch chr {
		case '}':
			selectorStart = idx + 1
			blockVals := parseBlock(val[blockStart:idx])
			for _, sel := range selectors {
				s = append(s, StyleRule{sel, blockVals})
			}
		case '{':
			blockStart = idx + 1
			selectors = parseSelectors(val[selectorStart:idx])

		}
	}
	return s
}

func (s CSSSelector) Matches(el *HTMLElement) bool {
	pieces := strings.Split(string(s), " ")
	if len(pieces) != 1 {
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
				//fmt.Printf("Matched by id: %s", s)
				return true
			}
		}
	}

	if el.Data == pieces[0] {
		return true
	}
	return false
}
func (r StyleRule) Matches(el *HTMLElement) bool {
	return r.Selector.Matches(el)
}
