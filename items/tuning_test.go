package items

import (
	"testing"
)

/*

#phyrigian@440:  a"=1,a#"=12/11,b"=8/7,c"'=6/5,c#"'=9/7,d"'=4/3,d#"'=16/11,e"'=3/2,f"'=8/5,f#"'=12/7,g"'=9/5,g#"'=64/33

*/

func TestTuningDefinition(t *testing.T) {
	var tests = []struct {
		definition string
		getNote    string
		expected   float64
	}{
		{
			// a"=1,a#"=12/11,b"=8/7,c"'=6/5,c#"'=9/7,d"'=4/3,d#"'=16/11,e"'=3/2,f"'=8/5
			`d":23,d#":12,e":-12,f#":40,g":30,a":4,c":10`,
			`d"`,
			23,
		},
		{
			`d":23,d#":12,e":-12,f#":40,g":30,a":4,c":10`,
			`d#"`,
			12,
		},
		{
			`d":23,d#":12,e":-12,f#":40,g":30,a":4,c":10`,
			`e"`,
			-12,
		},
		{
			`d":23,d#":12,e":-12,f#":40,g":30,a":4,c":10`,
			`c'`,
			10,
		},
		{
			`d#":12,e":-12,f#":40,g":30,a":4,c":10`,
			`c#'`,
			10,
		},
	}

	for i, test := range tests {

		// fmt.Println(test.definition)

		var def TuningCents
		err := def.Parse("test", test.definition)

		if err != nil {
			t.Errorf("[%v] error while parsing %#v: %s", i, test.definition, err.Error())
			continue
		}
		/*
			if i != 20 {
				continue
			}
		*/

		// fmt.Printf("def %#v\n", def)

		var mn Note

		mn, err = ParseNote(test.getNote)

		// fmt.Printf("mountpoint: %v\n", mn)

		if err != nil {
			t.Errorf("[%v] error in getNote %#v: %s", i, test.getNote, err.Error())
			continue
		}

		nn := mn.ToMIDI()

		got := def.CentsFor(nn)

		// fmt.Printf("halftones %v zeroIdx: %v\n", scDef.halftones, scDef.zeroIdx)

		// fmt.Printf("note: %s\n", nt.String())

		if got != test.expected {
			t.Errorf("[%v] Tuning(%q).CentsFor(%q) == %v // expected: %v", i, test.definition, test.getNote, got, test.expected)
		}
	}
}
