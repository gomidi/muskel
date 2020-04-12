package sketch

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"gitlab.com/gomidi/muskel/items"
	"gitlab.com/gomidi/muskel/track"
)

/*
ein kleines Modell zum testen


KLÄRUNG / REGELN

1. nur Templates können overrides sein
2. override templates enden, sobald sie enden, oder sobald irgendein event kommt (template oder nicht)
3. nicht override templates enden wie (2), außer wenn ein override template kommt, dann laufen sie weiter

*/

func templStr(start uint, s string) (p *positionedString) {
	p = &positionedString{}
	p.start = start
	p.s = s
	p.isTemplate = true
	p.isOverride = false
	return
}

func overStr(start uint, s string) (p *positionedString) {
	p = &positionedString{}
	p.start = start
	p.s = s
	p.isTemplate = true
	p.isOverride = true
	return
}

func blockStr(start uint, s string) (p *positionedString) {
	p = &positionedString{}
	p.start = start
	p.s = s
	p.isTemplate = false
	p.isOverride = false
	return
}

func resultStr(start uint, s string) (p *positionedString) {
	p = &positionedString{}
	p.start = start
	p.s = s
	p.isTemplate = false
	p.isOverride = false
	return
}

// positionedString simulates a sketch column of events
type positionedString struct {
	start      uint   // start position of the sketch column
	s          string // the events (each character is an event)
	isOverride bool   // simulates wether the sketch column call is an override
	isTemplate bool
}

func (p *positionedString) String() string {
	return fmt.Sprintf("{%v %v %v %q}", p.start, p.isTemplate, p.isOverride, p.s)
}

// end simulates the end position in absolute 32ths
func (p positionedString) end() uint {
	return p.start + uint(len(p.s))
}

func (p positionedString) inRange(from, to uint) bool {
	return p.start <= to && p.end() > from
}

func (l *positionedString) getEvents(from, to uint) string {
	if debug {
		fmt.Printf("[getEvents] from: %v to: %v\n", from, to)
	}

	diffStart := int(from - l.start)
	if diffStart < 0 {
		diffStart = 0
	}
	var diffEnd int
	if to < l.end() {
		diffEnd = int(to) - int(from)
		if from < l.start {
			diffEnd -= (int(l.start) - int(from))
		}
	}

	if debug {
		fmt.Printf("l.pos: %v l.end: %v diffStart: %v diffEnd: %v\n", l.start, l.end(), diffStart, diffEnd)
	}

	if diffEnd > 0 {
		return l.s[diffStart : diffStart+diffEnd]
	}
	return l.s[diffStart:]
}

var debug bool

func mergeStrings(mixed []*positionedString) (result []*positionedString) {
	var lastOPos uint
	var keepLooping bool = true
	var endPos uint //   mixed[len(mixed)-1].end() // start + uint(len(mixed[len(mixed)-1].s))
	var lastTemplateX *positionedString
	var lastTemplateIdx int

	lastTemplateX = nil

	for idx, o := range mixed {
		if o.end() > endPos {
			endPos = o.end()
		}

		if keepLooping && lastTemplateX != nil {
			//	fmt.Printf("we have a lastTemplate: %#v\n", lastTemplateX)
			if lastTemplateX.inRange(lastOPos, o.start) {
				var rrr = &positionedString{}
				if rrr.start == 0 {
					rrr.start = lastTemplateX.start
					if lastOPos > rrr.start {
						rrr.start = lastOPos
					}
				}

				if debug {
					fmt.Printf("lastTemplateIdx [%v]\n", lastTemplateIdx)
				}

				var toMod = o.start

				if idx+1 < len(mixed) {
					if debug {
						fmt.Printf("has [%v]\n", idx+1)
					}
					if mixed[idx+1].start <= o.start {
						toMod = mixed[idx+1].start
					}
				}

				rrr.s = string(lastTemplateX.getEvents(lastOPos, toMod))

				result = append(result, rrr)
			}
		}

		if o.isTemplate && !o.isOverride {
			keepLooping = true
			lastTemplateIdx = idx
			lastTemplateX = o
			//	fmt.Printf("setting lastTemplate to: %#v\n", lastTemplateX)
			continue
		}

		if !o.isTemplate {
			keepLooping = false
			lastTemplateX = nil
		}

		result = append(result, o)
		lastOPos = o.end()
	}

	//fmt.Printf("keepLooping: %v lastLoopedIdx: %v endPos: %v\n", keepLooping, lastLoopedIdx, endPos)

	if keepLooping && lastTemplateX != nil {
		if lastTemplateX.inRange(lastOPos, endPos) {
			var r = &positionedString{}
			if r.start == 0 {
				r.start = lastTemplateX.start
				if lastOPos > r.start {
					r.start = lastOPos
				}
			}

			r.s = lastTemplateX.getEvents(lastOPos, endPos)

			result = append(result, r)
		}
	}
	return
}

func TestMergeStrings(t *testing.T) {

	tests := []struct {
		mixed    []*positionedString
		expected []*positionedString
	}{
		{ // 0
			mixed: []*positionedString{
				templStr(2, "ABC-ABC-ABC-ABC-ABC"),
				overStr(7, "XYX/XX"),
			},

			expected: []*positionedString{
				resultStr(2, "ABC-A"),
				overStr(7, "XYX/XX"),
				resultStr(13, "-ABC-ABC"),
			},
		},
		{ // 1
			mixed: []*positionedString{
				templStr(2, "ABC-DEF-GHI-JKL-MNO"),
				overStr(7, "XYX/XX"),
			},

			expected: []*positionedString{
				resultStr(2, "ABC-D"),
				overStr(7, "XYX/XX"),
				resultStr(13, "-JKL-MNO"),
			},
		},
		{ // 2
			mixed: []*positionedString{
				templStr(2, "ABC-DEF-GHI-JKL-MNO"),
				blockStr(7, "XYX/XX"),
			},

			expected: []*positionedString{
				resultStr(2, "ABC-D"),
				blockStr(7, "XYX/XX"),
			},
		},
		{ // 3
			mixed: []*positionedString{
				templStr(2, "ABC-DEF-GHI-JKL-MNO"),
				overStr(7, "XX"),
				overStr(11, "YY"),
			},

			expected: []*positionedString{
				resultStr(2, "ABC-D"),
				overStr(7, "XX"),
				resultStr(9, "-G"),
				overStr(11, "YY"),
				resultStr(13, "-JKL-MNO"),
			},
		},
		{ // 4

			mixed: []*positionedString{
				templStr(2, "ABC-DEF-GHI-JKL-MNO"),
				overStr(7, "XX"),
				blockStr(11, "YY"),
			},

			expected: []*positionedString{
				resultStr(2, "ABC-D"),
				overStr(7, "XX"),
				resultStr(9, "-G"),
				blockStr(11, "YY"),
			},
		},

		{ // 5

			mixed: []*positionedString{
				templStr(2, "ABC-DEF-GHI-JKL-MNO"),
				blockStr(7, "XX"),
				overStr(11, "YY"),
			},

			expected: []*positionedString{
				resultStr(2, "ABC-D"),
				blockStr(7, "XX"),
				overStr(11, "YY"),
			},
		},

		{ // 6
			mixed: []*positionedString{
				templStr(2, "ABC-DEF-GHI-JKL-MNO"),
				overStr(7, "XX"),
				overStr(11, "YY"),
				overStr(15, "ZZ"),
			},

			expected: []*positionedString{
				resultStr(2, "ABC-D"),
				overStr(7, "XX"),
				resultStr(9, "-G"),
				overStr(11, "YY"),
				resultStr(13, "-J"),
				overStr(15, "ZZ"),
				resultStr(17, "-MNO"),
			},
		},

		{ // 7
			mixed: []*positionedString{
				templStr(2, "ABC-DEF-GHI-JKL-MNO"),
				overStr(7, "XX"),
				overStr(11, "YY"),
				blockStr(15, "ZZ"),
			},

			expected: []*positionedString{
				resultStr(2, "ABC-D"),
				overStr(7, "XX"),
				resultStr(9, "-G"),
				overStr(11, "YY"),
				resultStr(13, "-J"),
				blockStr(15, "ZZ"),
			},
		},
		{ // 8
			mixed: []*positionedString{
				templStr(2, "ABC-DEF-GHI-JKL-MNO"),
				overStr(7, "XX"),
				blockStr(11, "YY"),
				overStr(15, "ZZ"),
			},

			expected: []*positionedString{
				resultStr(2, "ABC-D"),
				overStr(7, "XX"),
				resultStr(9, "-G"),
				blockStr(11, "YY"),
				overStr(15, "ZZ"),
			},
		},
		{ // 9
			mixed: []*positionedString{
				overStr(0, "XX"),
				templStr(2, "ABC-DEF-GHI-JKL-MNO"),
				blockStr(11, "YY"),
				overStr(15, "ZZ"),
			},

			expected: []*positionedString{
				overStr(0, "XX"),
				resultStr(2, "ABC-DEF-G"),
				blockStr(11, "YY"),
				overStr(15, "ZZ"),
			},
		},
		{ // 10
			mixed: []*positionedString{
				overStr(0, "XX"),
				templStr(2, "ABC-DEF-GHI-JKL-MNO"),
				overStr(11, "YY"),
				overStr(15, "ZZ"),
			},

			expected: []*positionedString{
				overStr(0, "XX"),
				resultStr(2, "ABC-DEF-G"),
				overStr(11, "YY"),
				resultStr(13, "-J"),
				overStr(15, "ZZ"),
				resultStr(17, "-MNO"),
			},
		},
		{ // 11
			mixed: []*positionedString{
				overStr(0, "XX"),
				templStr(2, "ABC-DP"),
				templStr(7, "EF-GHI-JKL-MNO"),
				overStr(11, "YY"),
				overStr(15, "ZZ"),
			},

			expected: []*positionedString{
				overStr(0, "XX"),
				resultStr(2, "ABC-D"),
				resultStr(7, "EF-G"),
				overStr(11, "YY"),
				resultStr(13, "-J"),
				overStr(15, "ZZ"),
				resultStr(17, "-MNO"),
			},
		},
		{ // 12
			mixed: []*positionedString{
				overStr(0, "XX"),
				templStr(2, "ABC-DP"),
				overStr(11, "YY"),
				templStr(13, "EF-GHI-JKL-MNO"),
				overStr(15, "ZZ"),
			},

			expected: []*positionedString{
				overStr(0, "XX"),
				resultStr(2, "ABC-DP"),
				overStr(11, "YY"),
				resultStr(13, "EF"),
				overStr(15, "ZZ"),
				resultStr(17, "HI-JKL-MNO"),
			},
		},
		{ // 13
			mixed: []*positionedString{
				overStr(0, "XX"),
				templStr(2, "ABC-DP"),
				blockStr(11, "YY"),
				templStr(13, "EF-GHI-JKL-MNO"),
				overStr(15, "ZZ"),
			},

			expected: []*positionedString{
				overStr(0, "XX"),
				resultStr(2, "ABC-DP"),
				blockStr(11, "YY"),
				resultStr(13, "EF"),
				overStr(15, "ZZ"),
				resultStr(17, "HI-JKL-MNO"),
			},
		},
		/*
		 */
	}

	for i, test := range tests {

		// debug = i == 11

		/*
			if i > 3 {
				continue
			}
		*/

		//		result := mergeStrings(test.looped, test.other)
		result := mergeStrings(test.mixed)

		if !reflect.DeepEqual(result, test.expected) {
			//t.Errorf("[%v] \nmergeStrings(\n\t%v,\n\t%v\n) =     %v \n// exp. %v", i, test.looped, test.other, result, test.expected)
			t.Errorf("[%v] \nmergeStrings(\n\t%v\n) =     %v \n// exp. %v", i, test.mixed, result, test.expected)
		}
	}

}

type testScore struct {
	include   *Sketch
	getSketch func(name string) *Sketch
}

func (t testScore) Embed(p ...string) error {
	return nil
}

func (t testScore) Lyric(s string, from, to int) ([]string, error) {
	return nil, nil
}

func (t testScore) FilterTrack(s string, ev []*items.Event) []*items.Event {
	return ev
}

func (t testScore) GetIncludedSketch(a, b string, params []string) (*Sketch, error) {
	if t.include != nil && t.include.Name == a {
		return t.include, nil
	}
	return nil, fmt.Errorf("included sketch %q not found in file %q", b, a)
}

func (t *testScore) GetSketch(name string) (*Sketch, error) {
	return t.getSketch(name), nil
}

func (t *testScore) GetMode(name string) items.Mode {
	return nil
}

func (t *testScore) GetToken(name string) (string, error) {
	return "", fmt.Errorf("token %q not found", name)
}

func (t *testScore) GetTrack(name string) (*track.Track, error) {
	return nil, fmt.Errorf("track %q not found", name)
}

func (t *testScore) HasTrack(name string) bool {
	return false
}

func (t *testScore) Properties() map[string]interface{} {
	return nil
}

func TestCall(t *testing.T) {
	//t.Skip()
	var s = New("sk", nil)
	s.colOrder = []string{"col1", "col2"}
	b := items.NewBar()
	b.TimeSig = [2]uint8{4, 4}
	s.newBar(b)
	sc := &testScore{}
	sc.getSketch = func(name string) *Sketch {
		if name == "=sk" {
			return s
		}
		return nil
	}
	s.Score = sc
	pos := strings.Split("1\n1&\n2", "\n")

	for _, p := range pos {
		_, _p, _ := items.PositionTo32th(0, p)
		s.Positions = append(s.Positions, [2]uint{0, _p})
	}

	tests := []struct {
		col1     string // complete positions, lines separated by \n
		col2     string // complete col, lines separated by \n
		params   string // complete params, separated by ,
		expected string
	}{
		{"||", "c|d|e", "x,y", "1c|1&d|2e"},
		{"||", "#1|d|#1", "g,b", "1g|1&d|2g"},
		{"||", "#2|d|#1", "g,b", "1b|1&d|2g"},
		{"||", "{#2,#1}|d|#1", "g,b", "1{b,g}|1&d|2g"},
		{"||", "||", "g,b", ""},
		//{"||", "=sk.col2||", "g,b", ""}, // TODO make a sketch name a type and check for same sketchtable and same column
		{"a|b|c", "=sk.col1||", "g,b", "1a|1&b|2c"},
		{"a|b|c", "=sk.col1||#2", "g,b", "1a|1&b|2b"},
		{"a|#1|c", "=sk.col1(#2)||#2", "g,f", "1a|1&f|2f"},
		{"a|b|c", "=sk.col1||#2", "g,b", "1a|1&b|2b"},
		{"a|b|c", "=sk.col1|#1|", "g,b", "1a|1&g"},
		{"a|b|c", "=sk.col1|#1|d", "g,b", "1a|1&g|2d"},
		{"a|b|c", "d|=sk.col1|", "g,b", "1d|1&a|2b|2&c"},
		{"a|b|c", "d|=sk.col1|#1", "g,b", "1d|1&a|2g"},
		{"#1|#1|#1", "=sk.col1(#1)|=sk.col1(#2)|", "g,b", "1g|1&b|2b|2&b"},
		{"#1|#1|#1", "=sk.col1(#1)|=sk.col1(#2)[:2]|", "g,b", "1g|1&b|2b"},
		{"#1|#1|#1", "=sk.col1(#1)|=sk.col1(#2)[:1]|", "g,b", "1g|1&b"},
		{"#1|#1|#1", "=sk.col1(#1)|/=sk.col1(#2)[:1]|", "g,b", "1g|1&b"},
		{"|b|c", "=sk.col1||", "", "1&b|2c"},
		{"|b|c", "=!sk.col1|", "", "1b|1&c"},
		{"|b|c", "=!sk.col1||", "", "1b|1&c"},
		{"|b|c", "||=!sk.col1", "", "2b|2&c"},
		{"a||", "=sk.col1|*|", "", "1a|1&*"},
		{"a|*|", "=sk.col1|*|", "", "1a|1&*"},
		{"a|*|", "|=sk.col1|*|", "", "1&a|2*"},
		{"|a|", "|=sk.col1||", "", "2a"},
		/*
		 */
		{"|a|", "|=!sk.col1|*|", "", "1&a|2*"},
	}
	// TODO prevent endless loops from templates calling each other like col1 -> col2 -> col1 by keeping a stack of template calls
	// and checking them for duplicates

	//eventDebug = true

	for i, test := range tests {

		/*
			if i != 20 {
				continue
			}
		*/

		//fmt.Printf("-----------%v--------------------------------\n", i)

		s.Columns["col1"] = strings.Split(test.col1, "|")
		s.Columns["col2"] = strings.Split(test.col2, "|")
		params := strings.Split(test.params, ",")
		var patt column
		patt.sketch = s
		patt.name = "col2"
		evt, _, err := patt.Call(0, false, params...)

		if err != nil {
			t.Errorf("[%v](%q, %q).Call(\"col2\",0,%s) returns error: %s ", i, test.col1, test.col2, test.params, err.Error())
			continue
		}

		//fmt.Printf("evts: %v\n", evt)

		var bd strings.Builder

		for _, ev := range evt {
			bd.WriteString(items.Pos32thToString(ev.Position))
			bd.WriteString(ev.String() + "|")
		}

		res := bd.String()

		if len(res) > 1 {
			res = res[:len(res)-1]
		}

		if res != test.expected {
			t.Errorf("[%v](%q, %q).Call(\"col2\",0,%s) = %q // expected %q ", i, test.col1, test.col2, test.params, res, test.expected)
		}
	}
}

func TestReplaceParams(t *testing.T) {
	params := []string{"1", "2", "3", "4"}
	tests := []struct {
		input    string
		expected string
	}{
		{"a #1 #3 #2 #6", "a 1 3 2 2"},
		{"a #10", "a 2"},
	}

	for i, test := range tests {
		result := items.ReplaceParams(test.input, params)

		if result != test.expected {
			t.Errorf("[%v] replaceParams(%q, %v) = %q // expected %q ", i, test.input, params, result, test.expected)
		}
	}
}

func TestInjectParams(t *testing.T) {
	col := []string{
		"a #1 #3 #2 #6",
		"a #10",
	}
	params := []string{"1", "2", "3", "4"}

	var s Sketch

	res := s.injectParams(col, params)

	expected := []string{"a 1 3 2 2", "a 2"}

	if !reflect.DeepEqual(res, expected) {
		t.Errorf("injectParams(%v,%v) = %v // expected %v", col, params, res, expected)
	}
}
