package parser

import (
	"github.com/driusan/Gob/renderer"
	"net/url"
)

type Page struct {
	Content *renderer.RenderableDomElement
	URL     *url.URL
}
