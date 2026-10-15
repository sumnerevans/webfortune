package ogimage

import (
	"reflect"
	"strings"
	"testing"
)

func TestWrapColumnsPreservesFormatting(t *testing.T) {
	longLine := strings.Repeat("x", 72) + "y"
	input := "  first   line\n\n    /\\_/\\\n\tcat\n" + longLine
	want := []string{
		"  first   line",
		"",
		"    /\\_/\\",
		"        cat",
		strings.Repeat("x", 72),
		"y",
	}
	if got := wrapColumns(input, 72); !reflect.DeepEqual(got, want) {
		t.Errorf("wrapColumns() = %#v, want %#v", got, want)
	}
}
