package main

import (
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
	for i := 0; i < b.N; i++ {
		f := strings.NewReader(content)
		parsedhtml := loadHTML(f, nil)
		parsedhtml.Content.Render(1024)
	}
}

func BenchmarkParseOnly(b *testing.B) {
	for i := 0; i < b.N; i++ {
		f := strings.NewReader(content)
		loadHTML(f, nil)
	}
}
func BenchmarkRenderOnly(b *testing.B) {
	f := strings.NewReader(content)
	parsedhtml := loadHTML(f, nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parsedhtml.Content.Render(1024)
	}
}
