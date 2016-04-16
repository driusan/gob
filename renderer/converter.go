package renderer

import (
	"Gob/css"
	"Gob/dom"
	"golang.org/x/net/html"
)

func ConvertNodeToRenderableElement(root *html.Node) (*RenderableDomElement, error) {
	if root == nil {
		return nil, nil
	}

	element := &RenderableDomElement{
		(*dom.Element)(root),
		new(css.StyledElement),
		nil,
		nil,
		nil,
		nil,
		0,
		0,
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
