package renderer

import (
	"fmt"
	"image"
)

// A FloatStack represents the floating elements that haven't
// yet been cleared. There's generally two FloatStacks, one
// for the left elements, and one for the right elements.
// Once dot advances past the FloatStack, they're removed
// from the stack.
type FloatStack []*RenderableDomElement

func (f FloatStack) WidthAt(loc image.Point) int {
	if f == nil {
		return 0
	}
	var width int = 0
	for _, child := range f {
		bounds := child.BoxDrawRectangle
		if loc.Y >= bounds.Min.Y && loc.Y <= bounds.Max.Y {
			width += bounds.Size().X
		}
	}

	return width
}

// Remove any floats that are past dot from the float stack and return
// the floats that have not yet been cleared.
func (f FloatStack) ClearFloats(dot image.Point) FloatStack {
	var newstack = make(FloatStack, 0, len(f))

	for _, child := range f {
		if dot.Y < (child.BoxDrawRectangle.Max.Y) {
			newstack = append(newstack, child)
		}
	}
	return newstack
}

func (f FloatStack) NextFloatHeight() int {
	if len(f) == 0 {
		return 0
	}
	lastElem := f[len(f)-1]

	//size := lastElem.CSSOuterBox.Bounds().Size().Y
	size := lastElem.BoxDrawRectangle.Size().Y
	if size == 0 {
		fmt.Printf("Bounds size is 0?? %s %s %s", lastElem.CSSOuterBox.Bounds(), lastElem.Data, lastElem.GetTextContent())
	}
	return size
}
