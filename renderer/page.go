package renderer

import (
	"github.com/driusan/gob/css"

	"image/color"
	"net/url"
)

// Represents a page to be rendered
type Page struct {
	Content    *RenderableDomElement
	Background color.Color
	URL        *url.URL

	userAgentStyles, authorStyles css.Stylesheet
}

func (p *Page) getBody() *RenderableDomElement {
	if p == nil || p.Content == nil {
		return nil
	}
	if p.Content.Data != "html" {
		return nil
	}
	for c := p.Content.FirstChild; c != nil; c = c.NextSibling {
		if c.Data == "body" {
			return c
		}
	}
	return nil
}
