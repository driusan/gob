package parser

import (
	"github.com/driusan/gob/css"
	"github.com/driusan/gob/renderer"

	"image/color"
	"net/url"
)

type Page struct {
	Content    *renderer.RenderableDomElement
	Background color.Color
	URL        *url.URL

	userAgentStyles, authorStyles css.Stylesheet
}
