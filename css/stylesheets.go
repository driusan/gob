package css

import (
	"Gob/dom"
	"strings"
)

type StyleRule struct {
	Selector CSSSelector
	Name     StyleAttribute
	Value    StyleValue
	Src      StyleSource
}
type Stylesheet []StyleRule

type StyleAttribute string

type StyleValue struct {
	string
	Important bool
}

func (sv StyleValue) String() string {
	return sv.string
}
func ParseBlock(val string) map[StyleAttribute]StyleValue {
	m := make(map[StyleAttribute]StyleValue)
	pieces := strings.Split(val, ";")
	for _, attrib := range pieces {
		if strings.TrimSpace(attrib) == "" {
			continue
		}
		idx := strings.Index(attrib, ":")
		if idx < 0 {
			panic("Got a bad selector" + attrib)
		}
		selector := strings.TrimSpace(attrib[0:idx])
		value := strings.TrimSpace(attrib[idx+1:])

		var important bool
		if strings.HasSuffix(value, "important") {
			important = true
			value = value[0 : len(value)-len("important")]
		}
		m[StyleAttribute(selector)] = StyleValue{value, important}

	}
	return m
}

func ParseStylesheet(val string, src StyleSource) Stylesheet {
	s := make([]StyleRule, 0)
	selectorStart := 0
	blockStart := -1
	var selectors []CSSSelector
	for idx, chr := range val {
		switch chr {
		case '}':
			selectorStart = idx + 1
			blockVals := ParseBlock(val[blockStart:idx])
			for _, sel := range selectors {
				for name, val := range blockVals {
					s = append(s, StyleRule{
						Selector: sel,
						Name:     name,
						Value:    val,
						Src:      src,
					})

				}
			}
		case '{':
			blockStart = idx + 1
			selectors = parseSelectors(val[selectorStart:idx])

		}
	}
	return s
}

func (r StyleRule) Matches(el *dom.Element) bool {
	return r.Selector.Matches(el)
}
