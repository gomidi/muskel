package muskel_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"gitlab.com/gomidi/muskel_new_new/file"
	"gitlab.com/gomidi/muskel_new_new/score"
)

func init() {
	//score.DEBUG = true
}

func TestUnroll(t *testing.T) {
	// t.Skip()
	tests := []struct {
		input    string
		expected string
	}{
		{ // 0
			`
tracks: |Vocals|

score:   | Vocals |
#
    1    | e'     |
#
    1    | f#     |
`,
			`
score: | Vocals |
#
1      | e'     |
#
1      | f#     |

`,
		},
		{ // 1
			`
aaa: |x|
#
1   |a|
2   |b|

tracks: |Vocals|


score:   | Vocals  |
#
    1    | aaa_x   |
#
    1    |         |

`,
			`
score: | Vocals |
#
1      | a      |
2      | b      |
#
1      |        |

`,
		},
		{
			`
aa: 1a 2b
bb: 1aa/"hu"/"ho"/
=
         | <Vocals> | <Lyrics> |
    1    | aa       | bb       |

    1    |          |          |`,
			`
         | <Vocals> | <Lyrics> |
    1    | a        | "hu"     |
    2    | b        | "ho"     |

    1    |          |          |`,
		},
		{
			`
aa: 1a
bb: 2b
cc: 1*aa*bb*
=
         | <Vocals> |
    1    | cc       |

    1    |        |`,
			`
         | <Vocals> |
    1    | a        |
    2    | b        |

    1    |          |`,
		},
		{
			`
aa: 1a
=
         | <Vocals> |
    1    | *aa*b*  |

    1    |        |`,
			`
         | <Vocals> |
    1    | *a*b*    |

    1    |          |`,
		},
		{
			`
aa: 1a
=
         | <Vocals> | <piano> |
    1    | aa%2     | c       |

    1    |          | d       |

    1    | c        | e       |    
`,
			`
         | <Vocals> | <piano> |
    1    | a        | c       |

    1    | a        | d       |

    1    | c        | e       |   
`,
		},
		{
			`
aa@4/4: 1a | 1b
=
         | <Vocals> | <piano> |
    1    | aa%2     | c       |

    1    |          | d       |

    1    | c        | e       |    
`,
			`
         | <Vocals> | <piano> |
    1    | a        | c       |

    1    | b        | d       |

    1    | c        | e       |   
`,
		},
		{
			`
aa@4/4: 1a | 1b
=
         | <Vocals> | <piano> |
    1    | aa%3     | c       |

    1    |          | d       |

    1    |          |         |

    1    | c        | e       |    
`,
			`
         | <Vocals> | <piano> |
    1    | a        | c       |

    1    | b        | d       |

    1    | a        |         |

    1    | c        | e       |   
`,
		},
		{
			`
aa: 1a
=
         | <Vocals> | <piano> |
    1    | aa%2     | c       |

    1    |          | d       |

    1    |          | e       |    
`,
			`
         | <Vocals> | <piano> |
    1    | a        | c       |

    1    | a        | d       |

    1    | a        | e       |   
`,
		},
		{
			`
aa@4/4: 1a | 2b
=
         | <Vocals> | <piano> |
    1    | aa%2     | c       |

    2    |          | d       |

    1    |          | e       | 
    
    2    |          | f       |    
`,
			`
         | <Vocals> | <piano> |
    1    | a        | c       |

    2    | b        | d       |

    1    | a        | e       |

    2    | b        | f       |       
`,
		},
		{
			`
aa@4/4: 1a | 2b
=
         | <Vocals> | <piano> |
    1    | aa%1     | c       | A

    2    |          | d       | 

    1    |          |         |

    2    |          |         |
[A]
`,
			`
         | <Vocals> | <piano> |
    1    | a        | c       | A

    2    | b        | d       |

    1    | a        |         |

    2    | b        |         |

    1    | a        | c       | A

    2    | b        | d       |

    1    | a        |         |

    2    | b        |         |
`,
		},

		{
			`
aa@4/4: 1a | 2b | 3c
=
         | <Vocals> | <piano> |
    1    | a        | c       | A
*2
    1    |          |         |
[A]
  `,
			`
         | <Vocals> | <piano> |
    1    | a        | c       | A

    1    |          |         |

    1    |          |         |

    1    |          |         |

    1    | a        | c       | A

    1    |          |         |

    1    |          |         |

    1    |          |         |
		   `,
		},

		{
			`
aa@4/4: 1a | 2b | 3c
=
         | <Vocals> | <piano> |
    1    | a        | c       | A
*2
[A]
`,
			`
         | <Vocals> | <piano> |
    1    | a        | c       | A

    1    |          |         |

    1    |          |         |

    1    | a        | c       | A

    1    |          |         |

    1    |          |         |
`,
		},
		{
			`
aa@4/4: 1a | 2b
=
         | <Vocals> | <piano> |
    1    | aa%2     | c       |

    1    |          | d       |

    1    |          | e       | 
    
    1    |          | f       |    
`,
			`
         | <Vocals> | <piano> |
    1    | a        | c       |

    1    |          | d       |
    2    | b        |         |

    1    | a        | e       |

    1    |          | f       |
    2    | b        |         |       
`,
		},
		{
			`
aa@4/4: 1a | 2b
bb@4/4: 1c | 1d | 1e | 1f
=
         | <Vocals> | <piano> |
    1    | aa%2     | bb      |
*3
`,
			`
         | <Vocals> | <piano> |
    1    | a        | c       |

    1    |          | d       |
    2    | b        |         |

    1    | a        | e       |

    1    |          | f       |
    2    | b        |         | `,
		},
		{
			`
aa@4/4: 1a | 2b
bb@4/4: 1c | 1d | 1e | 1f
=
         | <Vocals> | <piano> |
    1    | aa%2     | bb      |
*3
    1    | g        | g       |
	   `,
			`
         | <Vocals> | <piano> |
    1    | a        | c       |

    1    |          | d       |
    2    | b        |         |

    1    | a        | e       |

    1    |          | f       |
    2    | b        |         |

    1    | g        | g       |
   `,
		},
		{
			`
aa@4/4: 1a | 2b
bb@4/4: 1c | 1d | 1e | 1f
=
         | <Vocals> | <piano> |
    1    | aa%2     | bb      |
*3
    1    |          | g       |
	   `,
			`
         | <Vocals> | <piano> |
    1    | a        | c       |

    1    |          | d       |
    2    | b        |         |

    1    | a        | e       |

    1    |          | f       |
    2    | b        |         |

    1    | a        | g       |
   `,
		},
		{
			`
aa:  1a' 2&*#1+:*#2-:*
=
  | <piano> | <vox> |
1 | aa(^1,^3) |   |

1 |    |  |
			`,
			`
         | <piano>     | <vox> |
    1    | a'          |       |
    2&   | *c'+:*e'-:* |       |

    1    |             |       |
			`,
		},
		{
			`
aa:  1a' 2&*#1+:*#2-:*
=
  | <piano> | <vox> |
1 | aa(\1,\3) |   |

1 |    |  |
			`,
			`
         | <piano>     | <vox> |
    1    | a'          |       |
    2&   | *c'+:*e'-:* |       |

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
    1    | templ^^3(^2)/:/^1/:/   |
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
    1    | *^2*"hi"* |
\dorian^e
    1    | ./.   |`,
			`
         | <Vocals>  |
4/4\dorian^d
    1    | *e*"hi"*  |
\dorian^e
    1    | *f#*"hi"* |`,
		},
		{
			`
=
         | <Vocals> |
4/4\dorian^d
    1    | {^2,a,b'} |
\dorian^e
    1    | ./.   |`,
			`
         | <Vocals>  |
4/4\dorian^d
    1    | {e,a,b'}  |
\dorian^e
    1    | {f#,a,b'} |`,
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
		{
			`
=
    |<piano>| <vox> |
1   |a'   | e   | A
 &  |c"   | %   |

2   |d    |     | B
*3
[A]
1&  |e    |     |
2   |     | g   |`,
			`
         | <piano> | <vox> |
    1    | a'      | e     | A
     &   | c"      | e     |

    2    | d       |       | B

    1    |         |       |

    1    |         |       |

    1    |         |       |

    1    | a'      | e     | A
     &   | c"      | e     |

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

2   |d    |     | B
[A]
*3
[A]
1&  |e    |     |
2   |     | g   |`,
			`
         | <piano> | <vox> |
    1    | a'      | e     | A
     &   | c"      | e     |

    2    | d       |       | B

    1    | a'      | e     | A
     &   | c"      | e     |

    1    |         |       |

    1    |         |       |

    1    |         |       |

    1    | a'      | e     | A
     &   | c"      | e     |

    1&   | e       |       |
    2    |         | g     |

		   `,
		},
	}

	for i, test := range tests {

		if i > 0 {
			continue
		}
		/*
			if i == 11 || i == 12 {
				continue
			}
		*/

		sc := score.New(fmt.Sprintf("test-%v", i), "@SCORE")
		f := file.FromReader(strings.NewReader(test.input), sc)

		sc.Files[fmt.Sprintf("test-%v", i)] = f

		err := f.Parse()
		if err != nil {
			t.Errorf("[%v] can't parse score: %s", i, err)
			continue
		}

		unrolled, err := sc.Unroll()
		if err != nil {
			t.Errorf("[%v] can't unroll score: %s", i, err)
			continue
		}

		_ = unrolled

		var bf bytes.Buffer

		//err = f.WriteTo(&bf)
		unrolled.WriteUnrolledTo(&bf)

		if err != nil {
			t.Errorf("[%v] error: %s", i, err)
			continue
		}

		got := strings.TrimSpace(bf.String())
		wanted := strings.TrimSpace(test.expected)

		if got != wanted {
			t.Errorf("[%v] score\n%s\n\nunrolled gives \n%s\n\nbut this was expected:\n%s\n%q\nvs\n%q\n", i, test.input, got, wanted, got, wanted)
		}
	}
}
