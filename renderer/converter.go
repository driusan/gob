package renderer

import (
	"strings"

	"github.com/driusan/gob/css"
	"github.com/driusan/gob/dom"
	"github.com/driusan/gob/net"
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
	if root.Type == html.ElementNode && strings.ToLower(root.Data) == "a" {
		if href := element.GetAttribute("href"); href != "" {
			if u, err := net.ParseURL(href); err == nil && loader.HasVisited(u) {
				element.State.Visited = true
			} else {
				element.State.Link = true
			}
		}
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
