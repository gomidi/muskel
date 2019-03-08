package muskel

import (
	"testing"
)

func TestNoteToStep(t *testing.T) {
	var tests = []struct {
		mode     mode
		baseNote uint8
		note     uint8
		expected int8
	}{
		{
			Ionian,
			60,
			60,
			0,
		},
		{
			Ionian,
			60,
			62,
			1,
		},
		{
			Ionian,
			60,
			64,
			2,
		},
		{
			Ionian,
			60,
			65,
			3,
		},
		{
			Ionian,
			60,
			67,
			4,
		},
		{
			Ionian,
			60,
			69,
			5,
		},
		{
			Ionian,
			60,
			71,
			6,
		},
		{
			Ionian,
			60,
			72,
			7,
		},
		{
			Ionian,
			60,
			73,
			8,
		},
		{
			Ionian,
			60,
			74,
			8,
		},
		{
			Ionian,
			60,
			59,
			-1,
		},
		{
			Ionian,
			60,
			57,
			-2,
		},
		{
			Ionian,
			60,
			55,
			-3,
		},
		{
			Ionian,
			60,
			53,
			-4,
		},
		{
			Ionian,
			60,
			52,
			-5,
		},
		{
			Ionian,
			60,
			50,
			-6,
		},
		{
			Ionian,
			60,
			49,
			-6,
		},
		{
			Ionian,
			60,
			48,
			-7,
		},
		{
			Ionian,
			60,
			47,
			-8,
		},
		{
			Dorian,
			50,
			48,
			-1,
		},
		{
			Dorian,
			50,
			49,
			0,
		},
	}

	for i, test := range tests {
		/*
			if i != 20 {
				continue
			}
		*/
		got := test.mode.NoteToStep(test.baseNote, test.note)
		// fmt.Printf("[%v] basenote: %v, test.note: %v: got: %v\n", i, test.baseNote, test.note, got)

		if got != test.expected {
			t.Errorf("[%v] mode(%q).NoteToStep(%v,%v) == %v // expected: %v", i, test.mode.name, test.baseNote, test.note, got, test.expected)
		}
	}
}

func TestStepToNote(t *testing.T) {
	var tests = []struct {
		mode     mode
		baseNote uint8
		step     int8
		expected uint8
	}{
		{
			Ionian,
			60,
			0,
			60,
		},
		{
			Ionian,
			60,
			1,
			62,
		},
		{
			Ionian,
			60,
			2,
			64,
		},
		{
			Ionian,
			60,
			3,
			65,
		},
		{
			Ionian,
			60,
			4,
			67,
		},
		{
			Ionian,
			60,
			5,
			69,
		},
		{
			Ionian,
			60,
			6,
			71,
		},
		{
			Ionian,
			60,
			7,
			72,
		},
		{
			Ionian,
			60,
			8,
			74,
		},
		{
			Ionian,
			60,
			-1,
			59,
		},
		{
			Ionian,
			60,
			-2,
			57,
		},
		{
			Ionian,
			60,
			-3,
			55,
		},
		{
			Ionian,
			60,
			-4,
			53,
		},
		{
			Ionian,
			60,
			-5,
			52,
		},
		{
			Ionian,
			60,
			-6,
			50,
		},
		{
			Ionian,
			60,
			-7,
			48,
		},
		{
			Ionian,
			60,
			-8,
			47,
		},
	}

	for i, test := range tests {
		got := test.mode.StepToNote(test.baseNote, test.step)

		if got != test.expected {
			t.Errorf("[%v] mode(%q).StepToNote(%v,%v) == %v // expected: %v", i, test.mode.name, test.baseNote, test.step, got, test.expected)
		}
	}
}
