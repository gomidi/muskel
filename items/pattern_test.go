package items

import (
	"reflect"
	"testing"
)

func TestSimpleTemplateParse(t *testing.T) {
	tests := []struct {
		data   string
		name   string
		params []string
	}{
		{"=hiho(a,d,e)", "=hiho", []string{"a", "d", "e"}},
		{"=hi.ho(a,d,e)", "=hi.ho", []string{"a", "d", "e"}},
		{"=hi.ho(=hu.ho,d,e)", "=hi.ho", []string{"=hu.ho", "d", "e"}},
		{"=.ho(=hu.ho,d,e)", "=.ho", []string{"=hu.ho", "d", "e"}},
		{"=hiho(a,d,(e))", "=hiho", []string{"a", "d", "(e)"}},
		{"=hiho(a,d,(e f g))", "=hiho", []string{"a", "d", "(e f g)"}},
	}

	//var pp Parser

	for i, test := range tests {
		var p Pattern
		//p.Parser = &pp

		err := p.Parse(test.data, 0)

		if err != nil {
			t.Errorf("[%v] PatternCall.SimpleParse(%q,0) returns error: %s", i, test.data, err.Error())
		}

		if p.Name != test.name {
			t.Errorf("[%v] PatternCall.SimpleParse(%q,0); name: %q // expected: %q", i, test.data, p.Name, test.name)
		}

		if !reflect.DeepEqual(p.Params, test.params) {
			t.Errorf("[%v] PatternCall.SimpleParse(%q,0); params: %#v // expected: %#v", i, test.data, p.Params, test.params)
		}
	}
}

func TestSimpleTokenParse(t *testing.T) {
	tests := []struct {
		data   string
		name   string
		params []string
		_slice [2]int
	}{
		{".hiho(a,d,e)", ".hiho", []string{"a", "d", "e"}, [2]int{-1, -1}},
		{".hi.ho(a,d,e)", ".hi.ho", []string{"a", "d", "e"}, [2]int{-1, -1}},
	}

	//var pp Parser

	for i, test := range tests {
		var p Token
		//p.Parser = &pp

		err := p.Parse(test.data, 0)

		if err != nil {
			t.Errorf("[%v] PatternCall.SimpleParse(%q,0) returns error: %s", i, test.data, err.Error())
		}

		if p.Name != test.name {
			t.Errorf("[%v] PatternCall.SimpleParse(%q,0); name: %q // expected: %q", i, test.data, p.Name, test.name)
		}

		if !reflect.DeepEqual(p.Params, test.params) {
			t.Errorf("[%v] PatternCall.SimpleParse(%q,0); params: %#v // expected: %#v", i, test.data, p.Params, test.params)
		}

		if !reflect.DeepEqual(p.Slice, test._slice) {
			t.Errorf("[%v] PatternCall.SimpleParse(%q,0); slice: %v // expected: %v", i, test.data, p.Slice, test._slice)
		}
	}
}

func TestSimpleLyricsParse(t *testing.T) {
	tests := []struct {
		data   string
		name   string
		params []string
		_slice [2]int
	}{
		{"@hiho(a,d,e)", "@hiho", []string{"a", "d", "e"}, [2]int{-1, -1}},
		{"@hi.ho(a,d,e)", "@hi.ho", []string{"a", "d", "e"}, [2]int{-1, -1}},
	}

	//var pp Parser

	for i, test := range tests {
		var p LyricsTable
		//p.Parser = &pp

		err := p.Parse(test.data, 0)

		if err != nil {
			t.Errorf("[%v] PatternCall.SimpleParse(%q,0) returns error: %s", i, test.data, err.Error())
		}

		if p.Name != test.name {
			t.Errorf("[%v] PatternCall.SimpleParse(%q,0); name: %q // expected: %q", i, test.data, p.Name, test.name)
		}

		if !reflect.DeepEqual(p.Params, test.params) {
			t.Errorf("[%v] PatternCall.SimpleParse(%q,0); params: %#v // expected: %#v", i, test.data, p.Params, test.params)
		}

		if !reflect.DeepEqual(p.Slice, test._slice) {
			t.Errorf("[%v] PatternCall.SimpleParse(%q,0); slice: %v // expected: %v", i, test.data, p.Slice, test._slice)
		}
	}
}

//func (p *PatternCall) SimpleParse(call string) error {
