package css

import (
	"fmt"
	"sort"
	"testing"
)

func TestCSSSourceSorting(t *testing.T) {
	var vals []StyleRule = []StyleRule{
		StyleRule{
			Src: UserAgentSrc,
		}, StyleRule{
			Src: UserSrc,
		}, StyleRule{
			Src: AuthorSrc,
		}}

	sort.Sort(byCSSPrecedence(vals))

	if vals[0].Src != AuthorSrc {
		fmt.Printf("Unexpected Source: %s, not author at index 0\n", vals[0].Src)
		t.Fail()
	}
	if vals[1].Src != UserSrc {
		fmt.Printf("Unexpected Source: %s, not UserSrc at index 1\n", vals[1].Src)
		t.Fail()
	}
	if vals[2].Src != UserAgentSrc {
		fmt.Printf("Unexpected Source: %s, not UserAgent at index 2\n", vals[2].Src)
		t.Fail()
	}
}

func TestSpecificitySorting(t *testing.T) {
	var vals []StyleRule = []StyleRule{
		StyleRule{Selector: CSSSelector{"foo", 0}, Src: AuthorSrc},
		StyleRule{Selector: CSSSelector{"#foo", 0}, Src: AuthorSrc},
		StyleRule{Selector: CSSSelector{"hello hello", 0}, Src: AuthorSrc},
		StyleRule{Selector: CSSSelector{"foo .hello", 0}, Src: AuthorSrc},
	}

	sort.Sort(byCSSPrecedence(vals))
	if vals[0].Selector.Selector != "#foo" {
		t.Errorf("Unexpected Selector at index 0: %s, not #foo\n", vals[0].Selector)

	}
	if vals[1].Selector.Selector != "foo .hello" {
		t.Errorf("Unexpected Selector at index 1: %s, not foo .hello\n", vals[1].Selector)
	}
	if vals[2].Selector.Selector != "hello hello" {
		t.Errorf("Unexpected Selector at index 2: %s, not hello hello\n", vals[2].Selector)

	}
	if vals[3].Selector.Selector != "foo" {
		t.Errorf("Unexpected Selector at index 3: %s, not foo\n", vals[3].Selector)

	}

	/*type StyleRule struct {
		Selector CSSSelector
		Name     StyleAttribute
		Value    StyleValue
		Src      StyleSource
	}*/

}
