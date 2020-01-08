package table

import (
	"testing"
)

func TestTableName(t *testing.T) {

	tests := []struct {
		name string
		ok   bool
	}{
		{"=hi", true},
		{"@ho", true},
		{"hu", true},
		{"=_hi", false},
		{"@_ho", false},
		{"_hu", false},
		{"=hi_", true},
		{"@ho_", true},
		{"hu_", true},
		{"=hi.ho", true},
		{"@ho.hi", true},
		{"hu.ho", true},
		{"=~hi.ho", false},
		{"@~ho.hi", false},
		{"~hu.ho", false},
		{"hi=", false},
		{"ho@", false},
		{"hi:", false},
		//{"=h=i", false},
	}

	for i, test := range tests {
		res := tableNameRegEx.MatchString(test.name)

		if test.ok != res {
			t.Errorf("[%v] tableNameRegEx.MatchString(%q) = %v // expected: %v", i, test.name, res, test.ok)
		}
	}
}
