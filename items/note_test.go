package items

import (
	"testing"
)

func TestScaleNote(t *testing.T) {
	tests := []string{
		"1",
		"2~",
		"-1:",
		"-10+",
		"10++",
		"10+++",
		"10++++",
		"10-",
		"10--",
		"10---",
		"10----",
		"10=",
		"10=:~",
	}

	for _, test := range tests {
		var nt = &Note{}
		err := nt.parseScale(test)

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
		"D#\"'::",
		"D#\"':::",
		"c\"\"+",
		"c\"\"\"++",
		"c\"'+++",
		"c\"'+++--",
		"c\"'++++",
		"c\"'-",
		"C\"'--",
		"C\"'---",
		"C#\"'----",
		"C#\"'=",
		"C#\"'=:~",
	}

	for _, test := range tests {
		nt, err := ParseNote(test)

		if err != nil {
			t.Errorf("could not parse note %q: %s", test, err.Error())
		}

		var n2 = nt.Dup().(*Note)
		//n2.Transposition = 0
		// n2.velocity = -1

		n2.Letter, n2.Augmenter, n2.Octave = KeyToNote(nt.ToMIDI())

		//fmt.Printf("%s: %v -> %v\n", test, nt.toMIDI(), n2.String())
		got1 := n2.String()

		if got1 != test && n2.ToMIDI() != nt.ToMIDI() {
			t.Errorf("keyToNote(%s.toMIDI()) == %q // expected %q", test, got1, test)
		}

		got2 := nt.String()

		if got2 != test {
			t.Errorf("parseNote(%q).String() == %q // expected %q", test, got2, test)
		}
	}
}
