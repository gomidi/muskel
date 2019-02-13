package muskel

import (
	"reflect"
	"strings"
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

type patternGetter struct {
	patterns map[string]*PatternDefinition
}

func (p *patternGetter) add(pd *PatternDefinition) {
	p.patterns[pd.Name] = pd
}

func (p *patternGetter) GetPatternDefinition(name string) *PatternDefinition {
	return p.patterns[name]
}

func newPatternGetter() *patternGetter {
	return &patternGetter{
		patterns: map[string]*PatternDefinition{},
	}
}

func TestCall(t *testing.T) {
	tests := []struct {
		definitions string
		call        string
		expected    string
		err         bool
	}{
		{
			"$pattA: 2a 2&b | 1c'",
			"$pattA",
			"2a 2&b|1c'",
			false,
		},
		{
			`$pattx: 2a 2&b | 1c'
			 $pattB: $pattx | $pattx`,
			"$pattB",
			"2a 2&b|1c'|2a 2&b|1c'",
			false,
		},
		{
			`$pattx: 2a 2&b | 1c'`,
			"$pattx[1:]",
			"2&b|1c'",
			false,
		},
		{
			`$pattx: 2a 2&b | 1c'`,
			"$pattx[2:]",
			"1c'",
			false,
		},
		{
			`$pattx: 2a 2&b | 1c'`,
			"$pattx[:2]",
			"2a 2&b",
			false,
		},
		{
			`$pattx: 2a 2&b | 1c'`,
			"$pattx[1:2]",
			"2&b",
			false,
		},
		{
			`$pattx: 2a 2&b | 1c'`,
			"$pattx/1 : d/",
			"1a 2&b|1d",
			false,
		},
		{
			`$pattx: 2a &b | 1c'`,
			"$pattx[:2]/: ;/",
			"2a ;b",
			false,
		},
		{
			`$pattx: 2#1 2&b | 1c'`,
			"$pattx(g)",
			"2g 2&b|1c'",
			false,
		},
		{
			`$pattx: 2#1 #2b | 1c'`,
			"$pattx(%,&)",
			"2% &b|1c'",
			false,
		},
		{
			`$pattx: 2#1 2&b | 1#1`,
			"$pattx(g'+)",
			"2g'+ 2&b|1g'+",
			false,
		},
		{
			`$pattx: 2#1 2&b | 1c'
			 $pattB: $pattx(#1) | $pattx(#2)`,
			"$pattB(g,f')",
			"2g 2&b|1c'|2f' 2&b|1c'",
			false,
		},
	}

	for i, test := range tests {

		get := newPatternGetter()

		var err error
		defs := strings.Split(test.definitions, "\n")

		for _, df := range defs {
			var pd PatternDefinition
			err = pd.Parse(strings.TrimSpace(df)[1:])
			if err != nil {
				break
			}
			get.add(&pd)
		}

		var pc PatternCall
		if err == nil {
			err = pc.Parse(test.call[1:])
		}
		/*
			if err != nil && !test.err {
				t.Errorf("[%v] Parse(%q) returned unexpected error: %s", i, test.input, err)
			}

			if err == nil && test.err {
				t.Errorf("[%v] Parse(%q) returned no error, but we expected one", i, test.input)
			}
		*/
		var res string

		var pdd = get.GetPatternDefinition(pc.Name)
		if pdd != nil {
			res, err = pdd.Call(&pc, get.GetPatternDefinition)
		}

		if res != test.expected {
			t.Errorf("[%v] Call(%q) resulted in %#v // expected: %#v", i, test.call, res, test.expected)
		}
	}
}

func TestParseDefinition(t *testing.T) {
	tests := []struct {
		input    string
		expected PatternDefinition
		err      bool
	}{
		// the prefix $ has already been handled before
		{"test : 2a' &kd", PatternDefinition{Name: "test", Original: "2a' &kd"}, false},
		{"test  2a' &kd", PatternDefinition{}, true},
		{"test : 2#1 &kd", PatternDefinition{Name: "test", Original: "2#1 &kd", NumParams: 1}, false},
		{"test : 2#1 #2kd", PatternDefinition{Name: "test", Original: "2#1 #2kd", NumParams: 2}, false},
		{"test : 2#1 &#1", PatternDefinition{Name: "test", Original: "2#1 &#1", NumParams: 1}, false},
	}

	for i, test := range tests {
		var pd PatternDefinition

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

func TestParseCall(t *testing.T) {

	tests := []struct {
		input    string
		expected PatternCall
		err      bool
	}{
		// the prefix $ has already been handled before
		{"test", PatternCall{Name: "test", Slice: [2]int{-1, -1}}, false},
		{"test[:2]", PatternCall{Name: "test", Slice: [2]int{0, 2}}, false},
		{"test(a,b)", PatternCall{Name: "test", Slice: [2]int{-1, -1}, Params: []string{"a", "b"}}, false},
		{"test/a : c/", PatternCall{Name: "test", Slice: [2]int{-1, -1}, Replacements: []string{"a", ":", "c"}}, false},
		{"test(a,b)[1:]/d : f/", PatternCall{
			Name:         "test",
			Slice:        [2]int{1, -1},
			Replacements: []string{"d", ":", "f"},
			Params:       []string{"a", "b"},
		}, false},
		{"\\first(a,2)[1:]/d : f/", PatternCall{
			Name:         "first",
			Slice:        [2]int{1, -1},
			Replacements: []string{"d", ":", "f"},
			Params:       []string{"a", "2"},
			SyncFirst:    true,
		}, false},
	}

	for i, test := range tests {
		var pc PatternCall

		err := pc.Parse(test.input)

		if err != nil && !test.err {
			t.Errorf("[%v] Parse(%q) returned unexpected error: %s", i, test.input, err)
		}

		if err == nil && test.err {
			t.Errorf("[%v] Parse(%q) returned no error, but we expected one", i, test.input)
		}

		if !reflect.DeepEqual(pc, test.expected) {
			t.Errorf("[%v] Parse(%q) resulted in %#v // expected: %#v", i, test.input, pc, test.expected)
		}
	}

}
