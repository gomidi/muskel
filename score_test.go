package muskel

import (
	"bytes"
	"strings"
	"testing"
)

func TestUnrollPattern(t *testing.T) {
	//t.Skip()
	tests := []struct {
		input    string
		expected string
	}{

		{
			`
$a:  1a' 2&b 5&c"
=
  | piano | vox |

1 | $a |    |
2 |    |  E  |

1 |    | F|
`, `



$a:             1a' 2&b 5&c"



=
         | piano | vox |
Ch       | -     | -   |
Bank     | -     | -   |
Prog     | -     | -   |
Vol      | -     | -   |

    1    | a'    |     |
    2    |       | E   |
    2&   | b     |     |

    1    |       | F   |
    1&   | c"    |     |
			`},

		{
			`
$a:  1a' 2&b 5g
=
  | piano |

1 | $a |

1 |  |
`, `



$a:             1a' 2&b 5g



=
         | piano |
Ch       | -     |
Bank     | -     |
Prog     | -     |
Vol      | -     |

    1    | a'    |
    2&   | b     |

    1    | g     |
		   		   `},
		{
			`
$a:  1a' 2&b 5g
=
  | piano |

1 | $a |

1 | e |
`, `



$a:             1a' 2&b 5g



=
         | piano |
Ch       | -     |
Bank     | -     |
Prog     | -     |
Vol      | -     |

    1    | a'    |
    2&   | b     |

    1    | e     |
		   `},
		{
			`
$a:  1a' 2&b 5g
=
  | piano |

1 | $a |
`, `



$a:             1a' 2&b 5g



=
         | piano |
Ch       | -     |
Bank     | -     |
Prog     | -     |
Vol      | -     |

    1    | a'    |
    2&   | b     |
		   		   `},
		{
			`
$a:  1a' 4&b 5&c"
=
  | piano |
3/4
1 | $a |

1& |    |
`, `



$a:             1a' 4&b 5&c"



=
         | piano |
Ch       | -     |
Bank     | -     |
Prog     | -     |
Vol      | -     |
3/4
    1    | a'    |

    1&   | b     |
    2&   | c"    |
		   `},

		{
			`
$a:  1a' 2&b 5&c"
=
  | piano | vox |

1 | $a |   |
2 |    | $a |

1 |    |  |
`, `



$a:             1a' 2&b 5&c"



=
         | piano | vox |
Ch       | -     | -   |
Bank     | -     | -   |
Prog     | -     | -   |
Vol      | -     | -   |

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
$a:  1a' 2b
=
  | piano |

2 | $a |
`, `



$a:             1a' 2b



=
         | piano |
Ch       | -     |
Bank     | -     |
Prog     | -     |
Vol      | -     |

    2    | a'    |
    3    | b     |
		   `},
	}

	for i, test := range tests {
		//		if i != 4 {
		//			continue
		//		}
		sc, err := Parse(strings.NewReader(strings.TrimSpace(test.input)))

		if err != nil {
			t.Errorf("[%v] could not parse score: %s\n%s\n", i, err.Error(), test.input)
			continue
		}

		unr, err := sc.Unroll()

		if err != nil {
			t.Errorf("[%v] could not unroll score: %s\n%s\n", i, err.Error(), test.input)
			continue
		}

		var bf bytes.Buffer

		err = unr.WriteTo(&bf)

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

func TestUnroll(t *testing.T) {
	//t.Skip()
	tests := []struct {
		input    string
		expected string
	}{

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

    1    | a'    |
     &   | c"    |

    2    | d     |

    1    | a'    |
     &   | c"    |

    2    | d     |

    1    | a'    |
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

    1    | a'    |
     &   | c"    |

    2    | d     |

    1    | a'    |
     &   | c"    |

    2    | d     |

    1&   | e     |

		   `,
		},
	}

	for i, test := range tests {
		sc, err := Parse(strings.NewReader(strings.TrimSpace(test.input)))

		if err != nil {
			t.Errorf("[%v] could not parse score: %s\n%s\n", i, err.Error(), test.input)
			continue
		}

		unr, err := sc.Unroll()

		if err != nil {
			t.Errorf("[%v] could not unroll score: %s\n%s\n", i, err.Error(), test.input)
			continue
		}

		var bf bytes.Buffer

		err = unr.WriteTo(&bf)

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
