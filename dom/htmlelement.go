package dom

import "golang.org/x/net/html"

type Element html.Node

func (e Element) GetTextContent() string {
	if e.Type == html.TextNode {
		return e.Data
	}
	var ret string
	for c := e.FirstChild; c != nil; c = c.NextSibling {
		switch c.Type {
		case html.TextNode:
			ret += c.Data
		case html.ElementNode:
			ret += (*Element)(c).GetTextContent()
		}
	}
	return ret
}

func (e Element) GetAttribute(attr string) string {
	for _, attrField := range e.Attr {
		if attrField.Key == attr {
			return attrField.Val
		}
	}
	return ""

}

func (e Element) OnClick() {
}
