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
	Value     string
	Important bool
}

func (sv StyleValue) String() string {
	return sv.Value
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

func (p parsingContext) String() string {
	switch p {
	case startContext:
		return "start"
	case matchingSelector:
		return "selector"
	case matchingAttribute:
		return "attribute"
	case matchingValue:
		return "value"
	case appendingSelector:
		return "appendingselector"
	case atImport:
		return "@import"
	default:
		panic("Unhandled parsingContext")
	}
}

func appendStyles(s []StyleRule, selectors []CSSSelector, attr StyleAttribute, val StyleValue, src StyleSource, order uint) ([]StyleRule, error) {
	if attr == "" {
		return s, nil
	}
	for i, sel := range selectors {
		s = append(s, StyleRule{
			Selector: CSSSelector{strings.TrimSpace(string(sel.Selector)), order + uint(i)},
			Name:     attr,
			Value:    val,
			Src:      src,
		})
	}
	return s, nil
}

func ParseStylesheet(val string, src StyleSource, importLoader net.URLReader, urlContext *url.URL, orderNo uint) (styles Stylesheet, nextOrderNoStart uint) {
	s := make([]StyleRule, 0)

	var blockSelectors []CSSSelector
	var curSelector CSSSelector
	var curAttribute StyleAttribute
	var curValue StyleValue
	var context parsingContext = startContext
	spaceIfMatch := false
	var importURL *url.URL
	scn := scanner.New(val)
	invalidSelector := false
	skipblocks := 0
	for {
		token := scn.Next()
		if token.Type == scanner.TokenEOF {
			break
		}
		if invalidSelector {
			if token.Value == "}" {
				skipblocks--
			} else if token.Value == "{" {
				skipblocks++
			}
			if skipblocks == 0 && (token.Value == "}" || token.Value == ";") {
				invalidSelector = false
			}
			curSelector = CSSSelector{}
			continue
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
				// if not, the "," char will set it back to matching
				context = appendingSelector
				spaceIfMatch = true
			case matchingValue:
				spaceIfMatch = true
			}
			continue
		case scanner.TokenIdent:
			switch context {
			case startContext, matchingSelector:
				curSelector = CSSSelector{token.Value, orderNo}
				spaceIfMatch = false
			case appendingSelector:
				if spaceIfMatch {
					curSelector.Selector += " "
					spaceIfMatch = false
				}
				curSelector.Selector += token.Value
				context = matchingSelector
			case matchingAttribute:
				curAttribute = StyleAttribute(token.Value)
				context = matchingValue
			case matchingValue:
				if token.Value == "important" {
					curValue.Important = true
				} else {
					if curValue.Value == "" {
						curValue.Value = token.Value
					} else {
						curValue.Value += " " + token.Value
					}
				}
			}
		case scanner.TokenChar:
			switch context {
			case startContext, matchingSelector, appendingSelector:
				switch token.Value {
				case ",":
					if curSelector.Selector != "" {
						blockSelectors = append(blockSelectors, curSelector)
					}
					curSelector.Selector = ""
					context = matchingSelector
					spaceIfMatch = false
				case "*":
					curSelector.Selector += token.Value
					context = matchingSelector
					spaceIfMatch = false
				case "=":
					curSelector.Selector += token.Value
					// the actual value is a TokenString, not a TokenIdent,
					// so the context doesn't need to change to appending.
					context = matchingSelector
					spaceIfMatch = false

				case ">", "[", "+", ":", ".", "~":
					curSelector.Selector += token.Value
					context = appendingSelector
					spaceIfMatch = false

				case "]", ")":
					curSelector.Selector += token.Value
					context = matchingSelector
					spaceIfMatch = false

				case "{":
					if curSelector.Selector != "" {
						blockSelectors = append(blockSelectors, curSelector)
					}
					context = matchingAttribute
					curValue = StyleValue{}
					curAttribute = ""
					spaceIfMatch = false
				case ";":
					continue
				default:
					//panic(fmt.Sprintf("Unhandled character %v in context %v", token.Value, context))
					invalidSelector = true
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
						s, _ = appendStyles(s, blockSelectors, curAttribute, curValue, src, orderNo)
						orderNo += uint(len(s))
					}
					// End of a block, so reset everything to the 0 value
					// after appending the rules.
					curSelector.Selector = ""
					blockSelectors = nil
					curSelector.OrderNumber = orderNo
					curAttribute = ""
					curValue = StyleValue{}

					context = matchingSelector
				}
			case matchingValue:
				switch token.Value {
				case "-":
					curValue.Value += token.Value
					spaceIfMatch = false
				case ",":
					curValue.Value += token.Value
					spaceIfMatch = true
				case ")":
					curValue.Value += token.Value
				case ";":
					if curAttribute != "" {
						s, _ = appendStyles(s, blockSelectors, curAttribute, curValue, src, orderNo)
						orderNo += uint(len(s))
					}
					curAttribute = ""
					curValue = StyleValue{}
					context = matchingAttribute

				case "}":

					if curAttribute != "" {
						s, _ = appendStyles(s, blockSelectors, curAttribute, curValue, src, orderNo)
						orderNo += uint(len(s))
					}
					// End of a block, so reset everything to the 0 value
					// after appending the rules.
					blockSelectors = nil
					curSelector.Selector = ""
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
					news, nextOrderNo := ParseStylesheet(string(styles), src, importLoader, importURL, orderNo)
					orderNo = nextOrderNo
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
		case scanner.TokenNumber:
			switch context {
			case startContext, matchingSelector, appendingSelector:
				invalidSelector = true
				continue
			}
			fallthrough
		case scanner.TokenHash:
			switch context {
			case startContext, matchingSelector, appendingSelector:
				curSelector.Selector += token.Value
				spaceIfMatch = false
			case matchingValue:
				if curValue.Value == "" {
					curValue.Value += token.Value
				} else {
					if spaceIfMatch {
						curValue.Value += " "
						spaceIfMatch = false
					}
					curValue.Value += token.Value
				}

			}
		case scanner.TokenIncludes, scanner.TokenPrefixMatch,
			scanner.TokenSuffixMatch, scanner.TokenSubstringMatch,
			scanner.TokenDashMatch, scanner.TokenFunction:
			switch context {
			case startContext, matchingSelector, appendingSelector:
				curSelector.Selector += token.Value
				spaceIfMatch = false
			case matchingValue:
				spaceIfMatch = false
				if curValue.Value == "" {
					curValue.Value = token.Value
				} else {
					curValue.Value += " " + token.Value
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
				if curValue.Value == "" {
					curValue.Value = token.Value
				} else {
					if spaceIfMatch {
						curValue.Value += " "
						spaceIfMatch = false
					}
					curValue.Value += token.Value
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
	return s, orderNo
}

func (r StyleRule) Matches(el *html.Node, st State) bool {
	return r.Selector.Matches(el, st)
}
