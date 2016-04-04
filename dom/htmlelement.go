package dom

import "golang.org/x/net/html"

type Element html.Node

func (e Element) GetTextContent() string {
	return ""
}
