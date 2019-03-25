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
		{"*", Rest},
		{"a'", Note{letter: "a", octave: 2, velocity: -1}},
		{"a'++", Note{letter: "a", octave: 2, velocity: 93}},
		{"^2", Note{velocity: -1, scaleNote: 2}},
		{"^2#", Note{velocity: -1, scaleNote: 2, augmenter: "#"}},
		{"^2++", Note{velocity: 93, scaleNote: 2}},
		{"c#'--~", Note{letter: "c", octave: 2, velocity: 33, glissandoStart: true, augmenter: "#"}},
		{":", Hold},
		{"./.", RepeatLastBar{}},
		{"...", RepeatLastBarUntilChange{}},
		{".4.", RepeatLastNBarsUntilChange(4)},
		{"%", RepeatLastEvent{}},
		{"{a,b}2", NTuple{items: []interface{}{Note{letter: "a", octave: 1, velocity: -1}, Note{letter: "b", octave: 1, velocity: -1}}, endPos: 8}},
		{"{a,b,*}2", NTuple{items: []interface{}{Note{letter: "a", octave: 1, velocity: -1}, Note{letter: "b", octave: 1, velocity: -1}, Rest}, endPos: 8}},
		{"\"hiho\"", Lyric("hiho")},
		//		{"Sa'", Note{letter: "a", octave: 2}},
		//		{"Za'", Note{letter: "a", octave: 2}},
		{"?30%a", &RandomProbability{prob: 30, item: Note{letter: "a", octave: 1, velocity: -1}, itemOriginalData: "a"}},
		{"?(a,b)", &RandomChooser{alternatives: []interface{}{Note{letter: "a", octave: 1, velocity: -1}, Note{letter: "b", octave: 1, velocity: -1}}, alternativesOriginalData: []string{"a", "b"}}},
		{"CC(23,100)", MIDICC{controller: 23, value: 100}},
		{"AT(54)", MIDIAftertouch{value: 54}},
		{"PB(-20)", MIDIPitchbend{value: -20}},
		{"PB(8100)", MIDIPitchbend{value: 8100}},
		{"PT(c',80)", MIDIPolyAftertouch{key: 60, value: 80}},
		{"~", GlissandoLinear},
		{"~~", GlissandoExponential},
		//	{"O[30]/a/b", &OSCMessage{}},
	}

	for i, test := range tests {
		var parser itemParser

		parser.GetDefinition = func(name string) *TemplateDefinition {
			if name == "pa" {
				return &TemplateDefinition{
					Name:      "pa",
					NumParams: 0,
					Original:  "2a'",
				}
			}
			if name == "pb" {
				return &TemplateDefinition{
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
