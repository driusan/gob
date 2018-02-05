package css

import (
	"io"
	"net/url"
	"reflect"
	"testing"
)

type noopURLer struct{}

func (l noopURLer) GetURL(u *url.URL) (body io.ReadCloser, resp int, err error) {
	return nil, 404, nil
}

func TestParseStylesheet(t *testing.T) {
	tests := []struct {
		Stylesheet string
		Expected   Stylesheet
	}{
		{
			"div { color: red }",
			Stylesheet{StyleRule{Selector: CSSSelector{"div", 0}, Name: "color", Value: StyleValue{"red", false}}},
		},
		{
			"html, body { color: red }",
			Stylesheet{
				StyleRule{Selector: CSSSelector{"html", 0}, Name: "color", Value: StyleValue{"red", false}},
				StyleRule{Selector: CSSSelector{"body", 1}, Name: "color", Value: StyleValue{"red", false}},
			},
		},
		{
			".header a { color: red }",
			Stylesheet{StyleRule{Selector: CSSSelector{".header a", 0}, Name: "color", Value: StyleValue{"red", false}}},
		},
		{
			"a { color: red; display: inline; }",
			Stylesheet{
				StyleRule{Selector: CSSSelector{"a", 0}, Name: "color", Value: StyleValue{"red", false}},
				StyleRule{Selector: CSSSelector{"a", 1}, Name: "display", Value: StyleValue{"inline", false}},
			},
		},
		{
			"p { margin: 1.12em 0; }",
			Stylesheet{
				// This gets expanded from the shorthand when applying to an element
				StyleRule{Selector: CSSSelector{"p", 0}, Name: "margin", Value: StyleValue{"1.12em 0", false}},
			},
		},
	}
	for i, tc := range tests {
		style, _ := ParseStylesheet(tc.Stylesheet, 0, noopURLer{}, nil, 0)
		if !reflect.DeepEqual(style, tc.Expected) {
			t.Errorf("Case %d: got %v want %v", i, style, tc.Expected)
		}
	}
}
