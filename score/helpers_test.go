package score

import "testing"

func TestPad(t *testing.T) {

	tests := []struct {
		s        string
		length   int
		expected string
	}{
		{"abc", 3, "abc"},
		{"abc", 4, "abc "},
		{"abc", 5, "abc  "},
	}

	for _, test := range tests {
		got := Pad(test.s, test.length)

		if got != test.expected {
			t.Errorf("pad(%#v, %v) = %#v // expected %#v", test.s, test.length, got, test.expected)
		}
	}
}
