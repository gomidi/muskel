package items

import (
	"testing"
)

func TestScaleDefinitionNoteToStep(t *testing.T) {
	//t.Skip()
	var tests = []struct {
		definition string
		baseNote   string
		note       string
		expected   int8
	}{
		{
			`c",d",e",f",g",a",b"`,
			`a"`,
			`a"`,
			0,
		},
		{
			`c",d",e",f",g",a",b"`,
			`a"`,
			`b"`,
			1,
		},
		{
			`c",d",e",f",g",a",b"`,
			`a"`,
			`c"'`,
			2,
		},

		{
			`c",d",e",f",g",a",b"`,
			`a"`,
			`d"'`,
			3,
		},

		{
			`c",d",e",f",g",a",b"`,
			`a"`,
			`e"'`,
			4,
		},
		{
			`c",d",e",f",g",a",b"`,
			`a"`,
			`f"'`,
			5,
		},
		{
			`c",d",e",f",g",a",b"`,
			`a"`,
			`g"'`,
			6,
		},
		{
			`c",d",e",f",g",a",b"`,
			`a"`,
			`a"'`,
			7,
		},
		{
			`c",d",e",f",g",a",b"`,
			`a"`,
			`b"'`,
			8,
		},
		{
			`c",d",e",f",g",a",b"`,
			`a"`,
			`c""`,
			9,
		},
		{
			`c",d",e",f",g",a",b"`,
			`a"`,
			`g"`,
			-1,
		},
		{
			`c",d",e",f",g",a",b"`,
			`a"`,
			`f"`,
			-2,
		},
		{
			`c",d",e",f",g",a",b"`,
			`a"`,
			`e"`,
			-3,
		},
		{
			`c",d",e",f",g",a",b"`,
			`a"`,
			`d"`,
			-4,
		},
		{
			`c",d",e",f",g",a",b"`,
			`a"`,
			`c"`,
			-5,
		},
		{
			`c",d",e",f",g",a",b"`,
			`a"`,
			`b'`,
			-6,
		},
		{
			`c",d",e",f",g",a",b"`,
			`a"`,
			`a'`,
			-7,
		},
		{
			`c",d",e",f",g",a",b"`,
			`a"`,
			`g'`,
			-8,
		},
		{
			`c",d",e",f",f#",a",b"`,
			`g"`,
			`f#"`,
			-1,
		},
	}

	for i, test := range tests {

		if i != 2 {
			continue
		}

		var def ScaleDefinition
		err := def.Parse("test", test.definition)

		if err != nil {
			t.Errorf("[%v] error while parsing %#v: %s", i, test.definition, err.Error())
			continue
		}

		// fmt.Printf("def %#v\n", def)

		var mn Note

		mn, err = ParseNote(test.baseNote)

		// fmt.Printf("mountpoint: %v\n", mn)

		if err != nil {
			t.Errorf("[%v] error in basenote %#v: %s", i, test.baseNote, err.Error())
			continue
		}

		var nn Note
		nn, err = ParseNote(test.note)

		if err != nil {
			t.Errorf("[%v] error in note %#v: %s", i, test.note, err.Error())
			continue
		}

		got := def.NoteToStep(mn.ToMIDI(), nn.ToMIDI())

		// fmt.Printf("note: %s\n", nt.String())

		if got != test.expected {
			t.Errorf("[%v] ScaleDefinition(%q).NoteToStep(%q, %q) == %v // expected: %v", i, test.definition, test.baseNote, test.note, got, test.expected)
		}
	}
}

func TestScaleDefinitionStepToNote(t *testing.T) {
	//t.Skip()
	var tests = []struct {
		definition string
		baseNote   string
		step       int8
		expected   string
	}{
		{ // 0
			`c",d",e",f",g",a",b"`,
			`a"`,
			0,
			`a"`,
		},
		{ // 1
			`c",d",e",f",g",a",b"`,
			`a"`,
			1,
			`b"`,
		},
		{ // 2
			`c",d",e",f",g",a",b"`,
			`a"`,
			2,
			`c"'`,
		},
		{ // 3
			`c",d",e",f",g",a",b"`,
			`a"`,
			3,
			`d"'`,
		},
		{ // 4
			`c",d",e",f",g",a",b"`,
			`a"`,
			4,
			`e"'`,
		},
		{ // 5
			`c",d",e",f",g",a",b"`,
			`a"`,
			5,
			`f"'`,
		},
		{ // 6
			`c",d",e",f",g",a",b"`,
			`a"`,
			6,
			`g"'`,
		},
		{ // 7
			`c",d",e",f",g",a",b"`,
			`a"`,
			7,
			`a"'`,
		},
		{ // 8
			`c",d",e",f",g",a",b"`,
			`a"`,
			8,
			`b"'`,
		},
		{ // 9
			`c",d",e",f",g",a",b"`,
			`a"`,
			9,
			`c""`,
		},
		{ // 10
			`c",d",e",f",g",a",b"`,
			`a"`,
			-1,
			`g"`,
		},
		{ // 11
			`c",d",e",f",g",a",b"`,
			`a"`,
			-2,
			`f"`,
		},
		{ // 12
			`c",d",e",f",g",a",b"`,
			`a"`,
			-3,
			`e"`,
		},
		{ // 13
			`c",d",e",f",g",a",b"`,
			`a"`,
			-4,
			`d"`,
		},
		{ // 14
			`c",d",e",f",g",a",b"`,
			`a"`,
			-5,
			`c"`,
		},
		{ // 15
			`c",d",e",f",g",a",b"`,
			`a"`,
			-6,
			`b'`,
		},
		{ // 16
			`c",d",e",f",g",a",b"`,
			`a"`,
			-7,
			`a'`,
		},
		{ // 17
			`c",d",e",f",g",a",b"`,
			`a"`,
			-8,
			`g'`,
		},
		{ // 18
			`c",d",e",f",f#",a",b"`,
			`g"`,
			-1,
			`e"`,
		},
	}

	for i, test := range tests {

		/*
			if i != 18 {
				continue
			}
		*/

		var def ScaleDefinition
		err := def.Parse("test", test.definition)

		if err != nil {
			t.Errorf("[%v] error while parsing %#v: %s", i, test.definition, err.Error())
			continue
		}

		// fmt.Printf("def %#v\n", def)

		var mn Note

		mn, err = ParseNote(test.baseNote)

		// fmt.Printf("mountpoint: %v\n", mn)

		if err != nil {
			t.Errorf("[%v] error in basenote %#v: %s", i, test.baseNote, err.Error())
			continue
		}

		nn := def.StepToNote(mn.ToMIDI(), test.step)

		var nt Note
		nt.Letter, nt.Augmenter, nt.Octave = KeyToNote(nn)
		//nt.Velocity = 63
		// fmt.Printf("note: %s\n", nt.String())

		got := nt.String()

		if got != test.expected {
			t.Errorf("[%v] ScaleDefinition(%q).StepToNote(%q,%v) == %v // expected: %v", i, test.definition, test.baseNote, test.step, got, test.expected)
		}
	}
}
