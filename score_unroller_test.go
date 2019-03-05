package muskel

import (
	"bytes"
	"strings"
	"testing"
)

func TestUnroll(t *testing.T) {
	//t.Skip()
	tests := []struct {
		input    string
		expected string
	}{
		{
			`
test: out of range
=
         | Vocals |
Ch       | 1      |
Bank     | -      |
Prog     | -      |
Vol      | -      |
    1    | e'     |

    1    |        |`,
			`
=
         | Vocals |
Ch       | 1      |
Bank     | -      |
Prog     | -      |
Vol      | -      |
PbRange  |        |
Trans    |        |
    1    | e'     |

    1    |        |`,
		},
		{
			`
=
         | Vocals | Piano | 
Ch       | -      | -     |
Bank     | -      | -     |
Prog     | -      | -     |
Vol      | -      | -     |
    3    |        |       |
     &   |        |       |
`,
			`
=
         | Vocals | Piano |
Ch       | -      | -     |
Bank     | -      | -     |
Prog     | -      | -     |
Vol      | -      | -     |
PbRange  |        |       |
Trans    |        |       |
    3    |        |       |
     &   |        |       |
`,
		},

		{
			`
=
   |piano|
1  |a'   |
						`,
			`
=
         | piano |
Ch       | -     |
Bank     | -     |
Prog     | -     |
Vol      | -     |
PbRange  |       |
Trans    |       |
    1    | a'    |
			`,
		},
		{
			`
=
   |piano|
1  |a'   |
&  |c"|
						`,
			`
=
         | piano |
Ch       | -     |
Bank     | -     |
Prog     | -     |
Vol      | -     |
PbRange  |       |
Trans    |       |
    1    | a'    |
     &   | c"    |
			`,
		},

		{
			`
=
   |piano|
1  |a'   |
&  |c"|

2  |d |
						`,
			`
=
         | piano |
Ch       | -     |
Bank     | -     |
Prog     | -     |
Vol      | -     |
PbRange  |       |
Trans    |       |
    1    | a'    |
     &   | c"    |

    2    | d     |
			`,
		},
		{
			`
=
   |piano|
1  |a'   |
&  |c"|

1  |./. |
						`,
			`
=
         | piano |
Ch       | -     |
Bank     | -     |
Prog     | -     |
Vol      | -     |
PbRange  |       |
Trans    |       |
    1    | a'    |
     &   | c"    |

    1    | a'    |
    1&   | c"    |
			`,
		},
		{
			`
=
   |piano|
1  |a'   |
&  |c"|

1  |... |

1  | |

1  |b |
						`,
			`
=
         | piano |
Ch       | -     |
Bank     | -     |
Prog     | -     |
Vol      | -     |
PbRange  |       |
Trans    |       |
    1    | a'    |
     &   | c"    |

    1    | a'    |
    1&   | c"    |

    1    | a'    |
    1&   | c"    |

    1    | b     |
			`,
		},

		{
			`
=
   |piano|
1  |a'   |
&  |c"|

1  |d'|
2& |e"|

1  |.2. |

1  | |

1  | |

1  |b |
		   		   		   			`,
			`
=
         | piano |
Ch       | -     |
Bank     | -     |
Prog     | -     |
Vol      | -     |
PbRange  |       |
Trans    |       |
    1    | a'    |
     &   | c"    |

    1    | d'    |
    2&   | e"    |

    1    | a'    |
    1&   | c"    |

    1    | d'    |
    2&   | e"    |

    1    | a'    |
    1&   | c"    |

    1    | b     |
		   		   		   `,
		},

		{
			`
=
   |piano|
1  |a'   | A
&  |c"|

2  |d |
[A]
[A]
1 |f"|
					   		   			`,
			`
=
         | piano |
Ch       | -     |
Bank     | -     |
Prog     | -     |
Vol      | -     |
PbRange  |       |
Trans    |       |
    1    | a'    | A
     &   | c"    |

    2    | d     |

    1    | a'    | A
     &   | c"    |

    2    | d     |

    1    | a'    | A
     &   | c"    |

    2    | d     |

    1    | f"    |
					   		   `,
		},
		{
			`
=
   |piano|
1  |a'   | A
&  |c"|

2  |d |
[A]
[A]
3/4
1 |f"|
					   		   			`,
			`
=
         | piano |
Ch       | -     |
Bank     | -     |
Prog     | -     |
Vol      | -     |
PbRange  |       |
Trans    |       |
    1    | a'    | A
     &   | c"    |

    2    | d     |

    1    | a'    | A
     &   | c"    |

    2    | d     |

    1    | a'    | A
     &   | c"    |

    2    | d     |
3/4
    1    | f"    |
					   		   `,
		},
		{
			`
=
   |piano|
1  |a'   | A
&  |c"|

2  |d |
[A]
[A]
3/4@120
1 |f"|
					   		   			`,
			`
=
         | piano |
Ch       | -     |
Bank     | -     |
Prog     | -     |
Vol      | -     |
PbRange  |       |
Trans    |       |
    1    | a'    | A
     &   | c"    |

    2    | d     |

    1    | a'    | A
     &   | c"    |

    2    | d     |

    1    | a'    | A
     &   | c"    |

    2    | d     |
3/4@120
    1    | f"    |
					   		   `,
		},
		{
			`
=
   |piano|
1  |a'   | A
&  |c"|

2  |d |
[A]
[A]
@100
1 |f"|
					   		   			`,
			`
=
         | piano |
Ch       | -     |
Bank     | -     |
Prog     | -     |
Vol      | -     |
PbRange  |       |
Trans    |       |
    1    | a'    | A
     &   | c"    |

    2    | d     |

    1    | a'    | A
     &   | c"    |

    2    | d     |

    1    | a'    | A
     &   | c"    |

    2    | d     |
@100
    1    | f"    |
					   		   `,
		},
		{
			`
=
   |piano|
1  |a'   | A
&  |c"|

2  |d |
[A]
[A]

1 |f"|
					   		   			`,
			`
=
         | piano |
Ch       | -     |
Bank     | -     |
Prog     | -     |
Vol      | -     |
PbRange  |       |
Trans    |       |
    1    | a'    | A
     &   | c"    |

    2    | d     |

    1    | a'    | A
     &   | c"    |

    2    | d     |

    1    | a'    | A
     &   | c"    |

    2    | d     |

    1    | f"    |
					   		   `,
		},
		{
			`
=
   |piano|
1  |a'   | A
&  |c"|

2  |d |
[A]
1&  |e |`,
			`
=
         | piano |
Ch       | -     |
Bank     | -     |
Prog     | -     |
Vol      | -     |
PbRange  |       |
Trans    |       |
    1    | a'    | A
     &   | c"    |

    2    | d     |

    1    | a'    | A
     &   | c"    |

    2    | d     |

    1&   | e     |

		   `,
		},
		{
			`
=
    |piano| vox |
1   |a'   | e   | A
 &  |c"   | %   |
2   |     | f   |

2   |d    |     |
[A]
1&  |e    |     |
2   |     | g   |`,
			`
=
         | piano | vox |
Ch       | -     | -   |
Bank     | -     | -   |
Prog     | -     | -   |
Vol      | -     | -   |
PbRange  |       |     |
Trans    |       |     |
    1    | a'    | e   | A
     &   | c"    | e   |
    2    |       | f   |

    2    | d     |     |

    1    | a'    | e   | A
     &   | c"    | e   |
    2    |       | f   |

    2    | d     |     |

    1&   | e     |     |
    2    |       | g   |

		   `,
		},
		{
			`
=
    |piano| vox |
1   |a'   | e   | A
 &  |c"   | %   |
2   |     | f   |

2   |d    |     | B
[A]
[A]
[B]
1&  |e    |     |
2   |     | g   |`,
			`
=
         | piano | vox |
Ch       | -     | -   |
Bank     | -     | -   |
Prog     | -     | -   |
Vol      | -     | -   |
PbRange  |       |     |
Trans    |       |     |
    1    | a'    | e   | A
     &   | c"    | e   |
    2    |       | f   |

    2    | d     |     | B

    1    | a'    | e   | A
     &   | c"    | e   |
    2    |       | f   |

    1    | a'    | e   | A
     &   | c"    | e   |
    2    |       | f   |

    2    | d     |     | B

    1&   | e     |     |
    2    |       | g   |

		   `,
		},
	}

	for i, test := range tests {
		/*
			if i != 15 {
				continue
			}
		*/
		sc, err := Parse(strings.NewReader(strings.TrimSpace(test.input)), "unroll")

		if err != nil {
			t.Errorf("[%v] could not parse score: %s\n%s\n", i, err.Error(), test.input)
			continue
		}

		// fmt.Printf("parts (before unrolling): %v\n", sc.Parts)

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

		result := bf.String()

		//		fmt.Println(result)

		res := strings.Split(result, "=\n")
		exp := strings.Split(test.expected, "=\n")

		got := strings.TrimSpace(res[1])
		wanted := strings.TrimSpace(exp[1])

		if got != wanted {
			t.Errorf("[%v] score\n%s\n\nunrolled gives \n%s\n\nbut this was expected:\n%s\n%q\nvs\n%q\n", i, test.input, got, wanted, got, wanted)
			//			t.Errorf("[%v] score\n%s\n\nunrolled gives \n%q\n\nbut this was expected:\n%q\n", i, test.input, got, wanted)
		}
	}
}