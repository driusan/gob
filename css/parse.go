package css

import (
	"github.com/driusan/Gob/net"
	"golang.org/x/net/html"
	"io/ioutil"
	"net/url"
)

// ExtractStyles takes an html Node as input, and extracts the unparsed text
// from any <style> elements in the HTML, returning the string of the style
// body.
func ExtractStyles(n *html.Node, context *url.URL) string {
	var style string
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
			return ""
		}
		newUrl, err := url.Parse(href)
		if err != nil {
			return ""
		}
		newAbsoluteURL := context.ResolveReference(newUrl)
		r, err := net.GetURLReader(newAbsoluteURL)
		if err != nil {
			return ""
		}
		defer r.Close()
		styles, err := ioutil.ReadAll(r)
		if err != nil {
			return ""
		}
		return string(styles)
	}

	if n.Type == html.ElementNode && n.Data == "style" {

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.TextNode {
				style += c.Data
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		style += ExtractStyles(c, context)
	}

	return style
}
