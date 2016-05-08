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

func (f FloatStack) Width() int {
	if f == nil {
		return 0
	}
	var width int = 0
	for _, child := range f {
		width += child.CSSOuterBox.Bounds().Size().X
	}

	return width
}

// Remove any floats that are past dot from the float stack
func (f FloatStack) ClearFloats(dot image.Point) FloatStack {
	var removed = f

	for i, child := range f {
		if (child.CSSOuterBox.Bounds().Size().Y) > dot.Y {
			removed = append(f[:i], f[i+1:]...)
		}
	}
	return removed
}

func (f FloatStack) NextFloatHeight() int {
	if f == nil || len(f) == 0 {
		return 0
	}
	lastElem := f[len(f)-1]

	size := lastElem.CSSOuterBox.Bounds().Size().Y
	if size == 0 {
		fmt.Printf("Bounds size is 0?? %s %s %s", lastElem.CSSOuterBox.Bounds(), lastElem.Data, lastElem.GetTextContent())
	}
	return 0
}
