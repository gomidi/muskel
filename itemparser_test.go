package muskel

import (
	"reflect"
	"testing"
)

func TestParseItem(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"_", Rest{}},
		{"a'", Note{letter: "a", octave: 2}},
		{"a'++", Note{letter: "a", octave: 2, dynamic: "++"}},
		{"c#'--~", Note{letter: "c", octave: 2, dynamic: "--", glissandoStart: true, augmenter: "#"}},
		{"kd", DrumNote{name: "kd"}},
		{"kd+", DrumNote{name: "kd", dynamic: "+"}},
		{"=", Hold},
		{"./.", RepeatLastBar{}},
		{"...", RepeatLastBarUntilChange{}},
		{".4.", RepeatLastNBarsUntilChange(4)},
		{"%", RepeatLastEvent{}},
		{"{kd,sn}2", NTuple{items: []interface{}{DrumNote{name: "kd"}, DrumNote{name: "sn"}}, endPos: 8}},
		{"\"hiho\"", Lyric("hiho")},
		{"Sa'", Note{letter: "a", octave: 2}},
		{"Za'", Note{letter: "a", octave: 2}},
		{"?[30]kd", &RandomProbability{prob: 30, item: DrumNote{name: "kd"}, itemOriginalData: "kd"}},
		{"?(kd,sn)", &RandomChooser{alternatives: []interface{}{DrumNote{name: "kd"}, DrumNote{name: "sn"}}, alternativesOriginalData: []string{"kd", "sn"}}},
		/*
			&PatternCall{Name:"pa", Params:[]string(nil), Replacements:[]string(nil), Slice:[2]int{-1, -1}, SyncFirst:false, result:"2c#", Events:[][]*muskel.positionedEvent{[]*muskel.positionedEvent{(*muskel.positionedEvent)(0xc00000a800)}}}
			Events:[][]*muskel.positionedEvent{[]*muskel.positionedEvent{(*muskel.positionedEvent)(0xc00000a800)}
		*/
		/*
			{"$pa",
				&PatternCall{
					Name:   "pa",
					Slice:  [2]int{-1, -1},
					result: "2a'",
					Events: []*positionedEvent{
						&positionedEvent{
							position:     "2",
							item:         Note{letter: "a", octave: 2},
							originalData: "a'",
						},
					},
				},
			},
			{"$pa/3/",
				&PatternCall{
					Name:         "pa",
					Replacements: []string{"3"},
					Slice:        [2]int{-1, -1},
					result:       "3a'",
					Events: []*positionedEvent{
						&positionedEvent{
							position:     "3",
							item:         Note{letter: "a", octave: 2},
							originalData: "a'",
						},
					},
				},
			},
			{"$pb(a')/: : 3&/",
				&PatternCall{
					Name:         "pb",
					Replacements: []string{":", ":", "3&"},
					Slice:        [2]int{-1, -1},
					result:       "1a' 2a' 3&a'",
					Params:       []string{"a'"},
					Events: []*positionedEvent{
						&positionedEvent{
							position:     "1",
							item:         Note{letter: "a", octave: 2},
							originalData: "a'",
						},
						&positionedEvent{
							position:     "2",
							item:         Note{letter: "a", octave: 2},
							originalData: "a'",
						},
						&positionedEvent{
							position:     "3&",
							item:         Note{letter: "a", octave: 2},
							originalData: "a'",
						},
					},
				},
			},
			{"$pb(a')[1:]/: 3&/",
				&PatternCall{
					Name:         "pb",
					Replacements: []string{":", "3&"},
					Slice:        [2]int{1, -1},
					result:       "2a' 3&a'",
					Params:       []string{"a'"},
					Events: []*positionedEvent{
						&positionedEvent{
							position:     "2",
							item:         Note{letter: "a", octave: 2},
							originalData: "a'",
						},
						&positionedEvent{
							position:     "3&",
							item:         Note{letter: "a", octave: 2},
							originalData: "a'",
						},
					},
				},
			},
		*/
		//	{"O[30]/a/b", &OSCMessage{}},
	}

	for i, test := range tests {
		var parser itemParser

		parser.GetDefinition = func(name string) *PatternDefinition {
			if name == "pa" {
				return &PatternDefinition{
					Name:      "pa",
					NumParams: 0,
					Original:  "2a'",
				}
			}
			if name == "pb" {
				return &PatternDefinition{
					Name:      "pb",
					NumParams: 1,
					Original:  "1#1 2a' 2&a'",
				}
			}
			return nil
		}

		result, err := parser.parseItem(test.input, 0)

		if err != nil {
			t.Errorf("[%v] could not parse item for %#v: %s", i, test.input, err.Error())
			continue
		}

		if !reflect.DeepEqual(result, test.expected) {
			t.Errorf("[%v] parseItem(%#v) = \n%#v \n// expected\n%#v", i, test.input, result, test.expected)
		}
	}
}
