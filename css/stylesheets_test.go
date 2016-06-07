package css

import (
	"testing"
)

var csscontent string = `
hello, abc[foo], what, a > b {
	display: none;
	foo: bar;
}
`

func assertSelector(t *testing.T, sty StyleRule, expected CSSSelector) {
	if s := sty.Selector; s != expected {
		t.Errorf("Incorrect selector for item 0. Got %v expected %v.", s, expected)
	}
}
func assertName(t *testing.T, sty StyleRule, expected StyleAttribute) {
	if s := sty.Name; s != expected {
		t.Errorf("Incorrect selector for item 0. Got %v expected %v.", s, expected)
	}
}
func TestCSSParser(t *testing.T) {
	sty := ParseStylesheet(csscontent, AuthorSrc)
	// 4 selectors to match, 2 attributes per selector = 6 elements in the stylesheet.
	if len(sty) != 8 {
		t.Errorf("Incorrect number of elements in stylesheet. Expected 3 got %v: %s", len(sty), sty)
	}
	assertSelector(t, sty[0], "hello")
	assertSelector(t, sty[1], "hello")
	assertSelector(t, sty[2], "abc[foo]")
	assertSelector(t, sty[3], "abc[foo]")
	assertSelector(t, sty[4], "what")
	assertSelector(t, sty[5], "what")
	assertSelector(t, sty[6], "a > b")

	assertName(t, sty[0], "display")
	assertName(t, sty[1], "foo")
	assertName(t, sty[2], "display")
	assertName(t, sty[3], "foo")
	assertName(t, sty[4], "display")
	assertName(t, sty[5], "foo")

	for i, s := range sty {
		if s.Src != AuthorSrc {
			t.Errorf("Incorrect source for element %d", i)
		}
	}
	/*
		type StyleRule struct {
			Selector CSSSelector
			Name     StyleAttribute
			Value    StyleValue
			Src      StyleSource*/
}
