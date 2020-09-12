package sketch

import (
	"fmt"
	"sort"
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

	inc := New("testinclude", nil)
	inc.AddColumn("col1")
	inc.AddColumn("col2")
	inc.ParseLine([]string{"# 2/4"})
	inc.ParseLine([]string{"1 | f | g |"})
	inc.ParseLine([]string{"# 3/4"})
	inc.ParseLine([]string{"1 | f | g |"})

	tests := []struct {
		lines    string
		expected []string
	}{
		{ // 0
			`
#
1 | a | c |
`,
			[]string{
				"[0] #0 @120.00",
			},
		},
		{ // 1
			`
# 4/4
1 | a | c |
`,
			[]string{
				"[0] #0 @120.00 4/4",
			},
		},
		{ // 2
			`
# 3/4
1 | a | c |
`,
			[]string{
				"[0] #0 @120.00 3/4",
			},
		},
		{ // 3
			`
# 3/6
1 | a | c |
`,
			[]string{
				"[0] #0 @120.00 3/6",
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
			[]string{
				"[0] #0 @120.00 [A]",
				"[32] #1 @120.00 3/4 [B]",
				"[56] #2 @0.00 [->A]",
				"[88] #3 @120.00 6/8",
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
			[]string{
				"[0] #0 @120.00 3/4 [A] // tester#1",
				"[24] #1 @120.00 4/4 [B]",
				"[56] #2 @0.00 [->A]",
				"[80] #3 @120.00 6/8",
			},
		},
		{ // 6
			`
#A 3/4
1 | a | c |
#B 4/4
1 | c | d |
'testinclude
#C 6/8
1 | f | g |
[B]
# 4/4
`,
			[]string{
				"[0] #0 @120.00 3/4 [A]",
				"[24] #1 @120.00 4/4 [B]",
				`[56] #2 @0.00 ("testinclude"."=testscore" len: 40)`,
				"[96] #3 @120.00 6/8 [C]",
				"[120] #4 @0.00 [->B]",
				"[192] #5 @120.00 4/4",
			},
		},
		{ // 7
			`
#A 3/4
1 | a | c |
#B 4/4
1 | c | d |
'testinclude
'testinclude
# 6/8
1 | f | g |
[B]
# 4/4
`,
			[]string{
				"[0] #0 @120.00 3/4 [A]",
				"[24] #1 @120.00 4/4 [B]",
				`[56] #2 @0.00 ("testinclude"."=testscore" len: 40)`,
				`[96] #3 @0.00 ("testinclude"."=testscore" len: 40)`,
				"[136] #4 @120.00 6/8",
				"[160] #5 @0.00 [->B]",
				"[296] #6 @120.00 4/4",
			},
		},
		{ // 8
			`
# 3/4
1& @130 | a | c |
`,
			[]string{
				"[0] #0 @120.00 3/4 {[04] @130.00}",
			},
		},
		{ // 9
			`
# 3/4
1& @130 | a | c |
2 \minor^a' |b | d |
4  @135 |   |  |
# 4/4 @140
`,
			[]string{
				`[0] #0 @120.00 3/4 {[04] @130.00,[24] @135.00} {[08] \minor^a'}`,
				"[24] #1 @140.00 4/4",
			},
		},
	}

	for i, test := range tests {

		sk := New("=testscore", &testScore{include: inc})
		sk.AddColumn("col1")
		sk.AddColumn("col2")
		lines := strings.Split(strings.TrimSpace(test.lines), "\n")
		for j, line := range lines {
			err := sk.ParseLine(strings.Split(line, "|"))
			if err != nil {
				t.Errorf("[%v] can't parse line %v %q: %s", i, j, line, err.Error())
			}
		}

		if len(sk.Bars) != len(test.expected) {
			t.Errorf("[%v] expected %v bars, got %v", i, len(test.expected), len(sk.Bars))
			continue
		}

		for k, bar := range sk.Bars {

			//if !reflect.DeepEqual(bar, sk.Bars[k]) {
			if barToString(bar) != test.expected[k] {
				t.Errorf(`
[%v] bar %v:
%s
%#v // expected
%#v
`, i, k, strings.TrimSpace(test.lines),
					barToString(bar), test.expected[k],
					//sk.Bars[k].Include, bar.Include,
					//sk.Bars[k].Position, bar.Position,
				)
			}
		}

	}
}

func internalScalesToString(scales map[uint]*items.Scale) string {
	var lines []string

	for pos, sc := range scales {
		lines = append(lines, fmt.Sprintf("[%02d] %s", pos, sc.String()))
	}

	sort.Strings(lines)

	return strings.Join(lines, ",")
}

func internalTempoChangesToString(tc map[uint]float64) string {
	var lines []string

	for pos, t := range tc {
		lines = append(lines, fmt.Sprintf("[%02d] @%0.2f", pos, t))
	}

	sort.Strings(lines)

	return strings.Join(lines, ",")
}

func barToString(b *items.Bar) string {
	/*
		&Bar{No: 0, Position: 0, TimeSig: [2]uint8{3, 4}, TimeSigChange: [2]uint8{3, 4}, Comment: "tester#1", Part: "A", TempoChange: 120},
						&Bar{No: 1, Position: 24, TimeSig: [2]uint8{4, 4}, TimeSigChange: [2]uint8{4, 4}, Part: "B", TempoChange: 120},
						&Bar{No: 2, Position: 56, TimeSig: [2]uint8{4, 4}, JumpTo: "A"},
	*/

	var bd strings.Builder
	bd.WriteString(fmt.Sprintf("[%v] #%v @%0.2f", b.Position, b.No, b.TempoChange))
	if b.TimeSigChange[0] > 0 && b.TimeSigChange[1] > 0 {
		bd.WriteString(fmt.Sprintf(" %v/%v", b.TimeSigChange[0], b.TimeSigChange[1]))
	}
	if b.Part != "" {
		bd.WriteString(" [" + b.Part + "]")
	}
	if b.JumpTo != "" {
		bd.WriteString(" [->" + b.JumpTo + "]")
	}

	if b.Include != nil {
		/*
			Include: &items.Include{
							File:        "testinclude",
							Sketch:      "=SCORE",
							Length32ths: 40,
						},
		*/
		bd.WriteString(fmt.Sprintf(" (%q.%q len: %v)", b.Include.File, b.Include.Sketch, b.Include.Length32ths))
	}

	if len(b.InnerTempoChanges) > 0 {
		bd.WriteString(" {" + internalTempoChangesToString(b.InnerTempoChanges) + "}")
	}

	if len(b.InnerScales) > 0 {
		bd.WriteString(" {" + internalScalesToString(b.InnerScales) + "}")
	}

	if b.Comment != "" {
		bd.WriteString(" // " + b.Comment)
	}

	return bd.String()
}
