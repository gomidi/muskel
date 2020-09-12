package file

import "testing"

func TestDashLine(t *testing.T) {
	// dashLine.MatchString(line)

	tests := []struct {
		str     string
		matches bool
	}{
		{"|--|bu", true},
		{" |-----|bu", true},
		{"| -- |bu", true},
		{"| -- |--- |", true},
		{"--|bu", false},
		{"bu", false},
		{"--", false},
		{"|--", false},
	}

	for i, test := range tests {
		does := dashLine.MatchString(test.str)

		if test.matches && !does {
			t.Errorf("[%v] error: %q must be a dash line", i, test.str)
		}

		if !test.matches && does {
			t.Errorf("[%v] error: %q must not be a dash line", i, test.str)
		}
	}
}
