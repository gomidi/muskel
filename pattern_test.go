package muskel

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
)

func TestParseFragment(t *testing.T) {
	//	t.Skip()
	tests := []struct {
		input    string
		expected patternFragment
	}{
		{"2a'", patternFragment{"2", "a'"}},
		{"a'", patternFragment{"", "a'"}},
		{"2", patternFragment{"2", ""}},
		{"2&", patternFragment{"2&", ""}},
		{"&", patternFragment{"&", ""}},
		{".", patternFragment{".", ""}},
		{";", patternFragment{";", ""}},
		{"2;", patternFragment{"2;", ""}},
		{"2.;", patternFragment{"2.;", ""}},
		{"2.", patternFragment{"2.", ""}},
		{"&.", patternFragment{"&.", ""}},
		{"&;", patternFragment{"&;", ""}},
		{"&.;", patternFragment{"&.;", ""}},

		{"2&a", patternFragment{"2&", "a"}},
		{"&a", patternFragment{"&", "a"}},
		{".a", patternFragment{".", "a"}},
		{";a", patternFragment{";", "a"}},
		{"2;a", patternFragment{"2;", "a"}},
		{"2.;a", patternFragment{"2.;", "a"}},
		{"2.a", patternFragment{"2.", "a"}},
		{"&.a", patternFragment{"&.", "a"}},
		{"&;a", patternFragment{"&;", "a"}},
		{"&.;a", patternFragment{"&.;", "a"}},
	}

	for i, test := range tests {
		var f patternFragment

		f.parse(test.input)

		if !reflect.DeepEqual(f, test.expected) {
			t.Errorf("[%v] parse(%q) resulted in %#v // expected: %#v", i, test.input, f, test.expected)
		}
	}
}

func TestUnrollPattern(t *testing.T) {
	//t.Skip()
	tests := []struct {
		input    string
		expected string
	}{
		{`
a_a:           f" 2g



=
         | Vocals |
Ch       | 1      |
Bank     |        |
Prog     |        |
Vol      |        |
    1    | e'     |
    2    | a_a     |

    1    | %      |
`,
			`
a_a:            f" 2g



=
         | Vocals |
Ch       | 1      |
Bank     |        |
Prog     |        |
Vol      |        |
PbRange  |        |
Trans    |        |
    1    | e'     |
    2    | f"     |
    3    | g      |

    1    | g      |
`,
		},
		{`
aa:           f" 2g



=
         | Vocals |
Ch       | 1      |
Bank     |        |
Prog     |        |
Vol      |        |
    1    | aa     |

    1    | ./.    |
`,
			`
aa:             f" 2g



=
         | Vocals |
Ch       | 1      |
Bank     |        |
Prog     |        |
Vol      |        |
PbRange  |        |
Trans    |        |
    1    | f"     |
    2    | g      |

    1    | f"     |
    2    | g      |
`,
		},
		{`
aa:           1f" 2g



=
         | Vocals |
Ch       | 1      |
Bank     |        |
Prog     |        |
Vol      |        |
    1    | e'     |
    2    | aa     |

    1    | %      |
`,
			`
aa:             1f" 2g



=
         | Vocals |
Ch       | 1      |
Bank     |        |
Prog     |        |
Vol      |        |
PbRange  |        |
Trans    |        |
    1    | e'     |
    2    | f"     |
    3    | g      |

    1    | g      |
`,
		},
		{
			`
aa: 1a' 2&#1 5&c"
bb~: 1aa(#1) 5aa(d)
=
  | piano | vox |
1 | aa(g) | bb~(f)   |

1 | ./.   |    |
`, `



aa:             1a' 2&#1 5&c"
bb~:            1aa(#1) 5aa(d)



=
         | piano | vox |
Ch       |       |     |
Bank     |       |     |
Prog     |       |     |
Vol      |       |     |
PbRange  |       |     |
Trans    |       |     |
    1    | a'    | a'  |
    2&   | g     | f   |

    1    | a'    | a'  |
    2&   | g     | d   |
			`},
		{
			`
aa:  1a' 2&#1 5&c"
=
  | piano | vox |
1 | aa(g) | aa(f)   |

1 | ./.   |    |
`, `



aa:             1a' 2&#1 5&c"



=
         | piano | vox |
Ch       |       |     |
Bank     |       |     |
Prog     |       |     |
Vol      |       |     |
PbRange  |       |     |
Trans    |       |     |
    1    | a'    | a'  |
    2&   | g     | f   |

    1    | a'    |     |
    1&   |       | c"  |
    2&   | g     |     |
			`},
		{
			`
aa:  1a' 2&b 5&c"
=
  | piano | vox |
1 | aa |    |
2 |    |  E  |

1 | ./.   | F|
`, `



aa:             1a' 2&b 5&c"



=
         | piano | vox |
Ch       |       |     |
Bank     |       |     |
Prog     |       |     |
Vol      |       |     |
PbRange  |       |     |
Trans    |       |     |
    1    | a'    |     |
    2    |       | E   |
    2&   | b     |     |

    1    | a'    | F   |
    2&   | b     |     |
			`},

		{
			`
aa:  1a' 2&b 5&c"
=
  | piano | vox |
1 | aa |    |
2 |    |  E  |

1 |    | F|
`, `



aa:             1a' 2&b 5&c"



=
         | piano | vox |
Ch       |       |     |
Bank     |       |     |
Prog     |       |     |
Vol      |       |     |
PbRange  |       |     |
Trans    |       |     |
    1    | a'    |     |
    2    |       | E   |
    2&   | b     |     |

    1    |       | F   |
    1&   | c"    |     |
			`},

		{
			`
aa:  1a' 2&b 5g
=
  | piano |
1 | aa |

1 |  |
`, `



aa:             1a' 2&b 5g



=
         | piano |
Ch       |       |
Bank     |       |
Prog     |       |
Vol      |       |
PbRange  |       |
Trans    |       |
    1    | a'    |
    2&   | b     |

    1    | g     |
		   		   `},
		{
			`
aa:  1a' 2&b 5g
=
  | piano |
1 | aa |

1 | e |
`, `



aa:             1a' 2&b 5g



=
         | piano |
Ch       |       |
Bank     |       |
Prog     |       |
Vol      |       |
PbRange  |       |
Trans    |       |
    1    | a'    |
    2&   | b     |

    1    | e     |
		   `},
		{
			`
aa:  1a' 2&b 5g
=
  | piano |
1 | aa |
`, `



aa:             1a' 2&b 5g



=
         | piano |
Ch       |       |
Bank     |       |
Prog     |       |
Vol      |       |
PbRange  |       |
Trans    |       |
    1    | a'    |
    2&   | b     |
		   		   `},
		{
			`
aa:  1a' 4&b 5&c"
=
  | piano |
3/4
1 | aa |

1& |    |
`, `



aa:             1a' 4&b 5&c"



=
         | piano |
Ch       |       |
Bank     |       |
Prog     |       |
Vol      |       |
PbRange  |       |
Trans    |       |
3/4
    1    | a'    |

    1&   | b     |
    2&   | c"    |
		   `},

		{
			`
aa:  1a' 2&b 5&c"
=
  | piano | vox |
1 | aa |   |
2 |    | aa |

1 |    |  |
`, `



aa:             1a' 2&b 5&c"



=
         | piano | vox |
Ch       |       |     |
Bank     |       |     |
Prog     |       |     |
Vol      |       |     |
PbRange  |       |     |
Trans    |       |     |
    1    | a'    |     |
    2    |       | a'  |
    2&   | b     |     |
    3&   |       | b   |

    1    |       |     |
    1&   | c"    |     |
    2&   |       | c"  |
		   		   `},

		{
			`
aa:  1a' 2b
=
  | piano |
2 | aa |
`, `



aa:             1a' 2b



=
         | piano |
Ch       |       |
Bank     |       |
Prog     |       |
Vol      |       |
PbRange  |       |
Trans    |       |
    2    | a'    |
    3    | b     |
		   `},
		{
			`
chords:         chords_swing(#1,#2)/:,=#1:=#2:,:,_,:,:/
chords_swing:   1&=#1:=#2: 2&=#1=#2   2&._ 2&.;=#1:=#2: 3&=#1:=#2: 4&=#1:=#2:
=
  | piano |
1 | chords(^1,^3) |
`, `



chords:         chords_swing(#1,#2)/:,=#1:=#2:,:,_,:,:/
chords_swing:   1&=#1:=#2: 2&=#1=#2   2&._ 2&.;=#1:=#2: 3&=#1:=#2: 4&=#1:=#2:



=
         | piano    |
Ch       |          |
Bank     |          |
Prog     |          |
Vol      |          |
PbRange  |          |
Trans    |          |
    1    |          |
    1&   | =c':=e': |
    2&   | =c':=e': |
    2&.  | _        |
    2&.; | _        |
    3&   | =c':=e': |
    4&   | =c':=e': |
		   `},
	}

	for i, test := range tests {
		// fmt.Printf("[%v]\n", i)
		if i != 13 {
			continue
		}

		sc, err := Parse(strings.NewReader(strings.TrimSpace(test.input)), "unroll pattern")

		if err != nil {
			t.Errorf("[%v] could not parse score: %s\n%s\n", i, err.Error(), test.input)
			continue
		}
		//		sc.SmallColumns = true

		unr, err := sc.Unroll()

		if err != nil {
			t.Errorf("[%v] could not unroll score: %s\n%s\n", i, err.Error(), test.input)
			continue
		}

		var bf bytes.Buffer

		_, err = unr.WriteTo(&bf)

		if err != nil {
			t.Errorf("[%v] could not format unrolled score: %s\n%s\n", i, err.Error(), test.input)
			continue
		}

		result := strings.TrimSpace(bf.String())
		expected := strings.TrimSpace(test.expected)

		//		fmt.Println(result)

		if result != expected {
			t.Errorf("[%v] score\n%s\n\nunrolled gives \n%s\n\nbut this was expected:\n%s\n%q\nvs\n%q\n", i, test.input, result, expected, result, expected)
			//			t.Errorf("[%v] score\n%s\n\nunrolled gives \n%q\n\nbut this was expected:\n%q\n", i, test.input, got, wanted)
		}
	}
}
