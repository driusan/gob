package parser

import (
	//	"fmt"
	"github.com/driusan/Gob/css"
	"github.com/driusan/Gob/net"
	"github.com/driusan/Gob/renderer"
	"golang.org/x/net/html"
	"image/color"
	"io"
	"io/ioutil"
	"net/url"
	"strconv"
	"strings"
)

func LoadPage(r io.Reader, loader net.URLReader, urlContext *url.URL) Page {
	parsedhtml, _ := html.Parse(r)
	styles, cssOrder := css.ExtractStyles(parsedhtml, loader, urlContext, 0)

	var body *html.Node
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
			break
		}
	}
	if body == nil {
		panic("Couldn't find body HTML element")
	}

	renderable, _ := renderer.ConvertNodeToRenderableElement(body, loader)

	sheet, _ := ioutil.ReadFile("useragent.css")
	userAgentStyles, cssOrder := css.ParseStylesheet(string(sheet), css.UserAgentSrc, loader, urlContext, cssOrder)

	p := Page{
		Content: renderable,
		URL:     nil,
	}

	renderable.Walk(func(el *renderer.RenderableDomElement) {
		el.PageLocation = urlContext
		for _, rule := range userAgentStyles {
			if rule.Matches((*html.Node)(el.Element)) {
				el.Styles.AddStyle(rule)
			}
		}

		for _, rule := range styles {
			if rule.Matches((*html.Node)(el.Element)) {
				el.Styles.AddStyle(rule)
			}
		}

		for _, attr := range el.Element.Attr {
			if attr.Key == "style" {
				vals := css.ParseBlock(attr.Val)
				for name, val := range vals {
					el.Styles.AddStyle(
						css.StyleRule{
							Selector: css.CSSSelector{"", cssOrder},
							Name:     name,
							Value:    val,
							Src:      css.InlineStyleSrc,
						})
					cssOrder++
				}
			}
		}

		// TODO(driusan): User styles too

		el.Styles.SortStyles()

		// Set the font size for this element, because em and ex
		// units depend on it.
		switch strVal := el.Styles.FontSize.GetValue(); strVal {
		case "":
			// nothing specified, so inherit from parent, or
			// fall back on default if there is no parent.
			if el.Parent == nil {
				el.Styles.SetFontSize(renderer.DefaultFontSize)
			} else {
				size, _ := el.Parent.Styles.GetFontSize()
				el.Styles.SetFontSize(size)
			}
		default:
			el.Styles.SetFontSize(fontSizeToPx(strVal, el.Parent))
		}

		if el.Type == html.ElementNode && strings.ToLower(el.Data) == "body" {
			background := el.GetBackgroundColor()
			if background == color.Transparent {
				background = color.RGBA{0xE0, 0xE0, 0xE0, 0xFF}
			}
			p.Background = background
		}
	})
	return p
}

func fontSizeToPx(val string, parent *renderer.RenderableDomElement) int {
	DefaultFontSize := renderer.DefaultFontSize
	if len(val) == 0 {
		psize, err := parent.Styles.GetFontSize()
		if err == nil {
			return psize
		}
		return DefaultFontSize
	}

	// "medium" is the default, and the CSS spec suggests a scaling factor
	// of 1.2.  1 / 1.2 = 0.833 for scaling downwards.
	// Use floating point arithmetric, then round
	switch val {
	// absolute size keywords
	case "xx-small":
		return int(float64(0.833*0.833*0.833) * float64(DefaultFontSize))
	case "x-small":
		return int(float64(0.833*0.833) * float64(DefaultFontSize))
	case "small":
		return int(float64(0.833) * float64(DefaultFontSize))
	case "medium":
		return DefaultFontSize
	case "large":
		return int(float64(1.2) * float64(DefaultFontSize))
	case "x-large":
		return int(1.2 * 1.2 * float64(DefaultFontSize))
	case "xx-large":
		return int(1.2 * 1.2 * 1.2 * float64(DefaultFontSize))
	// relative size keywords
	case "smaller":
		psize, _ := parent.Styles.GetFontSize()
		return int(0.833 * float64(psize))
	case "larger":
		psize, _ := parent.Styles.GetFontSize()
		return int(1.2 * float64(psize))
	// inherit
	case "inherit":
		psize, _ := parent.Styles.GetFontSize()
		return psize
	// 0 doesn't need a unit
	case "0":
		return 0
	}
	// handle percentages,
	if val[len(val)-1] == '%' {
		f, err := strconv.ParseFloat(string(val[0:len(val)-1]), 64)
		if err == nil {
			var psize int = DefaultFontSize
			if parent != nil {
				if psize, err = parent.Styles.GetFontSize(); err != nil {
					psize = DefaultFontSize
				}
			}

			size := int(f * float64(psize) / 100.0)
			return size
		}
		return DefaultFontSize
	}

	var psize int = DefaultFontSize
	if parent != nil {
		if ps, err := parent.Styles.GetFontSize(); err == nil {
			psize = ps
		}
	}
	size, err := css.ConvertUnitToPx(DefaultFontSize, psize, val)
	if err != nil {
		return DefaultFontSize
	}
	return size
}
