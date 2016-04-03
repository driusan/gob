package css

import (
	"golang.org/x/net/html"
)

// ExtractStyles takes an html Node as input, and extracts the unparsed text
// from any <style> elements in the HTML, returning the string of the style
// body.
// TODO(driusan): Extract styles from <link> elements too
func ExtractStyles(n *html.Node) string {
	var style string
	if n.Type == html.ElementNode && n.Data == "style" {
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.TextNode {
				style += c.Data
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		style += ExtractStyles(c)
	}

	return style
}
