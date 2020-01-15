package muskel

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"gitlab.com/gomidi/muskel/file"
)

type tabler interface {
	file.Part
	Name() string
	Cols() []string
}

func _newFile(s string) *file.File {
	return newFile("test", nil, strings.NewReader(s))
}

func TestFormatSimple(t *testing.T) {
	//t.Skip()
	tests := []struct {
		input    string
		expected string
	}{
		// 0
		{`
TRACK   | Voc | Git | Bass |
Channel | 2 | 3 | 4 |
Delay   | 1/8 | 1/5 | -1/5 |


PROPERTY |
x          | y

=texter | a | b | c
#
1 | |

`, `
TRACK   | Voc | Git | Bass |
 Channel| 2   | 3   | 4    |
 Delay  | 1/8 | 1/5 | -1/5 |


PROPERTY |
 x       | y

=texter | a | b | c |
 #
  1     |   |   |   |

`},

		// 1
		{`
//=a: b
=texter | a | b | c
#
1 | |

`, `
//=a: b
=texter | a | b | c |
 #
  1     |   |   |   |

`},

		// 2
		{`
PROPERTY
a           | b
/*
=texter | a | b | c
1 | |
*/

`, `
PROPERTY |
 a       | b
/*
=texter | a | b | c
1 | |
*/

`},

		// 3
		{`
PROPERTY
a          | b

PROPERTY
c          | d

/*
=texter | a | b | c
1 | |
*/

=test | c | b | c
 #
  1 | |
  2 |b|

`, `
PROPERTY |
 a       | b

PROPERTY |
 c       | d

/*
=texter | a | b | c
1 | |
*/

=test | c | b | c |
 #
  1   |   |   |   |
  2   | b |   |   |

`},

		// 4
		{`
PROPERTY |
a            | b    
 
/*
=texter | a | b | c
1 | |
*/

=test | c | b | c
#
//1 | |
2 |b|

`, `
PROPERTY |
 a       | b

/*
=texter | a | b | c
1 | |
*/

=test | c | b | c |
 #
//1 | |
  2   | b |   |   |

`},

		// 5
		{`
PROPERTY |
a            | b   
  
/*
=texter | a | b | c
1 | |
*/

=test | c | b | c |
#
//1 | |
2 |b|
//3 | |
4 |b|

`, `
PROPERTY |
 a       | b

/*
=texter | a | b | c
1 | |
*/

=test | c | b | c |
 #
//1 | |
  2   | b |   |   |
//3 | |
  4   | b |   |   |

`},
		// 6
		{`
PROPERTY | 
a        | b     

/*
=texter | a | b | c
1 | | | |C
*/

=test | c | b | c
#
//1 | |
2 |b|||D
//3 | |
4 |b|

`, `
PROPERTY |
 a       | b

/*
=texter | a | b | c
1 | | | |C
*/

=test | c | b | c |
 #
//1 | |
  2   | b |   |   | D
//3 | |
  4   | b |   |   |

`},
		// 7
		{`
PROPERTY |
a           | b     

TRACK |voc|git|bass|
Channel |1|2|3|
Program |10|20|30|

`, `
PROPERTY |
 a       | b

TRACK   | voc | git | bass |
 Channel| 1   | 2   | 3    |
 Program| 10  | 20  | 30   |

`},
		// 8
		{`
PROPERTY |
a        | b

SCALE |test|maj_or|min_or 
1|d#'|c|a"|
2|f'|d|b"|
3|a"|a"|c"'|
4|a#'|c"'|d"'|

`, `
PROPERTY |
 a       | b

SCALE | test | maj_or | min_or |
 1    | d#'  | c      | a"     |
 2    | f'   | d      | b"     |
 3    | a"   | a"     | c"'    |
 4    | a#'  | c"'    | d"'    |

`},
		// 9
		{`
PROPERTY |
 a       | b

TUNING |custom|
d"  | +23 |
d#"| +12 |
e"  | -12|
f#" | +40 |
g" | +30|
a" | 400Hz|
c"| +10 |

`, `
PROPERTY |
 a       | b

TUNING | custom |
 d"    | +23    |
 d#"   | +12    |
 e"    | -12    |
 f#"   | +40    |
 g"    | +30    |
 a"    | 400Hz  |
 c"    | +10    |

`},
		// 10
		{`
PROPERTY
a            | b

TIMBRE| legato[vocals] | staccato[vocals] |
On|@on-0.25        | @on              |
Off  |@off            | @on+0.25         |
Velocity |@vel-5          | @vel+5           |
PitchBend|@pitch          | @pitch+10        |
CC12  |125~;0          |                  |
CC35   |1~;120          |                  |

`, `
PROPERTY |
 a       | b

TIMBRE    | legato[vocals] | staccato[vocals] |
 On       | @on-0.25       | @on              |
 Off      | @off           | @on+0.25         |
 Velocity | @vel-5         | @vel+5           |
 PitchBend| @pitch         | @pitch+10        |
 CC12     | 125~;0         |                  |
 CC35     | 1~;120         |                  |

`},
		// 11
		{`
drums |
 kd   | MN34++

=rhythm |euclid |
#
1       |$_euclid(3,12,&) |

PROPERTY |
a            | b

FILTER | bass     |
source   | drums   |  
range    | *         |
srcrange | kd        |
thresh   | *         |
srcthresh| *         |
factor | *-0.5     |
target| cc_volume |
attack| 0.10ms    |
release| 0.20ms    |
max    | 100       |
min    | 20        |
type  | *         |

`, `
drums |
 kd   | MN34++

=rhythm | euclid           |
 #
  1     | $_euclid(3,12,&) |

PROPERTY |
 a       | b

FILTER    | bass      |
 source   | drums     |
 range    | *         |
 srcrange | kd        |
 thresh   | *         |
 srcthresh| *         |
 factor   | *-0.5     |
 target   | cc_volume |
 attack   | 0.10ms    |
 release  | 0.20ms    |
 max      | 100       |
 min      | 20        |
 type     | *         |

`},
		// 12
		{`
TRACK | voc |
ch  | 1 |

=SCORE | voc |
#
1    | c#" |
*3
#
2 | d' |
`, `
TRACK | voc |
 ch   | 1   |

=SCORE | voc |
 #
  1    | c#" |
 *3
 #
  2    | d'  |
`},
		// 13
		{`
TRACK | voc | bass | git |
ch  | 1 |

=tester | a | b |
#
1 | ^3 |

=SCORE | voc | bass |
#
1    | c#" |
*3
#
2 | d' |
`, `
TRACK | voc | bass | git |
 ch   | 1   |      |     |

=tester | a  | b |
 #
  1     | ^3 |   |

=SCORE | voc | bass |
 #
  1    | c#" |      |
 *3
 #
  2    | d'  |      |
`},
		// 14
		{`
TRACK | voc | bass |
ch  | 1 |

@lyrics |
hu ha ho | v1
hü hä hö   |

=SCORE | voc | bass |
#
1    | c#" |
*3
#
2 | d' |
`, `
TRACK | voc | bass |
 ch   | 1   |      |

@lyrics     |
 hu ha ho   | v1
 hü hä hö   |

=SCORE | voc | bass |
 #
  1    | c#" |      |
 *3
 #
  2    | d'  |      |
`},
		// 15
		{`
TRACK | voc | bass |
 // Volume | 100 | 80 |
ch  | 1 |

=SCORE | voc | bass |
#
1    | c#" |
*3
#
2 | d' |
`, `
TRACK | voc | bass |
// Volume | 100 | 80 |
 ch   | 1   |      |

=SCORE | voc | bass |
 #
  1    | c#" |      |
 *3
 #
  2    | d'  |      |
`},
		// 16
		{`
TRACK | voc | bass |
 // Volume | 100 | 80 |
ch  | 1 |

=SCORE | voc | bass |
#
1    | c#" |
*3
# // ##
2 | d' |
`, `
TRACK | voc | bass |
// Volume | 100 | 80 |
 ch   | 1   |      |

=SCORE | voc | bass |
 #
  1    | c#" |      |
 *3
 # // ## | voc | bass |
  2    | d'  |      |
`},
	}

	for i, test := range tests {

		/*
			if i > 10 {
				continue
			}
		*/
		//fmt.Printf("######### %v #############\n", i)

		fl := _newFile(test.input)
		err := fl.Parse()

		if err != nil {
			t.Errorf("[%v] error: %s", i, err)
			continue
		}

		var bf bytes.Buffer

		err = fl.WriteTo(&bf)

		if err != nil {
			t.Errorf("[%v] error: %s", i, err)
			continue
		}

		res := bf.String()

		if res != test.expected {
			t.Errorf("[%v] Parse(%q).WriteTo() = \n%s \n// expected\n%s\n%q vs\n%q\n", i, test.input, res, test.expected, res, test.expected)
		}

	}

}

func TestPropertie(t *testing.T) {
	//t.Skip()
	tests := []struct {
		input    string
		expected map[string]interface{}
	}{
		{
			`
PROPERTY |
a        | b     |

PROPERTY |
c        | d     |

`,
			map[string]interface{}{"a": "b", "c": "d"},
		},
		{
			`
PROPERTY |
a        | b     |

PROPERTY |
a        | d     |

`,
			map[string]interface{}{"a": "d"},
		},
	}

	for i, test := range tests {

		if i > 3 {
			continue
		}

		fl := _newFile(test.input)
		err := fl.Parse()

		if err != nil {
			t.Errorf("[%v] error: %s", i, err)
			continue
		}

		props := fl.Score.Properties()

		if !reflect.DeepEqual(props, test.expected) {
			t.Errorf("[%v] %q.Properties() = \n%#v \n// expected\n%#v", i, test.input, props, test.expected)
		}

	}

}

func TestParseParts(t *testing.T) {
	//t.Skip()
	tests := []struct {
		input string
		parts []string
	}{
		//	{"@a: b\n$include(\"main\",\"myfile\")\ntexter: | a | b | c\n1 | |\n\n", []string{"Property", "Include", "SketchTable", "EmptyLine"}},
		//	{"@a: b\ntexter: | a | b | c\n1 | |\n\n$include(\"my_file\")\n", []string{"Property", "SketchTable", "EmptyLine", "Include"}},
		{"//=a: b\n=texter | a | b | c\n#\n1 | |\n\n", []string{"Sketch", "EmptyLine"}},
		{"//=a: b\n=texter | a | b | c\n#\n1 | |\n\n=texter2 | a | b | c\n#\n1 | |\n", []string{"Sketch", "EmptyLine", "Sketch"}},
		{"//=a: b\n=texter | a | b | c\n#\n1 | |\n//c |d|\n2 | |\n\n", []string{"Sketch", "EmptyLine"}},
		{"//=a: b\n=texter | a | b | c\n#\n1 | |\n/*\nc |d|\n*/\n2 | |\n\n", []string{"Sketch", "EmptyLine"}},
		{"PROPERTY | \nwhow | yeah \n/*\n=texter | a | b | c\n#\n1 | |\n*/\n\n", []string{"Properties", "EmptyLine"}},
		{"drums | a |\nkd | MN12++|\n/*\n=texter | a | b | c\n#\n1 | |\n*/\n\n", []string{"Tokens", "EmptyLine"}},
		//{"rythm: $euclid(3,5,&)\n/*\ntexter: | a | b | c\n#\n1 | |\n*/\n\n", []string{"CommandSketch", "EmptyLine"}},
	}

	for i, test := range tests {
		// fl := NewFile("test.mskl", "", strings.NewReader(test.input))
		fl := _newFile(test.input)
		err := fl.Parse()

		if err != nil {
			t.Errorf("%v error: %s", i, err)
			continue
		}

		var partTypes []string

		for _, part := range fl.Parts {
			x := strings.Split(fmt.Sprintf("%T", part), ".")
			partTypes = append(partTypes, x[1])
		}

		if !reflect.DeepEqual(test.parts, partTypes) {
			t.Errorf("%v Parse(%q): table.Parts = %v // expected %v", i, test.input, partTypes, test.parts)
		}

	}

}

func TestParseTableHeader(t *testing.T) {
	//t.Skip()
	tests := []struct {
		input string
		name  string
		cols  []string
	}{
		{"=texter |a|b|c|\n#\n1 | | | |", "=texter", []string{"a", "b", "c"}},
		{"=tex_ter | a | B | c |\n#\n1 | | | |\n", "=tex_ter", []string{"a", "B", "c"}},
		{"test\nx|y\nc|d", "test", nil},
		{"test ||\nx|y|\nc|d|", "test", nil},
	}

	for i, test := range tests {
		// fl := NewFile("test.mskl", "", strings.NewReader(test.input))
		fl := _newFile(test.input)
		err := fl.Parse()

		if err != nil {
			t.Errorf("%v error: %s", i, err)
			continue
		}

		if len(fl.Parts) < 1 {
			t.Errorf("%v parsed less than one part: %#v", i, fl.Parts)
			continue
		}

		tbl, ok := fl.Parts[0].(tabler)

		if !ok {
			t.Errorf("%v part is no tabler, but %T", i, fl.Parts[0])
			continue
		}

		if tbl.Name() != test.name {
			t.Errorf("%v Parse(%q): table.Name = %q // expected %q", i, test.name, tbl.Name(), test.cols)
		}

		if !reflect.DeepEqual(test.cols, tbl.Cols()) {
			t.Errorf("%v Parse(%q): table.Cols = %#v // expected %#v", i, test.name, tbl.Cols(), test.cols)
		}

	}
}
