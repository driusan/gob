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

	for c := root.FirstChild; c != nil; c = c.NextSibling {
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
		Content:         renderable,
		URL:             urlContext,
		userAgentStyles: userAgentStyles,
		authorStyles:    styles,
	}
	p.ReapplyStyles()
	return p
}

func (p *Page) ReapplyStyles() {
	cssOrder := uint(0)
	p.Content.Walk(func(el *renderer.RenderableDomElement) {
		el.Styles.ClearStyles()
		el.ConditionalStyles = struct {
			Unconditional *css.StyledElement
			FirstLine     *css.StyledElement
			FirstLetter   *css.StyledElement
		}{
			new(css.StyledElement),
			new(css.StyledElement),
			new(css.StyledElement),
		}
		el.PageLocation = p.URL
		for _, rule := range p.userAgentStyles {
			if rule.Matches((*html.Node)(el.Element), el.State) {
				if strings.Index(rule.Selector.Selector, "first-line") >= 0 {
					el.ConditionalStyles.FirstLine.AddStyle(rule)
				} else if strings.Index(rule.Selector.Selector, "first-letter") >= 0 {
					el.ConditionalStyles.FirstLetter.AddStyle(rule)
				} else {
					el.ConditionalStyles.Unconditional.AddStyle(rule)
				}
			}
		}

		for _, rule := range p.authorStyles {
			if rule.Matches((*html.Node)(el.Element), el.State) {
				if strings.Index(rule.Selector.Selector, "first-line") >= 0 {
					el.ConditionalStyles.FirstLine.AddStyle(rule)
				} else if strings.Index(rule.Selector.Selector, "first-letter") >= 0 {
					el.ConditionalStyles.FirstLetter.AddStyle(rule)
				} else {
					el.ConditionalStyles.Unconditional.AddStyle(rule)
				}
			}
		}

		for _, attr := range el.Element.Attr {
			if strings.ToLower(attr.Key) == "style" {
				vals := css.ParseBlock(attr.Val)
				for name, val := range vals {
					el.ConditionalStyles.Unconditional.AddStyle(
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

		fl := el.ConditionalStyles.FirstLine.MergeStyles(
			el.ConditionalStyles.Unconditional,
		)
		el.ConditionalStyles.FirstLine = &fl
		flet := el.ConditionalStyles.FirstLetter.MergeStyles(
			el.ConditionalStyles.FirstLine,
		)
		el.ConditionalStyles.FirstLetter = &flet

		el.ConditionalStyles.Unconditional.SortStyles()
		el.ConditionalStyles.FirstLine.SortStyles()
		el.ConditionalStyles.FirstLetter.SortStyles()

		// Set the font size for this element, because em and ex
		// units depend on it.
		var base int
		switch strVal := el.ConditionalStyles.Unconditional.FontSize.GetValue(); strVal {
		case "":
			// nothing specified, so inherit from parent, or
			// fall back on default if there is no parent.
			if el.Parent == nil {
				el.ConditionalStyles.Unconditional.SetFontSize(css.DefaultFontSize)
				base = css.DefaultFontSize
			} else {
				size, _ := el.Parent.Styles.GetFontSize()
				el.ConditionalStyles.Unconditional.SetFontSize(size)
				base = size
			}
		default:
			// el.Styles.SetFontSize(fontSizeToPx(strVal, el.Parent))
			base = fontSizeToPx(strVal, el.Parent)
			el.ConditionalStyles.Unconditional.SetFontSize(base)
		}

		switch strVal := el.ConditionalStyles.FirstLine.FontSize.GetValue(); strVal {
		case "":
			el.ConditionalStyles.FirstLine.SetFontSize(base)
		default:
			base = fontSizeToPx(strVal, el.Parent)
			el.ConditionalStyles.FirstLine.SetFontSize(base)
		}

		switch strVal := el.ConditionalStyles.FirstLetter.FontSize.GetValue(); strVal {
		case "":
			el.ConditionalStyles.FirstLetter.SetFontSize(base)
		default:
			base = fontSizeToPx(strVal, el.Parent)
			el.ConditionalStyles.FirstLetter.SetFontSize(base)
		}

		el.Styles = el.ConditionalStyles.FirstLine

		if el.Type == html.ElementNode && strings.ToLower(el.Data) == "body" {
			background := el.GetBackgroundColor()
			if background == color.Transparent {
				background = color.RGBA{0xE0, 0xE0, 0xE0, 0xFF}
			}
			p.Background = background
		}
	})
}

func fontSizeToPx(val string, parent *renderer.RenderableDomElement) int {
	DefaultFontSize := css.DefaultFontSize
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
