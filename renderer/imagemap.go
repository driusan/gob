package renderer

import (
	"image"
	//"image/color"
)

type AreaMapping struct {
	Area    image.Rectangle
	Content *RenderableDomElement
}

type ImageMap []AreaMapping

func NewImageMap() ImageMap {
	return make([]AreaMapping, 0)
}

func (imap ImageMap) At(x, y int) *RenderableDomElement {
	var candidate *RenderableDomElement =nil
	for i := len(imap) - 1; i >= 0; i-- {
		area := imap[i]
		p := image.Point{x, y}
		if p.In(area.Area) {
			if candidate == nil {
				candidate = area.Content
			} else if area.Content.GetFloat() != "none" {
				candidate = area.Content
				return candidate.ImageMap.At(x-candidate.BoxDrawRectangle.Min.X, y-candidate.BoxDrawRectangle.Min.Y)
			}
		}
	}
	return candidate
}

func (imap *ImageMap) Add(el *RenderableDomElement, location image.Rectangle) {
	*imap = append(*imap, AreaMapping{
		Area:    location,
		Content: el,
	})
}
