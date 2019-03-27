package muskel_test

import (
	"reflect"
	"strings"
	"testing"

	"gitlab.com/gomidi/muskel/parser"
	"gitlab.com/gomidi/muskel/template"
)

type templateGetter struct {
	template map[string]*template.Definition
}

func (p *templateGetter) add(pd *template.Definition) {
	p.template[pd.Name] = pd
}

func (p *templateGetter) GetTemplateDefinition(name string) *template.Definition {
	return p.template[name]
}

func newTemplateGetter() *templateGetter {
	return &templateGetter{
		template: map[string]*template.Definition{},
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

	get := newTemplateGetter()

	var pd template.Definition
	err := pd.Parse("templA: 2a 2&b 3c' 5f 7&g 8a 13e")

	get.add(&pd)

	var pc = template.NewTemplateCall(&parser.Items{})
	pc.Getter = get.GetTemplateDefinition
	if err == nil {
		err = pc.ParseTemplate("templA", 0)
	}

	if err != nil {
		t.Errorf("parseTemplate() returned unexpected error: %s", err)
		return
	}

	var positionOfNextBar int
	var newPositionOfNextBar int
	var events []*template.PositionedEvent

	for i, test := range tests {
		if positionOfNextBar >= 0 {
			events, newPositionOfNextBar = template.TemplateEvents(pc.Events).Spread(positionOfNextBar, test.num, test.denom)
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

func TestParseTemplate(t *testing.T) {
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
			"templA: 2a 2&b 3c'",
			"templA",
			"2a 2&b 3c'",
			false,
		},
		{
			"templA: 12a 12&b 13c'",
			"templA",
			"12a 12&b 13c'",
			false,
		},
		{
			"templA: 2a 2&#1 #2c'",
			"templA(e',3&)",
			"2a 2&e' 3&c'",
			false,
		},
		{
			"templA: 2a 2&#1 #2c'",
			"templA(_e'_g'_,3&)",
			"2a 2&_e'_g'_ 3&c'",
			false,
		},
		{
			"templA: 2a 2&_#1_#2_",
			"templA(e',g')",
			"2a 2&_e'_g'_",
			false,
		},
		{
			"templA: 2a 2&#1 #2c'",
			"templA(e',3&)[1:]",
			"2&e' 3&c'",
			false,
		},
		{
			"templA: 1b 2a 2&#1 #2c'",
			"templA(e',3&)[:3]/1&,d,:/",
			"1&b 2d 2&e'",
			false,
		},
		{
			"templA: 1b 2a 3c'",
			"templA/g,d,4/",
			"1g 2d 4c'",
			false,
		},
		{
			"templA: 1b 1&* 2a 2&* 3c'",
			"templA/g,!d,!4/",
			"1g 1&* 2d 2&* 4c'",
			false,
		},
		{
			"templA: 1b 2a\ntemplB: 1templA",
			"templB",
			"1b 2a",
			false,
		},
		{
			"templA: 1b #1a\ntemplB: 1templA(3)",
			"templB",
			"1b 3a",
			false,
		},
		{
			"templA: 1b #1a\ntemplB: 1templA(#1)",
			"templB(3&)",
			"1b 3&a",
			false,
		},
		{
			"templA: 1#2 #1a\ntemplB: 1templA(#1,c)",
			"templB(3&)",
			"1c 3&a",
			false,
		},
		{
			"templA: 1b 2a\ntemplB: 2templA",
			"templB",
			"2b 3a",
			false,
		},
		{
			"templA: 1b 2a\ntemplB: 2&templA",
			"templB",
			"2&b 3&a",
			false,
		},
		{
			"templA: 1&b 2&a\ntemplB: 2&templA",
			"templB",
			"3b 4a",
			false,
		},
		{
			"templA: 1&b 2&a\ntemplB: 2&templA\ntemplC: 1templB",
			"templC",
			"3b 4a",
			false,
		},
		{
			"templA: 1&b 2&a\ntemplB: 2&templA\ntemplC: 1&templB",
			"templC",
			"3&b 4&a",
			false,
		},
		{
			"templA: 1&b 2&a\ntemplB: 2&!templA",
			"templB",
			"2&b 3&a",
			false,
		},
		{
			"templA: 1&b 2&a\ntemplB: 2&templA\ntemplC: 1&!templB",
			"templC",
			"2b 3a",
			false,
		},
	}

	for i, test := range tests {

		get := newTemplateGetter()

		var err error
		defs := strings.Split(test.definitions, "\n")

		for _, df := range defs {
			var pd template.Definition
			err = pd.Parse(strings.TrimSpace(df))
			if err != nil {
				break
			}
			get.add(&pd)
		}

		var pc = template.NewTemplateCall(&parser.Items{})
		pc.Getter = get.GetTemplateDefinition
		if err == nil {
			err = pc.ParseTemplate(test.call, 0)
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
		expected template.Call
		err      bool
	}{
		// the prefix $ has already been handled before
		{"test", template.Call{Name: "test", Slice: [2]int{-1, -1}}, false},
		{"test1", template.Call{Name: "test1", Slice: [2]int{-1, -1}}, false},
		{"te_st", template.Call{Name: "te_st", Slice: [2]int{-1, -1}}, false},
		{"test^2", template.Call{Name: "test", Slice: [2]int{-1, -1}, ScaleMove: 2, ScaleMoveMode: 1}, false},
		{"test^^-11", template.Call{Name: "test", Slice: [2]int{-1, -1}, ScaleMove: -11, ScaleMoveMode: 2}, false},
		{"test~+", template.Call{Name: "test~", Slice: [2]int{-1, -1}, VelocityAdd: "+"}, false},
		{"test++", template.Call{Name: "test", Slice: [2]int{-1, -1}, VelocityAdd: "++"}, false},
		{"test~~=", template.Call{Name: "test~~", Slice: [2]int{-1, -1}, VelocityAdd: "="}, false},
		{"test-", template.Call{Name: "test", Slice: [2]int{-1, -1}, VelocityAdd: "-"}, false},
		{"test--", template.Call{Name: "test", Slice: [2]int{-1, -1}, VelocityAdd: "--"}, false},
		{"test[:2]", template.Call{Name: "test", Slice: [2]int{0, 2}}, false},
		{"test(a,b)", template.Call{Name: "test", Slice: [2]int{-1, -1}, Params: []string{"a", "b"}}, false},
		{"test(_a_b_)", template.Call{Name: "test", Slice: [2]int{-1, -1}, Params: []string{"_a_b_"}}, false},
		{"test(_^1_^3_)", template.Call{Name: "test", Slice: [2]int{-1, -1}, Params: []string{"_^1_^3_"}}, false},
		{"test/a,:,c/", template.Call{Name: "test", Slice: [2]int{-1, -1}, Replacements: []string{"a", ":", "c"}}, false},
		{"test(a,b)[1:]/d,:,f/", template.Call{
			Name:         "test",
			Slice:        [2]int{1, -1},
			Replacements: []string{"d", ":", "f"},
			Params:       []string{"a", "b"},
		}, false},
		{"!first(a,2)[1:]/d,:,f/", template.Call{
			Name:         "first",
			Slice:        [2]int{1, -1},
			Replacements: []string{"d", ":", "f"},
			Params:       []string{"a", "2"},
			SyncFirst:    true,
		}, false},
	}

	for i, test := range tests {
		var pc template.Call

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
