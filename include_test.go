package muskel

import (
	"strings"
	"testing"
)

func TestInclude(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			`
$include("drumnotes")

=
   | <Voc> | <Drums> | <Piano> |
Ch |     |       |       |
1  | a"  | kd    | b"    |
2& | c   | sn    | d     |

1  |     | ho    |       |
`,
			`


$include("drumnotes")





=
         | <Voc> | <Drums> | <Piano> |
File     |       |         |         |
Ch       |       |         |         |
Bank     |       |         |         |
Prog     |       |         |         |
Vol      |       |         |         |
PbRange  |       |         |         |
Trans    |       |         |         |
Delay    |       |         |         |
    1    | a"    | MN36=:  | b"      |
    2&   | c     | MN40=:  | d       |

    1    |       | MN46=:  |         |
`,
		},
		{
			`
$include("testdata/includes/header")
@other: test
hu: 1sn 2kd

=
   | <Voc> | <Drums> | <Piano> |
Ch |     |       |       |
1  | a"  | kd    | b"    |
2& | c   | sn    | d     |

1  |     | hu    |       |
`,
			`
@other:         test


$include("testdata/includes/header")


hu:             1sn 2kd



=
         | <Voc> | <Drums> | <Piano> |
File     |       |         |         |
Ch       |       |         |         |
Bank     |       |         |         |
Prog     |       |         |         |
Vol      |       |         |         |
PbRange  |       |         |         |
Trans    |       |         |         |
Delay    |       |         |         |
    1    | a"    | e"      | b"      |
    2&   | c     | f"      | d       |

    1    |       | f"      |         |
    2    |       | e"      |         |
`,
		},
		{
			`
=
   | <Voc> | <Drums> | <Piano> |
Ch |     |       |       |
4/4
1  | a"  | d    | b"    |
2& | c   |      | d     |
$include("testdata/includes/score")
1  | d   |      |       | B
`,
			`
=
         | <Voc> | <Drums> | <Piano> |
File     |       |         |         |
Ch       |       |         |         |
Bank     |       |         |         |
Prog     |       |         |         |
Vol      |       |         |         |
PbRange  |       |         |         |
Trans    |       |         |         |
Delay    |       |         |         |
4/4
    1    | a"    | d       | b"      |
    2&   | c     |         | d       |

    1    | e"    |         | f"      | A
    2&   | c     |         | d       |

    1    | d     |         |         | B
`,
		},
		{
			`
=
   | <Voc> | <Drums> | <Piano> |
Ch |     |       |       |
$include("testdata/includes/score")
1  | d   |      |       | B
`,
			`
=
         | <Voc> | <Drums> | <Piano> |
File     |       |         |         |
Ch       |       |         |         |
Bank     |       |         |         |
Prog     |       |         |         |
Vol      |       |         |         |
PbRange  |       |         |         |
Trans    |       |         |         |
Delay    |       |         |         |
    1    | e"    |         | f"      | A
    2&   | c     |         | d       |

    1    | d     |         |         | B
`,
		},
		{
			`
=
   | <Voc> | <Drums> | <Piano> |
Ch |     |       |       |
$include("testdata/includes/score")
$include("testdata/includes/score")
`,
			`
=
         | <Voc> | <Drums> | <Piano> |
File     |       |         |         |
Ch       |       |         |         |
Bank     |       |         |         |
Prog     |       |         |         |
Vol      |       |         |         |
PbRange  |       |         |         |
Trans    |       |         |         |
Delay    |       |         |         |
    1    | e"    |         | f"      | A
    2&   | c     |         | d       |

    1    | e"    |         | f"      | A
    2&   | c     |         | d       |
`,
		},
		{
			`
=
   | <Voc> | <Drums> | <Piano> |
Ch |     |       |       |
4/4
1  | a"  | d    | b"    |
2& | c   |      | d     |
$include("testdata/includes/score")
$include("testdata/includes/score")
1  | d   |      |       | B
`,
			`
=
         | <Voc> | <Drums> | <Piano> |
File     |       |         |         |
Ch       |       |         |         |
Bank     |       |         |         |
Prog     |       |         |         |
Vol      |       |         |         |
PbRange  |       |         |         |
Trans    |       |         |         |
Delay    |       |         |         |
4/4
    1    | a"    | d       | b"      |
    2&   | c     |         | d       |

    1    | e"    |         | f"      | A
    2&   | c     |         | d       |

    1    | e"    |         | f"      | A
    2&   | c     |         | d       |

    1    | d     |         |         | B
`,
		},
	}

	for i, test := range tests {
		/*
			if i != 2 {
				continue
			}
		*/
		sc, err := Parse(strings.NewReader(strings.TrimSpace(test.input)), "include-main")

		if err != nil {
			t.Errorf("[%v] could not parse score: %s\n%s\n", i, err.Error(), test.input)
			continue
		}

		unr, err := sc.Unroll()

		if err != nil {
			t.Errorf("[%v] could not unroll score: %s\n%s\n", i, err.Error(), test.input)
			continue
		}

		var bf strings.Builder

		err = unr.WriteTo(&bf)

		if err != nil {
			t.Errorf("[%v] could not format unrolled score: %s\n%s\n", i, err.Error(), test.input)
			continue
		}

		result := bf.String()

		//		fmt.Println(result)

		//res := strings.Split(result, "=\n")
		//exp := strings.Split(test.expected, "=\n")

		got := strings.TrimSpace(result)
		wanted := strings.TrimSpace(test.expected)

		if got != wanted {
			t.Errorf("[%v] score\n%s\n\nunrolled gives \n%s\n\nbut this was expected:\n%s\n%q\nvs\n%q\n", i, test.input, got, wanted, got, wanted)
			//			t.Errorf("[%v] score\n%s\n\nunrolled gives \n%q\n\nbut this was expected:\n%q\n", i, test.input, got, wanted)
		}
	}
}

func TestInclude2(t *testing.T) {
	//	t.Skip()
	tests := []struct {
		input    string
		expected string
	}{
		{
			`
$include("drumnotes")

=
   | <Voc> | <Drums> | <Piano> |
Ch |     |       |       |
1  | a"  | kd    | b"    |
2& | c   | sn    | d     |

1  |     | ho    |       |
`,
			`


$include("drumnotes")





=
         | <Voc> | <Drums> | <Piano> |
File     |       |         |         |
Ch       |       |         |         |
Bank     |       |         |         |
Prog     |       |         |         |
Vol      |       |         |         |
PbRange  |       |         |         |
Trans    |       |         |         |
Delay    |       |         |         |
    1    | a"    | kd      | b"      |
    2&   | c     | sn      | d       |

    1    |       | ho      |         |
`,
		},
		{
			`
$include("testdata/includes/header")
@other: test
hu: 1sn 2kd

=
   | <Voc> | <Drums> | <Piano> |
Ch |     |       |       |
1  | a"  | kd    | b"    |
2& | c   | sn    | d     |

1  |     | hu    |       |
`,
			`
@other:         test


$include("testdata/includes/header")


hu:             1sn 2kd



=
         | <Voc> | <Drums> | <Piano> |
File     |       |         |         |
Ch       |       |         |         |
Bank     |       |         |         |
Prog     |       |         |         |
Vol      |       |         |         |
PbRange  |       |         |         |
Trans    |       |         |         |
Delay    |       |         |         |
    1    | a"    | kd      | b"      |
    2&   | c     | sn      | d       |

    1    |       | hu      |         |
`,
		},
		{
			`
=
   | <Voc> | <Drums> | <Piano> |
Ch |     |       |       |
4/4
1  | a"  | d    | b"    |
2& | c   |      | d     |
$include("testdata/includes/score")
1  | d   |      |       |
`,
			`
=
         | <Voc> | <Drums> | <Piano> |
File     |       |         |         |
Ch       |       |         |         |
Bank     |       |         |         |
Prog     |       |         |         |
Vol      |       |         |         |
PbRange  |       |         |         |
Trans    |       |         |         |
Delay    |       |         |         |
4/4
    1    | a"    | d       | b"      |
    2&   | c     |         | d       |
$include("testdata/includes/score")
    1    | d     |         |         |
`,
		},
		{
			`
=
   | <Voc> | <Drums> | <Piano> |
Ch |     |       |       |
4/4
1  | a"  | d    | b"    |
2& | c   |      | d     |
$include("testdata/includes/score")
$include("testdata/includes/score")
1  | d   |      |       |
`,
			`
=
         | <Voc> | <Drums> | <Piano> |
File     |       |         |         |
Ch       |       |         |         |
Bank     |       |         |         |
Prog     |       |         |         |
Vol      |       |         |         |
PbRange  |       |         |         |
Trans    |       |         |         |
Delay    |       |         |         |
4/4
    1    | a"    | d       | b"      |
    2&   | c     |         | d       |
$include("testdata/includes/score")
$include("testdata/includes/score")
    1    | d     |         |         |
`,
		},
	}

	for i, test := range tests {
		/*
			if i != 0 {
				continue
			}
		*/
		sc, err := Parse(strings.NewReader(strings.TrimSpace(test.input)), "include main")

		if err != nil {
			t.Errorf("[%v] could not parse score: %s\n%s\n", i, err.Error(), test.input)
			continue
		}

		var bf strings.Builder

		err = sc.WriteTo(&bf)

		if err != nil {
			t.Errorf("[%v] could not format score: %s\n%s\n", i, err.Error(), test.input)
			continue
		}

		result := bf.String()

		//		fmt.Println(result)

		//res := strings.Split(result, "=\n")
		//exp := strings.Split(test.expected, "=\n")

		got := strings.TrimSpace(result)
		wanted := strings.TrimSpace(test.expected)

		if got != wanted {
			t.Errorf("[%v] score\n%s\nformatted gives \n%s\n\nbut this was expected:\n%s\n%q\nvs\n%q\n", i, test.input, got, wanted, got, wanted)
			//			t.Errorf("[%v] score\n%s\n\nunrolled gives \n%q\n\nbut this was expected:\n%q\n", i, test.input, got, wanted)
		}
	}
}
