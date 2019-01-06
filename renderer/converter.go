package renderer

import (
	"strings"

	"github.com/driusan/gob/css"
	"github.com/driusan/gob/dom"
	"github.com/driusan/gob/net"
	"golang.org/x/net/html"
)

func convertNodeToRenderableElement(root *html.Node, loader net.URLReader) *RenderableDomElement {
	if root == nil {
		return nil
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

	element.FirstChild = convertNodeToRenderableElement(root.FirstChild, loader)
	element.NextSibling = convertNodeToRenderableElement(root.NextSibling, loader)

	var prev *RenderableDomElement = nil
	for c := element.FirstChild; c != nil; c = c.NextSibling {
		c.PrevSibling = prev
		c.Parent = element
		prev = c
	}
	return element
}
