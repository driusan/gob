package main

import (
	"Gob/css"
	"Gob/renderer"
	"fmt"
	"golang.org/x/net/html"
	"strconv"
	//	"strings"
	"io"
	"io/ioutil"
	"net/url"
	"os"
)

func loadHTML(r io.Reader, urlContext *url.URL) *Page {
	parsedhtml, _ := html.Parse(r)
	styles := css.ExtractStyles(parsedhtml, urlContext)

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

	styles2 := css.ParseStylesheet(styles, css.AuthorSrc)

	renderable, _ := renderer.ConvertNodeToRenderableElement(body)

	sheet, _ := ioutil.ReadFile("useragent.css")
	userAgentStyles := css.ParseStylesheet(string(sheet), css.UserAgentSrc)

	renderable.Walk(func(el *renderer.RenderableDomElement) {
		el.PageLocation = urlContext
		for _, rule := range userAgentStyles {
			if rule.Matches((*html.Node)(el.Element)) {
				el.Styles.AddStyle(rule)
			}
		}

		for _, rule := range styles2 {
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
							Selector: "",
							Name:     name,
							Value:    val,
							Src:      css.InlineStyleSrc,
						})
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

		if el.Type == html.ElementNode && el.Data == "body" {
			background = el.GetBackgroundColor()
		}
	})

	return &Page{
		Content: renderable,
		URL:     nil,
	}

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
	}
	// handle percentages,
	if val[len(val)-1] == '%' {
		f, err := strconv.ParseFloat(string(val[0:len(val)-1]), 64)
		if err == nil {
			psize, err := parent.Styles.GetFontSize()
			if err != nil {
				return DefaultFontSize
			}
			size := int(f * float64(psize) / 100.0)
			return size
		}
		return DefaultFontSize
	}

	// all other units are 2 characters long
	switch unit := string(val[len(val)-2:]); unit {
	case "em":
		// 1em is basically a scaling factor for the parent font
		// when calculating font size
		f, err := strconv.ParseFloat(string(val[0:len(val)-2]), 64)
		if err == nil {
			psize, _ := parent.Styles.GetFontSize()
			return int(f * float64(psize))
		}
		return DefaultFontSize
	case "ex":
		// 1ex is supposed to be the height of a lower case x, but
		// the spec says you can use 1ex = 0.5em if calculating
		// the size of an x is impossible or impracticle. Since
		// I'm too lazy to figure out how to do that, it's impracticle.
		f, err := strconv.ParseFloat(string(val[0:len(val)-2]), 64)
		if err == nil {
			psize, _ := parent.Styles.GetFontSize()
			return int(f * float64(psize) / 2.0)
		}
		return DefaultFontSize
	case "px":
		// parse px as a float then cast, just in case someone
		// used a decimal.
		f, err := strconv.ParseFloat(string(val[0:len(val)-2]), 64)
		if err == nil {
			return int(f)
		}
		return DefaultFontSize
	case "in", "cm", "mm", "pt", "pc":
		fmt.Fprintf(os.Stderr, "Unhandled unit for font size: %s. Using default.\n", unit)
		return DefaultFontSize
	}

	// If nothing's been handled and there's no parent, use the default.
	if parent == nil {
		return DefaultFontSize
	}

	// if there is a parent, use it if we can.
	switch psize, err := parent.Styles.GetFontSize(); err {
	case nil:
		return psize
	default:
		return DefaultFontSize
	}

}
