package muskel

import "testing"

//calc32thsAdd

//halfTonesToPitchbend

func TestCalc32thsAdd(t *testing.T) {
	tests := []struct {
		distance32th int64
		notediff     int8
		expected     float64
	}{
		{32, 32, 1},
		{32, -32, -1},
		{32, 64, 2},
		{32, -64, -2},
	}

	for i, test := range tests {
		got := calcAdd(test.distance32th, test.notediff)

		if got != test.expected {
			t.Errorf("[%v] calc32thsAdd(%v, %v) == %v // expected %v", i, test.distance32th, test.notediff, got, test.expected)
		}
	}
}

func TestHalfTonesToPitchbend(t *testing.T) {
	tests := []struct {
		halftones float64
		expected  int16
	}{
		{0, 0},
		{1, 4096},
		{2, 8192},
		{-2, -8191},
		{-1, -4096},
	}

	for i, test := range tests {
		got := halfTonesToPitchbend(test.halftones, 2)

		if got != test.expected {
			t.Errorf("[%v] halfTonesToPitchbend(%v) == %v // expected %v", i, test.halftones, got, test.expected)
		}
	}
}
