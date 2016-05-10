package renderer

import (
	"github.com/driusan/Gob/css"
	"github.com/driusan/Gob/dom"
	"golang.org/x/net/html"
)

func ConvertNodeToRenderableElement(root *html.Node) (*RenderableDomElement, error) {
	if root == nil {
		return nil, nil
	}

	element := &RenderableDomElement{
		Element: (*dom.Element)(root),
		Styles:  new(css.StyledElement),
	}

	element.FirstChild, _ = ConvertNodeToRenderableElement(root.FirstChild)
	element.NextSibling, _ = ConvertNodeToRenderableElement(root.NextSibling)

	var prev *RenderableDomElement = nil
	for c := element.FirstChild; c != nil; c = c.NextSibling {
		c.PrevSibling = prev
		c.Parent = element
		prev = c
	}
	return element, nil
}
