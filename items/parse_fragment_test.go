package items

import (
	"reflect"
	"testing"
)

func TestParseFragment(t *testing.T) {
	//	t.Skip()
	tests := []struct {
		input    string
		expected templateFragment
	}{
		{"2a'", templateFragment{"2", "a'"}},
		{"a'", templateFragment{"", "a'"}},
		{"2", templateFragment{"2", ""}},
		{"2&", templateFragment{"2&", ""}},
		{"&", templateFragment{"&", ""}},
		{".", templateFragment{".", ""}},
		{";", templateFragment{";", ""}},
		{"2;", templateFragment{"2;", ""}},
		{"2.;", templateFragment{"2.;", ""}},
		{"2.", templateFragment{"2.", ""}},
		{"&.", templateFragment{"&.", ""}},
		{"&;", templateFragment{"&;", ""}},
		{"&.;", templateFragment{"&.;", ""}},

		{"2&a", templateFragment{"2&", "a"}},
		{"&a", templateFragment{"&", "a"}},
		{".a", templateFragment{".", "a"}},
		{";a", templateFragment{";", "a"}},
		{"2;a", templateFragment{"2;", "a"}},
		{"2.;a", templateFragment{"2.;", "a"}},
		{"2.a", templateFragment{"2.", "a"}},
		{"&.a", templateFragment{"&.", "a"}},
		{"&;a", templateFragment{"&;", "a"}},
		{"&.;a", templateFragment{"&.;", "a"}},
	}

	for i, test := range tests {
		var f templateFragment

		f.parse(test.input)

		if !reflect.DeepEqual(f, test.expected) {
			t.Errorf("[%v] parse(%q) resulted in %#v // expected: %#v", i, test.input, f, test.expected)
		}
	}
}
