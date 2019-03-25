package items

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
		completed, num32th, _ := PositionTo32th(test.lastPos, test.pos)

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

func TestVelocityFromDynamic(t *testing.T) {
	tests := []struct {
		dyn string
		vel int8
	}{
		{"++++", 123},
		{"+++", 108},
		{"++", 93},
		{"+", 78},
		{"=", 63},
		{"", -1},
		{"-", 48},
		{"--", 33},
		{"---", 18},
		{"----", 5},
	}

	for _, test := range tests {
		got := DynamicToVelocity(test.dyn)

		if got != test.vel {
			t.Errorf("velocityFromDynamic(%#v) = %v // expected %v", test.dyn, got, test.vel)
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
		got := Pos32thToString(test.pos)

		if got != test.expected {
			t.Errorf("pos32thToString(%v) = %#v // expected %#v", test.pos, got, test.expected)
		}
	}
}
