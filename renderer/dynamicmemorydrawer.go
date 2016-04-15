package renderer

import (
	"image"
	"image/color"
)

func NewDynamicMemoryDrawer(r image.Rectangle) *DynamicMemoryDrawer {
	return &DynamicMemoryDrawer{
		src: make(map[image.Point]color.RGBA),
		min: image.Point{0, 0},
		max: image.Point{r.Dx(), r.Dy()},
		model: color.RGBAModel,
	}
}
type DynamicMemoryDrawer struct {
	src map[image.Point]color.RGBA
	model color.Model
	min image.Point
	max image.Point
}

func (d *DynamicMemoryDrawer) ColorModel() color.Model {
	return d.model
}

func (d *DynamicMemoryDrawer) At(x, y int) color.Color {
	point := image.Point{X: x, Y: y}
	if c, ok := d.src[point]; ok {
		return c
	}
	return color.Transparent
	//return &color.RGBA{25, 50, 75, 255}
}

func (d *DynamicMemoryDrawer) Bounds() image.Rectangle {
	return image.Rectangle{
		Min: d.min,
		Max: d.max,
	}
}

func (d *DynamicMemoryDrawer) Set(x, y int, c color.Color) {
	if x < d.min.X {
		d.min.X = x
	}
	if y < d.min.Y {
		d.min.Y = y
	}
	if x > d.max.X {
		d.max.X = x
	}
	if y > d.max.Y {
		d.max.Y = y
	}
	point := image.Point{X: x, Y: y}
	if _, _, _, alpha := c.RGBA(); alpha == 0 {
		delete(d.src, point)
	} else {
		cl := d.model.Convert(c)
		if clr, ok := cl.(color.RGBA); ok {
			d.src[point] = clr
		}
	}
}
func (d *DynamicMemoryDrawer) GrowBounds(r image.Rectangle) {
	if r.Min.X < d.min.X {
		d.min.X = r.Min.X
	}
	if r.Min.Y < d.min.Y {
		d.min.Y = r.Min.Y
	}
	if r.Max.X > d.max.X {
		d.max.X = r.Max.X
	}
	if r.Max.Y > d.max.Y {
		d.max.Y = r.Max.Y
	}
}
