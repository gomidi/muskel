package items

import (
	"testing"
)

func TestNoteToMIDI(t *testing.T) {
	tests := []struct {
		note     string
		midinote uint8
	}{
		{"C''", 12},
		{"C#''", 13},
		{"D''", 14},
		{"D#''", 15},
		{"E''", 16},
		{"F''", 17},
		{"F#''", 18},
		{"G''", 19},
		{"G#''", 20},
		{"A''", 21},
		{"A#''", 22},
		{"B''", 23},
		{"C'", 24},
		{"C#'", 25},
		{"D'", 26},
		{"D#'", 27},
		{"E'", 28},
		{"F'", 29},
		{"F#'", 30},
		{"G'", 31},
		{"G#'", 32},
		{"A'", 33},
		{"A#'", 34},
		{"B'", 35},
		{"C", 36},
		{"C#", 37},
		{"D", 38},
		{"D#", 39},
		{"E", 40},
		{"F", 41},
		{"F#", 42},
		{"G", 43},
		{"G#", 44},
		{"A", 45},
		{"A#", 46},
		{"B", 47},
		{"c", 48},
		{"c#", 49},
		{"d", 50},
		{"d#", 51},
		{"e", 52},
		{"f", 53},
		{"f#", 54},
		{"g", 55},
		{"g#", 56},
		{"a", 57},
		{"a#", 58},
		{"b", 59},
		{"c'", 60},
		{"c#'", 61},
		{"d'", 62},
		{"d#'", 63},
		{"e'", 64},
		{"f'", 65},
		{"f#'", 66},
		{"g'", 67},
		{"g#'", 68},
		{"a'", 69},
		{"a#'", 70},
		{"b'", 71},
	}

	for i, test := range tests {
		nt, err := ParseNote(test.note)

		if err != nil {
			t.Errorf("[%v] can't parse note %#v: %s", i, test.note, err)
			continue
		}

		got := nt.ToMIDI()

		if got != test.midinote {
			t.Errorf("[%v] noteToMIDI(%#v) = %v // expected %v", i, test.note, got, test.midinote)
		}
	}
}
