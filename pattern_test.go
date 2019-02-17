package muskel

import (
	"reflect"
	"testing"
)

func TestParseFragment(t *testing.T) {
	tests := []struct {
		input    string
		expected patternFragment
	}{
		{"2a'", patternFragment{"2", "a'"}},
		{"a'", patternFragment{"", "a'"}},
		{"2", patternFragment{"2", ""}},
		{"2&", patternFragment{"2&", ""}},
		{"&", patternFragment{"&", ""}},
		{".", patternFragment{".", ""}},
		{";", patternFragment{";", ""}},
		{"2;", patternFragment{"2;", ""}},
		{"2.;", patternFragment{"2.;", ""}},
		{"2.", patternFragment{"2.", ""}},
		{"&.", patternFragment{"&.", ""}},
		{"&;", patternFragment{"&;", ""}},
		{"&.;", patternFragment{"&.;", ""}},

		{"2&a", patternFragment{"2&", "a"}},
		{"&a", patternFragment{"&", "a"}},
		{".a", patternFragment{".", "a"}},
		{";a", patternFragment{";", "a"}},
		{"2;a", patternFragment{"2;", "a"}},
		{"2.;a", patternFragment{"2.;", "a"}},
		{"2.a", patternFragment{"2.", "a"}},
		{"&.a", patternFragment{"&.", "a"}},
		{"&;a", patternFragment{"&;", "a"}},
		{"&.;a", patternFragment{"&.;", "a"}},
	}

	for i, test := range tests {
		var f patternFragment

		f.parse(test.input)

		if !reflect.DeepEqual(f, test.expected) {
			t.Errorf("[%v] parse(%q) resulted in %#v // expected: %#v", i, test.input, f, test.expected)
		}
	}
}
