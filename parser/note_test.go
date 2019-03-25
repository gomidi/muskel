package parser

import (
	"testing"

	"gitlab.com/gomidi/muskel/muskellib"
)

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
		"^10=",
		"^10=:~",
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
		"c#'",
		"C~",
		"D#\"':",
		"c\"\"+",
		"c\"\"\"++",
		"c\"'+++",
		"c\"'++++",
		"c\"'-",
		"C\"'--",
		"C\"'---",
		"C#\"'----",
		"C#\"'=",
		"C#\"'=:~",
	}

	for _, test := range tests {
		nt, err := parseNote(test)

		if err != nil {
			t.Errorf("could not parse note %q: %s", test, err.Error())
		}

		var n2 = nt.Dup()
		// n2.velocity = -1

		n2.Letter, n2.Augmenter, n2.Octave = muskellib.KeyToNote(nt.ToMIDI())

		//fmt.Printf("%s: %v -> %v\n", test, nt.toMIDI(), n2.String())
		got1 := n2.String()

		if got1 != test {
			t.Errorf("keyToNote(%s.toMIDI()) == %q // expected %q", test, got1, test)
		}

		got2 := nt.String()

		if got2 != test {
			t.Errorf("parseNote(%q).String() == %q // expected %q", test, got2, test)
		}
	}
}
