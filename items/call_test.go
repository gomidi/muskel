package items

import (
	"reflect"
	"testing"
)

func TestSimpleParse(t *testing.T) {
	tests := []struct {
		data   string
		name   string
		params []string
		_slice [2]int
	}{
		{"hiho(a,d,e)", "hiho", []string{"a", "d", "e"}, [2]int{-1, -1}},
		{"hi.ho(a,d,e)", "hi.ho", []string{"a", "d", "e"}, [2]int{-1, -1}},
		{"=hiho(a,d,e)", "=hiho", []string{"a", "d", "e"}, [2]int{-1, -1}},
		{"=hi.ho(a,d,e)", "=hi.ho", []string{"a", "d", "e"}, [2]int{-1, -1}},
		{"@hiho(a,d,e)", "@hiho", []string{"a", "d", "e"}, [2]int{-1, -1}},
		{"@hi.ho(a,d,e)", "@hi.ho", []string{"a", "d", "e"}, [2]int{-1, -1}},
		{"=hi.ho(=hu.ho,d,e)", "=hi.ho", []string{"=hu.ho", "d", "e"}, [2]int{-1, -1}},
		{"=.ho(=hu.ho,d,e)", "=.ho", []string{"=hu.ho", "d", "e"}, [2]int{-1, -1}},
		{"=hiho(a,d,e)[:2]", "=hiho", []string{"a", "d", "e"}, [2]int{0, 2}},
		{"=hiho(a,d,(e))[:2]", "=hiho", []string{"a", "d", "(e)"}, [2]int{0, 2}},
		{"=hiho(a,d,(e))[2:]", "=hiho", []string{"a", "d", "(e)"}, [2]int{2, -1}},
		{"=hiho(a,d,(e))[2:4]", "=hiho", []string{"a", "d", "(e)"}, [2]int{2, 4}},
		{"=hiho(a,d,(e f g))[:2]", "=hiho", []string{"a", "d", "(e f g)"}, [2]int{0, 2}},
	}

	//var pp Parser

	for i, test := range tests {
		var p Call
		//p.Parser = &pp

		err := p.SimpleParse(test.data)

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
