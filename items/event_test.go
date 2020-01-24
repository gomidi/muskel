package items

import (
	"fmt"
	"testing"
)

func TestEventParse(t *testing.T) {
	tests := []struct {
		input    string
		itemType string
		posShift int
	}{
		{"a<<", "*items.Note", -2},
		{"MN32>>>", "*items.MIDINote", 3},
		{"MN32>>><<", "*items.MIDINote", 1},
		{"MN32>>><<<", "*items.MIDINote", 0},
	}

	for i, test := range tests {
		var ev Event
		err := ev.Parse(test.input)

		if err != nil {
			t.Errorf("[%v] error in ev.Parse(%q): %s ", i, test.input, err.Error())
			continue
		}

		itemType := fmt.Sprintf("%T", ev.Item)

		if itemType != test.itemType || ev.PosShift != test.posShift {
			t.Errorf("[%v] ev.Parse(%q): type: %s posshift: %v // expected  type: %s posshift: %v", i, test.input, itemType, ev.PosShift, test.itemType, test.posShift)
		}
	}
}
