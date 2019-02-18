package muskel

import (
	"testing"
)

func TestPositionTo32th(t *testing.T) {
	tests := []struct {
		lastPos   string
		pos       string
		completed string
		num32th   uint
	}{
		{"", "1", "1", 0},
		{"1", ";", "1;", 1},
		{"1", ".", "1.", 2},
		{"1", "&", "1&", 4},
		{"1", "&;", "1&;", 5},
		{"1&", ";", "1&;", 5},
		{"1", "&.", "1&.", 6},
		{"1&", ".", "1&.", 6},
		{"1", "&.;", "1&.;", 7},
		{"1&", ".;", "1&.;", 7},
		{"1&.", ";", "1&.;", 7},
		{"", "2", "2", 8},
		{"2", "&", "2&", 12},
	}

	for i, test := range tests {
		completed, num32th, _ := positionTo32th(test.lastPos, test.pos)

		if completed != test.completed || num32th != test.num32th {
			t.Errorf("[%v] positionTo32th(%#v, %#v) = %#v, %v // expected %#v, %v", i, test.lastPos, test.pos, completed, num32th, test.completed, test.num32th)
		}
	}
}

func TestGetQNNumberFromPos(t *testing.T) {
	tests := []struct {
		position string
		qnumber  int
		rest     string
	}{
		{"1", 1, ""},
		{"2", 2, ""},
		{"3", 3, ""},
		{"4", 4, ""},
		{"5", 5, ""},
		{"1&", 1, "&"},
		{"1&;", 1, "&;"},
		{"1&.;", 1, "&.;"},
		{"1.;", 1, ".;"},
		{"1.", 1, "."},
		{"1&.", 1, "&."},
		{"1;", 1, ";"},
		{"2&", 2, "&"},
		{"3&", 3, "&"},
		{"4&", 4, "&"},
		{"5&", 5, "&"},
	}

	for i, test := range tests {
		qn, rest := getQNNumberFromPos(test.position)

		if qn != test.qnumber || rest != test.rest {
			t.Errorf("[%v] getQNNumberFromPos(%#v) = %v, %#v // expected %v, %#v", i, test.position, qn, rest, test.qnumber, test.rest)
		}
	}
}

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
		nt, err := parseNote(test.note)

		if err != nil {
			t.Errorf("[%v] can't parse note %#v: %s", i, test.note, err)
			continue
		}

		got := noteToMIDI(nt.(Note))

		if got != test.midinote {
			t.Errorf("[%v] noteToMIDI(%#v) = %v // expected %v", i, test.note, got, test.midinote)
		}
	}
}

func TestVelocityFromDynamic(t *testing.T) {
	tests := []struct {
		dyn string
		vel int8
	}{
		{"++++", 123},
		{"+++", 108},
		{"++", 93},
		{"+", 78},
		{"*", 63},
		{"", -1},
		{"-", 48},
		{"--", 33},
		{"---", 18},
		{"----", 5},
	}

	for _, test := range tests {
		got := velocityFromDynamic(test.dyn)

		if got != test.vel {
			t.Errorf("velocityFromDynamic(%#v) = %v // expected %v", test.dyn, got, test.vel)
		}
	}
}

func TestPad(t *testing.T) {

	tests := []struct {
		s        string
		length   int
		expected string
	}{
		{"abc", 3, "abc"},
		{"abc", 4, "abc "},
		{"abc", 5, "abc  "},
	}

	for _, test := range tests {
		got := pad(test.s, test.length)

		if got != test.expected {
			t.Errorf("pad(%#v, %v) = %#v // expected %#v", test.s, test.length, got, test.expected)
		}
	}
}

func TestPos32thToString(t *testing.T) {

	tests := []struct {
		pos      uint
		expected string
	}{
		{0, "1"},
		{1, "1;"},
		{2, "1."},
		{3, "1.;"},
		{4, "1&"},
		{5, "1&;"},
		{6, "1&."},
		{7, "1&.;"},
		{8, "2"},
		{9, "2;"},
		{10, "2."},
		{11, "2.;"},
		{12, "2&"},
		{13, "2&;"},
		{14, "2&."},
		{15, "2&.;"},
		{16, "3"},
		{24, "4"},
		{32, "5"},
		{40, "6"},
	}

	for _, test := range tests {
		got := pos32thToString(test.pos)

		if got != test.expected {
			t.Errorf("pos32thToString(%v) = %#v // expected %#v", test.pos, got, test.expected)
		}
	}
}
