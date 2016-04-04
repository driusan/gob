package css

import (
	"errors"
	"image/color"
	"strconv"
	"strings"
)

var (
	NoStyles       = errors.New("No styles to apply")
	InvalidUnit    = errors.New("Invalid CSS unit")
	NotImplemented = errors.New("Support not yet implemented")
)

func ConvertUnitToPx(basis int, cssString string) (int, error) {
	if len(cssString) < 2 {
		return basis, InvalidUnit
	}
	if cssString[len(cssString)-2:] == "px" {
		val, _ := strconv.Atoi(cssString[0 : len(cssString)-2])
		return val, nil

	}
	return basis, NotImplemented
}

func ConvertColorToRGBA(cssString string) (*color.RGBA, error) {
	if cssString[0:3] == "rgb" {
		tuple := cssString[4 : len(cssString)-1]
		pieces := strings.Split(tuple, ",")
		if len(pieces) != 3 {
			panic("wrong number of colors")
		}

		rint, _ := strconv.Atoi(strings.TrimSpace(pieces[0]))
		gint, _ := strconv.Atoi(strings.TrimSpace(pieces[1]))
		bint, _ := strconv.Atoi(strings.TrimSpace(pieces[2]))
		return &color.RGBA{uint8(rint), uint8(gint), uint8(bint), 255}, nil

	}
	switch cssString {
	case "maroon":
		return &color.RGBA{0x80, 0, 0, 255}, nil
	case "red":
		return &color.RGBA{0xff, 0, 0, 255}, nil
	case "orange":
		return &color.RGBA{0xff, 0xa5, 0, 255}, nil
	case "yellow":
		return &color.RGBA{0xff, 0xff, 0, 255}, nil
	case "olive":
		return &color.RGBA{0x80, 0x80, 0, 255}, nil
	case "purple":
		return &color.RGBA{0x80, 0, 0x80, 255}, nil
	case "fuchsia":
		return &color.RGBA{0xff, 0, 0xff, 255}, nil
	case "white":
		return &color.RGBA{0xff, 0xff, 0xff, 255}, nil
	case "lime":
		return &color.RGBA{0, 0xff, 0, 255}, nil
	case "green":
		return &color.RGBA{0, 0x80, 0, 255}, nil
	case "navy":
		return &color.RGBA{0, 0, 0x80, 255}, nil
	case "blue":
		return &color.RGBA{0, 0, 0xff, 255}, nil
	case "aqua":
		return &color.RGBA{0, 0xff, 0xff, 255}, nil
	case "teal":
		return &color.RGBA{0, 0x80, 0x80, 255}, nil
	case "black":
		return &color.RGBA{0, 0, 0, 255}, nil
	case "silver":
		return &color.RGBA{0xc0, 0xc0, 0xc0, 255}, nil
	case "gray", "grey":
		return &color.RGBA{0x80, 0x80, 0x80, 255}, nil
	}
	return nil, NoStyles
}
