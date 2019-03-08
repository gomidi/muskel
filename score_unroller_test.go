package muskel

import (
	"bytes"
	"strings"
	"testing"
)

func TestUnroll(t *testing.T) {
	// t.Skip()
	tests := []struct {
		input    string
		expected string
	}{
		{
			`
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
         | Vocals |
Ch       | 1      |
Bank     | -      |
Prog     | -      |
Vol      | -      |
    1    | ^2     |

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
    1    | d'     |

    1    |        |`,
		},
		{
			`
=
         | Vocals |
Ch       | 1      |
Bank     | -      |
Prog     | -      |
Vol      | -      |
4/4
$scale(d,dorian)
    1    | ^2     |

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
4/4
$scale(d,dorian)
    1    | e      |

    1    |        |`,
		},
		{
			`
patt: 1^2 2&^3
=
         | Vocals |
Ch       | 1      |
Bank     | -      |
Prog     | -      |
Vol      | -      |
4/4
$scale(d,dorian)
    1    | patt   |

$scale(e,dorian)
    1    | ./.   |`,
			`
=
         | Vocals |
Ch       | 1      |
Bank     | -      |
Prog     | -      |
Vol      | -      |
PbRange  |        |
Trans    |        |
4/4
$scale(d,dorian)
    1    | e      |
    2&   | f      |

$scale(e,dorian)
    1    | f#     |
    2&   | g      |`,
		},
		{
			`
patt: 1^2 2&^3
=
         | Vocals |
Ch       | 1      |
Bank     | -      |
Prog     | -      |
Vol      | -      |
4/4
$scale(d,dorian)
    1    | patt^2   |

$scale(e,dorian)
    1    | ./.   |`,
			`
=
         | Vocals |
Ch       | 1      |
Bank     | -      |
Prog     | -      |
Vol      | -      |
PbRange  |        |
Trans    |        |
4/4
$scale(d,dorian)
    1    | g      |
    2&   | a      |

$scale(e,dorian)
    1    | a      |
    2&   | b      |`,
		},
		{
			`
patt: 1^2 2&^3
=
         | Vocals |
Ch       | 1      |
Bank     | -      |
Prog     | -      |
Vol      | -      |
4/4
$scale(d,dorian)
    1    | patt^-3   |

$scale(e,dorian)
    1    | ./.   |`,
			`
=
         | Vocals |
Ch       | 1      |
Bank     | -      |
Prog     | -      |
Vol      | -      |
PbRange  |        |
Trans    |        |
4/4
$scale(d,dorian)
    1    | B      |
    2&   | c      |

$scale(e,dorian)
    1    | c#     |
    2&   | d      |`,
		},
		{
			`
pattA: 1^2 2&^3
pattB: 1pattA^-3
=
         | Vocals |
Ch       | 1      |
Bank     | -      |
Prog     | -      |
Vol      | -      |
4/4
$scale(d,dorian)
    1    | pattB   |

$scale(e,dorian)
    1    | ./.   |`,
			`
=
         | Vocals |
Ch       | 1      |
Bank     | -      |
Prog     | -      |
Vol      | -      |
PbRange  |        |
Trans    |        |
4/4
$scale(d,dorian)
    1    | B      |
    2&   | c      |

$scale(e,dorian)
    1    | c#     |
    2&   | d      |`,
		},
		{
			`
patt: 1^2 2&^3
=
         | Vocals |
Ch       | 1      |
Bank     | -      |
Prog     | -      |
Vol      | -      |
4/4
$scale(d,dorian)
    1    | patt^^3   |

$scale(e,dorian)
    1    | ./.   |`,
			`
=
         | Vocals |
Ch       | 1      |
Bank     | -      |
Prog     | -      |
Vol      | -      |
PbRange  |        |
Trans    |        |
4/4
$scale(d,dorian)
    1    | f      |
    2&   | g      |

$scale(e,dorian)
    1    | g      |
    2&   | a      |`,
		},
		// the following is not allowed and should return an error (therefor is uncommented)
		/*
		   		{
		   			`
		   patt: 1^2 2&^3
		   pattB: 1patt^^3
		   =
		            | Vocals |
		   Ch       | 1      |
		   Bank     | -      |
		   Prog     | -      |
		   Vol      | -      |
		   4/4
		   $scale(d,dorian)
		       1    | pattB   |

		   $scale(e,dorian)
		       1    | ./.   |`,
		   			`
		   =
		            | Vocals |
		   Ch       | 1      |
		   Bank     | -      |
		   Prog     | -      |
		   Vol      | -      |
		   PbRange  |        |
		   Trans    |        |
		   4/4
		   $scale(d,dorian)
		       1    | f      |
		       2&   | g      |

		   $scale(e,dorian)
		       1    | g      |
		       2&   | a      |`,
		   		},
		*/
		{
			`
patt: 1^2 2&^3
=
         | Vocals |
Ch       | 1      |
Bank     | -      |
Prog     | -      |
Vol      | -      |
4/4
$scale(d,dorian)
    1    | patt^^-2   |

$scale(e,dorian)
    1    | ./.   |`,
			`
=
         | Vocals |
Ch       | 1      |
Bank     | -      |
Prog     | -      |
Vol      | -      |
PbRange  |        |
Trans    |        |
4/4
$scale(d,dorian)
    1    | B      |
    2&   | c      |

$scale(e,dorian)
    1    | c#     |
    2&   | d      |`,
		},
		{
			`
@test: mount to step 3
patt: 1e 2&f
=
         | Vocals |
Ch       | 1      |
Bank     | -      |
Prog     | -      |
Vol      | -      |
4/4
$scale(d,dorian)
    1    | patt^^3   |

$scale(e,dorian)
    1    | ./.   |`,
			`
@test: mount to step 3
=
         | Vocals |
Ch       | 1      |
Bank     | -      |
Prog     | -      |
Vol      | -      |
PbRange  |        |
Trans    |        |
4/4
$scale(d,dorian)
    1    | f      |
    2&   | g      |

$scale(e,dorian)
    1    | g      |
    2&   | a      |`,
		},
		{
			`
@test: mount to step 3
patt: 1#1 2&f 3^1
=
         | Vocals |
Ch       | 1      |
Bank     | -      |
Prog     | -      |
Vol      | -      |
4/4
$scale(d,dorian)
    1    | patt^^3(e)   |

$scale(e,dorian)
    1    | ./.   |`,
			`
@test: mount to step 3
=
         | Vocals |
Ch       | 1      |
Bank     | -      |
Prog     | -      |
Vol      | -      |
PbRange  |        |
Trans    |        |
4/4
$scale(d,dorian)
    1    | f      |
    2&   | g      |
    3    | d      |

$scale(e,dorian)
    1    | g      |
    2&   | a      |
    3    | e      |`,
		},
		{
			`
@test: mount to step 3
patt: 1#1 2&f 3^1
=
         | Vocals |
Ch       | 1      |
Bank     | -      |
Prog     | -      |
Vol      | -      |
4/4
$scale(d,dorian)
    1    | patt^^3(^2)   |

$scale(e,dorian)
    1    | ./.   |`,
			`
@test: mount to step 3
=
         | Vocals |
Ch       | 1      |
Bank     | -      |
Prog     | -      |
Vol      | -      |
PbRange  |        |
Trans    |        |
4/4
$scale(d,dorian)
    1    | f      |
    2&   | f      |
    3    | e      |

$scale(e,dorian)
    1    | g      |
    2&   | f      |
    3    | f#     |`,
		},
		{
			`
@test: mount to step 3
patt: 1#1 2&f 3^1
=
         | Vocals |
Ch       | 1      |
Bank     | -      |
Prog     | -      |
Vol      | -      |
4/4
$scale(d,dorian)
    1    | patt^^3(^2)/:,^1,:/   |

$scale(e,dorian)
    1    | ./.   |`,
			`
@test: mount to step 3
=
         | Vocals |
Ch       | 1      |
Bank     | -      |
Prog     | -      |
Vol      | -      |
PbRange  |        |
Trans    |        |
4/4
$scale(d,dorian)
    1    | f      |
    2&   | d      |
    3    | e      |

$scale(e,dorian)
    1    | g      |
    2&   | e      |
    3    | f#     |`,
		},
		// the following is not allowed and should return an error (therefor is uncommented)
		/*
		   		{
		   			`
		   @test: mount to step 3
		   patt: 1e 2&f
		   pattB: 1patt^^3
		   =
		            | Vocals |
		   Ch       | 1      |
		   Bank     | -      |
		   Prog     | -      |
		   Vol      | -      |
		   4/4
		   $scale(d,dorian)
		       1    | pattB   |

		   $scale(e,dorian)
		       1    | ./.   |`,
		   			`
		   @test: mount to step 3
		   =
		            | Vocals |
		   Ch       | 1      |
		   Bank     | -      |
		   Prog     | -      |
		   Vol      | -      |
		   PbRange  |        |
		   Trans    |        |
		   4/4
		   $scale(d,dorian)
		       1    | f      |
		       2&   | g      |

		   $scale(e,dorian)
		       1    | g      |
		       2&   | a      |`,
		   		},
		*/
		{
			`
@test: mount to step 1
patt: 1e 2&f
=
         | Vocals |
Ch       | 1      |
Bank     | -      |
Prog     | -      |
Vol      | -      |
4/4
$scale(d,dorian)
    1    | patt^^1   |

$scale(e,dorian)
    1    | ./.   |`,
			`
@test: mount to step 1
=
         | Vocals |
Ch       | 1      |
Bank     | -      |
Prog     | -      |
Vol      | -      |
PbRange  |        |
Trans    |        |
4/4
$scale(d,dorian)
    1    | d      |
    2&   | e      |

$scale(e,dorian)
    1    | e      |
    2&   | f#     |`,
		},
		{
			`
@test: mount to step 2
patt: 1e 2&f#
=
         | Vocals |
Ch       | 1      |
Bank     | -      |
Prog     | -      |
Vol      | -      |
4/4
$scale(d,dorian)
    1    | patt^^2   |

$scale(e,dorian)
    1    | ./.   |`,
			`
@test: mount to step 2
=
         | Vocals |
Ch       | 1      |
Bank     | -      |
Prog     | -      |
Vol      | -      |
PbRange  |        |
Trans    |        |
4/4
$scale(d,dorian)
    1    | e      |
    2&   | g      |

$scale(e,dorian)
    1    | f#     |
    2&   | a      |`,
		},
		{
			`
=
         | Vocals |
Ch       | 1      |
Bank     | -      |
Prog     | -      |
Vol      | -      |
4/4
$scale(d,dorian)
    1    | ^2  "hi" |

$scale(e,dorian)
    1    | ./.   |`,
			`
=
         | Vocals   |
Ch       | 1        |
Bank     | -        |
Prog     | -        |
Vol      | -        |
PbRange  |          |
Trans    |          |
4/4
$scale(d,dorian)
    1    | e  "hi"  |

$scale(e,dorian)
    1    | f#  "hi" |`,
		},
		{
			`
=
         | Vocals |
Ch       | 1      |
Bank     | -      |
Prog     | -      |
Vol      | -      |
4/4
$scale(d,dorian)
    1    | {^2,a,b'}2 |

$scale(e,dorian)
    1    | ./.   |`,
			`
=
         | Vocals     |
Ch       | 1          |
Bank     | -          |
Prog     | -          |
Vol      | -          |
PbRange  |            |
Trans    |            |
4/4
$scale(d,dorian)
    1    | {e,a,b'}2  |

$scale(e,dorian)
    1    | {f#,a,b'}2 |`,
		},
		{
			`
=
         | Vocals | Keys |
Ch       | 1      | -    |
Bank     | -      | -    |
Prog     | -      | -    |
Vol      | -      | -    |
4/4
$scale(d,dorian)
    1    | ^2     | ^3   |

$scale(e,dorian)
    1    | ^2     | ^3   |`,
			`
=
         | Vocals | Keys |
Ch       | 1      | -    |
Bank     | -      | -    |
Prog     | -      | -    |
Vol      | -      | -    |
PbRange  |        |      |
Trans    |        |      |
4/4
$scale(d,dorian)
    1    | e      | f    |

$scale(e,dorian)
    1    | f#     | g    |`,
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
			if i != 12 {
				continue
			}
		*/
		//fmt.Printf("unroll [%v]\n", i)

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
