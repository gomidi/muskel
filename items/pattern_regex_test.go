package items

import (
	"reflect"
	"testing"
)

func TestPatternRegexp(t *testing.T) {
	tests := []struct {
		input       string
		shouldMatch bool
		matches     []string
	}{
		{"hi", true, []string{"hi", "hi", "", "", "", ""}},
		{"=hi", true, []string{"=hi", "=hi", "", "", "", ""}},
		{"=.hi", true, []string{"=.hi", "=.hi", "", "", "", ""}},
		{"@hi", true, []string{"@hi", "@hi", "", "", "", ""}},
		{"=hi()", true, []string{"=hi", "=hi", "", "", "", ""}},
		{"=hi.ho()", true, []string{"=hi.ho", "=hi.ho", "", "", "", ""}},
		{"=hi(hui)", true, []string{"=hi", "=hi", "", "", "", ""}},
		{"=hi(hui)[0:3]", true, []string{"=hi", "=hi", "", "", "", ""}},
		{"=hi(hui)[0:3]@ho", true, []string{"=hi", "=hi", "", "", "", ""}},
	}

	for i, test := range tests {
		// TemplateReg
		res := regTemplateCallNameDyn.MatchString(test.input)

		if res != test.shouldMatch {
			var doesNot string
			var shouldNot string
			if !test.shouldMatch {
				shouldNot = " not"
			}

			if !res {
				doesNot = " not"
			}
			t.Errorf("[%v] %q does%s match as pattern, but should%s", i, test.input, doesNot, shouldNot)

		}

		mt := regTemplateCallNameDyn.FindStringSubmatch(test.input)

		if !reflect.DeepEqual(mt, test.matches) {
			t.Errorf("[%v] %q matches %#v // expected %#v", i, test.input, mt, test.matches)
		}
	}
}
