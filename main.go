package main

import (
	//"runtime/pprof"
	"Gob/net"
	"Gob/renderer"
	"fmt"
	"golang.org/x/exp/shiny/driver"
	"golang.org/x/exp/shiny/screen"
	"golang.org/x/mobile/event/key"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/mouse"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
	"golang.org/x/mobile/event/touch"
	"golang.org/x/net/html"
	"image"
	"image/color"
	"image/draw"
	"net/url"
	"os"
)

var (
	background = color.Color(color.RGBA{0xE0, 0xE0, 0xE0, 0xFF})
	//	background = color.RGBA{0xFF, 0xFF, 0xFF, 0xFF}
)

type Viewport struct {
	// The size of the viewport
	Size size.Event

	// The whole, source image to be displayed in the viewport. It will be clipped
	// and displayed in the viewport according to the Size and Cursor
	Content image.Image

	// The location of the image to be displayed into the viewpart.
	Cursor image.Point
}
type Page struct {
	Content *renderer.RenderableDomElement
	URL     *url.URL
}

func paintWindow(s screen.Screen, w screen.Window, v *Viewport, page *Page) {
	viewport := v.Size.Bounds()

	if v.Content != nil {
		b, err := s.NewBuffer(v.Size.Size())
		dst := b.RGBA()

		// Fill the buffer with the window background colour before
		// drawing the web page on top of it.
		draw.Draw(dst, dst.Bounds(), &image.Uniform{background}, image.ZP, draw.Src)

		// Draw the clipped portion of the page that is within view
		draw.Draw(dst, viewport, v.Content, v.Cursor, draw.Over)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s", err)
			return
		}
		defer b.Release()
		draw.Draw(dst, viewport, v.Content, v.Cursor, draw.Over)
		w.Upload(image.Point{0, 0}, b, viewport)
	} else {
		fmt.Fprintf(os.Stderr, "No body to render!\n")
	}
	w.Publish()
}

func main() {
	/*
		f, _ := os.Create("test.profile")
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	*/

	filename := "test.html"
	if len(os.Args) > 1 {
		filename = os.Args[1]

	}

	page, err := loadPage(filename)
	if err != nil || page == nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		return
	}

	driver.Main(func(s screen.Screen) {
		w, err := s.NewWindow(nil)
		if err != nil {
			panic(err)
		}
		defer w.Release()

		var v Viewport
		// there will be a size event immediately after creating
		// the window which will trigger this.
		for {
			switch e := w.NextEvent().(type) {
			case lifecycle.Event:
				if e.To == lifecycle.StageDead {
					return
				}
			case key.Event:
				switch e.Code {
				case key.CodeEscape:
					return
				case key.CodeLeftArrow:
					if e.Direction == key.DirPress {
						scrollSize := 50
						v.Cursor.X -= scrollSize
						if v.Cursor.X > v.Content.Bounds().Max.X {
							v.Cursor.X = v.Content.Bounds().Max.X - 10
						}
						paintWindow(s, w, &v, page)
					}
				case key.CodeRightArrow:
					if e.Direction == key.DirPress {
						scrollSize := 50
						v.Cursor.X += scrollSize
						if v.Cursor.X > v.Content.Bounds().Max.X {
							v.Cursor.X = v.Content.Bounds().Max.X - 10
						}
						paintWindow(s, w, &v, page)
					}
				case key.CodeDownArrow:
					if e.Direction == key.DirPress {
						scrollSize := v.Size.Size().Y / 2
						v.Cursor.Y += scrollSize
						if v.Cursor.Y > v.Content.Bounds().Max.Y {
							v.Cursor.Y = v.Content.Bounds().Max.Y - 10
						}
						paintWindow(s, w, &v, page)
					}
				case key.CodeUpArrow:
					if e.Direction == key.DirPress {
						scrollSize := v.Size.Size().Y / 2
						v.Cursor.Y -= scrollSize
						if v.Cursor.Y < 0 {
							v.Cursor.Y = 0
						}
						paintWindow(s, w, &v, page)
					}
				default:
					fmt.Printf("Unknown key: %s", e.Code)
				}
			case paint.Event:
				paintWindow(s, w, &v, page)
			case size.Event:
				v.Size = e
				v.Content = page.Content.Render(e.Size().X)
			case touch.Event:
				fmt.Printf("Touch event!")
			case mouse.Event:
				//fmt.Printf("Mouse event at %d, %d! %e", e.X, e.Y, e)
				if page.Content != nil && page.Content.ImageMap != nil {

					el := page.Content.ImageMap.At(int(e.X)+v.Cursor.X, int(e.Y)+v.Cursor.Y)
					if el != nil {
						switch e.Direction {
						case mouse.DirRelease:
							el.OnClick()
							if el.Type == html.ElementNode && el.Data == "a" {
								p, err := loadNewPage(page.URL, el.GetAttribute("href"))
								if err == nil && p != nil {
									page = p
									v.Content = p.Content.Render(v.Size.Size().X)
								}
							}
						default:
							if el.Type == html.ElementNode && el.Data == "a" {
								fmt.Printf("Hovering over link %s\n", el.GetAttribute("href"))
							}
						}
					}
				}
			default:
				//	fmt.Printf("%s\n", e)
			}
		}
	})
}
func loadNewPage(context *url.URL, path string) (*Page, error) {
	fmt.Printf("Inew newPage\n")
	u, err := url.Parse(path)
	if err != nil {
		return nil, err
	}
	newURL := context.ResolveReference(u)
	r, err := net.GetURLReader(newURL)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	fmt.Printf("Loading new page %s\n", newURL)
	p := loadHTML(r, newURL)
	p.URL = newURL
	fmt.Printf("Loaded new page %s\n", newURL)
	return p, nil
}
