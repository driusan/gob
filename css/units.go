package css

import (
	"encoding/hex"
	"errors"
	"fmt"
	"image/color"
	"strconv"
	"strings"
)

var (
	NoStyles     = errors.New("No styles to apply")
	InheritValue = errors.New("Value should be inherited")
)

var PixelsPerPt float64

func init() {
	// Assume 96 DPI unless someone tells us otherwise at this resolution.
	PixelsPerPt = (96.0 / 72.0)
}

func NewPxValue(x int) StyleValue {
	return StyleValue{fmt.Sprintf("%dpx", x), false}
}
func NewValue(val string) StyleValue {
	return StyleValue{val, false}
}

func ConvertUnitToPx(fontsize int, percentbasis int, cssString string) (int, error) {
	if cssString == "0" {
		return 0, nil
	}
	if len(cssString) < 2 {
		return fontsize, fmt.Errorf("Invalid CSS Unit or value: %v", cssString)
	}

	val := cssString
	// handle percentages,
	if val[len(val)-1] == '%' {
		f, err := strconv.ParseFloat(string(val[0:len(val)-1]), 64)
		if err == nil {
			size := int(f * float64(percentbasis) / 100.0)
			return size, nil
		}
		return fontsize, fmt.Errorf("Invalid CSS Unit or value: %v", cssString)
	}

	// all other units are 2 characters long
	switch unit := string(val[len(val)-2:]); unit {
	case "em":
		// 1em is basically a scaling factor for the parent font
		// when calculating font size
		f, err := strconv.ParseFloat(string(val[0:len(val)-2]), 64)
		if err == nil {
			return int(f * float64(fontsize)), nil
		}
		return fontsize, fmt.Errorf("Invalid CSS Unit or value: %v", cssString)
	case "ex":
		// 1ex is supposed to be the height of a lower case x, but
		// the spec says you can use 1ex = 0.5em if calculating
		// the size of an x is impossible or impracticle. Since
		// I'm too lazy to figure out how to do that, it's impracticle.
		f, err := strconv.ParseFloat(string(val[0:len(val)-2]), 64)
		if err == nil {
			return int(f * float64(fontsize) / 2.0), nil
		}
		return fontsize, fmt.Errorf("Invalid CSS Unit or value: %v", cssString)
	case "px":
		// parse px as a float then cast, just in case someone
		// used a decimal.
		f, err := strconv.ParseFloat(string(val[0:len(val)-2]), 64)
		if err == nil {
			return int(f * PixelsPerPt * 0.75), nil
		}
		return fontsize, fmt.Errorf("Invalid CSS Unit or value: %v", cssString)
	case "in":
		f, err := strconv.ParseFloat(string(val[0:len(val)-2]), 64)
		if err == nil {
			return int(f * PixelsPerPt * 72), nil
		}
		return int(fontsize), fmt.Errorf("Invalid CSS Unit or value: %v", cssString)
	case "cm":
		f, err := strconv.ParseFloat(string(val[0:len(val)-2]), 64)
		if err == nil {
			return int(f * PixelsPerPt * 72.0 / 2.54), nil
		}
		return int(fontsize), fmt.Errorf("Invalid CSS Unit or value: %v", cssString)
	case "mm":
		f, err := strconv.ParseFloat(string(val[0:len(val)-2]), 64)
		if err == nil {
			return int(f * PixelsPerPt * 72.0 / 25.4), nil
		}
		return int(fontsize), fmt.Errorf("Invalid CSS Unit or value: %v", cssString)
	case "pt":
		f, err := strconv.ParseFloat(string(val[0:len(val)-2]), 64)
		if err == nil {
			return int(f * PixelsPerPt), nil
		}
		return int(fontsize), fmt.Errorf("Invalid CSS Unit or value: %v", cssString)
	case "pc":
		f, err := strconv.ParseFloat(string(val[0:len(val)-2]), 64)
		if err == nil {
			return int(f * PixelsPerPt * 12), nil
		}
		return int(fontsize), fmt.Errorf("Invalid CSS Unit or value: %v", cssString)
	}
	return fontsize, fmt.Errorf("Unimplemented CSS Unit or invalid value: %v", cssString)
}

func hexToUint8(val string) uint8 {
	if len(val) != 2 {
		panic("Invalid input")
	}
	r, err := hex.DecodeString(val)
	if err != nil {
		panic(err)
	}
	return uint8(r[0])
}
func sHexToUint8(val byte) uint8 {
	switch val {
	case '0': // 0x00
		return 0
	case '1': // 0x11
		return 0x11
	case '2':
		return 0x22
	case '3':
		return 0x33
	case '4':
		return 0x44
	case '5':
		return 0x55
	case '6':
		return 0x66
	case '7':
		return 0x77
	case '8':
		return 0x88
	case '9':
		return 0x99
	case 'a', 'A':
		return 0xAA
	case 'b', 'B':
		return 0xBB
	case 'c', 'C':
		return 0xCC
	case 'd', 'D':
		return 0xDD
	case 'e', 'E':
		return 0xEE
	case 'f', 'F':
		return 0xFF
	}
	panic("Invalid character")
}

func ConvertColorToRGBA(cssString string) (*color.RGBA, error) {

	black := &color.RGBA{0, 0, 0, 255}
	if len(cssString) > 3 && cssString[0:3] == "rgb" {
		tuple := cssString[4 : len(cssString)-1]
		pieces := strings.Split(tuple, ",")
		if len(pieces) != 3 {
			return black, fmt.Errorf("Invalid colour: %v", cssString)
		}

		rint, _ := strconv.Atoi(strings.TrimSpace(pieces[0]))
		gint, _ := strconv.Atoi(strings.TrimSpace(pieces[1]))
		bint, _ := strconv.Atoi(strings.TrimSpace(pieces[2]))
		return &color.RGBA{uint8(rint), uint8(gint), uint8(bint), 255}, nil

	} else if len(cssString) > 1 && cssString[0] == '#' {
		switch len(cssString) {
		case 7:
			// #RRGGBB
			return &color.RGBA{hexToUint8(cssString[1:3]), hexToUint8(cssString[3:5]), hexToUint8(cssString[5:]), 255}, nil
		case 4:
			// #RGB
			return &color.RGBA{sHexToUint8(cssString[1]), sHexToUint8(cssString[2]), sHexToUint8(cssString[3]), 255}, nil
		}
		return black, fmt.Errorf("Invalid colour: %v", cssString)
	}
	switch cssString {
	case "inherit":
		return black, InheritValue
	case "transparent":
		return &color.RGBA{0x80, 0, 0, 0}, nil
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
	return black, NoStyles
}

func IsColor(c string) bool {
	c = strings.TrimSpace(c)
	switch c {
	case "inherit", "transparent", "maroon", "red", "orange", "yellow", "olive", "purple",
		"fuchsia", "white", "lime", "green", "navy", "blue", "aqua", "teal",
		"black", "silver", "gray", "grey":
		return true
	}
	switch length := len(c); length {
	case 0:
		return false
	case 4, 7:
		if c[0] == '#' {
			for _, letter := range c[1:] {
				if (letter >= '0' && letter <= '9') ||
					(letter >= 'a' && letter <= 'f') ||
					(letter >= 'A' && letter <= 'F') {
					continue
				}
			}
			return true
		}
		return false
	default:
		if length > 4 && c[0:4] == "rgb(" {
			return true
		}
		return false
	}
}
func IsURL(u string) bool {
	u = strings.TrimSpace(u)
	if len(u) <= 4 {
		return false
	}
	return u[0:4] == "url("
}
func IsPercentage(p string) bool {
	p = strings.TrimSpace(p)
	if p == "" {
		return false
	}
	return p[len(p)-1] == '%'
}

func IsLength(l string) bool {
	l = strings.TrimSpace(l)
	if l == "0" {
		return true
	}
	if len(l) < 2 {
		return false
	}
	switch l[len(l)-2:] {
	case "in", "cm", "mm", "pt", "pc", "px":
		return true
	case "em", "ex":
		return true
	}
	return false
}

func IsBorderStyle(s string) bool {
	switch s {
	case "none", "hidden", "dotted", "dashed", "solid", "double",
		"groove", "ridge", "inset", "outset":
		return true
	}
	return false
}
