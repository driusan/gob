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
