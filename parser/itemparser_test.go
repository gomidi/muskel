package parser

import (
	"reflect"
	"testing"

	. "gitlab.com/gomidi/muskel/items"
	"gitlab.com/gomidi/muskel/template"
)

func TestParseItem(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"*", Rest},
		{"a'", Note{Letter: "a", Octave: 2, Velocity: -1}},
		{"a'++", Note{Letter: "a", Octave: 2, Velocity: 93}},
		{"^2", Note{Velocity: -1, ScaleNote: 2}},
		{"^2#", Note{Velocity: -1, ScaleNote: 2, Augmenter: "#"}},
		{"^2++", Note{Velocity: 93, ScaleNote: 2}},
		{"c#'--~", Note{Letter: "c", Octave: 2, Velocity: 33, GlissandoStart: true, Augmenter: "#"}},
		{":", Hold},
		{"./.", RepeatLastBar{}},
		{"...", RepeatLastBarUntilChange{}},
		{".4.", RepeatLastNBarsUntilChange(4)},
		{"%", RepeatLastEvent{}},
		{"{a,b}2", NTuple{Items: []interface{}{Note{Letter: "a", Octave: 1, Velocity: -1}, Note{Letter: "b", Octave: 1, Velocity: -1}}, EndPos: 8}},
		{"{a,b,*}2", NTuple{Items: []interface{}{Note{Letter: "a", Octave: 1, Velocity: -1}, Note{Letter: "b", Octave: 1, Velocity: -1}, Rest}, EndPos: 8}},
		{"\"hiho\"", Lyric("hiho")},
		//		{"Sa'", Note{letter: "a", octave: 2}},
		//		{"Za'", Note{letter: "a", octave: 2}},
		{"?30%a", &RandomProbability{Prob: 30, Item: Note{Letter: "a", Octave: 1, Velocity: -1}, ItemOriginalData: "a"}},
		{"?(a,b)", &RandomChooser{Alternatives: []interface{}{Note{Letter: "a", Octave: 1, Velocity: -1}, Note{Letter: "b", Octave: 1, Velocity: -1}}, AlternativesOriginalData: []string{"a", "b"}}},
		{"CC(23,100)", MIDICC{Controller: 23, Value: 100}},
		{"AT(54)", MIDIAftertouch{Value: 54}},
		{"PB(-20)", MIDIPitchbend{Value: -20}},
		{"PB(8100)", MIDIPitchbend{Value: 8100}},
		{"PT(c',80)", MIDIPolyAftertouch{Key: 60, Value: 80}},
		{"~", GlissandoLinear},
		{"~~", GlissandoExponential},
		//	{"O[30]/a/b", &OSCMessage{}},
	}

	for i, test := range tests {
		var parser Items

		parser.GetDefinition = func(name string) *template.Definition {
			if name == "pa" {
				return &template.Definition{
					Name:      "pa",
					NumParams: 0,
					Original:  "2a'",
				}
			}
			if name == "pb" {
				return &template.Definition{
					Name:      "pb",
					NumParams: 1,
					Original:  "1#1 2a' 2&a'",
				}
			}
			return nil
		}

		result, err := parser.ParseItem(test.input, 0)

		if err != nil {
			t.Errorf("[%v] could not parse item for %#v: %s", i, test.input, err.Error())
			continue
		}

		if !reflect.DeepEqual(result, test.expected) {
			t.Errorf("[%v] parseItem(%#v) = \n%#v \n// expected\n%#v", i, test.input, result, test.expected)
		}
	}
}
