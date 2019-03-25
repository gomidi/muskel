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

func TestSpread(t *testing.T) {
	//	t.Skip()
	tests := []struct {
		num   uint8
		denom uint8

		expectedEvents            string
		expectedPositionOfNextBar int
	}{
		{4, 4, "2a 2&b 3c'", 32},
		{3, 4, "1f 3&g", 56},
		{3, 4, "1a", 80},
		{2, 4, "", 96},
		{4, 4, "1e", -1},
	}

	get := newPatternGetter()

	var pd PatternDefinition
	err := pd.Parse("pattA: 2a 2&b 3c' 5f 7&g 8a 13e")

	get.add(&pd)

	var pc = &PatternCall{}
	pc.getter = get.GetPatternDefinition
	if err == nil {
		err = pc.parsePattern("pattA", 0)
	}

	if err != nil {
		t.Errorf("parsePattern() returned unexpected error: %s", err)
		return
	}

	var positionOfNextBar int
	var newPositionOfNextBar int
	var events []*positionedEvent

	for i, test := range tests {
		if positionOfNextBar >= 0 {
			events, newPositionOfNextBar = PatternEvents(pc.Events).Spread(positionOfNextBar, test.num, test.denom)
		}

		if newPositionOfNextBar != test.expectedPositionOfNextBar {
			t.Errorf("[%v] Spread(%v, %v, %v) returned positionOfNextBar %v // expected: %v", i, positionOfNextBar, test.num, test.denom, newPositionOfNextBar, test.expectedPositionOfNextBar)
		}

		positionOfNextBar = newPositionOfNextBar

		var bf strings.Builder

		for _, event := range events {
			bf.WriteString(event.String() + " ")
		}

		res := strings.TrimSpace(bf.String())

		if res != test.expectedEvents {
			t.Errorf("[%v] Spread(%v, %v, %v) returned events = %q // expected %q", i, positionOfNextBar, test.num, test.denom, res, test.expectedEvents)
		}
	}

}

func TestParsePattern(t *testing.T) {
	//	t.Skip()
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
			"pattA: 2a 2&b 3c'",
			"pattA",
			"2a 2&b 3c'",
			false,
		},
		{
			"pattA: 12a 12&b 13c'",
			"pattA",
			"12a 12&b 13c'",
			false,
		},
		{
			"pattA: 2a 2&#1 #2c'",
			"pattA(e',3&)",
			"2a 2&e' 3&c'",
			false,
		},
		{
			"pattA: 2a 2&#1 #2c'",
			"pattA(e',3&)[1:]",
			"2&e' 3&c'",
			false,
		},
		{
			"pattA: 1b 2a 2&#1 #2c'",
			"pattA(e',3&)[:3]/1&,d,:/",
			"1&b 2d 2&e'",
			false,
		},
		{
			"pattA: 1b 2a 3c'",
			"pattA/g,d,4/",
			"1g 2d 4c'",
			false,
		},
		{
			"pattA: 1b 1&* 2a 2&* 3c'",
			"pattA/g,!d,!4/",
			"1g 1&* 2d 2&* 4c'",
			false,
		},
		{
			"pattA: 1b 2a\npattB: 1pattA",
			"pattB",
			"1b 2a",
			false,
		},
		{
			"pattA: 1b #1a\npattB: 1pattA(3)",
			"pattB",
			"1b 3a",
			false,
		},
		{
			"pattA: 1b #1a\npattB: 1pattA(#1)",
			"pattB(3&)",
			"1b 3&a",
			false,
		},
		{
			"pattA: 1#2 #1a\npattB: 1pattA(#1,c)",
			"pattB(3&)",
			"1c 3&a",
			false,
		},
		{
			"pattA: 1b 2a\npattB: 2pattA",
			"pattB",
			"2b 3a",
			false,
		},
		{
			"pattA: 1b 2a\npattB: 2&pattA",
			"pattB",
			"2&b 3&a",
			false,
		},
		{
			"pattA: 1&b 2&a\npattB: 2&pattA",
			"pattB",
			"3b 4a",
			false,
		},
		{
			"pattA: 1&b 2&a\npattB: 2&pattA\npattC: 1pattB",
			"pattC",
			"3b 4a",
			false,
		},
		{
			"pattA: 1&b 2&a\npattB: 2&pattA\npattC: 1&pattB",
			"pattC",
			"3&b 4&a",
			false,
		},
		{
			"pattA: 1&b 2&a\npattB: 2&!pattA",
			"pattB",
			"2&b 3&a",
			false,
		},
		{
			"pattA: 1&b 2&a\npattB: 2&pattA\npattC: 1&!pattB",
			"pattC",
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
			err = pd.Parse(strings.TrimSpace(df))
			if err != nil {
				break
			}
			get.add(&pd)
		}

		var pc = &PatternCall{}
		pc.getter = get.GetPatternDefinition
		if err == nil {
			err = pc.parsePattern(test.call, 0)
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

func TestParseCall(t *testing.T) {
	//	t.Skip()
	tests := []struct {
		input    string
		expected PatternCall
		err      bool
	}{
		// the prefix $ has already been handled before
		{"test", PatternCall{Name: "test", Slice: [2]int{-1, -1}}, false},
		{"test1", PatternCall{Name: "test1", Slice: [2]int{-1, -1}}, false},
		{"te_st", PatternCall{Name: "te_st", Slice: [2]int{-1, -1}}, false},
		{"test^2", PatternCall{Name: "test", Slice: [2]int{-1, -1}, scaleMove: 2, scaleMoveMode: 1}, false},
		{"test^^-11", PatternCall{Name: "test", Slice: [2]int{-1, -1}, scaleMove: -11, scaleMoveMode: 2}, false},
		{"test~+", PatternCall{Name: "test~", Slice: [2]int{-1, -1}, velocityAdd: "+"}, false},
		{"test++", PatternCall{Name: "test", Slice: [2]int{-1, -1}, velocityAdd: "++"}, false},
		{"test~~=", PatternCall{Name: "test~~", Slice: [2]int{-1, -1}, velocityAdd: "="}, false},
		{"test-", PatternCall{Name: "test", Slice: [2]int{-1, -1}, velocityAdd: "-"}, false},
		{"test--", PatternCall{Name: "test", Slice: [2]int{-1, -1}, velocityAdd: "--"}, false},
		{"test[:2]", PatternCall{Name: "test", Slice: [2]int{0, 2}}, false},
		{"test(a,b)", PatternCall{Name: "test", Slice: [2]int{-1, -1}, Params: []string{"a", "b"}}, false},
		{"test/a,:,c/", PatternCall{Name: "test", Slice: [2]int{-1, -1}, Replacements: []string{"a", ":", "c"}}, false},
		{"test(a,b)[1:]/d,:,f/", PatternCall{
			Name:         "test",
			Slice:        [2]int{1, -1},
			Replacements: []string{"d", ":", "f"},
			Params:       []string{"a", "b"},
		}, false},
		{"!first(a,2)[1:]/d,:,f/", PatternCall{
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
