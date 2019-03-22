package muskel

import (
	"strings"
	"testing"
)

func init() {
	DEBUG = true
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
patt: 1^2 2&^3
=
         | <Vocals> |
4/4\dorian^d
    1    | patt   |
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
patt: 1^2 2&^3
=
         | <Vocals> |
4/4\dorian^d
    1    | patt^2   |
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
patt: 1^2 2&^3
=
         | <Vocals> |
4/4\dorian^d
    1    | patt^-3   |
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
pattA: 1^2 2&^3
pattB: 1pattA^-3
=
         | <Vocals> |
4/4\dorian^d
    1    | pattB   |
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
patt: 1^2 2&^3
=
         | <Vocals> |
4/4\dorian^d
    1    | patt^^3   |
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
patt: 1^2 2&^3
=
         | <Vocals> |
4/4\dorian^d
    1    | patt^^-2   |
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
patt: 1e 2&f
=
         | <Vocals> |
4/4\dorian^d
    1    | patt^^3   |
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
patt: 1#1 2&f 3^1
=
         | <Vocals> |
4/4\dorian^d
    1    | patt^^3(e)   |
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
patt: 1#1 2&f 3^1
=
         | <Vocals> |
4/4\dorian^d
    1    | patt^^3(^2)   |
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
patt: 1#1 2&f 3^1
=
         | <Vocals> |
4/4\dorian^d
    1    | patt^^3(^2)/:,^1,:/   |
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
patt: 1e 2&f
=
         | <Vocals> |
4/4\dorian^d
    1    | patt^^1   |
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
patt: 1e 2&f#
=
         | <Vocals> |
4/4\dorian^d
    1    | patt^^2   |
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

		sc, err := Parse(strings.NewReader(strings.TrimSpace(test.input)), "unroll")

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

		fm := unr.Formatter()
		fm.hideInstrumentProps = true
		fm.hideHeader = true
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
