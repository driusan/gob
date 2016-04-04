package dom

// A TextElement is a renderable TextNode from an HTML document.
type TextElement struct {
	TextContent string
}

func (e TextElement) GetTextContent() string {
	return e.TextContent
}
