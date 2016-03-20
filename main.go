package main

import (
	"errors"
	"fmt"
	"golang.org/x/exp/shiny/driver"
	"golang.org/x/exp/shiny/screen"
	"golang.org/x/mobile/event/key"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
	"golang.org/x/mobile/event/touch"
	"golang.org/x/net/html"
	"image"
	"image/color"
	"io"
	"os"
	"strconv"
	"strings"
)

var (
	background = color.RGBA{0xE0, 0xE0, 0xE0, 0xFF}
	//	background = color.RGBA{0xFF, 0xFF, 0xFF, 0xFF}

	NoStyles     = errors.New("No styles to apply")
	NotAnElement = errors.New("Not an element node")
)

type Viewport struct {
	Size size.Event
}
type Page struct {
	//*html.Node
	Body *HTMLElement
}

func convertUnitToPx(basis int, cssString string) int {
	//fmt.Printf("Attempting to interpret '%s'\n", cssString)
	if cssString[len(cssString)-2:] == "px" {
		val, _ := strconv.Atoi(cssString[0 : len(cssString)-2])
		return val

	}
	panic("aaaah")
}
func convertUnitToColor(cssString string) (*color.RGBA, error) {
	//background: rgb(0, 0, 255);
	//fmt.Printf("Attempting to interpret '%s'\n", cssString)
	if cssString[0:3] == "rgb" {
		tuple := cssString[4 : len(cssString)-1]
		pieces := strings.Split(tuple, ",")
		if len(pieces) != 3 {
			panic("wrong number of colors")
		}
		//for i, val := range pieces {

		//fmt.Printf("%d: %s\n", i, val)
		//}
		rint, _ := strconv.Atoi(strings.TrimSpace(pieces[0]))
		gint, _ := strconv.Atoi(strings.TrimSpace(pieces[1]))
		bint, _ := strconv.Atoi(strings.TrimSpace(pieces[2]))
		return &color.RGBA{uint8(rint), uint8(gint), uint8(bint), 255}, nil

	}
	return nil, NoStyles
}

func extractStyles(n *html.Node) string {
	var style string
	if n.Type == html.ElementNode && n.Data == "style" {
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.TextNode {
				style += c.Data
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		style += extractStyles(c)
	}

	return style
}

func convertNodeToHTMLElement(root *html.Node) (*HTMLElement, error) {
	switch root.Type {
	case html.ElementNode:
		fmt.Printf("Convertin an element %s\n", root.Data)
		var textContent string
		var children []*HTMLElement
		var lastError error
		for c := root.FirstChild; c != nil; c = c.NextSibling {
			switch c.Type {
			case html.ElementNode:
				newChild, err := convertNodeToHTMLElement(c)
				if err != nil {
					lastError = err
					continue
				}
				children = append(children, newChild)
			case html.TextNode:
				textContent += c.Data
			}
		}

		return &HTMLElement{root, nil, textContent, children}, lastError
	case html.TextNode:
		return &HTMLElement{nil, nil, root.Data, nil}, NotAnElement
	default:
		return nil, NotAnElement
	}
	fmt.Printf("This should not happen.\n")
	return nil, NotAnElement
}
func parseHTML(r io.Reader) (*Page, Stylesheet) {
	parsedhtml, _ := html.Parse(r)
	styles := extractStyles(parsedhtml)

	var body *HTMLElement
	var root = parsedhtml.FirstChild
	for c := root.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == "body" {
			body, _ = convertNodeToHTMLElement(c)
			break
		}
	}
	return &Page{body},
		ParseStylesheet(styles)

}

func realWalkBody(n *HTMLElement, callback func(e *HTMLElement)) {
	if n == nil {
		return
	}
	if n.Type == html.ElementNode {
		callback(n)
	}
	for _, c := range n.Children {
		realWalkBody(c, callback)
	}
}
func (p Page) WalkBody(callback func(*HTMLElement)) {
	if p.Body == nil {
		panic("Nothing to walk")
	}
	realWalkBody(p.Body, callback)
}

func paintWindow(s screen.Screen, w screen.Window, v *Viewport, page *Page, sty Stylesheet) {
	page.WalkBody(func(n *HTMLElement) {
		for _, rule := range sty {
			if rule.Matches(n) {
				n.AddStyle(rule)
			}
		}
	})

	// Fill the window background with gray
	w.Fill(v.Size.Bounds(), background, screen.Src)

	if page.Body != nil {
		b, err := s.NewBuffer(v.Size.Size())
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s", err)
			return
		}
		defer b.Release()
		page.Body.Render(b.RGBA())
		w.Upload(image.Point{0, 0}, b, v.Size.Bounds())
	} else {
		fmt.Fprintf(os.Stderr, "No body to render!\n")
	}
	w.Publish()
}
func main() {
	driver.Main(func(s screen.Screen) {
		w, err := s.NewWindow(nil)
		if err != nil {
			panic(err)
		}
		defer w.Release()

		f, err := os.Open("test.html")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not open test.html\n")
			return
		}
		parsedhtml, sty := parseHTML(f)
		f.Close()

		var v Viewport
		for {
			switch e := w.NextEvent().(type) {
			case lifecycle.Event:
				if e.To == lifecycle.StageDead {
					return
				}
			case key.Event:
				if e.Code == key.CodeEscape {
					return
				}
			case paint.Event:
				fmt.Printf("Painting\n")
				paintWindow(s, w, &v, parsedhtml, sty)
			case size.Event:
				fmt.Printf("Resizing window\n")
				v.Size = e
			case touch.Event:
				fmt.Printf("Touch event!")
			default:
				//	fmt.Printf("%s\n", e)
			}
		}
	})
}
