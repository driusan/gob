package css

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"strings"

	"github.com/driusan/Gob/net"
	"github.com/gorilla/css/scanner"
	"golang.org/x/net/html"
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
	startContext
	matchingSelector
	matchingAttribute
	matchingValue
	appendingSelector // The next TokenIdent should be appended to the current selector
	atImport
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

func ParseStylesheet(val string, src StyleSource, importLoader net.URLReader, urlContext *url.URL) Stylesheet {
	s := make([]StyleRule, 0)

	var blockSelectors []CSSSelector
	var curSelector CSSSelector
	var curAttribute StyleAttribute
	var curValue StyleValue
	var context parsingContext = startContext
	spaceIfMatch := false
	var importURL *url.URL
	scn := scanner.New(val)
	for {
		token := scn.Next()
		if token.Type == scanner.TokenEOF {
			break
		}
		switch token.Type {
		// Different kinds of comments
		case scanner.TokenCDO, scanner.TokenCDC:
			continue
		case scanner.TokenComment:
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
			case startContext, matchingSelector:
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
		case scanner.TokenChar:
			switch context {
			case startContext, matchingSelector, appendingSelector:
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
			case atImport:
				switch token.Value {
				case ";":
					// Pretend we're at the start so future
					// import lines succeed
					context = startContext
					r, resp, err := importLoader.GetURL(importURL)
					if err != nil {
						continue
					}
					defer r.Close()
					if resp < 200 || resp >= 300 {
						continue
					}
					styles, err := ioutil.ReadAll(r)
					if err != nil {
						continue
					}
					news := ParseStylesheet(string(styles), src, importLoader, importURL)
					s = append(s, news...)
				}
			}
		case scanner.TokenString:
			if context == atImport {
				undecorated := token.Value[1 : len(token.Value)-1]
				iu, err := url.Parse(undecorated)
				if err != nil {
					continue
				}
				importURL = urlContext.ResolveReference(iu)
			}
			fallthrough
		case scanner.TokenHash, scanner.TokenNumber:
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
			case startContext, matchingSelector, appendingSelector:
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
		case scanner.TokenURI:
			if context == atImport {
				undecorated := strings.TrimSuffix(strings.TrimPrefix(token.Value, "url("), ")")
				iu, err := url.Parse(undecorated)
				if err != nil {
					continue
				}
				importURL = urlContext.ResolveReference(iu)
				continue
			}
			fallthrough
		case scanner.TokenDimension, scanner.TokenPercentage:
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
		case scanner.TokenAtKeyword:
			if token.Value == "@import" && context == startContext {
				context = atImport
			}
		case scanner.TokenError:
			fallthrough
		default:
			fmt.Printf("%s = %s: %v\n", token.Type, token.Value, token)
		}
	}
	return s
}

func (r StyleRule) Matches(el *html.Node) bool {
	return r.Selector.Matches(el)
}
