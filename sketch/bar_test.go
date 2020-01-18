package sketch

import (
	"reflect"
	"strings"
	"testing"

	"gitlab.com/gomidi/muskel/items"
)

/*
rules:

- a bar must always know its proper length:
  - normal bar: based on timesignature
  - jumps based on parts length
  - imports based on length of import column
- a bar will get its correct position from the very beginning (no injected bars)
- empty position lines take the previous position and result in multiple events on the same position and column
- commands, background-patterns and scale-changes take percedence when sorting before other items on the same position
*/

func TestBars(t *testing.T) {

	inc := NewSketch("testinclude", nil)
	inc.AddColumn("col1")
	inc.AddColumn("col2")
	inc.ParseLine([]string{"# 2/4"})
	inc.ParseLine([]string{"1 | f | g |"})
	inc.ParseLine([]string{"# 3/4"})
	inc.ParseLine([]string{"1 | f | g |"})

	tests := []struct {
		lines    string
		expected []*Bar
	}{
		{ // 0
			`
#
1 | a | c |
`,
			[]*Bar{
				&Bar{Position: 0, TimeSig: [2]uint8{4, 4}, TempoChange: 120},
			},
		},
		{ // 1
			`
# 4/4
1 | a | c |
`,
			[]*Bar{
				&Bar{Position: 0, TimeSig: [2]uint8{4, 4}, TimeSigChange: [2]uint8{4, 4}, TempoChange: 120},
			},
		},
		{ // 2
			`
# 3/4
1 | a | c |
`,
			[]*Bar{
				&Bar{Position: 0, TimeSig: [2]uint8{3, 4}, TimeSigChange: [2]uint8{3, 4}, TempoChange: 120},
			},
		},
		{ // 3
			`
# 3/6
1 | a | c |
`,
			[]*Bar{
				&Bar{Position: 0, TimeSig: [2]uint8{3, 6}, TimeSigChange: [2]uint8{3, 6}, TempoChange: 120},
			},
		},
		{ // 4
			`
#A
1 | a | c |
#B 3/4
1 | c | d |
[A]
# 6/8
1 | f | g |
`,
			[]*Bar{
				&Bar{No: 0, Position: 0, TimeSig: [2]uint8{4, 4}, Part: "A", TempoChange: 120},
				&Bar{No: 1, Position: 32, TimeSig: [2]uint8{3, 4}, TimeSigChange: [2]uint8{3, 4}, Part: "B", TempoChange: 120},
				&Bar{No: 2, Position: 56, TimeSig: [2]uint8{4, 4}, JumpTo: "A"},
				&Bar{No: 3, Position: 88, TimeSig: [2]uint8{6, 8}, TimeSigChange: [2]uint8{6, 8}, TempoChange: 120},
			},
		},
		{ // 5
			`
#A 3/4     // tester#1
1 | a | c |
#B 4/4
1 | c | d |
[A]
# 6/8
1 | f | g |
`,
			[]*Bar{
				&Bar{No: 0, Position: 0, TimeSig: [2]uint8{3, 4}, TimeSigChange: [2]uint8{3, 4}, Comment: "tester#1", Part: "A", TempoChange: 120},
				&Bar{No: 1, Position: 24, TimeSig: [2]uint8{4, 4}, TimeSigChange: [2]uint8{4, 4}, Part: "B", TempoChange: 120},
				&Bar{No: 2, Position: 56, TimeSig: [2]uint8{4, 4}, JumpTo: "A"},
				&Bar{No: 3, Position: 80, TimeSig: [2]uint8{6, 8}, TimeSigChange: [2]uint8{6, 8}, TempoChange: 120},
			},
		},
		{ // 6
			`
#A 3/4
1 | a | c |
#B 4/4
1 | c | d |
$include("testinclude")
#C 6/8
1 | f | g |
[B]
# 4/4
`,
			[]*Bar{
				&Bar{No: 0, Position: 0, TimeSig: [2]uint8{3, 4}, TimeSigChange: [2]uint8{3, 4}, Part: "A", TempoChange: 120},
				&Bar{No: 1, Position: 24, TimeSig: [2]uint8{4, 4}, TimeSigChange: [2]uint8{4, 4}, Part: "B", TempoChange: 120},
				&Bar{No: 2, Position: 56, TimeSig: [2]uint8{4, 4}, Include: &items.Include{
					File:        "testinclude",
					Sketch:      "=SCORE",
					Length32ths: 40,
				},
				},
				&Bar{No: 3, Position: 96, TimeSig: [2]uint8{6, 8}, TimeSigChange: [2]uint8{6, 8}, Part: "C", TempoChange: 120},
				&Bar{No: 4, Position: 120, TimeSig: [2]uint8{4, 4}, JumpTo: "B"},
				&Bar{No: 5, Position: 192, TimeSig: [2]uint8{4, 4}, TimeSigChange: [2]uint8{4, 4}, TempoChange: 120},
			},
		},
		{ // 7
			`
#A 3/4
1 | a | c |
#B 4/4
1 | c | d |
$include("testinclude")
$include("testinclude")
# 6/8
1 | f | g |
[B]
# 4/4
`,
			[]*Bar{
				&Bar{No: 0, Position: 0, TimeSig: [2]uint8{3, 4}, TimeSigChange: [2]uint8{3, 4}, Part: "A", TempoChange: 120},
				&Bar{No: 1, Position: 24, TimeSig: [2]uint8{4, 4}, TimeSigChange: [2]uint8{4, 4}, Part: "B", TempoChange: 120},
				&Bar{No: 2, Position: 56, TimeSig: [2]uint8{4, 4},
					Include: &items.Include{
						File:        "testinclude",
						Sketch:      "=SCORE",
						Length32ths: 40,
					},
				},
				&Bar{No: 3, Position: 96, TimeSig: [2]uint8{4, 4},
					Include: &items.Include{
						File:        "testinclude",
						Sketch:      "=SCORE",
						Length32ths: 40,
					},
				},
				&Bar{No: 4, Position: 136, TimeSig: [2]uint8{6, 8}, TimeSigChange: [2]uint8{6, 8}, TempoChange: 120},
				&Bar{No: 5, Position: 160, TimeSig: [2]uint8{4, 4}, JumpTo: "B"},
				&Bar{No: 6, Position: 296, TimeSig: [2]uint8{4, 4}, TimeSigChange: [2]uint8{4, 4}, TempoChange: 120},
			},
		},
	}

	for i, test := range tests {

		sk := NewSketch("testscore", &testScore{include: inc})
		sk.AddColumn("col1")
		sk.AddColumn("col2")
		lines := strings.Split(strings.TrimSpace(test.lines), "\n")
		for j, line := range lines {
			err := sk.ParseLine(strings.Split(line, "|"))
			if err != nil {
				t.Errorf("[%v] can't parse line %v %q: %s", i, j, line, err.Error())
			}
		}

		for k, bar := range test.expected {
			if !reflect.DeepEqual(bar, sk.Bars[k]) {
				t.Errorf(`
[%v] bar %v:
%s
%#v // expected
%#v
include:
%#v vs
%#v
Position:
%v vs
%v`, i, k, strings.TrimSpace(test.lines),
					sk.Bars[k], bar,
					sk.Bars[k].Include, bar.Include,
					sk.Bars[k].Position, bar.Position,
				)
			}
		}

	}
}
