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
