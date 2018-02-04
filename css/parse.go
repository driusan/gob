package css

import (
	"io/ioutil"
	"net/url"

	"github.com/driusan/Gob/net"
	"golang.org/x/net/html"
)

// ExtractStyles takes an html Node as input, and extracts the unparsed text
// from any <style> elements in the HTML, returning the string of the style
// body.
func ExtractStyles(n *html.Node, loader net.URLReader, context *url.URL) Stylesheet {
	if n.Type == html.ElementNode && n.Data == "link" {
		var href, rel string
		for _, attr := range n.Attr {
			switch attr.Key {
			case "href":
				href = attr.Val
			case "rel":
				rel = attr.Val
			}
		}
		if href == "" || rel != "stylesheet" {
			return nil
		}
		newUrl, err := url.Parse(href)
		if err != nil {
			return nil
		}
		newAbsoluteURL := context.ResolveReference(newUrl)
		r, resp, err := loader.GetURL(newAbsoluteURL)
		if err != nil {
			return nil
		}

		// Only parse the stylesheet if it's found, otherwise we extract
		// styles from 404 error pages
		if resp < 200 || resp >= 300 {
			return nil
		}
		defer r.Close()
		styles, err := ioutil.ReadAll(r)
		if err != nil {
			return nil
		}
		return ParseStylesheet(string(styles), AuthorSrc, loader, newAbsoluteURL)
	}

	var styleElem string
	if n.Type == html.ElementNode && n.Data == "style" {
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.TextNode {
				styleElem += c.Data
			}
		}
	}
	style := ParseStylesheet(styleElem, AuthorSrc, loader, context)
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		style = append(style, ExtractStyles(c, loader, context)...)
	}
	return style
}
