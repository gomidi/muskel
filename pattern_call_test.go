package muskel

import (
	"reflect"
	"strings"
	"testing"
)

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

func TestParsePattern(t *testing.T) {
	//parsePattern(data string, positionIn32th uint, getter func(name string) *PatternDefinition) error

	tests := []struct {
		definitions string
		call        string
		expected    string
		err         bool
	}{
		/*
		 */
		{
			"$pattA: 2a 2&b 3c'",
			"$pattA",
			"2a 2&b 3c'",
			false,
		},
		{
			"$pattA: 12a 12&b 13c'",
			"$pattA",
			"12a 12&b 13c'",
			false,
		},
		{
			"$pattA: 2a 2&#1 #2c'",
			"$pattA(e',3&)",
			"2a 2&e' 3&c'",
			false,
		},
		{
			"$pattA: 2a 2&#1 #2c'",
			"$pattA(e',3&)[1:]",
			"2&e' 3&c'",
			false,
		},
		{
			"$pattA: 1b 2a 2&#1 #2c'",
			"$pattA(e',3&)[:3]/1& d :/",
			"1&b 2d 2&e'",
			false,
		},
		{
			"$pattA: 1b 2a\n$pattB: 1$pattA",
			"$pattB",
			"1b 2a",
			false,
		},
		{
			"$pattA: 1b #1a\n$pattB: 1$pattA(3)",
			"$pattB",
			"1b 3a",
			false,
		},
		{
			"$pattA: 1b #1a\n$pattB: 1$pattA(#1)",
			"$pattB(3&)",
			"1b 3&a",
			false,
		},
		{
			"$pattA: 1#2 #1a\n$pattB: 1$pattA(#1,c)",
			"$pattB(3&)",
			"1c 3&a",
			false,
		},
		{
			"$pattA: 1b 2a\n$pattB: 2$pattA",
			"$pattB",
			"2b 3a",
			false,
		},
		{
			"$pattA: 1b 2a\n$pattB: 2&$pattA",
			"$pattB",
			"2&b 3&a",
			false,
		},
		{
			"$pattA: 1&b 2&a\n$pattB: 2&$pattA",
			"$pattB",
			"3b 4a",
			false,
		},
		{
			"$pattA: 1&b 2&a\n$pattB: 2&$pattA\n$pattC: 1$pattB",
			"$pattC",
			"3b 4a",
			false,
		},
		{
			"$pattA: 1&b 2&a\n$pattB: 2&$pattA\n$pattC: 1&$pattB",
			"$pattC",
			"3&b 4&a",
			false,
		},
		{
			"$pattA: 1&b 2&a\n$pattB: 2&$\\pattA",
			"$pattB",
			"2&b 3&a",
			false,
		},
		{
			"$pattA: 1&b 2&a\n$pattB: 2&$pattA\n$pattC: 1&$\\pattB",
			"$pattC",
			"2b 3a",
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

		var pc = &PatternCall{}
		pc.getter = get.GetPatternDefinition
		if err == nil {
			err = pc.parsePattern(test.call[1:], 0)
			//			err = pc.Parse(test.call[1:])
		}

		if err != nil && !test.err {
			t.Errorf("[%v] parsePattern(%q) returned unexpected error: %s", i, test.call, err)
		}

		if err == nil && test.err {
			t.Errorf("[%v] parsePattern(%q) returned no error, but we expected one", i, test.call)
		}

		var res = strings.TrimSpace(pc.String())

		if res != test.expected {
			t.Errorf("[%v] parsePattern(%q) resulted in %#v // expected: %#v", i, test.call, res, test.expected)
		}
	}

}

func TestCall(t *testing.T) {
	t.Skip()
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
			res, err = pdd.Call(pc.Params...)
		}

		if res != test.expected {
			t.Errorf("[%v] Call(%q) resulted in %#v // expected: %#v", i, test.call, res, test.expected)
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
