package muskel

import (
	"reflect"
	"testing"
)

func TestBarEnsurePositionExist(t *testing.T) {
	tests := []struct {
		positions         []uint
		originalPositions []string
		ensure            uint
		resultPos         []uint
		resultOrig        []string
	}{
		{[]uint{0, 8}, []string{"1", "2"}, 4, []uint{0, 4, 8}, []string{"1", "1&", "2"}},
		{[]uint{0, 8}, []string{"1", "2"}, 12, []uint{0, 8, 12}, []string{"1", "2", "2&"}},
		{[]uint{8, 16}, []string{"2", "3"}, 12, []uint{8, 12, 16}, []string{"2", "2&", "3"}},
		{[]uint{}, []string{}, 4, []uint{4}, []string{"1&"}},
	}

	for i, test := range tests {
		var b Bar
		b.positions = test.positions
		b.originalPositions = test.originalPositions

		b.ensurePositionExist(test.ensure)

		if !reflect.DeepEqual(b.positions, test.resultPos) || !reflect.DeepEqual(b.originalPositions, test.resultOrig) {
			t.Errorf(
				"[%v] (Bar{positions: %v, originalPositions: %v}).ensurePositionExist(%v) = %v, %v // expected %v, %v",
				i, test.positions, test.originalPositions, test.ensure, b.positions, b.originalPositions, test.resultPos, test.resultOrig,
			)
		}
	}
}
