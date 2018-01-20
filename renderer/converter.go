package renderer

import (
	"github.com/driusan/Gob/css"
	"github.com/driusan/Gob/dom"
	"github.com/driusan/Gob/net"
	"golang.org/x/net/html"
)

func ConvertNodeToRenderableElement(root *html.Node, loader net.URLReader) (*RenderableDomElement, error) {
	if root == nil {
		return nil, nil
	}

	element := &RenderableDomElement{
		Element:  (*dom.Element)(root),
		Styles:   new(css.StyledElement),
		resolver: loader,
	}

	element.FirstChild, _ = ConvertNodeToRenderableElement(root.FirstChild, loader)
	element.NextSibling, _ = ConvertNodeToRenderableElement(root.NextSibling, loader)

	var prev *RenderableDomElement = nil
	for c := element.FirstChild; c != nil; c = c.NextSibling {
		c.PrevSibling = prev
		c.Parent = element
		prev = c
	}
	return element, nil
}
