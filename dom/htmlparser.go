package dom

import (
	"errors"
	//"golang.org/x/net/html"
	//"strings"
)

var (
	NotAnElement = errors.New("Not a valid element")
)

/*
func ConvertNodeToHTMLElement(root *html.Node) (*Element, error) {
	switch root.Type {
	case html.ElementNode:
		var textContent string
		var children []DomElement
		var lastError error
		for c := root.FirstChild; c != nil; c = c.NextSibling {
			switch c.Type {
			case html.ElementNode:
				newChild, err := ConvertNodeToHTMLElement(c)
				if err != nil {
					lastError = err
					continue
				}
				children = append(children, newChild)
			case html.TextNode:
				if trimmed := strings.TrimSpace(c.Data); trimmed != "" {
					children = append(children, TextElement{trimmed})
					textContent += trimmed
				}
			}
		}

		return &Element{root, children}, lastError
		//	case html.TextNode:
		//		return &HTMLElement{nil, nil, root.Data, nil}, NotAnElement
	default:
		return nil, NotAnElement
	}
	panic("This should not happen")
	return nil, NotAnElement
}
*/
