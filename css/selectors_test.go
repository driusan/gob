package css

import (
	"golang.org/x/net/html"
	"strings"
	"testing"
	//"fmt"
	//"golang.org/x/mobile/event/size"
	//"os"
)

var content string = `<!DOCTYPE html>
<html>
    <head>
        <meta charset="utf-8">
        <title>Gob's Selector Test Document</title>
    </head>
    <body>
        <div class="site otherclass" id="sitediv">
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

func TestCSS1SimpleSelector(t *testing.T) {
	f := strings.NewReader(content)
	doc, err := html.Parse(f)
	if err != nil {
		print("Could not parse sample document\n")
		t.Fail()
	}

	head := doc.FirstChild.NextSibling.FirstChild
	body := head.NextSibling.NextSibling // the first sibling is a whitespace text node
	sitediv := body.FirstChild.NextSibling

	// test some basic type selectors
	rule := StyleRule{Selector: "body"}
	if rule.Matches(body) != true {
		print("body did not match body element\n")
		t.Fail()
	}
	rule.Selector = "h1"
	if rule.Matches(body) != false {
		print("body incorrectly matched h1 element\n")
		t.Fail()
	}

	rule.Selector = "div"
	if rule.Matches(sitediv) != true {
		print("div did not match \"div\" selector\n")
		t.Fail()
	}

	// test variations of class selectors
	rule.Selector = ".site"
	if rule.Matches(sitediv) != true {
		print("div did not match \".site\" class selector\n")
		t.Fail()
	}
	rule.Selector = "div.site"
	if rule.Matches(sitediv) != true {
		print("div did not match \"div.site\" class selector\n")
		t.Fail()
	}
	rule.Selector = "*.site"
	if rule.Matches(sitediv) != true {
		print("div did not match \"*.site\" class selector\n")
		t.Fail()
	}
	// make sure the id isn't interpreted as a class
	rule.Selector = ".sitediv"
	if rule.Matches(sitediv) != false {
		print("div incorrectly matched id sitediv as a class\n")
		t.Fail()
	}

	// test variations of id selectors
	rule.Selector = "#sitediv"
	if rule.Matches(sitediv) != true {
		print("div did not match \"#sitediv\" id selector\n")
		t.Fail()
	}
	rule.Selector = "div#sitediv"
	if rule.Matches(sitediv) != true {
		print("div did not match \"div#sitediv\" id selector\n")
		t.Fail()
	}
	rule.Selector = "h1#sitediv"
	if rule.Matches(sitediv) != false {
		print("div with sitediv id incorrectly matched wrong element type (\"h1#sitediv\") selector\n")
		t.Fail()
	}
	rule.Selector = "*#sitediv"
	if rule.Matches(sitediv) != true {
		print("div with sitediv id did not match \"*#sitediv\" selector\n")
		t.Fail()
	}

	// test both class and id
	rule.Selector = "#sitediv.site"
	if rule.Matches(sitediv) != true {
		print("div with sitediv id and site class did not match \"#sitediv.site\" selector\n")
		t.Fail()
	}
	rule.Selector = ".site#sitediv"
	if rule.Matches(sitediv) != true {
		print("div with sitediv id and site class did not match \".site#sitediv\" selector\n")
		t.Fail()
	}
	rule.Selector = "div#sitediv.site"
	if rule.Matches(sitediv) != true {
		print("div with sitediv id and site class did not match \"div#sitediv.site\" selector\n")
		t.Fail()
	}
	rule.Selector = "div.site#sitediv"
	if rule.Matches(sitediv) != true {
		print("div with sitediv id and site class did not match \"div.site#sitediv\" selector\n")
		t.Fail()
	}
	rule.Selector = "*#sitediv.site"
	if rule.Matches(sitediv) != true {
		print("div with sitediv id and site class did not match \"*#sitediv.site\" selector\n")
		t.Fail()
	}
	rule.Selector = "*.site#sitediv"
	if rule.Matches(sitediv) != true {
		print("div with sitediv id and site class did not match \"*.site#sitediv\" selector\n")
		t.Fail()
	}
	rule.Selector = "h1.site#sitediv"
	if rule.Matches(sitediv) != false {
		print("div with sitediv id and site class incorrectly matched h1 tag \"h1.site#sitediv\" selector\n")
		t.Fail()
	}
}
func TestCSS1ParentSelector(t *testing.T) {
	f := strings.NewReader(content)
	doc, err := html.Parse(f)
	if err != nil {
		print("Could not parse sample document\n")
		t.Fail()
	}

	head := doc.FirstChild.NextSibling.FirstChild
	body := head.NextSibling.NextSibling // the first sibling is a whitespace text node
	sitediv := body.FirstChild.NextSibling
	headerdiv := sitediv.FirstChild.NextSibling
	h1 := headerdiv.FirstChild.NextSibling

	// The portion of the document that we're looking at looks like:
	//<div class="site otherclass" id="sitediv">
	//  <div class="header">
	//    <h1 class="title"><a href="/">Gob Benchmark Test</a></h1>
	//    <a class="extra" href="/">home</a>
	//  </div>
	// [...]
	//</div>
	rule := StyleRule{Selector: "h1"}
	if rule.Matches(h1) != true {
		t.Error("Did not match simple h1 selector")
	}
	rule.Selector = "div h1"
	if rule.Matches(h1) != true {
		t.Error("Did not match parent div selector")
	}
	rule.Selector = ".header h1"
	if rule.Matches(h1) != true {
		t.Error("Did not match parent div selector by class")
	}
	rule.Selector = ".site h1"
	if rule.Matches(h1) != true {
		t.Error("Ancestor selector did not grandparent")
	}
}
