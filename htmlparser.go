package main

import (
	//"fmt"
	"golang.org/x/net/html"
	"io"
	"strings"
)

func convertNodeToHTMLElement(root *html.Node) (*HTMLElement, error) {
	switch root.Type {
	case html.ElementNode:
		var textContent string
		var children []RenderableElement
		var lastError error
		for c := root.FirstChild; c != nil; c = c.NextSibling {
			switch c.Type {
			case html.ElementNode:
				newChild, err := convertNodeToHTMLElement(c)
				if err != nil {
					lastError = err
					continue
				}
				children = append(children, newChild)
			case html.TextNode:
				if trimmed := strings.TrimSpace(c.Data); trimmed != "" {
					children = append(children, TextElement{StyledElement{make([]StyleRule, 0)}, trimmed})
					textContent += trimmed
				}
			}
		}

		rules := make([]StyleRule, 0)
		return &HTMLElement{root, StyledElement{rules}, textContent, children}, lastError
		//	case html.TextNode:
		//		return &HTMLElement{nil, nil, root.Data, nil}, NotAnElement
	default:
		return nil, NotAnElement
	}
	panic("This should not happen")
	return nil, NotAnElement
}

func parseHTML(r io.Reader) (*Page, Stylesheet) {
	parsedhtml, _ := html.Parse(r)
	styles := extractStyles(parsedhtml)

	var body *HTMLElement
	var root *html.Node
	for c := parsedhtml.FirstChild; c != nil; c = c.NextSibling {
		if c.Data == "html" && c.Type == html.ElementNode {
			root = c
			break
		}
	}
	if root == nil {
		panic("Couldn't find HTML element")
	}
	//fmt.Printf("root: %s\n", root)
	for c := root.FirstChild; c != nil; c = c.NextSibling {
		//fmt.Printf("Investigating %s\n", c)
		if c.Type == html.ElementNode && c.Data == "body" {
			body, _ = convertNodeToHTMLElement(c)
			break
		}
	}
	return &Page{body},
		ParseStylesheet(styles)

}
