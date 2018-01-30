package parser

import (
	"github.com/driusan/Gob/renderer"
	"image/color"
	"net/url"
)

type Page struct {
	Content    *renderer.RenderableDomElement
	Background color.Color
	URL        *url.URL
}
