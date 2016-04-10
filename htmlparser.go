package main

import (
	//"fmt"
	"Gob/css"
	"Gob/dom"
	"Gob/renderer"
	"golang.org/x/net/html"
	//	"strings"
	"io"
	"io/ioutil"
)

/*
type RenderableDomElement struct {
	*dom.Element
	Styles *css.StyledElement

	FirstChild *RenderableDomElement
	NextSibling *RenderableDomElement
}
*/
func convertNodeToRenderableElement(root *html.Node) (*renderer.RenderableDomElement, error) {
	if root == nil {
		return nil, nil
	}

	element := &renderer.RenderableDomElement{
		(*dom.Element)(root),
		new(css.StyledElement),
		nil,
		nil,
		nil,
		nil,
	}
	element.FirstChild, _ = convertNodeToRenderableElement(root.FirstChild)
	element.NextSibling, _ = convertNodeToRenderableElement(root.NextSibling)

	var prev *renderer.RenderableDomElement = nil
	for c := element.FirstChild; c != nil; c = c.NextSibling {
		c.PrevSibling = prev
		c.Parent = element
		prev = c
	}
	return element, nil

	/*
		switch root.Type {
		case html.ElementNode:
			var textContent string
			var lastError error
			for c := root.FirstChild; c != nil; c = c.NextSibling {
				switch c.Type {
				case html.ElementNode:
					//newChild, err := convertNodeToRenderableElement(c)
					if err != nil {
						lastError = err
						continue
					}
				case html.TextNode:
					if trimmed := strings.TrimSpace(c.Data); trimmed != "" {
						textContent += trimmed
					}
				}
			}

			return element, lastError
		}
		panic("This should not happen")
		//return nil, NotAnElement
	*/
}

func parseHTML(r io.Reader) *Page {
	parsedhtml, _ := html.Parse(r)
	styles := css.ExtractStyles(parsedhtml)

	var body *html.Node // renderer.RenderableDomElement
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
			body = c
			//body, _ = convertNodeToRenderableElement(c)
			//body = (*dom.Element)(c)
			break
		}
	}

	styles2 := css.ParseStylesheet(styles, css.AuthorSrc)

	renderable, _ := convertNodeToRenderableElement(body)

	sheet, _ := ioutil.ReadFile("useragent.css")
	userAgentStyles := css.ParseStylesheet(string(sheet), css.UserAgentSrc)

	renderable.Walk(func(el *renderer.RenderableDomElement) {
		for _, rule := range userAgentStyles {
			if rule.Matches(el.Element) {
				el.Styles.AddStyle(rule)
			}
		}

		for _, rule := range styles2 {
			if rule.Matches(el.Element) {
				el.Styles.AddStyle(rule)
			}
		}

		// TODO(driusan): Add inline and user styles too
		el.Styles.SortStyles()
	})

	return &Page{renderable}

}
