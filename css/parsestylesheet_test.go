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
			Stylesheet{StyleRule{Selector: "div", Name: "color", Value: StyleValue{"red", false}}},
		},
		{
			"html, body { color: red }",
			Stylesheet{
				StyleRule{Selector: "html", Name: "color", Value: StyleValue{"red", false}},
				StyleRule{Selector: "body", Name: "color", Value: StyleValue{"red", false}},
			},
		},
		{
			".header a { color: red }",
			Stylesheet{StyleRule{Selector: ".header a", Name: "color", Value: StyleValue{"red", false}}},
		},
		{
			"a { color: red; display: inline; }",
			Stylesheet{
				StyleRule{Selector: "a", Name: "color", Value: StyleValue{"red", false}},
				StyleRule{Selector: "a", Name: "display", Value: StyleValue{"inline", false}},
			},
		},
		{
			"p { margin: 1.12em 0; }",
			Stylesheet{
				// This gets expanded from the shorthand when applying to an element
				StyleRule{Selector: "p", Name: "margin", Value: StyleValue{"1.12em 0", false}},
			},
		},
	}
	for i, tc := range tests {
		style := ParseStylesheet(tc.Stylesheet, 0, noopURLer{}, nil)
		if !reflect.DeepEqual(style, tc.Expected) {
			t.Errorf("Case %d: got %v want %v", i, style, tc.Expected)
		}
	}
}
