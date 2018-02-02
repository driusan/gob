package css

import (
	"fmt"
	//"Gob/dom"
	"github.com/gorilla/css/scanner"
	"golang.org/x/net/html"
	"strings"
)

type StyleRule struct {
	Selector CSSSelector
	Name     StyleAttribute
	Value    StyleValue
	Src      StyleSource
}

func (sr StyleRule) String() string {
	return fmt.Sprintf("(%s { %s: %s (%s)})", sr.Selector, sr.Name, sr.Value, sr.Src)
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
func (sv StyleValue) GetValue() string {
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
			continue //panic("Got a bad selector" + attrib)
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

type parsingContext uint8

const (
	matchingUnknown = parsingContext(iota)
	matchingSelector
	matchingAttribute
	matchingValue
	appendingSelector // The next TokenIdent should be appended to the current selector
)

func appendStyles(s []StyleRule, selectors []CSSSelector, attr StyleAttribute, val StyleValue, src StyleSource) ([]StyleRule, error) {
	if attr == "" {
		return s, nil
	}
	for _, sel := range selectors {
		s = append(s, StyleRule{
			Selector: CSSSelector(strings.TrimSpace(string(sel))),
			Name:     attr,
			Value:    val,
			Src:      src,
		})
	}
	return s, nil
}

func ParseStylesheet(val string, src StyleSource) Stylesheet {
	s := make([]StyleRule, 0)

	var blockSelectors []CSSSelector
	var curSelector CSSSelector
	var curAttribute StyleAttribute
	var curValue StyleValue
	var context parsingContext = matchingSelector
	spaceIfMatch := false
	scn := scanner.New(val)
	for {
		token := scn.Next()
		if token.Type == scanner.TokenEOF {
			break
		}
		//fmt.Printf("Token: %v\n", token)
		switch token.Type {
		// Different kinds of comments
		case scanner.TokenCDO, scanner.TokenCDC:
			continue
		case scanner.TokenComment:
			//fmt.Printf("Comment: %v\n")
			continue
		case scanner.TokenS:
			switch context {
			case matchingSelector:
				// if we get another Ident, it's an E F type selector.
				// if not, the , will set it back to matching
				context = appendingSelector
				spaceIfMatch = true
			case matchingValue:
				spaceIfMatch = true
			}
			// whitespace tokens. Ignore them.
			continue
		case scanner.TokenIdent:
			switch context {
			case matchingSelector:
				curSelector = CSSSelector(token.Value)
				spaceIfMatch = false
			case appendingSelector:
				if spaceIfMatch {
					curSelector += " "
					spaceIfMatch = false
				}
				curSelector += CSSSelector(token.Value)
				context = matchingSelector
			case matchingAttribute:
				curAttribute = StyleAttribute(token.Value)
				context = matchingValue
			case matchingValue:
				if token.Value == "important" {
					curValue.Important = true
				} else {
					if curValue.string == "" {
						curValue.string = token.Value
					} else {
						curValue.string += " " + token.Value
					}
				}
			}
			//fmt.Printf("TokenIdent: %s\n", token.Value)
			//curSelector = CSSSelector(token.Value)
			//matchingSelectors = true
		case scanner.TokenChar:
			switch context {
			case matchingSelector, appendingSelector:
				switch token.Value {
				case ",":
					if curSelector != "" {
						blockSelectors = append(blockSelectors, curSelector)
					}
					curSelector = ""
					context = matchingSelector
					spaceIfMatch = false
				case "*":
					curSelector += CSSSelector(token.Value)
					context = matchingSelector
					spaceIfMatch = false

				case "=":
					curSelector += CSSSelector(token.Value)
					// the actual value is a TokenString, not a TokenIdent,
					// so the context doesn't need to change to appending.
					context = matchingSelector
					spaceIfMatch = false

				case ">", "[", "+", ":", ".", "~":
					curSelector += CSSSelector(token.Value)
					context = appendingSelector
					spaceIfMatch = false

				case "]", ")":
					curSelector += CSSSelector(token.Value)
					context = matchingSelector
					spaceIfMatch = false

				case "{":
					if curSelector != "" {
						blockSelectors = append(blockSelectors, curSelector)
					}
					context = matchingAttribute
					curValue = StyleValue{}
					curAttribute = ""
					spaceIfMatch = false

					//fmt.Printf("Selectors for block: %s\n", blockSelectors)
				}
			case matchingAttribute:
				switch token.Value {
				case ":":
					curValue = StyleValue{}
					context = matchingValue
					spaceIfMatch = false
				case "}":
					spaceIfMatch = false

					if curAttribute != "" {
						s, _ = appendStyles(s, blockSelectors, curAttribute, curValue, src)
					}
					// End of a block, so reset everything to the 0 value
					// after appending the rules.
					blockSelectors = nil
					curSelector = ""
					curAttribute = ""
					curValue = StyleValue{}
					context = matchingSelector
				}
			case matchingValue:
				switch token.Value {
				case ",":
					curValue.string += token.Value
					spaceIfMatch = true
				case ")":
					curValue.string += token.Value
				case ";":
					if curAttribute != "" {
						s, _ = appendStyles(s, blockSelectors, curAttribute, curValue, src)
					}
					curAttribute = ""
					curValue = StyleValue{}
					context = matchingAttribute

				case "}":

					if curAttribute != "" {
						s, _ = appendStyles(s, blockSelectors, curAttribute, curValue, src)
					}
					// End of a block, so reset everything to the 0 value
					// after appending the rules.
					blockSelectors = nil
					curSelector = ""
					curAttribute = ""
					curValue = StyleValue{}
					context = matchingSelector
				}
			}
		case scanner.TokenString, scanner.TokenHash, scanner.TokenNumber:
			switch context {
			case matchingSelector, appendingSelector:
				curSelector += CSSSelector(token.Value)
				spaceIfMatch = false
			case matchingValue:
				if curValue.string == "" {
					curValue.string += token.Value
				} else {
					if spaceIfMatch {
						curValue.string += " "
						spaceIfMatch = false
					}
					curValue.string += token.Value

				}

			}
		case scanner.TokenIncludes, scanner.TokenPrefixMatch,
			scanner.TokenSuffixMatch, scanner.TokenSubstringMatch,
			scanner.TokenDashMatch, scanner.TokenFunction:
			switch context {
			case matchingSelector, appendingSelector:
				curSelector += CSSSelector(token.Value)
				spaceIfMatch = false
			case matchingValue:
				spaceIfMatch = false
				if curValue.string == "" {
					curValue.string = token.Value
				} else {
					curValue.string += " " + token.Value
				}

			}
		case scanner.TokenDimension, scanner.TokenPercentage, scanner.TokenURI:
			switch context {
			case matchingValue:
				if curValue.string == "" {
					curValue.string = token.Value
				} else {
					if spaceIfMatch {
						curValue.string += " "
						spaceIfMatch = false
					}
					curValue.string += token.Value
				}
			}
		default:
			fmt.Printf("%s = %s: %v\n", token.Type, token.Value, token)
		}

	}
	//fmt.Printf("Selectors for block: %s\n", blockSelectors)
	return s
}

func (r StyleRule) Matches(el *html.Node) bool {
	return r.Selector.Matches(el)
}
