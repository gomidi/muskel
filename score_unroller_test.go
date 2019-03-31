package muskel_test

import (
	"strings"
	"testing"

	"gitlab.com/gomidi/muskel"
	"gitlab.com/gomidi/muskel/formatter"
	"gitlab.com/gomidi/muskel/score"
	"gitlab.com/gomidi/muskel/unroller"
)

func init() {
	score.DEBUG = true
}

func TestUnroll(t *testing.T) {
	// t.Skip()
	tests := []struct {
		input    string
		expected string
	}{
		{
			`
=
         | <Vocals> |
    1    | e'     |

    1    |        |`,
			`
         | <Vocals> |
    1    | e'       |

    1    |          |`,
		},
		{
			`
aa:  1a' 2&_#1+:_#2-:_
=
  | <piano> | <vox> |
1 | aa(^1,^3) |   |

1 |    |  |
			`,
			`
         | <piano>     | <vox> |
    1    | a'          |       |
    2&   | _c'+:_e'-:_ |       |

    1    |             |       |
			`,
		},
		{
			`
=
         | <Vocals> |
    1    | ^2     |

    1    |        |`,
			`
         | <Vocals> |
    1    | d'       |

    1    |          |`,
		},
		{
			`
=
         | <Vocals> |
4/4\dorian^d
    1    | ^2     |

    1    |        |`,
			`
         | <Vocals> |
4/4\dorian^d
    1    | e        |

    1    |          |`,
		},
		{
			`
templ: 1^2 2&^3
=
         | <Vocals> |
4/4\dorian^d
    1    | templ   |
\dorian^e
    1    | ./.   |`,
			`
         | <Vocals> |
4/4\dorian^d
    1    | e        |
    2&   | f        |
\dorian^e
    1    | f#       |
    2&   | g        |`,
		},
		{
			`
templ: 1^2 2&^3
=
         | <Vocals> |
4/4\dorian^d
    1    | templ^2   |
\dorian^e
    1    | ./.   |`,
			`
         | <Vocals> |
4/4\dorian^d
    1    | g        |
    2&   | a        |
\dorian^e
    1    | a        |
    2&   | b        |`,
		},
		{
			`
templ: 1^2 2&^3
=
         | <Vocals> |
4/4\dorian^d
    1    | templ^-3   |
\dorian^e
    1    | ./.   |`,
			`
         | <Vocals> |
4/4\dorian^d
    1    | B        |
    2&   | c        |
\dorian^e
    1    | c#       |
    2&   | d        |`,
		},
		{
			`
templA: 1^2 2&^3
templB: 1templA^-3
=
         | <Vocals> |
4/4\dorian^d
    1    | templB   |
\dorian^e
    1    | ./.   |`,
			`
         | <Vocals> |
4/4\dorian^d
    1    | B        |
    2&   | c        |
\dorian^e
    1    | c#       |
    2&   | d        |`,
		},
		{
			`
templ: 1^2 2&^3
=
         | <Vocals> |
4/4\dorian^d
    1    | templ^^3   |
\dorian^e
    1    | ./.   |`,
			`
         | <Vocals> |
4/4\dorian^d
    1    | f        |
    2&   | g        |
\dorian^e
    1    | g        |
    2&   | a        |`,
		},
		{
			`
templ: 1^2 2&^3
=
         | <Vocals> |
4/4\dorian^d
    1    | templ^^-2   |
\dorian^e
    1    | ./.   |`,
			`
         | <Vocals> |
4/4\dorian^d
    1    | B        |
    2&   | c        |
\dorian^e
    1    | c#       |
    2&   | d        |`,
		},
		{
			`
templ: 1e 2&f
=
         | <Vocals> |
4/4\dorian^d
    1    | templ^^3   |
\dorian^e
    1    | ./.   |`,
			`
         | <Vocals> |
4/4\dorian^d
    1    | f        |
    2&   | g        |
\dorian^e
    1    | g        |
    2&   | a        |`,
		},
		{
			`
templ: 1#1 2&f 3^1
=
         | <Vocals> |
4/4\dorian^d
    1    | templ^^3(e)   |
\dorian^e
    1    | ./.   |`,
			`
         | <Vocals> |
4/4\dorian^d
    1    | f        |
    2&   | g        |
    3    | d        |
\dorian^e
    1    | g        |
    2&   | a        |
    3    | e        |`,
		},
		{
			`
templ: 1#1 2&f 3^1
=
         | <Vocals> |
4/4\dorian^d
    1    | templ^^3(^2)   |
\dorian^e
    1    | ./.   |`,
			`
         | <Vocals> |
4/4\dorian^d
    1    | f        |
    2&   | f        |
    3    | e        |
\dorian^e
    1    | g        |
    2&   | f        |
    3    | f#       |`,
		},
		{
			`
@test: mount to step 3
templ: 1#1 2&f 3^1
=
         | <Vocals> |
4/4\dorian^d
    1    | templ^^3(^2)/:,^1,:/   |
\dorian^e
    1    | ./.   |`,
			`
         | <Vocals> |
4/4\dorian^d
    1    | f        |
    2&   | d        |
    3    | e        |
\dorian^e
    1    | g        |
    2&   | e        |
    3    | f#       |`,
		},
		{
			`
templ: 1e 2&f
=
         | <Vocals> |
4/4\dorian^d
    1    | templ^^1   |
\dorian^e
    1    | ./.   |`,
			`
         | <Vocals> |
4/4\dorian^d
    1    | d        |
    2&   | e        |
\dorian^e
    1    | e        |
    2&   | f#       |`,
		},
		{
			`
templ: 1e 2&f#
=
         | <Vocals> |
4/4\dorian^d
    1    | templ^^2   |
\dorian^e
    1    | ./.   |`,
			`
         | <Vocals> |
4/4\dorian^d
    1    | e        |
    2&   | g        |
\dorian^e
    1    | f#       |
    2&   | a        |`,
		},
		{
			`
=
         | <Vocals> |
4/4\dorian^d
    1    | _^2_"hi"_ |
\dorian^e
    1    | ./.   |`,
			`
         | <Vocals>  |
4/4\dorian^d
    1    | _e_"hi"_  |
\dorian^e
    1    | _f#_"hi"_ |`,
		},
		{
			`
=
         | <Vocals> |
4/4\dorian^d
    1    | {^2,a,b'}2 |
\dorian^e
    1    | ./.   |`,
			`
         | <Vocals>   |
4/4\dorian^d
    1    | {e,a,b'}2  |
\dorian^e
    1    | {f#,a,b'}2 |`,
		},
		{
			`
=
         | <Vocals> | <Keys> |
4/4\dorian^d
    1    | ^2       | ^3     |
\dorian^e
    1    | ^2       | ^3     |`,
			`
         | <Vocals> | <Keys> |
4/4\dorian^d
    1    | e        | f      |
\dorian^e
    1    | f#       | g      |`,
		},
		{
			`
=
         | <Vocals> | <Keys> |
4/4\dorian^d
    1    | ^2     | ^3#  |
\dorian^e
    1    | ^2     | ^3## |`,
			`
         | <Vocals> | <Keys> |
4/4\dorian^d
    1    | e        | f#     |
\dorian^e
    1    | f#       | a      |`,
		},
		{
			`
=
         | <Vocals> | <Piano> | 
    3    |          |         |
     &   |          |         |
`,
			`
         | <Vocals> | <Piano> |
    3    |          |         |
     &   |          |         |
`,
		},

		{
			`
=
   |<piano>|
1  |a'   |
						`,
			`
         | <piano> |
    1    | a'      |
			`,
		},
		{
			`
=
   |<piano>|
1  |a'   |
&  |c"|
						`,
			`
         | <piano> |
    1    | a'      |
     &   | c"      |
			`,
		},

		{
			`
=
   |<piano>|
1  |a'   |
&  |c"|

2  |d |
						`,
			`
         | <piano> |
    1    | a'      |
     &   | c"      |

    2    | d       |
			`,
		},
		{
			`
=
   |<piano>|
1  |a'   |
&  |c"|

1  |./. |
						`,
			`
         | <piano> |
    1    | a'      |
     &   | c"      |

    1    | a'      |
    1&   | c"      |
			`,
		},
		{
			`
=
   |<piano>|
1  |a'   |
&  |c"|

1  |... |

1  | |

1  |b |
						`,
			`
         | <piano> |
    1    | a'      |
     &   | c"      |

    1    | a'      |
    1&   | c"      |

    1    | a'      |
    1&   | c"      |

    1    | b       |
			`,
		},

		{
			`
=
   |<piano>|
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
         | <piano> |
    1    | a'      |
     &   | c"      |

    1    | d'      |
    2&   | e"      |

    1    | a'      |
    1&   | c"      |

    1    | d'      |
    2&   | e"      |

    1    | a'      |
    1&   | c"      |

    1    | b       |
		   		   		   `,
		},

		{
			`
=
   |<piano>|
1  |a'   | A
&  |c"|

2  |d |
[A]
[A]
1 |f"|
					   		   			`,
			`
         | <piano> |
    1    | a'      | A
     &   | c"      |

    2    | d       |

    1    | a'      | A
     &   | c"      |

    2    | d       |

    1    | a'      | A
     &   | c"      |

    2    | d       |

    1    | f"      |
					   		   `,
		},
		{
			`
=
   |<piano>|
1  |a'   | A
&  |c"|

2  |d |
[A]
[A]
3/4
1 |f"|
					   		   			`,
			`
         | <piano> |
    1    | a'      | A
     &   | c"      |

    2    | d       |

    1    | a'      | A
     &   | c"      |

    2    | d       |

    1    | a'      | A
     &   | c"      |

    2    | d       |
3/4
    1    | f"      |
					   		   `,
		},
		{
			`
=
   |<piano>|
1  |a'   | A
&  |c"|

2  |d |
[A]
[A]
3/4@120
1 |f"|
					   		   			`,
			`
         | <piano> |
    1    | a'      | A
     &   | c"      |

    2    | d       |

    1    | a'      | A
     &   | c"      |

    2    | d       |

    1    | a'      | A
     &   | c"      |

    2    | d       |
3/4@120
    1    | f"      |
					   		   `,
		},
		{
			`
=
   |<piano>|
1  |a'   | A
&  |c"|

2  |d |
[A]
[A]
@100
1 |f"|
					   		   			`,
			`
         | <piano> |
    1    | a'      | A
     &   | c"      |

    2    | d       |

    1    | a'      | A
     &   | c"      |

    2    | d       |

    1    | a'      | A
     &   | c"      |

    2    | d       |
@100
    1    | f"      |
					   		   `,
		},
		{
			`
=
   |<piano>|
1  |a'   | A
&  |c"|

2  |d |
[A]
[A]

1 |f"|
					   		   			`,
			`
         | <piano> |
    1    | a'      | A
     &   | c"      |

    2    | d       |

    1    | a'      | A
     &   | c"      |

    2    | d       |

    1    | a'      | A
     &   | c"      |

    2    | d       |

    1    | f"      |
					   		   `,
		},
		{
			`
=
   |<piano>|
1  |a'   | A
&  |c"|

2  |d |
[A]
1&  |e |`,
			`
         | <piano> |
    1    | a'      | A
     &   | c"      |

    2    | d       |

    1    | a'      | A
     &   | c"      |

    2    | d       |

    1&   | e       |

		   `,
		},
		{
			`
=
    |<piano>| <vox> |
1   |a'   | e   | A
 &  |c"   | %   |
2   |     | f   |

2   |d    |     |
[A]
1&  |e    |     |
2   |     | g   |`,
			`
         | <piano> | <vox> |
    1    | a'      | e     | A
     &   | c"      | e     |
    2    |         | f     |

    2    | d       |       |

    1    | a'      | e     | A
     &   | c"      | e     |
    2    |         | f     |

    2    | d       |       |

    1&   | e       |       |
    2    |         | g     |

		   `,
		},
		{
			`
=
    |<piano>| <vox> |
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
         | <piano> | <vox> |
    1    | a'      | e     | A
     &   | c"      | e     |
    2    |         | f     |

    2    | d       |       | B

    1    | a'      | e     | A
     &   | c"      | e     |
    2    |         | f     |

    1    | a'      | e     | A
     &   | c"      | e     |
    2    |         | f     |

    2    | d       |       | B

    1&   | e       |       |
    2    |         | g     |

		   `,
		},
	}

	for i, test := range tests {
		/*
			if i != 12 {
				continue
			}
		*/

		sc, err := muskel.Parse(strings.NewReader(strings.TrimSpace(test.input)), "unroll")

		if err != nil {
			t.Errorf("[%v] could not parse score: %s\n%s\n", i, err.Error(), test.input)
			continue
		}

		unr, err := unroller.Unroll(sc)

		if err != nil {
			t.Errorf("[%v] could not unroll score: %s\n%s\n", i, err.Error(), test.input)
			continue
		}

		var bf strings.Builder

		fm := formatter.New(unr)
		fm.HideTrackProperties = true
		fm.HideHeader = true
		fm.WriteTo(&bf)

		if err != nil {
			t.Errorf("[%v] could not format unrolled score: %s\n%s\n", i, err.Error(), test.input)
			continue
		}

		got := strings.TrimSpace(bf.String())
		wanted := strings.TrimSpace(test.expected)

		if got != wanted {
			t.Errorf("[%v] score\n%s\n\nunrolled gives \n%s\n\nbut this was expected:\n%s\n%q\nvs\n%q\n", i, test.input, got, wanted, got, wanted)
		}
	}
}
