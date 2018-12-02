package main

import (
	"github.com/driusan/Gob/net"
	"github.com/driusan/Gob/parser"

	"context"
	"image"
	"strings"
	"testing"
	//"fmt"
	//"golang.org/x/mobile/event/size"
	//"os"
)

var content string = `
<!DOCTYPE html>
<html>
    <head>
        <meta charset="utf-8">
        <title>Gob's Basic Benchmark</title>
<style>
html, body {
    height: 100%; 
    width: 100%;
}

body {
  color: black;
  background: #115050;
}

.site {
    margin-left: auto;
    margin-right: auto;
    width: 80%;
    background: white;
    padding: 3%;
}

a {
    color: #155050;

}

a:visited {
    color: #5B6868;
}

a:hover {
    color: black;
}

/* Header */
.header a {
    color: #115050;
    text-decoration: none;
}

.header h1 {
    display: inline;
    padding-right: 10px;

}

</style>
    </head>
    <body>
        <div class="site">
          <div class="header">
            <h1 class="title"><a href="/">Gob Benchmark Test</a></h1>
            <a class="extra" href="/">home</a>
          </div>

          <div id="content">
          <p style="border-width: 5px; border-color: green; margin: 5px; padding: 10px; border-style: solid;">
		This is a simple benchmark which <a href="/">contains</a>various inline styles</p>
	  </p>

<p>It's all HTML and CSS1</p>
          </div>
        </div>

    </body>
</html>
`

func BenchmarkParseAndRender(b *testing.B) {
	loader := net.DefaultReader{}
	for i := 0; i < b.N; i++ {
		f := strings.NewReader(content)
		parsedhtml := parser.LoadPage(f, loader, nil)
		parsedhtml.Content.Render(1024)
	}
}
func BenchmarkParseAndRenderInto(b *testing.B) {
	loader := net.DefaultReader{}
	dst := image.NewRGBA(image.Rectangle{image.ZP, image.Point{1024, 768}})
	for i := 0; i < b.N; i++ {
		f := strings.NewReader(content)
		parsedhtml := parser.LoadPage(f, loader, nil)
		parsedhtml.Content.RenderInto(context.TODO(), dst)
	}
}

func BenchmarkParseOnly(b *testing.B) {
	loader := net.DefaultReader{}
	for i := 0; i < b.N; i++ {
		f := strings.NewReader(content)
		parser.LoadPage(f, loader, nil)
	}
}
func BenchmarkRenderOnly(b *testing.B) {
	loader := net.DefaultReader{}
	f := strings.NewReader(content)
	parsedhtml := parser.LoadPage(f, loader, nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parsedhtml.Content.Render(1024)
	}
}

func BenchmarkRenderLayoutOnly(b *testing.B) {
	loader := net.DefaultReader{}
	f := strings.NewReader(content)
	parsedhtml := parser.LoadPage(f, loader, nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var lh int
		parsedhtml.Content.LayoutPass(1024, image.ZR, &image.Point{0, 0}, &lh)
	}
}
func BenchmarkRenderDrawOnly(b *testing.B) {
	loader := net.DefaultReader{}
	f := strings.NewReader(content)
	parsedhtml := parser.LoadPage(f, loader, nil)
	var lh int
	parsedhtml.Content.LayoutPass(1024, image.ZR, &image.Point{0, 0}, &lh)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parsedhtml.Content.DrawPass()
	}
}
