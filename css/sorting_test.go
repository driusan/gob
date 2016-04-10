package css

import (
	"fmt"
	"sort"
	"testing"
)

func TestCSSSourceSorting(t *testing.T) {
	var vals []StyleRule = []StyleRule{
		StyleRule{
			src: UserAgentSrc,
		}, StyleRule{
			src: UserSrc,
		}, StyleRule{
			src: AuthorSrc,
		}}

	sort.Sort(byCSSPrecedence(vals))

	if vals[0].src != AuthorSrc {
		fmt.Printf("Unexpected Source: %s, not author at index 0\n", vals[0].src)
		t.Fail()
	}
	if vals[1].src != UserSrc {
		fmt.Printf("Unexpected Source: %s, not UserSrc at index 1\n", vals[1].src)
		t.Fail()
	}
	if vals[2].src != UserAgentSrc {
		fmt.Printf("Unexpected Source: %s, not UserAgent at index 2\n", vals[2].src)
		t.Fail()
	}
}
