package muskel

import (
	"reflect"
	"testing"
)

func TestParseDefinition(t *testing.T) {
	//	t.Skip()
	tests := []struct {
		input    string
		expected TemplateDefinition
		err      bool
	}{
		// the prefix $ has already been handled before
		{"test : 2a' &kd", TemplateDefinition{Name: "test", Original: "2a' &kd"}, false},
		{"test : a' &kd", TemplateDefinition{Name: "test", Original: "a' &kd"}, false},
		{"test  2a' &kd", TemplateDefinition{}, true},
		{"test~ : 2#1 &kd", TemplateDefinition{Name: "test~", Original: "2#1 &kd", NumParams: 1}, false},
		{"test : 2#1 #2kd", TemplateDefinition{Name: "test", Original: "2#1 #2kd", NumParams: 2}, false},
		{"test~~ : 2#1 &#1", TemplateDefinition{Name: "test~~", Original: "2#1 &#1", NumParams: 1}, false},
	}

	for i, test := range tests {
		var pd TemplateDefinition

		err := pd.Parse(test.input)

		if err != nil && !test.err {
			t.Errorf("[%v] Parse(%q) returned unexpected error: %s", i, test.input, err)
		}

		if err == nil && test.err {
			t.Errorf("[%v] Parse(%q) returned no error, but we expected one", i, test.input)
		}

		if !reflect.DeepEqual(pd, test.expected) {
			t.Errorf("[%v] Parse(%q) resulted in %#v // expected: %#v", i, test.input, pd, test.expected)
		}
	}
}

func TestCallDefinition(t *testing.T) {
	//	t.Skip()
	tests := []struct {
		input    string
		params   []string
		expected string
		err      bool
	}{
		// the prefix $ has already been handled before
		{"test: 2#1 &kd", []string{"a'"}, "2a' &kd", false},
		{"test1: 2#1 &kd", []string{"a'"}, "2a' &kd", false},
		{"test: #1 &kd", []string{"a'"}, "a' &kd", false},
		{"test: 2#1 &kd", []string{"a'", "b"}, "", true},
		{"test: 2#1 &#2", []string{"a'", "b"}, "2a' &b", false},
		{"test: 2#1 &#2 3$x(#1,#2)", []string{"a'", "b"}, "2a' &b 3$x(a',b)", false},
	}

	for i, test := range tests {
		var pd TemplateDefinition

		err := pd.Parse(test.input)

		if err != nil && !test.err {
			t.Errorf("[%v] Parse(%q) returned unexpected error: %s", i, test.input, err)
		}

		res, err := pd.Call(test.params...)

		if err != nil && !test.err {
			t.Errorf("[%v] Call(%v) returned unexpected error: %s", i, test.params, err)
		}

		if err == nil && test.err {
			t.Errorf("[%v] Call(%v) returned no error, but we expected one", i, test.params)
		}

		if res != test.expected {
			t.Errorf("[%v] Parse(%q).Call(%v) resulted in %#v // expected: %#v", i, test.input, test.params, res, test.expected)
		}
	}
}
