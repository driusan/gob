package renderer

import (
	"strings"

	"net/url"

	"github.com/driusan/gob/css"
	"github.com/driusan/gob/dom"
	"github.com/driusan/gob/net"

	"golang.org/x/net/html"

	"golang.org/x/exp/shiny/screen"
)

func ConvertNodeToRenderableElement(root *html.Node, location *url.URL, window screen.Window, parent *RenderableDomElement, loader net.URLReader) (*RenderableDomElement, error) {
	if root == nil {
		return nil, nil
	}

	element := &RenderableDomElement{
		Element:  &dom.Element{Node: root, Window: window, Location: location},
		Styles:   new(css.StyledElement),
		resolver: loader,
	}
	if parent != nil {
		element.setParent(parent)
	}
	if root.Type == html.ElementNode {
		switch strings.ToLower(root.Data) {
		case "a":
			if href := element.GetAttribute("href"); href != "" {
				if u, err := net.ParseURL(href); err == nil && loader.HasVisited(u) {
					element.State.Visited = true
				} else {
					element.State.Link = true
				}
			}
		case "input":
			// Set the default value, it will change with user input
			element.Value = element.GetAttribute("value")
		}
	}

	fc, _ := ConvertNodeToRenderableElement(root.FirstChild, location, window, element, loader)
	ns, _ := ConvertNodeToRenderableElement(root.NextSibling, location, window, parent, loader)

	element.setFirstChild(fc)
	element.setNextSibling(ns)

	//	var prev *RenderableDomElement = nil
	for c := element.FirstChild; c != nil; c = c.NextSibling {
		//c.PrevSibling.setNextSibling(prev)
		c.setParent(element)
		//	prev = c
	}
	return element, nil
}

func (c *RenderableDomElement) setParent(parent *RenderableDomElement) {
	c.Parent = parent
	c.Element.Parent = parent.Element
}

func (e *RenderableDomElement) setFirstChild(child *RenderableDomElement) {
	if child == nil {
		return
	}
	e.FirstChild = child
	e.Element.FirstChild = child.Element
}

func (e *RenderableDomElement) setNextSibling(sibl *RenderableDomElement) {
	if sibl == nil {
		return
	}
	e.NextSibling = sibl
	e.Element.NextSibling = sibl.Element
}
