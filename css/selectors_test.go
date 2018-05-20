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
	    <a name="label">I am a label</a>
          </div>

          <div id="content">
          <p style="border-width: 5px; border-color: green; margin: 5px; padding: 10px; border-style: solid;">
		This is a simple benchmark which <a href="/">contains</a>various inline styles</p>
	  </p>

	  <p>It's all HTML and CSS1</p>
          </div>
          <ul>
	  	<li>Text Content
			<ul>
				<li>This is a regression test</li>
			</ul>
		</li>
		</ul>
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

	var st State
	// test some basic type selectors
	rule := StyleRule{Selector: CSSSelector{"body", 0}}

	if rule.Matches(body, st) != true {
		print("body did not match body element\n")
		t.Fail()
	}
	rule.Selector.Selector = "h1"
	if rule.Matches(body, st) != false {
		print("body incorrectly matched h1 element\n")
		t.Fail()
	}

	rule.Selector.Selector = "div"
	if rule.Matches(sitediv, st) != true {
		print("div did not match \"div\" selector\n")
		t.Fail()
	}

	// test variations of class selectors
	rule.Selector.Selector = ".site"
	if rule.Matches(sitediv, st) != true {
		print("div did not match \".site\" class selector\n")
		t.Fail()
	}
	rule.Selector.Selector = "div.site"
	if rule.Matches(sitediv, st) != true {
		print("div did not match \"div.site\" class selector\n")
		t.Fail()
	}
	rule.Selector.Selector = "*.site"
	if rule.Matches(sitediv, st) != true {
		print("div did not match \"*.site\" class selector\n")
		t.Fail()
	}
	// make sure the id isn't interpreted as a class
	rule.Selector.Selector = ".sitediv"
	if rule.Matches(sitediv, st) != false {
		print("div incorrectly matched id sitediv as a class\n")
		t.Fail()
	}

	// test variations of id selectors
	rule.Selector.Selector = "#sitediv"
	if rule.Matches(sitediv, st) != true {
		print("div did not match \"#sitediv\" id selector\n")
		t.Fail()
	}
	rule.Selector.Selector = "div#sitediv"
	if rule.Matches(sitediv, st) != true {
		print("div did not match \"div#sitediv\" id selector\n")
		t.Fail()
	}
	rule.Selector.Selector = "h1#sitediv"
	if rule.Matches(sitediv, st) != false {
		print("div with sitediv id incorrectly matched wrong element type (\"h1#sitediv\") selector\n")
		t.Fail()
	}
	rule.Selector.Selector = "*#sitediv"
	if rule.Matches(sitediv, st) != true {
		print("div with sitediv id did not match \"*#sitediv\" selector\n")
		t.Fail()
	}

	// test both class and id
	rule.Selector.Selector = "#sitediv.site"
	if rule.Matches(sitediv, st) != true {
		print("div with sitediv id and site class did not match \"#sitediv.site\" selector\n")
		t.Fail()
	}
	rule.Selector.Selector = ".site#sitediv"
	if rule.Matches(sitediv, st) != true {
		print("div with sitediv id and site class did not match \".site#sitediv\" selector\n")
		t.Fail()
	}
	rule.Selector.Selector = "div#sitediv.site"
	if rule.Matches(sitediv, st) != true {
		print("div with sitediv id and site class did not match \"div#sitediv.site\" selector\n")
		t.Fail()
	}
	rule.Selector.Selector = "div.site#sitediv"
	if rule.Matches(sitediv, st) != true {
		print("div with sitediv id and site class did not match \"div.site#sitediv\" selector\n")
		t.Fail()
	}
	rule.Selector.Selector = "*#sitediv.site"
	if rule.Matches(sitediv, st) != true {
		print("div with sitediv id and site class did not match \"*#sitediv.site\" selector\n")
		t.Fail()
	}
	rule.Selector.Selector = "*.site#sitediv"
	if rule.Matches(sitediv, st) != true {
		print("div with sitediv id and site class did not match \"*.site#sitediv\" selector\n")
		t.Fail()
	}
	rule.Selector.Selector = "h1.site#sitediv"
	if rule.Matches(sitediv, st) != false {
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
	var st State

	// The portion of the document that we're looking at looks like:
	//<div class="site otherclass" id="sitediv">
	//  <div class="header">
	//    <h1 class="title"><a href="/">Gob Benchmark Test</a></h1>
	//    <a class="extra" href="/">home</a>
	//    <a name="label">I am a label</a>
	//  </div>
	// [...]
	//</div>
	rule := StyleRule{Selector: CSSSelector{"h1", 0}}
	if rule.Matches(h1, st) != true {
		t.Error("Did not match simple h1 selector")
	}
	rule.Selector.Selector = "div h1"
	if rule.Matches(h1, st) != true {
		t.Error("Did not match parent div selector")
	}
	rule.Selector.Selector = ".header h1"
	if rule.Matches(h1, st) != true {
		t.Error("Did not match parent div selector by class")
	}
	rule.Selector.Selector = ".site h1"
	if rule.Matches(h1, st) != true {
		t.Error("Ancestor selector did not select grandparent")
	}
	rule.Selector.Selector = "div div h1"
	if rule.Matches(h1, st) != true {
		t.Error("Did not match multilevel selector")
	}

	ul := headerdiv.NextSibling.NextSibling.NextSibling.NextSibling
	li1 := ul.FirstChild.NextSibling
	ul2 := li1.FirstChild.NextSibling
	li2 := ul2.FirstChild.NextSibling

	rule.Selector.Selector = "UL LI LI"
	if rule.Matches(li2, st) != true {
		t.Error("Did not match deeper multilevel selector")
	}
}

/*
func TestCSS1LinkSelector(t *testing.T) {
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
	//    <a name="label">I am a label</a>
	//  </div>
	// [...]
	//</div>
	var st State
	rule := StyleRule{Selector: CSSSelector{":link", 0}}
	if rule.Matches(h1, st) != false {
		t.Error("h1 is not a link")
	}
	if rule.Matches(h1.FirstChild, st) != true {
		t.Error("h1's child should be a link")
	}
	// h1.NextSibling = whitespace, NextSibling.NextSibling = <a class="extra" ...
	if rule.Matches(h1.NextSibling.NextSibling, st) != true {
		t.Errorf("h1's sibling should be a link")
	}
	if rule.Matches(h1.NextSibling.NextSibling.NextSibling.NextSibling, st) != false {
		t.Error("h1's second sibling is a named anchor, not a link")
	}
}
*/
