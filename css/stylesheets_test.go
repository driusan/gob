package css

import (
	"testing"
)

var basiccsscontent string = `
hello, abc[foo], what, a > b {
	display: none;
	foo: bar;
}
`

var multiplecsscontent string = `
hello {
	display: what; }

goodbye {
	display: yay;
}
`

func assertSelector(t *testing.T, sty StyleRule, expected string) {
	t.Helper()
	if s := sty.Selector.Selector; s != expected {
		t.Errorf("Incorrect selector for item. Got %v expected %v.", s, expected)
	}
}
func assertName(t *testing.T, sty StyleRule, expected StyleAttribute) {
	t.Helper()
	if s := sty.Name; s != expected {
		t.Errorf("Incorrect property name for item. Got %v expected %v.", s, expected)
	}
}

func assertValue(t *testing.T, sty StyleRule, expected StyleValue) {
	t.Helper()
	if s := sty.Value; s != expected {
		t.Errorf("Incorrect property value for item. Got %v expected %v.", s, expected)
	}
}

// Tests basic usage of CSS parser
func TestCSSParser(t *testing.T) {
	sty, _ := ParseStylesheet(basiccsscontent, AuthorSrc, noopURLer{}, nil, 0)
	// 4 selectors to match, 2 attributes per selector = 6 elements in the stylesheet.
	if len(sty) != 8 {
		t.Fatalf("Incorrect number of elements in stylesheet. Expected 8 got %v: %s", len(sty), sty)
	}
	assertSelector(t, sty[0], "hello")
	assertSelector(t, sty[1], "abc[foo]")
	assertSelector(t, sty[2], "what")
	assertSelector(t, sty[3], "a>b")
	assertSelector(t, sty[4], "hello")
	assertSelector(t, sty[5], "abc[foo]")
	assertSelector(t, sty[6], "what")
	assertSelector(t, sty[7], "a>b")

	assertName(t, sty[0], "display")
	assertName(t, sty[1], "display")
	assertName(t, sty[2], "display")
	assertName(t, sty[3], "display")
	assertName(t, sty[4], "foo")
	assertName(t, sty[5], "foo")
	assertName(t, sty[6], "foo")
	assertName(t, sty[7], "foo")

	assertValue(t, sty[0], StyleValue{"none", false})
	assertValue(t, sty[1], StyleValue{"none", false})
	assertValue(t, sty[2], StyleValue{"none", false})
	assertValue(t, sty[3], StyleValue{"none", false})
	assertValue(t, sty[4], StyleValue{"bar", false})
	assertValue(t, sty[5], StyleValue{"bar", false})
	assertValue(t, sty[6], StyleValue{"bar", false})
	assertValue(t, sty[7], StyleValue{"bar", false})

	for i, s := range sty {
		if s.Src != AuthorSrc {
			t.Errorf("Incorrect source for element %d", i)
		}
	}
}

// Tests that when there are multiple blocks in a CSS file, it's parsed correctly
func TestMultipleCSSBlocks(t *testing.T) {
	sty, _ := ParseStylesheet(multiplecsscontent, AuthorSrc, noopURLer{}, nil, 0)
	if len(sty) != 2 {
		t.Fatalf("Incorrect number of elements in stylesheet. Expected 4 got %v: %s", len(sty), sty)
	}
	assertSelector(t, sty[0], "hello")
	assertSelector(t, sty[1], "goodbye")

	assertName(t, sty[0], "display")
	assertName(t, sty[1], "display")

	assertValue(t, sty[0], StyleValue{"what", false})
	assertValue(t, sty[1], StyleValue{"yay", false})

	for i, s := range sty {
		if s.Src != AuthorSrc {
			t.Errorf("Incorrect source for element %d", i)
		}
	}

}

// Tests that all types of basic selectors in CSS Level 1, 2, and 3 are parsed
// correctly.
func TestAllBasicSelectorTypes(t *testing.T) {
	sty, _ := ParseStylesheet(
		`*, E, E[foo], E[foo="bar"], E[foo~="bar"], E[foo^="bar"], E[foo$="bar"], /* 7 */
E[foo*="bar"], E[foo|="bar"], E:root, E:nth-child(1), E:nth-last-child(2), /* 12 */
E:nth-of-type(1), E:nth-last-of-type(3), /* 14 */
E:first-child, E:last-child, E:first-of-type, E:last-of-type, E:only-child, /* 19 */
E:only-of-type, E:empty, E:link, E:visited, E:active, E:hover, E:focus, /* 26 */
E:target, E:lang(fr), E:enabled, E:disabled, E:checked, E:first-line, /* 32 */
E::first-line, E:first-letter, E::first-letter, E:before, E::before, /* 37 */
E:after,E::after, E.warning, .warning, E#myid, #myid, E:not(F), E	 F, /* 45 */
E > F, E + F, E ~ F /* 48 */ {display: none }`,
		AuthorSrc,
		noopURLer{},
		nil,
		0,
	)
	if len(sty) != 48 {
		t.Fatalf("Incorrect number of elements. Expected 48 got %d", len(sty))
	}

	assertSelector(t, sty[0], "*")
	assertSelector(t, sty[1], "E")
	assertSelector(t, sty[2], "E[foo]")
	assertSelector(t, sty[3], `E[foo="bar"]`)
	assertSelector(t, sty[4], `E[foo~="bar"]`)
	assertSelector(t, sty[5], `E[foo^="bar"]`)
	assertSelector(t, sty[6], `E[foo$="bar"]`)
	assertSelector(t, sty[7], `E[foo*="bar"]`)
	assertSelector(t, sty[8], `E[foo|="bar"]`)
	assertSelector(t, sty[9], "E:root")
	assertSelector(t, sty[10], `E:nth-child(1)`)
	assertSelector(t, sty[11], `E:nth-last-child(2)`)
	assertSelector(t, sty[12], `E:nth-of-type(1)`)
	assertSelector(t, sty[13], `E:nth-last-of-type(3)`)
	assertSelector(t, sty[14], `E:first-child`)
	assertSelector(t, sty[15], `E:last-child`)
	assertSelector(t, sty[16], `E:first-of-type`)
	assertSelector(t, sty[17], `E:last-of-type`)
	assertSelector(t, sty[18], `E:only-child`)
	assertSelector(t, sty[19], `E:only-of-type`)
	assertSelector(t, sty[20], `E:empty`)
	assertSelector(t, sty[21], `E:link`)
	assertSelector(t, sty[22], `E:visited`)
	assertSelector(t, sty[23], `E:active`)
	assertSelector(t, sty[24], `E:hover`)
	assertSelector(t, sty[25], `E:focus`)
	assertSelector(t, sty[26], `E:target`)
	assertSelector(t, sty[27], `E:lang(fr)`)
	assertSelector(t, sty[28], `E:enabled`)
	assertSelector(t, sty[29], `E:disabled`)
	assertSelector(t, sty[30], `E:checked`)
	assertSelector(t, sty[31], `E:first-line`)
	assertSelector(t, sty[32], `E::first-line`)
	assertSelector(t, sty[33], `E:first-letter`)
	assertSelector(t, sty[34], `E::first-letter`)
	assertSelector(t, sty[35], `E:before`)
	assertSelector(t, sty[36], `E::before`)
	assertSelector(t, sty[37], `E:after`)
	assertSelector(t, sty[38], `E::after`)
	assertSelector(t, sty[39], `E.warning`)
	assertSelector(t, sty[40], `.warning`)
	assertSelector(t, sty[41], `E#myid`)
	assertSelector(t, sty[42], `#myid`)
	assertSelector(t, sty[43], `E:not(F)`)
	assertSelector(t, sty[44], "E F")
	assertSelector(t, sty[45], `E>F`)
	assertSelector(t, sty[46], `E+F`)
	assertSelector(t, sty[47], `E~F`)

	for _, el := range sty {
		assertName(t, el, "display")
	}
}

// Tests that different types of units in values get parsed correctly.
func TestCSSUnits(t *testing.T) {
	sty, _ := ParseStylesheet(`
a {
	height: 300px;
	width: 5%;
	foo: 2pt;
	opacity: 3;
	other: 3.0;
	hello: "string" important;
	abc: url("yay");
	multi: fff f		 fff;
	c: rgb(255, 255, 255);
	margin-bottom: -2cm;
	alpha: rgba(0, 255, 255, 255);
}`, AuthorSrc, noopURLer{}, nil, 0)
	assertName(t, sty[0], "height")
	assertName(t, sty[1], "width")
	assertName(t, sty[2], "foo")
	assertName(t, sty[3], "opacity")
	assertName(t, sty[4], "other")
	assertName(t, sty[5], "hello")
	assertName(t, sty[6], "abc")
	assertName(t, sty[7], "multi")
	assertName(t, sty[8], "c")
	assertName(t, sty[9], "margin-bottom")
	assertName(t, sty[10], "alpha")

	assertValue(t, sty[0], StyleValue{"300px", false})
	assertValue(t, sty[1], StyleValue{"5%", false})
	assertValue(t, sty[2], StyleValue{"2pt", false})
	assertValue(t, sty[3], StyleValue{"3", false})
	assertValue(t, sty[4], StyleValue{"3.0", false})
	assertValue(t, sty[5], StyleValue{"\"string\"", true})
	assertValue(t, sty[6], StyleValue{"url(\"yay\")", false})
	assertValue(t, sty[7], StyleValue{"fff f fff", false})
	assertValue(t, sty[8], StyleValue{"rgb(255, 255, 255)", false})
	assertValue(t, sty[9], StyleValue{"-2cm", false})
	assertValue(t, sty[10], StyleValue{"rgba(0, 255, 255, 255)", false})
}

// Tests basic usage of CSS parser
func TestMultpleSelectors(t *testing.T) {
	sty, _ := ParseStylesheet(`em, ul li li { color: green }`, AuthorSrc, noopURLer{}, nil, 0)
	assertSelector(t, sty[0], `em`)
	// BUG(driusan): The final tie break is not implemented
	assertSelector(t, sty[1], `ul li li`)
}

// There was a regression at some point where if the first selector was an ID selector, it
// wouldn't get parsed. This ensures that it's fixed.
func TestIdRegression(t *testing.T) {
	sty, _ := ParseStylesheet(`#one { display: what; } #two { display: yay; }`, AuthorSrc, noopURLer{}, nil, 0)
	assertSelector(t, sty[0], `#one`)
	assertSelector(t, sty[1], `#two`)
}

func TestInvalidIdentifier(t *testing.T) {
	sty, _ := ParseStylesheet(`.one {color: green;}
	.1 {color: red;}
	.a1 {color: green;}
	P.two {color: purple;}`, AuthorSrc, noopURLer{}, nil, 0)
	if len(sty) != 3 {
		t.Fatalf("Parsed invalid identifier which starts with digit: got %v", sty)
	}
	assertSelector(t, sty[0], ".one")
	assertSelector(t, sty[1], ".a1")
	assertSelector(t, sty[2], "P.two")
}
