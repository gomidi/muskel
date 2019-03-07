package muskel

import "testing"

func TestScaleNote(t *testing.T) {
	tests := []string{
		"^1",
		"^2~",
		"^-1:",
		"^-10+",
		"^10++",
		"^10+++",
		"^10++++",
		"^10-",
		"^10--",
		"^10---",
		"^10----",
		"^10*",
		"^10*:~",
	}

	for _, test := range tests {
		nt, err := parseScaleNote(test[1:])

		if err != nil {
			t.Errorf("could not parse note %q: %s", test, err.Error())
		}

		got := nt.String()

		if got != test {
			t.Errorf("parseNote(%q).String() == %q // expected %q", test, got, test)
		}
	}
}

func TestNote(t *testing.T) {
	tests := []string{
		"a#'",
		"C~",
		"D#\"':",
		"c\"\"+",
		"c\"\"\"++",
		"c\"\"'+++",
		"c\"\"'++++",
		"c\"'-",
		"C\"\"--",
		"C\"'---",
		"C#\"'----",
		"C#\"'*",
		"C#\"'*:~",
	}

	for _, test := range tests {
		nt, err := parseNote(test)

		if err != nil {
			t.Errorf("could not parse note %q: %s", test, err.Error())
		}

		got := nt.String()

		if got != test {
			t.Errorf("parseNote(%q).String() == %q // expected %q", test, got, test)
		}
	}
}
