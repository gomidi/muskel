package muskel_test

import (
	"strings"
	"testing"

	"gitlab.com/gomidi/muskel"
	"gitlab.com/gomidi/muskel/items"
	"gitlab.com/gomidi/muskel/score"
)

func TestInclude(t *testing.T) {
	//t.Skip()
	tests := []struct {
		input    string
		expected string
	}{
		{ // 0
			`
'drumnote

|TRACK | Voc | Drums | Piano |

| =SCORE | Voc | Drums | Piano |
|#
|1  | a"  | .drums.kd    | b"    |
|2& | c   | .drums.sn    | d     |
|#
|1  |     | .drums.ho    |       |
`,
			`
| =SCORE                   | Drums   | Piano | Voc |
| ------------------------ | ------- | ----- | --- |
|  # 4/4 @120.00 \major^c' |         |       |     |
|     1                    | MN36::: | b"    | a"  |
|     2&                   | MN40::: | d     | c   |
|  #                       |         |       |     |
|     1                    | MN46::: |       |     |
`,
		},
		{ // 1
			`
'testdata/includes/header
'drumnote

=hu |a|
#
1   |.drums.sn| 
2   |.drums.kd|

TRACK | Voc | Drums | Piano |

=SCORE   | Voc | Drums   | Piano |
#
1  | a"  | .drums.kd    | b"    |
2& | c   | .drums.sn    | d     |
#
1  |     | =hu.a    |       |
`,
			`
=SCORE                   | Drums   | Piano | Voc |
 # 4/4 @120.00 \major^c'
    1                    | MN36::: | b"    | a"  |
    2&                   | MN40::: | d     | c   |
 #
    1                    | MN40::: |       |     |
    2                    | MN36::: |       |     |
`,
		},
		{ // 2
			`
'testdata/includes/header
'drumnote

=rhythm |a|
#
1   |$euclid(3,12,&)| 
*3

TRACK | Voc | Drums | Piano |

=SCORE | Voc | Drums                      | Piano |
#
1      | a"  | =rhythm(f",e") | b"    |
2&     | c   |                            | d     |
#
1      |     |                            |       |
`,
			`
=SCORE                   | Drums | Piano | Voc |
 # 4/4 @120.00 \major^c'
    1                    | f"    | b"    | a"  |
    1&                   | e"    |       |     |
    2                    | e"    |       |     |
    2&                   | e"    | d     | c   |
    3                    | f"    |       |     |
    3&                   | e"    |       |     |
    4                    | e"    |       |     |
    4&                   | e"    |       |     |
 #
    1                    | f"    |       |     |
    1&                   | e"    |       |     |
    2                    | e"    |       |     |
    2&                   | e"    |       |     |
`,
		},
		{ // 3
			`
'testdata/includes/header
'drumnote

=rhythm |a|
#
1&  | $euclid(3,12,&) | 
*3

TRACK | Voc | Drums | Piano |

=SCORE | Voc | Drums                      | Piano |
#
1      | a"  | =rhythm(f",e") | b"    |
2&     | c   |                            | d     |
#
1      |     |                            |       |
`,
			`
=SCORE                   | Drums | Piano | Voc |
 # 4/4 @120.00 \major^c'
    1                    |       | b"    | a"  |
    1&                   | f"    |       |     |
    2                    | e"    |       |     |
    2&                   | e"    | d     | c   |
    3                    | e"    |       |     |
    3&                   | f"    |       |     |
    4                    | e"    |       |     |
    4&                   | e"    |       |     |
 #
    1                    | e"    |       |     |
    1&                   | f"    |       |     |
    2                    | e"    |       |     |
    2&                   | e"    |       |     |
    3                    | e"    |       |     |
`,
		},
		{ // 4
			`
TRACK | Voc | Drums | Piano |

=SCORE   | Voc | Drums  | Piano |
# 3/4
1        | a"  | d    | b"    |
2&       | c   |      | d     |
'testdata/includes/score
#
1        | d   |      |       | B
`,
			`
=SCORE                   | Drums | Piano | Voc |
 # 3/4 @120.00 \major^c'
    1                    | d     | b"    | a"  |
    2&                   |       | d     | c   |
 # 4/4 \major^c'
    1                    |       | f"    | e"  |
    2&                   |       | d     | c   |
 # 3/4 \major^c'
    1                    |       |       | d   |
`,
		},
		{ // 5
			`
TRACK | Voc | Drums | Piano |

=SCORE  | Voc | Drums | Piano |
'testdata/includes/score
#
1       | d   |       |       | B
`,
			`
=SCORE                   | Drums | Piano | Voc |
 # 4/4 @120.00 \major^c'
    1                    |       | f"    | e"  |
    2&                   |       | d     | c   |
 # \major^c'
    1                    |       |       | d   |
`,
		},
		{ // 6
			`
TRACK | Voc | Drums | Piano |

=SCORE | Voc | Drums   | Piano |
'testdata/includes/score
'testdata/includes/score
`,
			`
=SCORE                   | Drums | Piano | Voc |
 # 4/4 @120.00 \major^c'
    1                    |       | f"    | e"  |
    2&                   |       | d     | c   |
 #
    1                    |       | f"    | e"  |
    2&                   |       | d     | c   |
`,
		},
		{ // 7
			`
TRACK | Voc | Drums | Piano |

=SCORE   | Voc | Drums   | Piano |
# 4/4
1        | a"  | d    | b"    |
2&       | c   |      | d     |
'testdata/includes/score
'testdata/includes/score
#
1        | d   |      |       | B

`,
			`
=SCORE                   | Drums | Piano | Voc |
 # 4/4 @120.00 \major^c'
    1                    | d     | b"    | a"  |
    2&                   |       | d     | c   |
 # \major^c'
    1                    |       | f"    | e"  |
    2&                   |       | d     | c   |
 #
    1                    |       | f"    | e"  |
    2&                   |       | d     | c   |
 # \major^c'
    1                    |       |       | d   |

`,
		},
		{ // 8
			`
TRACK | Voc |

=pattern | a |  b  |
#
1        | c |     |
1&       | d | =.a |

=SCORE   | Voc |
# 4/4
1        | =pattern.b  |

`,
			`
=SCORE                   | Voc |
 # 4/4 @120.00 \major^c'
    1&                   | c   |
    2                    | d   |

`,
		},
		{ // 9
			`
TRACK | Voc |

=pattern | a |  b  |
#
1        | c |     |
1&       | d | =.a |

=SCORE   | Voc |
# 4/4
1        | =pattern.b%3  |
*2

`,
			`
=SCORE                   | Voc |
 # 4/4 @120.00 \major^c'
    1&                   | c   |
    2                    | d   |
 #
    1&                   | c   |
    2                    | d   |
 #
    1&                   | c   |
    2                    | d   |
`,
		},
		{ // 10
			`
TRACK | Voc |

=pattern | a |  b  |
#
1        | c |     |
1&       | d | =.a |

=SCORE   | Voc |
#A 4/4
1        | =pattern.b%3  |
*2
#B 3/4    
1        | C |
[A]

`,
			`
=SCORE                    | Voc |
 #A 4/4 @120.00 \major^c'
    1&                    | c   |
    2                     | d   |
 #
    1&                    | c   |
    2                     | d   |
 #
    1&                    | c   |
    2                     | d   |
 #B 3/4
    1                     | C   |
 # 4/4
    1&                    | c   |
    2                     | d   |
 #
    1&                    | c   |
    2                     | d   |
 #
    1&                    | c   |
    2                     | d   |
`,
		},
		{ // 11
			`
TRACK | flute |

=motiv | X |
# 4/4 @144
 1     | a |

=SCORE | flute    |
#A 4/4
 1     | =motiv.X |
*3
[A]
# 5/6
 1     | B        |
# 4/4
 1     | G        |
`,
			`
=SCORE                    | flute |
 #A 4/4 @120.00 \major^c'
    1                     | a     |
 #
 #
 #
 #
    1                     | a     |
 #
 #
 #
 # 5/6
    1                     | B     |
 # 4/4
    1                     | G     |
`,
		},
		{ // 12
			`
TRACK | left |
 Ch   | 1    |

=SCORE | left |
#A
 1     | a    |
#B
 1     | E   |
#C
 1     | B   |
[B]
#
 1     | g    |
`,
			`
=SCORE                    | left |
 #A 4/4 @120.00 \major^c'
    1                     | a    |
 #B
    1                     | E    |
 #C
    1                     | B    |
 #
    1                     | E    |
 #
    1                     | g    |
`,
		},
		{ // 13
			`
TRACK | left |
 Ch   | 1    |

=SCORE | left |
# // hi
 1     | a    |
`,
			`
=SCORE                   | left |
 # 4/4 @120.00 \major^c'
    1                    | a    |
`,
		},
		{ // 14
			`
TRACK | left |
 Ch   | 1    |

=patt | a |
#
1     | (c e) |
3     | (g a) |


=SCORE | left |
#
1  | d |
#
 1     | *    |
 2     | =!patt.a |
`,
			`
=SCORE                   | left  |
 # 4/4 @120.00 \major^c'
    1                    | d     |
 #
    1                    | *     |
    2                    | (c e) |
    4                    | (g a) |
`,
		},
		{ // 15
			`
TRACK | left |
 Ch   | 1    |

=SCORE | left |
#
1  | d |
   | e |
`,
			`
=SCORE                   | left |
 # 4/4 @120.00 \major^c'
    1                    | d    |
    1                    | e    |
`,
		},
		{ // 16
			`
TRACK | left |
 Ch   | 1    |

=SCORE | left |
# \ionian^d#
1      | 1   |
2      | 2   |
`,
			`
=SCORE                    | left |
 # 4/4 @120.00 \ionian^d#
    1                     | d#   |
    2                     | f    |
`,
		},
		{ // 17
			`
TRACK | left |
 Ch   | 1    |

=SCORE | left |
#
1      | 1   |
2      | 2   |
#  \ionian^d#
1      | 1   |
2      | 2   |
#  \ionian^f#
1      | 1   |
2      | 2   |
`,
			`
=SCORE                   | left |
 # 4/4 @120.00 \major^c'
    1                    | c'   |
    2                    | d'   |
 # \ionian^d#
    1                    | d#   |
    2                    | f    |
 # \ionian^f#
    1                    | f#   |
    2                    | g#   |
`,
		},
		{ // 18
			`
TRACK | left |
 Ch   | 1    |

=SCORE | left |
#
1      | 1   |
2      | 2   |
#  \ionian^d#
1      | 1   |
2      | \ionian^g |
2      | 2   |
#  \ionian^f#
1      | 1   |
2      | 2   |
`,
			`
=SCORE                   | left |
 # 4/4 @120.00 \major^c'
    1                    | c'   |
    2                    | d'   |
 # \ionian^d#
    1                    | d#   |
    2                    | a    |
 # \ionian^f#
    1                    | g    |
    2                    | a    |
`,
		},
		{ // 19
			`
TRACK | left |
 Ch   | 1    |

=patt | a |
#
1      | 1   |
2      | \ionian^g |
2      | 2   |

=SCORE | left |
#
1      | 1   |
2      | 2   |
#  \ionian^d#
1      | =patt  |
#  \ionian^f#
1      | 1   |
2      | 2   |
`,
			`
=SCORE                   | left |
 # 4/4 @120.00 \major^c'
    1                    | c'   |
    2                    | d'   |
 # \ionian^d#
    1                    | d#   |
    2                    | a    |
 # \ionian^f#
    1                    | g    |
    2                    | a    |
`,
		},
		{ // 20
			`
TRACK | left |
 Ch   | 1    |

SCALE | myscale |
      | d"      |
      | g"      |
      | a"      |
      | c"'     |
      | d"'     |

=patt | a |
#
1      | 1   |
2      | \myscale^a" |
2      | 2   |

=SCORE | left |
#
1      | 1   |
2      | 2   |
#  \myscale^d#
1      | =patt  |
#  \myscale^d"
1      | 1   |
2      | 2   |
`,
			`
=SCORE                   | left |
 # 4/4 @120.00 \major^c'
    1                    | c'   |
    2                    | d'   |
 # \myscale^d#
    1                    | d#   |
    2                    | c"'  |
 # \myscale^d"
    1                    | a"   |
    2                    | c"'  |
`,
		},
		{ // 21
			`
TRACK | left | right |
 Ch   | 1    | 2     |

=patt | |
#
1    |1 |
2    |2 |
#
1    |1 |
2    |2 |
#
1    |1 |
2    |2 |


=SCORE | left | right |
# \minor^b'
1      | 2   | =patt |
2      | 1 | |
#
1      |=patt |  | 
*2
#
1      | 2   | 1  |
2      | 1   | 1  |
`,
			`
=SCORE                   | left | right |
 # 4/4 @120.00 \minor^b'
    1                    | c#"  | b'    |
    2                    | b'   | c#"   |
 #
    1                    | b'   | b'    |
    2                    | c#"  | c#"   |
 #
    1                    | b'   | b'    |
    2                    | c#"  | c#"   |
 #
    1                    | b'   |       |
    2                    | c#"  |       |
 #
    1                    | c#"  | b'    |
    2                    | b'   | b'    |
`,
		},
		{ // 22
			`
TRACK | left | right |
 Ch   | 1    | 2     |

=patt | |
#
1    |1 |
2    |2 |
#
1    |3 |
2    |3 |


=SCORE | right   |
#  \major^c"
1      | =patt%3 |
*3
#
1      | =patt%2 |
*4
`,
			`
=SCORE                   | right |
 # 4/4 @120.00 \major^c"
    1                    | c"    |
    2                    | d"    |
 #
    1                    | e"    |
    2                    | e"    |
 #
    1                    | c"    |
    2                    | d"    |
 #
    1                    | e"    |
    2                    | e"    |
 #
    1                    | c"    |
    2                    | d"    |
 #
    1                    | e"    |
    2                    | e"    |
 #
    1                    | c"    |
    2                    | d"    |
 #
    1                    | e"    |
    2                    | e"    |
 #
`,
		},
		{ // 23
			`
TRACK | x | y |
 Ch   | 1 |   |

=patt | 2    | 1   | 0  |
#
 1    |      | =.2 | 1 |
 4&   | 4   |     |    |
#
 1    | 3   |     | 2 |
 2    | 2   |     |    |

=SCORE | x          | y          |
#  \major^c"
 1     |            | =patt.1%3  |
 4&    | =!patt.1%3 |            |
*5
`,
			`
=SCORE                   | x  | y  |
 # 4/4 @120.00 \major^c"
    4&                   | f" | f" |
 #
    1                    | e" | e" |
    2                    | d" | d" |
 #
    1                    | f" |    |
    1&                   | e" |    |
    2&                   | d" |    |
    4&                   |    | f" |
 #
    1                    |    | e" |
    1&                   | f" |    |
    2                    | e" | d" |
    3                    | d" |    |
 #
    4&                   |    | f" |
 #
    1                    |    | e" |
    2                    |    | d" |
`,
		},
		{ // 24
			`
TRACK | x | y |
 Ch   | 1 |   |

=patt | 2    | 
#
 1    |      |
 4&   | 4   |     
#
 1    | 3   |   
 2    | 2   |     

=SCORE | x            | y           |
#  \major^c"
 1     |              | $from(=patt.2,2) |
 4&    | $from(=!patt.2,2) |             |
*1
`,
			`
=SCORE                   | x  | y  |
 # 4/4 @120.00 \major^c"
    1                    |    | e" |
    2                    |    | d" |
    4&                   | e" |    |
 #
    1&                   | d" |    |
`,
		},
		{ // 25
			`
TRACK | x | y |
 Ch   | 1 |   |


=SCORE | x  | y  |
#start  \major^c"
 1     |    | c" |
 4&    | d" |    |
[start]
`,
			`
=SCORE                        | x  | y  |
 #start 4/4 @120.00 \major^c"
    1                         |    | c" |
    4&                        | d" |    |
 #
    1                         |    | c" |
    4&                        | d" |    |
`,
		},
		{ // 26
			`
TRACK | x | 
 Ch   | 1 |   

=long | p |
#
1     | c' |
2     | d' |
3     | e' |
4     | f' |
#
1     | .1. |

=short | q |
# 1/4
1     | g' |
1&    | g" |


=SCORE | x  | 
#
 1     | =long.p%2| 
#
 1     | | 
 2     | /=short.q | 
#
 1     | | 
`,
			`
=SCORE                   | x  |
 # 4/4 @120.00 \major^c'
    1                    | c' |
    2                    | d' |
    3                    | e' |
    4                    | f' |
 #
    1                    | c' |
    2                    | g' |
    2&                   | g" |
    3                    | e' |
    4                    | f' |
 #
    1                    | c' |
    2                    | d' |
    3                    | e' |
    4                    | f' |
`,
		},
		{ // 27
			`
TRACK | x | 
 Ch   | 1 |   

=long | p |
#
1     | c' |
2     | d' |
3     | e' |
4     | f' |
#
1     | .1. |

=short | q |
# 2/4
2     | g' |
2&    | g" |


=SCORE | x  | 
#
 1     | =long.p%2| 
#
 1     | | 
 2     | /=!short.q | 
#
 1     | | 
`,
			`
=SCORE                   | x  |
 # 4/4 @120.00 \major^c'
    1                    | c' |
    2                    | d' |
    3                    | e' |
    4                    | f' |
 #
    1                    | c' |
    2                    | g' |
    2&                   | g" |
    4                    | f' |
 #
    1                    | c' |
    2                    | d' |
    3                    | e' |
    4                    | f' |
`,
		},
		{ // 28
			`
TRACK | x | 
 Ch   | 1 |   

=long | p |
#
1     | c' |
2     | d' |
3     | e' |
4     | f' |
#
1     | .1. |

=short | q |
# 1/4
1     | g' |
1&    | g" |


=SCORE | x  | 
#
 1     | =long.p%2| 
#
 1     | /=short.q | 
#
 1     | | 
`,
			`
=SCORE                   | x  |
 # 4/4 @120.00 \major^c'
    1                    | c' |
    2                    | d' |
    3                    | e' |
    4                    | f' |
 #
    1                    | g' |
    1&                   | g" |
    2                    | d' |
    3                    | e' |
    4                    | f' |
 #
    1                    | c' |
    2                    | d' |
    3                    | e' |
    4                    | f' |
`,
		},
		{ // 29
			`
TRACK | x | 
 Ch   | 1 |   

=long | p |
#
1     | c' |
2     | d' |
3     | e' |
4     | f' |
#
1     | .1. |

=short | q |
# 1/4
1     | g' |
1&    | g" |


=SCORE | x  | 
#
 1     | =long.p%2| 
       | /=short.q | 
#
#
 1     | | 
`,
			`
=SCORE                   | x  |
 # 4/4 @120.00 \major^c'
    1                    | g' |
    1&                   | g" |
    2                    | d' |
    3                    | e' |
    4                    | f' |
 #
    1                    | c' |
    2                    | d' |
    3                    | e' |
    4                    | f' |
 #
    1                    | c' |
    2                    | d' |
    3                    | e' |
    4                    | f' |
`,
		},
		{ // 30
			`
TRACK | x | 
 Ch   | 1 |   

=long | p |
#
1     | c' |
2     | d' |
3     | e' |
4     | f' |
#
1     | .1. |


=SCORE | x  | 
#
 1     | =long.p%2| 
 2     | /g' |
#
#
 1     | | 
`,
			`
=SCORE                   | x  |
 # 4/4 @120.00 \major^c'
    1                    | c' |
    2                    | g' |
    3                    | e' |
    4                    | f' |
 #
    1                    | c' |
    2                    | d' |
    3                    | e' |
    4                    | f' |
 #
    1                    | c' |
    2                    | d' |
    3                    | e' |
    4                    | f' |
`,
		},
		{ // 31
			`
TRACK | x | 
 Ch   | 1 |   

=long | p |
#
1     | c' |
2     | d' |
3     | e' |
4     | f' |
#
1     | .1. |


=SCORE | x  | 
#
 1     | =long.p%2| 
 2     | /* |
#
#
 1     | | 
`,
			`
=SCORE                   | x  |
 # 4/4 @120.00 \major^c'
    1                    | c' |
    2                    | *  |
    3                    | e' |
    4                    | f' |
 #
    1                    | c' |
    2                    | d' |
    3                    | e' |
    4                    | f' |
 #
    1                    | c' |
    2                    | d' |
    3                    | e' |
    4                    | f' |
`,
		},
		{ // 32
			`
TRACK | x | 
 Ch   | 1 |   

=long | p |
#
1     | c' |
2     | d' |
3     | e' |
4     | f' |
#
1     | .1. |


=SCORE | x  | 
#
 1     | =long.p%2| 
       | /g' |
#
#
 1     | | 
`,
			`
=SCORE                   | x  |
 # 4/4 @120.00 \major^c'
    1                    | g' |
    2                    | d' |
    3                    | e' |
    4                    | f' |
 #
    1                    | c' |
    2                    | d' |
    3                    | e' |
    4                    | f' |
 #
    1                    | c' |
    2                    | d' |
    3                    | e' |
    4                    | f' |
`,
		},
		{ // 33
			`
TRACK | x | 
 Ch   | 1 |   

=long | p |
#
1     | c' |
2     | d' |
3     | e' |
4     | f' |
#
1     | .1. |


=SCORE | x  | 
#
 1     | =long.p%2| 
       | /* |
#
#
 1     | | 
`,
			`
=SCORE                   | x  |
 # 4/4 @120.00 \major^c'
    1                    | *  |
    2                    | d' |
    3                    | e' |
    4                    | f' |
 #
    1                    | c' |
    2                    | d' |
    3                    | e' |
    4                    | f' |
 #
    1                    | c' |
    2                    | d' |
    3                    | e' |
    4                    | f' |
`,
		},
		{ // 34
			`
TRACK | x | 
 Ch   | 1 |   

=long | p |
#
1     | c' |
2     | d' |
3     | e' |
4     | f' |
#
1     | .1. |

.short |
ba    | (b' A)


=SCORE | x  | 
#
 1     | =long.p%2| 
       | /.short.ba |
#
#
 1     | | 
`,
			`
=SCORE                   | x      |
 # 4/4 @120.00 \major^c'
    1                    | (b' A) |
    2                    | d'     |
    3                    | e'     |
    4                    | f'     |
 #
    1                    | c'     |
    2                    | d'     |
    3                    | e'     |
    4                    | f'     |
 #
    1                    | c'     |
    2                    | d'     |
    3                    | e'     |
    4                    | f'     |
`,
		},
		{ // 35
			`
TRACK | x | 
 Ch   | 1 |   

=long | p |
#
1     | c' |
2     | d' |
3     | e' |
4     | f' |
#
1     | .1. |

.short |
ba    | (b' A)


=SCORE | x  | 
#
 1     | =long.p%2| 
 2     | /.short.ba |
#
#
 1     | | 
`,
			`
=SCORE                   | x      |
 # 4/4 @120.00 \major^c'
    1                    | c'     |
    2                    | (b' A) |
    3                    | e'     |
    4                    | f'     |
 #
    1                    | c'     |
    2                    | d'     |
    3                    | e'     |
    4                    | f'     |
 #
    1                    | c'     |
    2                    | d'     |
    3                    | e'     |
    4                    | f'     |
`,
		},
		{ // 36
			`
TRACK | x | 
 Ch   | 1 |   

=long | p |
#
1     | c' |
2     | d' |
3     | e' |
4     | f' |
#
1     | .1.|

.short |
ba    | (b' A)


=SCORE | x  | 
#
 1     | =long| 
#
#
 1     | .2. |
 2     | /.short.ba |
#
  
`,
			`
=SCORE                   | x      |
 # 4/4 @120.00 \major^c'
    1                    | c'     |
    2                    | d'     |
    3                    | e'     |
    4                    | f'     |
 #
    1                    | c'     |
    2                    | d'     |
    3                    | e'     |
    4                    | f'     |
 #
    1                    | c'     |
    2                    | (b' A) |
    3                    | e'     |
    4                    | f'     |
 #
    1                    | c'     |
    2                    | d'     |
    3                    | e'     |
    4                    | f'     |
`,
		},
		{ // 37
			`
TRACK | x | 
 Ch   | 1 |   

=patt | p |
#
1     | #1 |
2     | #2 |
3     | #1 |
4     | #3 |

=SCORE | x  | 
#
 1     | =patt((a b c)...)| 
  
`,
			`
=SCORE                   | x |
 # 4/4 @120.00 \major^c'
    1                    | a |
    2                    | b |
    3                    | a |
    4                    | c |
 `,
		},
		{ // 38
			`
TRACK | x | 
 Ch   | 1 |   

=patt | p |
#
1     | #1 |
2     | #2 |
3     | #1 |
4     | #3 |

.chords |
test   | (a b c)

=SCORE | x  | 
#
 1     | =patt(.chords.test...)| 
  
`,
			`
=SCORE                   | x |
 # 4/4 @120.00 \major^c'
    1                    | a |
    2                    | b |
    3                    | a |
    4                    | c |
 `,
		},
		{ // 39
			`
TRACK | x | 
 Ch   | 1 |   


=SCORE | x  | 
# \minor^e
 1     | 1 |
 2     | 2 |
 3     | 3 | 
  
`,
			`
=SCORE                  | x  |
 # 4/4 @120.00 \minor^e
    1                   | e  |
    2                   | f# |
    3                   | g  |
 `,
		},
		{ // 40
			`
TRACK | x |
 Ch   | 1 |

=patt | p |
#
1     | #1 |
2     | #2 |
3     | #1 |
4     | #3 |

=SCORE | x  |
#
 1     | =patt(\major^c'...)|

  `,
			`
=SCORE                   | x  |
 # 4/4 @120.00 \major^c'
    1                    | c' |
    2                    | d' |
    3                    | c' |
    4                    | e' |
   `,
		},
		{ // 41
			`
TRACK | x |
 Ch   | 1 |

=patt | p |
#
1     | #1 |
2     | #2 |
3     | #3 |
4     | #4 |
#
1     | #5 |
2     | #6 |
3     | #7 |
4     | #8 |

=SCORE | x  |
#
 1     | =patt(\minor^a...)|
#

  `,
			`
=SCORE                   | x  |
 # 4/4 @120.00 \major^c'
    1                    | a  |
    2                    | b  |
    3                    | c' |
    4                    | d' |
 #
    1                    | e' |
    2                    | f' |
    3                    | g' |
    4                    | a  |  
   `,
		},
		{ // 42
			`
'drumnote

=rhythm |a|
#
1   |$euclid(3,12,&)| 
*3

TRACK | Drums |

=SCORE |  Drums                      |
#
1      |  =rhythm(f",.drums.sn) |
2&     |                             |
#
1      |                             |
`,
			`
=SCORE                   | Drums   |
 # 4/4 @120.00 \major^c'
    1                    | f"      |
    1&                   | MN40::: |
    2                    | MN40::: |
    2&                   | MN40::: |
    3                    | f"      |
    3&                   | MN40::: |
    4                    | MN40::: |
    4&                   | MN40::: |
 #
    1                    | f"      |
    1&                   | MN40::: |
    2                    | MN40::: |
    2&                   | MN40::: |
`,
		},
		{ // 43
			`
'drumnote

=rhythm |a|
#
1   |$euclid(3,12,&)| 
*3

TRACK | Drums |

=SCORE |  Drums                      |
#
1      |  =rhythm((f" g),.drums.sn) |
2&     |                             |
#
1      |                             |
`,
			`
=SCORE                   | Drums   |
 # 4/4 @120.00 \major^c'
    1                    | (f" g)  |
    1&                   | MN40::: |
    2                    | MN40::: |
    2&                   | MN40::: |
    3                    | (f" g)  |
    3&                   | MN40::: |
    4                    | MN40::: |
    4&                   | MN40::: |
 #
    1                    | (f" g)  |
    1&                   | MN40::: |
    2                    | MN40::: |
    2&                   | MN40::: |
`,
		},
		{ // 44
			`
'drumnote

=rhythm |a|
#
1   |$euclid(3,12,&)| 
*3

TRACK | Drums |

=SCORE |  Drums                      |
#
1      |  =rhythm(.drums.kd,(.drums.sn .drums.kd)) |
2&     |                             |
#
1      |                             |
`,
			`
=SCORE                   | Drums             |
 # 4/4 @120.00 \major^c'
    1                    | MN36:::           |
    1&                   | (MN40::: MN36:::) |
    2                    | (MN40::: MN36:::) |
    2&                   | (MN40::: MN36:::) |
    3                    | MN36:::           |
    3&                   | (MN40::: MN36:::) |
    4                    | (MN40::: MN36:::) |
    4&                   | (MN40::: MN36:::) |
 #
    1                    | MN36:::           |
    1&                   | (MN40::: MN36:::) |
    2                    | (MN40::: MN36:::) |
    2&                   | (MN40::: MN36:::) |
`,
		},
		{ // 45
			`
'drumnote

=rhythm |a|
#
1   |$euclid(3,12,&)| 
*3

TRACK | Drums |

=SCORE |  Drums                      |
#
1      |  =rhythm(.drums.kd,{a,b}) |
2&     |                             |
#
1      |                             |
`,
			`
=SCORE                   | Drums   |
 # 4/4 @120.00 \major^c'
    1                    | MN36::: |
    1&                   | {a,b}   |
    2                    | {a,b}   |
    2&                   | {a,b}   |
    3                    | MN36::: |
    3&                   | {a,b}   |
    4                    | {a,b}   |
    4&                   | {a,b}   |
 #
    1                    | MN36::: |
    1&                   | {a,b}   |
    2                    | {a,b}   |
    2&                   | {a,b}   |
`,
		},
		{ // 46
			`
'drumnote

=rhythm |a|
#
1   |$euclid(3,12,&)| 
*3

TRACK | Drums |

=SCORE |  Drums                      |
#
1      |  =rhythm(.drums.kd,{.drums.kd,b}) |
2&     |                             |
#
1      |                             |
`,
			`
=SCORE                   | Drums       |
 # 4/4 @120.00 \major^c'
    1                    | MN36:::     |
    1&                   | {MN36:::,b} |
    2                    | {MN36:::,b} |
    2&                   | {MN36:::,b} |
    3                    | MN36:::     |
    3&                   | {MN36:::,b} |
    4                    | {MN36:::,b} |
    4&                   | {MN36:::,b} |
 #
    1                    | MN36:::     |
    1&                   | {MN36:::,b} |
    2                    | {MN36:::,b} |
    2&                   | {MN36:::,b} |
`,
		},
		{ // 47
			`
=melody | Verse1 |
 #
  1     | a      |
  1     | b>     |
  2     | c      |

TRACK   | Vocals |
 Channel| 1      |

@lyr  |
 a b c| V1

=SCORE | Vocals         |
 #
  1    | =melody@lyr.V1 |
`,
			`
=SCORE                   | Vocals   |
 # 4/4 @120.00 \major^c'
    1                    | (a "a")  |
    1                    | (b "b")> |
    2                    | (c "c")  |
`,
		},
		{ // 48
			`
=melody | Verse1 |
 #
  1     | a      |
  1     | b>     |
  2     | c      |

TRACK   | Vocals | Drums |
 Channel| 1      | 2     |

@lyr  |
 a b c| V1

=SCORE | Vocals         | Drums |
 #
  1    | =melody@lyr.V1 |       |
`,
			`
=SCORE                   | Drums | Vocals   |
 # 4/4 @120.00 \major^c'
    1                    |       | (a "a")  |
    1                    |       | (b "b")> |
    2                    |       | (c "c")  |
`,
		},
		{ // 49
			`
=melody  | Verse1 |
 #
    1    | a      |
    1    | b>     |
    2    | c      |

TRACK    | Vocals | Drums |
 Channel | 1      | 2     |

@lyr     |
 a b c   | V1

=SCORE   | Drums          | Vocals |
 #
    1    |           |   =melody@lyr.V1    |
`,
			`
=SCORE                   | Drums | Vocals   |
 # 4/4 @120.00 \major^c'
    1                    |       | (a "a")  |
    1                    |       | (b "b")> |
    2                    |       | (c "c")  |
`,
		},
		{ // 50
			`
=melody | Verse1 |
 #
  1     | 6^2      |
  2     | 6^-2     |

TRACK   | Vocals |
 Channel| 1      |

=SCORE  | Vocals |
 #
  1     |   =melody    |
`,
			`
=SCORE                   | Vocals |
 # 4/4 @120.00 \major^c'
    1                    | c"     |
    2                    | f'     |
`,
		},
		{ // 51
			`
      
'testdata/includes/lyrics

=melody | Verse1 |
 #
  1     | a      |
  2     | c      |

TRACK   | Vocals |
 Channel| 1      |

=SCORE       | Vocals         |
 # \major^c'
  1          | =melody@lyrics.Verse |
`,
			`
=SCORE                   | Vocals    |
 # 4/4 @120.00 \major^c'
    1                    | (a "Hi-") |
    2                    | (c "ho")  |
`,
		},
		{ // 52
			`
      

TRACK   | Voc    | Piano | Drums |
 Channel| 1      |       |       |

=SCORE | Voc       | Piano | Drums |
 # 4/4 @100
  1    | (a "Hi-") |       |       |
  2    | (c "ho")  |       |       |
'testdata/includes/score2
`,
			`
=SCORE                    | Drums | Piano | Voc       |
 # 4/4 @100.00 \major^c'
    1                     |       |       | (a "Hi-") |
    2                     |       |       | (c "ho")  |
 #C 3/4 @134.00 \major^c'
    1                     |       | f"    | e"        |
    2&                    |       | d     | c         |
 # 6/4 @100.00
    4&                    |       | f     | g'        |
`,
		},
		{ // 53
			`
TRACK | Voc | Drums | Piano |

=SCORE   | Voc | Drums   | Piano |
'testdata/includes/score2
#
1        | d   |      |     B  |

`,
			`
=SCORE                    | Drums | Piano | Voc |
 #C 3/4 @134.00 \major^c'
    1                     |       | f"    | e"  |
    2&                    |       | d     | c   |
 # 6/4 @100.00
    4&                    |       | f     | g'  |
 # 4/4 @120.00 \major^c'
    1                     |       | B     | d   |

`,
		},
		{ // 54
			`
TRACK | Voc | Piano |

=SCORE   | Voc | Piano |
'testdata/includes/score3
#
1        | d   |      |    

`,
			`
=SCORE                   | Piano | Voc |
 #A 4/4 @140.00 \major^g
    1                    | f"    | e"  |
 #
    1                    | G     | F   |
 #B 2/4 @130.00 \minor^a
    2                    | f     | g'  |
 # 4/4 @140.00 \major^g
    1                    | f"    | e"  |
 #
    1                    | G     | F   |
 # 2/4 @130.00 \minor^a
    1                    | *     | b   |
 # 4/4 @120.00 \major^c'
    1                    |       | d   |
`,
		},
		{ // 55
			`
TRACK | Voc | Piano |

=SCORE   | Voc | Piano |
# \minor^c
1        | 1  | 2    |
'testdata/includes/score3=something
#
1        | 1   |  2    |    

`,
			`
=SCORE                  | Piano | Voc |
 # 4/4 @120.00 \minor^c
    1                   | d     | c   |
 #
    1                   | d     | c   |
 #
    1                   | d     | c   |
`,
		},
		{ // 56
			`
TRACK | Voc | Piano |

=SCORE   | Voc | Piano |
# \minor^c
1        | 1  | 2    |
'testdata/includes/score3=something2
#
1        | 1   |  2    |    

`,
			`
=SCORE                  | Piano | Voc |
 # 4/4 @120.00 \minor^c
    1                   | d     | c   |
 # \major^g
    1                   | a     | g   |
 # \minor^c
    1                   | d     | c   |
`,
		},
		{ // 57
			`
TRACK | Voc | Piano |

=SCORE   | Voc | Piano |
#
1        | c  | d    |
&        | g  | a    |
#
2,        | b   |  g    |    
&        | d | f |

`,
			`
=SCORE                   | Piano | Voc |
 # 4/4 @120.00 \major^c'
    1                    | d     | c   |
    1&                   | a     | g   |
 #
    2,                   | g     | b   |
    2&                   | f     | d   |
`,
		},
		{ // 58
			`
TRACK | drums |

'drumnote
'testdata/includes/score3

=SCORE   | drums |
#
1        | =template.drums(.drums.kd,.drums.sn) |

`,
			`
=SCORE                   | drums   |
 # 4/4 @120.00 \major^c'
    1                    | MN36::: |
    2                    | MN40::: |
`,
		},
		{ // 59
			`
TRACK | drums |

'drumnote

$$embed(".drums.*")

'testdata/includes/score3

=SCORE   | drums |
#
1        | =template.drums(.kd,.sn) |

`,
			`
=SCORE                   | drums   |
 # 4/4 @120.00 \major^c'
    1                    | MN36::: |
    2                    | MN40::: |
`,
		},
		{ // 60
			`
TRACK | flute |

=SCORE | flute |
#A 4/4
 1     | a     |
 2     | c     |
#
 3     | d     |
#B 
 1     | [A]   |
*3

`,
			`
=SCORE                    | flute |
 #A 4/4 @120.00 \major^c'
    1                     | a     |
    2                     | c     |
 #
    3                     | d     |
 #B
    1                     | a     |
    2                     | c     |
 #
    3                     | d     |
 #
 #
 `,
		},
		{ // 61
			`
TRACK | flute | gita |

=SCORE | flute | gita |
#A 4/4
 1     | a     | g    |
 2     | c     | f#   |
#
 3     | d     | e    |
#B 
 1     | [A]   | d#   |
#
 2     |       | c    |
*2

`,
			`
=SCORE                    | flute | gita |
 #A 4/4 @120.00 \major^c'
    1                     | a     | g    |
    2                     | c     | f#   |
 #
    3                     | d     | e    |
 #B
    1                     | a     | d#   |
    2                     | c     |      |
 #
    2                     |       | c    |
    3                     | d     |      |
 #
 #
 `,
		},
		{ // 62
			`
TRACK | flute | gita |

=SCORE | flute | gita |
#A 4/4
 1     | a     | g    |
 2     | c     | f#   |
#
 3     | d     | e    |
#B 
 1     | [A]   | d#   |
#
 2     | G      | c    |
*2

`,
			`
=SCORE                    | flute | gita |
 #A 4/4 @120.00 \major^c'
    1                     | a     | g    |
    2                     | c     | f#   |
 #
    3                     | d     | e    |
 #B
    1                     | a     | d#   |
    2                     | c     |      |
 #
    2                     | G     | c    |
 #
 #
 `,
		},
		{ // 63
			`
TRACK | flute | gita |

=SCORE | flute | gita |
#A 4/4
 1     | a     | g    |
 2     | c     | f#   |
#
 3     | d     | e    |
#B 
 1     | [A]   | d#   |
#
 2     | /G    | c    |
*2

`,
			`
=SCORE                    | flute | gita |
 #A 4/4 @120.00 \major^c'
    1                     | a     | g    |
    2                     | c     | f#   |
 #
    3                     | d     | e    |
 #B
    1                     | a     | d#   |
    2                     | c     |      |
 #
    2                     | G     | c    |
    3                     | d     |      |
 #
 #
 `,
		},
		{ // 64
			`
TRACK | flute | gita |

=SCORE | flute | gita |
#A 4/4
 1     | a     | g    |
 2     | c     | f#   |
#
 3     | d     | e    |
#B 
 1     | [A]   | d#   |
 1&    |       | [!A] |
#
 2     | /G    | c    |
*2

`,
			`
=SCORE                    | flute | gita |
 #A 4/4 @120.00 \major^c'
    1                     | a     | g    |
    2                     | c     | f#   |
 #
    3                     | d     | e    |
 #B
    1                     | a     | d#   |
    1&                    |       | g    |
    2                     | c     |      |
    2&                    |       | f#   |
 #
    2                     | G     | c    |
    3                     | d     |      |
 #
 #
 `,
		},
		{ // 65
			`
TRACK | flute | gita |

=SCORE | flute | gita |
#A 4/4
 1     | a     | g    |
 2     | c     | f#   |
#
 3     | d     | e    |
#B 
 1     | [A]%2   | d#   |
 1&    |       | [!A] |
#
 2     |       | c    |
*2

`,
			`
=SCORE                    | flute | gita |
 #A 4/4 @120.00 \major^c'
    1                     | a     | g    |
    2                     | c     | f#   |
 #
    3                     | d     | e    |
 #B
    1                     | a     | d#   |
    1&                    |       | g    |
    2                     | c     |      |
    2&                    |       | f#   |
 #
    2                     |       | c    |
    3                     | d     |      |
 #
    1                     | a     |      |
    2                     | c     |      |
 #
    3                     | d     |      |
 `,
		},
		{ // 66
			`
TRACK | flute | gita |

=SCORE | flute | gita |
#A 4/4
 1     | a     | g    |
 2     | c     | f#   |
#
 3     | d     | e    |
#B 
 1     | [A]%2   | d#   |
 1&    |       | [!A] |
#
 2     | /G    | c    |
*2

`,
			`
=SCORE                    | flute | gita |
 #A 4/4 @120.00 \major^c'
    1                     | a     | g    |
    2                     | c     | f#   |
 #
    3                     | d     | e    |
 #B
    1                     | a     | d#   |
    1&                    |       | g    |
    2                     | c     |      |
    2&                    |       | f#   |
 #
    2                     | G     | c    |
    3                     | d     |      |
 #
    1                     | a     |      |
    2                     | c     |      |
 #
    3                     | d     |      |
 `,
		},
		{ // 67
			`
TRACK | flute | 

=SCORE | flute | 
# \major^g'
 1     | 1    |
 1& \major^c' | 1 |
 2     | 2     | 
#
 3     | 2     | 


`,
			`
=SCORE                   | flute |
 # 4/4 @120.00 \major^g'
    1                    | g'    |
    1& \major^c'         |       |
    1&                   | c'    |
    2                    | d'    |
 #
    3                    | d'    |
 `,
		},
		{ // 68
			`
TRACK | flute | 

=SCORE | flute | 
# @110
 1     | 1    |
 1& @115 | 1 |
 2     | 2     | 
#
 3     | 2     | 


`,
			`
=SCORE                   | flute |
 # 4/4 @110.00 \major^c'
    1                    | c'    |
    1& @115.00           |       |
    1&                   | c'    |
    2                    | d'    |
 #
    3                    | d'    |
 `,
		},
		{ // 69
			`
TRACK | flute | 

=SCORE | flute | 
# @110
 1     | 1    |
 1& @115 |  |
 2     | 2     | 
#
 3     | 2     | 


`,
			`
=SCORE                   | flute |
 # 4/4 @110.00 \major^c'
    1                    | c'    |
    1& @115.00           |       |
    2                    | d'    |
 #
    3                    | d'    |
 `,
		},
		{ // 70
			`
TRACK | flute | 

=SCORE | flute | 
# \major^g'
 1     | 1    |
 1& \major^c' |  |
 2     | 2     | 
#
 3     | 2     | 


`,
			`
=SCORE                   | flute |
 # 4/4 @120.00 \major^g'
    1                    | g'    |
    1& \major^c'         |       |
    2                    | d'    |
 #
    3                    | d'    |
 `,
		},
		{ // 71
			`
TRACK | flute |

=test | flute |
#
1     | c     |
2     | d     |
3     | e     |
3&    | f     |
4     | g     | 

=SCORE | flute | 
#
 1     | $reverse(=test.flute) |


`,
			`
=SCORE                   | flute |
 # 4/4 @120.00 \major^c'
    1                    | g     |
    2                    | f     |
    3                    | e     |
    3&                   | d     |
    4                    | c     |
 `,
		},
		{ // 72
			`
TRACK | flute |

=test | flute |
#
1     | c     |
2     | d     |
#
3     | e     |
3&    | f     |
4     | g     | 

=SCORE | flute | 
#
 1     | $reverse(=test.flute) |
#

`,
			`
=SCORE                   | flute |
 # 4/4 @120.00 \major^c'
    1                    | g     |
    2                    | f     |
 #
    3                    | e     |
    3&                   | d     |
    4                    | c     |
 `,
		},
		{ // 73
			`
TRACK | flute |

=test | a  | b  |
#
1     | a  |    |
1&    |    | b  |
2     | A  |    |
2&    |    | B  |
3     | a  |    |
3&    |    | b  |
4     | a' |    |
4&    |    | b' |

=SCORE | flute | 
#
 1     | $merge(=test.a,=test.b) |


`,
			`
=SCORE                   | flute |
 # 4/4 @120.00 \major^c'
    1                    | a     |
    1&                   | b     |
    2                    | A     |
    2&                   | B     |
    3                    | a     |
    3&                   | b     |
    4                    | a'    |
    4&                   | b'    |
 `,
		},
		{ // 74
			`
TRACK | flute |

=test | a  | b  |
#
1     | a  |    |
1&    |    | b  |
2     | A  |    |
2&    |    | B  |
3     | a  |    |
3&    |    | b  |
4     | a' |    |
4&    |    | b' |

=SCORE | flute | 
#
 1     | $merge(=test.a,=test.b)/$reverse(=) |


`,
			`
=SCORE                   | flute |
 # 4/4 @120.00 \major^c'
    1                    | b'    |
    1&                   | a'    |
    2                    | b     |
    2&                   | a     |
    3                    | B     |
    3&                   | A     |
    4                    | b     |
    4&                   | a     |
 `,
		},
		{ // 75
			`
TRACK | flute |

=arp |   |
#
1     | #1  |
1&    | #3  |
2     | #2  |

.chords |
 T      | (1 3 5)

=SCORE | flute | 
#
 1     | =arp(.chords.T...) |


`,
			`
=SCORE                   | flute |
 # 4/4 @120.00 \major^c'
    1                    | c'    |
    1&                   | g'    |
    2                    | e'    |
 `,
		},
		{ // 76
			`
TRACK | flute |

=arp |   |
1     | #1  |
1&    | #3  |
2     | #2  |

.chords |
 T      | (1 3 5)

=SCORE | flute | 
 1     | =arp(.chords.T...) |


`,
			`
=SCORE                   | flute |
 # 4/4 @120.00 \major^c'
    1                    | c'    |
    1&                   | g'    |
    2                    | e'    |
 `,
		},
		{ // 77
			`
TRACK | flute |

=SCORE | flute | 
 1     | (cb c#) |
 2     | c#      |
 3     | db      |


`,
			`
=SCORE                   | flute   |
 # 4/4 @120.00 \major^c'
    1                    | (cb c#) |
    2                    | c#      |
    3                    | db      |
 `,
		},
		{ // 78
			`
TRACK | flute |

.test     | a  | b   |
X         | a" | b"' |

=SCORE | flute | 
 1     | .test.X.a |
 2     | .test.X.b      |
 3     | *      |


`,
			`
=SCORE                   | flute |
 # 4/4 @120.00 \major^c'
    1                    | a"    |
    2                    | b"'   |
    3                    | *     |
 `,
		},
		{ // 79
			`
TRACK | bass | flute |

.test     | flute  | bass   |
X         | a" | b"' |

=SCORE | bass     | flute    |
 1     | .test.X. | .test.X. |
 2     | *        | *        |


`,
			`
=SCORE                   | bass | flute |
 # 4/4 @120.00 \major^c'
    1                    | b"'  | a"    |
    2                    | *    | *     |
 `,
		},
		{ // 80
			`
TRACK  | bass | flute |
Import |      | bass  |

.test     | flute  | bass   |
X         | a" | b"' |

=SCORE | bass     |
 1     | .test.X. |
 2     | *        |


`,
			`
=SCORE                   | bass | flute |
 # 4/4 @120.00 \major^c'
    1                    | b"'  | a"    |
    2                    | *    | *     |
 `,
		},
		{ // 81
			`
'drumnote
			
TRACK  | dr |

=test     | 1                | 2                             |
   1      | $euclid(3,8,0.5) | =.1(.drums.kd++,.drums.sn+)   |
 #
   1      |                  | ...                           |
#
#

=SCORE | dr     |
 1     | =test.2 |
 *3

`,
			`
=SCORE                   | dr        |
 # 4/4 @120.00 \major^c'
    1                    | MN36++::: |
    1&                   | MN40+:::  |
    2                    | MN40+:::  |
    2&                   | MN36++::: |
    3                    | MN40+:::  |
    3&                   | MN40+:::  |
    4                    | MN36++::: |
    4&                   | MN40+:::  |
 # \major^c'
    1                    | MN36++::: |
    1&                   | MN40+:::  |
    2                    | MN40+:::  |
    2&                   | MN36++::: |
    3                    | MN40+:::  |
    3&                   | MN40+:::  |
    4                    | MN36++::: |
    4&                   | MN40+:::  |
 #
    1                    | MN36++::: |
    1&                   | MN40+:::  |
    2                    | MN40+:::  |
    2&                   | MN36++::: |
    3                    | MN40+:::  |
    3&                   | MN40+:::  |
    4                    | MN36++::: |
    4&                   | MN40+:::  |
 #
    1                    | MN36++::: |
    1&                   | MN40+:::  |
    2                    | MN40+:::  |
    2&                   | MN36++::: |
    3                    | MN40+:::  |
    3&                   | MN40+:::  |
    4                    | MN36++::: |
    4&                   | MN40+:::  |
 `,
		},
		{ // 82
			`
TRACK  | voc |

=patt |  |
#
  1   | #1 |
  2   | #3 |
#
  1&  | #4 |

=SCORE | voc     |
 1     | $rabbit(=patt,c',3,80%,5) |
 *2

`,
			`
=SCORE                   | voc |
 # 4/4 @120.00 \major^c'
    1                    | c'  |
    2                    | A#" |
 # \major^c'
    1&                   | C#" |
 #
  `,
		},
		{ // 83
			`
TRACK  | voc |

=patt |  |
#
  1   | e' |
  2   | f' |
#
  1&  | c' |

=SCORE | voc                       |
 1     | $mirror(=patt,d') |
 *2

`,
			`
=SCORE                   | voc |
 # 4/4 @120.00 \major^c'
    1                    | c'  |
    2                    | b   |
 # \major^c'
    1&                   | e'  |
 #
  `,
		},
		{ // 84
			`
TRACK  | voc |

=patt |  |
#
  1   | (e' f' c') |

=SCORE | voc                       |
 1     | $mirror(=patt,d') |

`,
			`
=SCORE                   | voc       |
 # 4/4 @120.00 \major^c'
    1                    | (c' b e') |
  `,
		},
		{ // 85
			`
TRACK  | voc |

=patt |  |
#
  1   | {e',f',c'} |
  1&  | *          |

=SCORE | voc                       |
 1     | $mirror(=patt,d') |

`,
			`
=SCORE                   | voc       |
 # 4/4 @120.00 \major^c'
    1                    | {c',b,e'} |
    1&                   | *         |
  `,
		},
		{ // 86
			`
TRACK  | voc |

=patt |  |
#
  1   | (f a c') |

=SCORE | voc                       |
 1     | $negative(=patt,c') |

`,
			`
=SCORE                   | voc         |
 # 4/4 @120.00 \major^c'
    1                    | (d" a#' g') |
  `,
		},
		{ // 87
			`
TRACK  | voc |

=patt |  |
#
  1   | (A C# E G) |

=SCORE | voc                       |
 1     | $negative(=patt,c') |

`,
			`
=SCORE                   | voc                 |
 # 4/4 @120.00 \major^c'
    1                    | (a#" f#"' d#"' c"') |
  `,
		},
		{ // 88
			`
TRACK  | voc |

=patt |  |
#
  1   | (E B e g) |
  2   | (G B d f) |
  3   | (C G c e) |

=SCORE | voc                       |
 1     | $negative(=patt,c) |

`,
			`
=SCORE                   | voc           |
 # 4/4 @120.00 \major^c'
    1                    | (d#' g# d# c) |
    2                    | (c' g# f d)   |
    3                    | (g' c' g d#)  |
  `,
		},
		{ // 89
			`
TRACK  | voc |

=patt |  |
#
  1   | {C,D,E,F,G,A,B,c} |

=SCORE | voc                       |
 1     | $negative(=patt,C) |

`,
			`
=SCORE                   | voc                     |
 # 4/4 @120.00 \major^c'
    1                    | {G,F,D#,D,C,A#',G#',G'} |
  `,
		},
		{ // 90
			`
TRACK  | voc | git |

=patt | 1 | 2 |
#
  1   | c | d |
  2   | d |   |
#
  1   | _ | e |
#
  1   |   | f |

=SCORE | git       | voc       |
#
 1     | =patt.1%5 | =patt.2%2 |
*4
#
 1     | *         | *         |

`,
			`
=SCORE                   | git | voc |
 # 4/4 @120.00 \major^c'
    1                    | c   | d   |
    2                    | d   |     |
 #
    1                    | c   | e   |
    2                    | d   |     |
 #
    1                    | c   | f   |
    2                    | d   |     |
 #
    1                    | c   | d   |
    2                    | d   |     |
 #
    1                    | c   | e   |
    2                    | d   |     |
 #
    1                    | *   | *   |
  `,
		},
		{ // 91
			`
TRACK  | voc | git |

=patt | 1  |
#
  1   | c  |
#
  1   | d  |
#
  1   | e  |

=SCORE | voc     |
#A
 1     | =patt.1 |
*1
#
 1     | .2.     |
*1
[A]

`,
			`
=SCORE                    | voc |
 #A 4/4 @120.00 \major^c'
    1                     | c   |
 #
    1                     | d   |
 #
    1                     | c   |
 #
    1                     | d   |
 #
    1                     | c   |
 #
    1                     | d   |
 #
    1                     | c   |
 #
    1                     | d   |
  `,
		},
		{ // 92
			`
TRACK  | voc | git |

=patt | 1  |
#
  1   | c  |
#
  1   | d  |
#
  1   | e  |
#
  1   | f  |
#
  1   | g  |
#
  1   | b  |
#
  1   | c# |
#
  1   | d# |

=SCORE | voc     | git        |
#A
 1     | =patt.1 | =patt.1%10 |
*3
#
 1     | .4.     |            |
*3
[A]

`,
			`
=SCORE                    | git | voc |
 #A 4/4 @120.00 \major^c'
    1                     | c   | c   |
 #
    1                     | d   | d   |
 #
    1                     | e   | e   |
 #
    1                     | f   | f   |
 #
    1                     | g   | c   |
 #
    1                     | b   | d   |
 #
    1                     | c#  | e   |
 #
    1                     | d#  | f   |
 #
    1                     | c   | c   |
 #
    1                     | d   | d   |
 #
    1                     | e   | e   |
 #
    1                     | f   | f   |
 #
    1                     | g   | c   |
 #
    1                     | b   | d   |
 #
    1                     | c#  | e   |
 #
    1                     | d#  | f   |
  `,
		},
		{ // 93
			`
TRACK |p|
Ch |1  |

=SCORE |p|
#A
1  |c" |
2  |d  |
[A]
#
1  | c' |
2 |* |
[A]
			`,
			`
=SCORE                    | p  |
 #A 4/4 @120.00 \major^c'
    1                     | c" |
    2                     | d  |
 #
    1                     | c" |
    2                     | d  |
 #
    1                     | c' |
    2                     | *  |
 #
    1                     | c" |
    2                     | d  |
`,
		},
		{ // 94
			`
TRACK |p|
Ch |1  |

=SCORE |p|
#
1  | \(c e g) |
2  | (1 2 3)   |
#
1  | \(g a b) |
2  | (2 3)     |
3  | *         |
			`,
			`
=SCORE                   | p       |
 # 4/4 @120.00 \major^c'
    2                    | (c e g) |
 #
    2                    | (a b)   |
    3                    | *       |
`,
		},
		{ // 95
			`
TRACK |p|
Ch |1  |

=SCORE |p|
# \(c e g)
1  |  |
2  | (1 2 3)   |
# \(g a b)
1  |  |
2  | (2 3)     |
3  | *         |
			`,
			`
=SCORE                  | p       |
 # 4/4 @120.00 \(c e g)
    2                   | (c e g) |
 # \(g a b)
    2                   | (a b)   |
    3                   | *       |
`,
		},
		{ // 96
			`
TRACK |p|
Ch |1  |

=SCORE |p|
# @100~
1  | 1 |
2  | 2 |
# @140
2  | (2 3)     |
3  | *         |
			`,
			`
=SCORE                    | p       |
 # 4/4 @100.00~ \major^c'
    1                     | c'      |
    2                     | d'      |
 # @140.00
    2                     | (d' e') |
    3                     | *       |
`,
		},
		{ // 97
			`
TRACK |p|
Ch |1  |
Ambitus |from:C' to:c' cut:false|

=SCORE |p|
#
1  | D" |
2  | d' |
			`,
			`
=SCORE                   | p  |
 # 4/4 @120.00 \major^c'
    1                    | D' |
    2                    | d  |
`,
		},
		{ // 98
			`
TRACK |p|
Ch |1  |
Ambitus |from:C' to:c' cut:true|

=SCORE |p|
#
1  | D" |
1& | c  |
2  | d' |
			`,
			`
=SCORE                   | p |
 # 4/4 @120.00 \major^c'
    1&                   | c |
`,
		},
		{ // 99
			`
TRACK | p |
Ch    | 1 |

=t  | 1  |
#
 1  | #1 |
 1& | #2 |
 2  | #3 |
 3  | _  |

=SCORE | p            |
#
    1  | =t(a,c,c)%3  |
#
    1  |              |
    3  | *            |
			`,
			`
=SCORE                   | p |
 # 4/4 @120.00 \major^c'
    1                    | a |
    1&                   | c |
    2                    | c |
    3                    | a |
    3&                   | c |
    4                    | c |
 #
    1                    | a |
    1&                   | c |
    2                    | c |
    3                    | * |
`,
		},
		{ // 100
			`
TRACK    | piano |
 Channel | 1     |

=motiv   |    |
 #
    1    | #1 |
    1&   | #2 |
    2    | #3 |
    3    | _  |

=SCORE       | piano              |
 # \minor^c'
    4        | =!motiv(1,-1,-1)%4 |
 *1
 #
    1&       | 5                  |


			`,
			`
=SCORE                   | piano |
 # 4/4 @120.00 \minor^c'
    4                    | c'    |
    4&                   | a#    |
 #
    1                    | a#    |
    2                    | c'    |
    2&                   | a#    |
    3                    | a#    |
    4                    | c'    |
    4&                   | a#    |
 #
    1                    | a#    |
    1&                   | g'    |
`,
		},
		{ // 101
			`
TRACK    | piano |
 Channel | 1     |

=motiv   |    |
 #
    1    |    |
    1&   |    |
    2    |    |
    3    |    |
    4    | #1 |
    4&   | #2 |
 #
    1    | #3 |
    2    | _  |

=SCORE       | piano              |
 # \minor^c'
    4        | =!motiv(1,-1,-1)%3 |
 *2
 #
    1        | 5                  |


			`,
			`
=SCORE                   | piano |
 # 4/4 @120.00 \minor^c'
    4                    | c'    |
    4&                   | a#    |
 #
    1                    | a#    |
    2                    | c'    |
    2&                   | a#    |
    3                    | a#    |
    4                    | c'    |
    4&                   | a#    |
 #
    1                    | a#    |
 #
    1                    | g'    |
`,
		},
		{ // 102
			`
TRACK    | piano |
 Channel | 1     |

=motiv   |    |
 #
    1    |    |
    1&   |    |
    2    |    |
    3    |    |
    4    | #1 |
    4&   | #2 |
 #
    1    | #3 |
    2    | _  |

=SCORE       | piano              |
 # \minor^c'
    1        | =motiv(1,-1,-1)%3 |
 *2
 #
    1        | 5                  |


			`,
			`
=SCORE                   | piano |
 # 4/4 @120.00 \minor^c'
    4                    | c'    |
    4&                   | a#    |
 #
    1                    | a#    |
 #
    1                    | c'    |
    1&                   | a#    |
    2                    | a#    |
 #
    1                    | g'    |
`,
		},
		{ // 103
			`
'drumnote.drums

TRACK | Voc | Drums | Piano |

=SCORE | Voc | Drums | Piano |
#
1  | a"  | .drums.kd    | b"    |
2& | c   | .drums.sn    | d     |
#
1  |     | .drums.ho    |       |
`,
			`
=SCORE                   | Drums   | Piano | Voc |
 # 4/4 @120.00 \major^c'
    1                    | MN36::: | b"    | a"  |
    2&                   | MN40::: | d     | c   |
 #
    1                    | MN46::: |       |     |
`,
		},
		{ // 104
			`
'testdata/includes/score3=template

TRACK | Drums |

=SCORE | Drums |
#
1  |  =template(MN36:::,MN40:::)    | 
`,
			`
=SCORE                   | Drums   |
 # 4/4 @120.00 \major^c'
    1                    | MN36::: |
    2                    | MN40::: |
`,
		},
		{ // 105
			`

TRACK | Drums |
Channel | 10  |

'drumnote

.short | |
a | .drums.kd |

=SCORE | Drums |
#
1  |  .short.a    | 
`,
			`
=SCORE                   | Drums   |
 # 4/4 @120.00 \major^c'
    1                    | MN36::: |
`,
		},
		{ // 106
			`

TRACK | Drums |
Channel | 10  |

=SCORE | Drums |
#
1  |  'drumnote.drums.kd    | 
`,
			`
=SCORE                   | Drums   |
 # 4/4 @120.00 \major^c'
    1                    | MN36::: |
`,
		},
		{ // 107
			`

TRACK | Drums |
Channel | 10  |

.short | |
kick      | 'drumnote.drums.kd |

=SCORE | Drums |
#
1  |  .short.kick    | 
`,
			`
=SCORE                   | Drums   |
 # 4/4 @120.00 \major^c'
    1                    | MN36::: |
`,
		},
		{ // 108
			`

TRACK | Drums |

=SCORE | Drums |
#
1  |  'testdata/includes/score3=template.drums(MN36:::,MN40:::)    | 
`,
			`
=SCORE                   | Drums   |
 # 4/4 @120.00 \major^c'
    1                    | MN36::: |
    2                    | MN40::: |
`,
		},
		{ // 109
			`

TRACK | drums |

=SCORE | drums |
#
1  |  'testdata/includes/score3=template.(MN36:::,MN40:::)    | 
`,
			`
=SCORE                   | drums   |
 # 4/4 @120.00 \major^c'
    1                    | MN36::: |
    2                    | MN40::: |
`,
		},
		{ // 110
			`

TRACK | vocals |

=SCORE | vocals |
#
1  |  'testdata/includes/score   | 
`,
			`
=SCORE                   | vocals |
 # 4/4 @120.00 \major^c'
    1                    | e"     |
    2&                   | c      |
`,
		},
		{ // 111
			`

TRACK | Piano |

=SCORE | Piano |
#
1  |  'testdata/includes/score=.   | 
`,
			`
=SCORE                   | Piano |
 # 4/4 @120.00 \major^c'
    1                    | f"    |
    2&                   | d     |
`,
		},
		{ // 112
			`

TRACK | Piano | Voc

=SCORE | Piano Voc |
#
1  |  'testdata/includes/score=.   | 
`,
			`
=SCORE                   | Piano | Voc |
 # 4/4 @120.00 \major^c'
    1                    | f"    | e"  |
    2&                   | d     | c   |
`,
		},
		{ // 113
			`

TRACK | Piano | Voc

=SCORE | Piano Voc |
#
1  |  f"   | 
`,
			`
=SCORE                   | Piano | Voc |
 # 4/4 @120.00 \major^c'
    1                    | f"    | f"  |
`,
		},
		{ // 114 TODO: fix this!
			`

TRACK | Piano | Voc

=SCORE | Piano Voc |
#
1  |  'testdata/includes/score4=.   | 
`,
			`
=SCORE                   | Piano | Voc |
 # 4/4 @120.00 \major^c'
    1                    | e"    | e"  |
    2&                   | c     | c   |
`,
		},
		{ // 115 solo < 0
			`

| TRACK | Piano | Voc |
| solo  |  -1   |     |

| =SCORE | Piano Voc |
| #      |           |
| 1  |  'testdata/includes/score4=.   | 
`,
			`
| =SCORE                  | Piano | Voc |
| # 4/4 @120.00 \major^c' |       |     |
|    1                    | e"    | e"  |
|    2&                   | c     | c   |
`,
		},
		{ // 116 solo < 0
			`
'drumnote.drums

| TRACK | Voc | Drums | Piano |
| solo  | -1  |       |       |

| =SCORE | Voc | Drums | Piano |
| #   |      |           |       |
| 1  | a"  | .drums.kd    | b"    |
| 2& | c   | .drums.sn    | d     |
| #
| 1  |     | .drums.ho    |       |
`,
			`
| =SCORE                   | Drums   | Piano |
| ------------------------ | ------- | ----- |
|  # 4/4 @120.00 \major^c' |         |       |
|     1                    | MN36::: | b"    |
|     2&                   | MN40::: | d     |
|  #                       |         |       |
|     1                    | MN46::: |       |
`,
		},

		/*
		    - firstsync scheint nicht richtig zu funktionieren
		   (z.B. einbetten von =!patt1 auf 4&, wobei patt1 patt2 einbindet ohne ! auf 1 und patt2 erste note auf 4& hat)
		*/
		/*
		       { // 19
		         `
		   TRACK | left |
		    Ch   | 1    |

		   =patt  | a    |
		   #
		   1      | ^1   |
		   2      | ^2   |

		   =SCORE | left |
		   #
		   1      | =patt.a^1   |
		   #
		   1      | =patt.a^2   |
		   `,
		         `
		   =SCORE | left |
		   # 4/4 @120.00
		    1     | c'   |
		    2     | d'   |
		   #
		    1     | d'   |
		    2     | e'   |
		   `,
		       },
		*/
	}

	items.DEBUG = false

	skip := map[int]bool{
		//36: true,
		//91: true,
		//92: true,
		//90:  true,
		//99:  true,
		//100: true,
		//101: true,
		//24: true,
		//112: true,
		114: true,
	}

	for i, test := range tests {

		//fmt.Printf("############ %v ###############\n", i)

		/*
			if i != 114 {
				continue
			}
		*/

		if i > 0 && i < 116 {
			continue
		}
		/*
		 */
		if skip[i] {
			continue
		}

		/*
		   sc, err := muskel.Parse(strings.NewReader(strings.TrimSpace(test.input)), "include-main")

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

		   err = muskel.WriteFormattedTo(unr, &bf)

		   if err != nil {
		     t.Errorf("[%v] could not format unrolled score: %s\n%s\n", i, err.Error(), test.input)
		     continue
		   }

		   result := bf.String()
		*/

		var bf strings.Builder

		err := muskel.Unroll("include-main", nil, strings.NewReader(strings.TrimSpace(test.input)), &bf)

		if err != nil {
			t.Errorf("[%v] could not format score: %s\n%s\n", i, err.Error(), test.input)
			continue
		}

		result := bf.String()

		//    fmt.Println(result)

		//res := strings.Split(result, "=\n")
		//exp := strings.Split(test.expected, "=\n")

		got := "| =SCORE" + strings.Split(strings.TrimSpace(result), "=SCORE")[1]
		wanted := strings.TrimSpace(test.expected)

		if got != wanted {
			t.Errorf("[%v] score\n%s\n\nunrolled gives \n%s\n\nbut this was expected:\n%s\n%q\nvs\n%q\n", i, test.input, got, wanted, got, wanted)
			//      t.Errorf("[%v] score\n%s\n\nunrolled gives \n%q\n\nbut this was expected:\n%q\n", i, test.input, got, wanted)
		}
	}

	items.DEBUG = false
}

func TestSolo(t *testing.T) {
	tests := []struct {
		sologroupID uint
		input       string
		expected    string
	}{
		{ // 0
			1,
			`
'drumnote.drums

| TRACK | Voc | Drums | Piano |
| solo  | 1  |    2   |   1    |

| =SCORE | Voc | Drums | Piano |
| #   |      |           |       |
| 1  | a"  | .drums.kd    | b"    |
| 2& | c   | .drums.sn    | d     |
| #
| 1  |     | .drums.ho    |       |
`,
			`
| =SCORE                   | Piano | Voc |
| ------------------------ | ----- | --- |
|  # 4/4 @120.00 \major^c' |       |     |
|     1                    | b"    | a"  |
|     2&                   | d     | c   |
|  #                       |       |     |
`,
		},
		{ // 1
			0,
			`
'drumnote.drums

| TRACK | Voc | Drums | Piano |
| solo  | 1  |    2   |   1    |

|=SCORE | Voc | Drums | Piano |
|#
|1  | a"  | .drums.kd    | b"    |
|2& | c   | .drums.sn    | d     |
|#
|1  |     | .drums.ho    |       |
`,
			`
| =SCORE                   | Drums   | Piano | Voc |
| ------------------------ | ------- | ----- | --- |
|  # 4/4 @120.00 \major^c' |         |       |     |
|     1                    | MN36::: | b"    | a"  |
|     2&                   | MN40::: | d     | c   |
|  #                       |         |       |     |
|     1                    | MN46::: |       |     |
`,
		},
	}

	items.DEBUG = false

	skip := map[int]bool{}

	for i, test := range tests {

		/*
		 */
		if skip[i] {
			continue
		}

		var bf strings.Builder

		err := muskel.Unroll("include-main", nil, strings.NewReader(strings.TrimSpace(test.input)), &bf, score.SoloGroup(test.sologroupID))

		if err != nil {
			t.Errorf("[%v] could not format score: %s\n%s\n", i, err.Error(), test.input)
			continue
		}

		result := bf.String()

		got := "| =SCORE" + strings.Split(strings.TrimSpace(result), "=SCORE")[1]
		wanted := strings.TrimSpace(test.expected)

		if got != wanted {
			t.Errorf("[%v] score\n%s\n\nunrolled gives \n%s\n\nbut this was expected:\n%s\n%q\nvs\n%q\n", i, test.input, got, wanted, got, wanted)
			//      t.Errorf("[%v] score\n%s\n\nunrolled gives \n%q\n\nbut this was expected:\n%q\n", i, test.input, got, wanted)
		}
	}

	items.DEBUG = false
}

func TestStartEnd(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{ // 0
			`
'drumnote.drums

| TRACK | Voc | Drums | Piano |

| =SCORE | Voc | Drums | Piano |
| #   |      |           |       |
| 1  | a"  | .drums.kd    | g"    |
| 2& | d   | .drums.sn    | d     |
| #>   |      |           |       |
| 1  | a"  | .drums.kd    | b"    |
| 2& | c   | .drums.sn    | d     |
| #
| 1  |     | .drums.ho    |       |
`,
			`
| =SCORE                    | Drums   | Piano | Voc |
| ------------------------- | ------- | ----- | --- |
|  #> 4/4 @120.00 \major^c' |         |       |     |
|     1                     | MN36::: | b"    | a"  |
|     2&                    | MN40::: | d     | c   |
|  #                        |         |       |     |
|     1                     | MN46::: |       |     |
`,
		},
		{ // 1
			`
'drumnote.drums

| TRACK | Voc | Drums | Piano |
| solo  | 1  |    2   |   1    |

|=SCORE | Voc | Drums | Piano |
|#
|1  | a"  | .drums.kd    | b"    |
|2& | c   | .drums.sn    | d     |
|#_ |    |               |       |
|1  | f"  | .drums.kd    | g"    |
|2& | c   | .drums.sn    | d     |
|#
|1  |     | .drums.ho    |       |
`,
			`
| =SCORE                   | Drums   | Piano | Voc |
| ------------------------ | ------- | ----- | --- |
|  # 4/4 @120.00 \major^c' |         |       |     |
|     1                    | MN36::: | b"    | a"  |
|     2&                   | MN40::: | d     | c   |
`,
		},
		{ // 2
			`
'drumnote.drums

| TRACK | Voc | Drums | Piano |

| =SCORE | Voc | Drums | Piano |
| #   |      |           |       |
| 1  | a"  | .drums.kd    | g"    |
| 2& | d   | .drums.sn    | d     |
| #   |      |           |       |
| 1  | a"  | .drums.kd    | g"    |
| 2& | d   | .drums.sn    | d     |
| #>   |      |           |       |
| 1  | a"  | .drums.kd    | b"    |
| 2& | c   | .drums.sn    | d     |
| #
| 1  |     | .drums.ho    |       |
`,
			`
| =SCORE                    | Drums   | Piano | Voc |
| ------------------------- | ------- | ----- | --- |
|  #> 4/4 @120.00 \major^c' |         |       |     |
|     1                     | MN36::: | b"    | a"  |
|     2&                    | MN40::: | d     | c   |
|  #                        |         |       |     |
|     1                     | MN46::: |       |     |
`,
		},
	}

	items.DEBUG = false

	skip := map[int]bool{
		// 2: true,
	}

	for i, test := range tests {

		/*
		 */
		if skip[i] {
			continue
		}

		var bf strings.Builder

		err := muskel.Unroll("include-main", nil, strings.NewReader(strings.TrimSpace(test.input)), &bf, score.CutOut())

		if err != nil {
			t.Errorf("[%v] could not format score: %s\n%s\n", i, err.Error(), test.input)
			continue
		}

		result := bf.String()

		got := "| =SCORE" + strings.Split(strings.TrimSpace(result), "=SCORE")[1]
		wanted := strings.TrimSpace(test.expected)

		if got != wanted {
			t.Errorf("[%v] score\n%s\n\nunrolled gives \n%s\n\nbut this was expected:\n%s\n%q\nvs\n%q\n", i, test.input, got, wanted, got, wanted)
			//      t.Errorf("[%v] score\n%s\n\nunrolled gives \n%q\n\nbut this was expected:\n%q\n", i, test.input, got, wanted)
		}
	}

	items.DEBUG = false
}

func TestInclude2(t *testing.T) {
	//t.Skip()
	tests := []struct {
		input    string
		expected string
	}{
		{ // 0
			`
'drumnote

|TRACK | Voc | Drums | Piano |

|=SCORE | Voc | Drums | Piano |
|#
|1      | a"  | .drums.kd    | b"    |
|2&     | c   | .drums.sn    | d     |
|#
|1      |     | .drums.ho    |       |
`,
			`
'drumnote

| TRACK    | Voc | Drums | Piano |
| -------- | --- | ----- | ----- |

| =SCORE   | Voc | Drums     | Piano |
| -------- | --- | --------- | ----- |
|  #       |     |           |       |
|     1    | a"  | .drums.kd | b"    |
|     2&   | c   | .drums.sn | d     |
|  #       |     |           |       |
|     1    |     | .drums.ho |       |

`,
		},
		{ // 1
			`
'testdata/includes/header
'drumnote

|=hu | a        |
|#
|1   | .drums.sn |
|2   | .drums.kd |

|TRACK | Voc | Drums | Piano |

|=SCORE   | Voc | Drums | Piano |
|#
| 1       | a"  | .drums.kd    | b"    |
| 2&      | c   | .drums.sn    | d     |
|#
| 1       |     | =hu.a       |       |

`,
			`
'testdata/includes/header
'drumnote

| =hu      | a         |
| -------- | --------- |
|  #       |           |
|     1    | .drums.sn |
|     2    | .drums.kd |

| TRACK    | Voc | Drums | Piano |
| -------- | --- | ----- | ----- |

| =SCORE   | Voc | Drums     | Piano |
| -------- | --- | --------- | ----- |
|  #       |     |           |       |
|     1    | a"  | .drums.kd | b"    |
|     2&   | c   | .drums.sn | d     |
|  #       |     |           |       |
|     1    |     | =hu.a     |       |

`,
		},
		{ // 2
			`
|TRACK | Voc | Drums | Piano |

|=SCORE | Voc | Drums | Piano |
|# 4/4
|1  | a"  | d    | b"    |
|2& | c   |      | d     |
|'testdata/includes/score
|#
|1  | d   |      |       |

`,
			`
| TRACK    | Voc | Drums | Piano |
| -------- | --- | ----- | ----- |

| =SCORE                    | Voc | Drums | Piano |
| ------------------------- | --- | ----- | ----- |
|  # 4/4                    |     |       |       |
|     1                     | a"  | d     | b"    |
|     2&                    | c   |       | d     |
|  'testdata/includes/score |     |       |       |
|  #                        |     |       |       |
|     1                     | d   |       |       |

`,
		},
		{ // 3
			`
|TRACK | Voc | Drums | Piano |

|=SCORE   | Voc | Drums | Piano |
|# 4/4
|1  | a"  | d    | b"    |
|2& | c   |      | d     |
|'testdata/includes/score
|'testdata/includes/score
|#
|1  | d   |      |       |

`,
			`
| TRACK    | Voc | Drums | Piano |
| -------- | --- | ----- | ----- |

| =SCORE                    | Voc | Drums | Piano |
| ------------------------- | --- | ----- | ----- |
|  # 4/4                    |     |       |       |
|     1                     | a"  | d     | b"    |
|     2&                    | c   |       | d     |
|  'testdata/includes/score |     |       |       |
|  'testdata/includes/score |     |       |       |
|  #                        |     |       |       |
|     1                     | d   |       |       |
`,
		},
		{ // 4
			`
'drumnote.drums

|TRACK | Voc | Drums | Piano |

|=SCORE | Voc | Drums | Piano |
|#
|1      | a"  | .drums.kd    | b"    |
|2&     | c   | .drums.sn    | d     |
|#
|1      |     | .drums.ho    |       |
`,
			`
'drumnote.drums

| TRACK    | Voc | Drums | Piano |
| -------- | --- | ----- | ----- |

| =SCORE   | Voc | Drums     | Piano |
| -------- | --- | --------- | ----- |
|  #       |     |           |       |
|     1    | a"  | .drums.kd | b"    |
|     2&   | c   | .drums.sn | d     |
|  #       |     |           |       |
|     1    |     | .drums.ho |       |

`,
		},
		{ // 5
			`
'testdata/includes/score3=template
'drumnote.drums

|TRACK | Voc | Drums | Piano |

|=SCORE | Voc | Drums | Piano |
|#
|1      | a"  | .drums.kd    | b"    |
|2&     | c   | .drums.sn    | d     |
|#
|1      |     | .drums.ho    |       |
`,
			`
'testdata/includes/score3=template
'drumnote.drums

| TRACK    | Voc | Drums | Piano |
| -------- | --- | ----- | ----- |

| =SCORE   | Voc | Drums     | Piano |
| -------- | --- | --------- | ----- |
|  #       |     |           |       |
|     1    | a"  | .drums.kd | b"    |
|     2&   | c   | .drums.sn | d     |
|  #       |     |           |       |
|     1    |     | .drums.ho |       |

`,
		},
		{ // 6
			`

|TRACK | Piano | Voc

|=SCORE | Piano Voc |
|#
|1  |  f"   | 
		`,
			`
| TRACK    | Piano | Voc |
| -------- | ----- | --- |

| =SCORE   | Piano Voc |
| -------- | --------- |
|  #       |           |
|     1    | f"        | 
		`,
		},
		{ // 7
			`

|TRACK | Piano | Voc

|=SCORE | Piano Voc |
|#
|1  |  'drumnote.drums.kd   | 
		`,
			`
| TRACK    | Piano | Voc |
| -------- | ----- | --- |

| =SCORE   | Piano Voc          |
| -------- | ------------------ |
|  #       |                    |
|     1    | 'drumnote.drums.kd | 
		`,
		},
	}

	//items.DEBUG = true

	for i, test := range tests {

		/*
			if i > 3 {
				continue
			}
		*/
		/*
		   sc, err := muskel.Parse(strings.NewReader(strings.TrimSpace(test.input)), "include main")

		   if err != nil {
		     t.Errorf("[%v] could not parse score: %s\n%s\n", i, err.Error(), test.input)
		     continue
		   }

		   var bf strings.Builder

		   err = muskel.WriteFormattedTo(sc, &bf)

		   if err != nil {
		     t.Errorf("[%v] could not format score: %s\n%s\n", i, err.Error(), test.input)
		     continue
		   }

		   result := bf.String()
		*/
		var bf strings.Builder

		err := muskel.Format("include-main", nil, strings.NewReader(strings.TrimSpace(test.input)), &bf)

		if err != nil {
			t.Errorf("[%v] could not format score YYY: %s\n%s\n", i, err.Error(), test.input)
			continue
		}

		result := bf.String()

		//    fmt.Println(result)

		//res := strings.Split(result, "=\n")
		//exp := strings.Split(test.expected, "=\n")

		got := strings.TrimSpace(result)
		wanted := strings.TrimSpace(test.expected)

		if got != wanted {
			t.Errorf("[%v] score\n%s\nformatted gives \n%s\n\nbut this was expected:\n%s\n%q\nvs\n%q\n", i, test.input, got, wanted, got, wanted)
			//      t.Errorf("[%v] score\n%s\n\nunrolled gives \n%q\n\nbut this was expected:\n%q\n", i, test.input, got, wanted)
		}
	}
}

func TestInclude3(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{ // 0
			`
|TRACK | Voc |
|VelocityScale | min:0 max:0 random:0 step:0 center:0 |

|=SCORE | Voc |
|#
|1      | a"  |
|2&     | c   |
|#
|1      |     |
`,
			`| TRACK           | Voc                                  |
| --------------- | ------------------------------------ |
|  File           |                                      |
|  Channel        |                                      |
|  Program        |                                      |
|  Bank           |                                      |
|  Transpose      |                                      |
|  Volume         |                                      |
|  Delay          |                                      |
|  PitchbendRange | 2                                    |
|  VelocityScale  | min:0 max:0 random:0 step:0 center:0 |
|  Ambitus        |                                      |
|  Import         |                                      |

| =SCORE                   | Voc |
| ------------------------ | --- |
|  # 4/4 @120.00 \major^c' |     |
|     1                    | a"  |
|     2&                   | c   |
|  #                       |     |
`,
		},
	}

	//items.DEBUG = true

	for i, test := range tests {

		var bf strings.Builder

		err := muskel.Unroll("include-main", nil, strings.NewReader(strings.TrimSpace(test.input)), &bf)

		if err != nil {
			t.Errorf("[%v] could not unroll score: %s\n%s\n", i, err.Error(), test.input)
			continue
		}

		result := bf.String()

		//    fmt.Println(result)

		//res := strings.Split(result, "=\n")
		//exp := strings.Split(test.expected, "=\n")

		got := strings.TrimSpace(result)
		wanted := strings.TrimSpace(test.expected)

		if got != wanted {
			t.Errorf("[%v] score\n%s\n\nunrolled gives \n%s\n\nbut this was expected:\n%s\n%q\nvs\n%q\n", i, test.input, got, wanted, got, wanted)
			//      t.Errorf("[%v] score\n%s\n\nunrolled gives \n%q\n\nbut this was expected:\n%q\n", i, test.input, got, wanted)
		}
	}

	//items.DEBUG = false
}
