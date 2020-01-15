package muskel_test

import (
	"strings"
	"testing"

	"gitlab.com/gomidi/muskel"
)

func TestInclude(t *testing.T) {
	//t.Skip()
	tests := []struct {
		input    string
		expected string
	}{
		{ // 0
			`
$include("drumnote")

TRACK | Voc | Drums | Piano |

=SCORE | Voc | Drums | Piano |
#
1  | a"  | drums.kd    | b"    |
2& | c   | drums.sn    | d     |
#
1  |     | drums.ho    |       |
`,
			`
=SCORE         | Drums   | Piano | Voc |
 # 4/4 @120.00
    1          | MN36::: | b"    | a"  |
    2&         | MN40::: | d     | c   |
 #
    1          | MN46::: |       |     |
`,
		},
		{ // 1
			`
$include("testdata/includes/header")
$include("drumnote")

=hu |a|
#
1   |drums.sn| 
2   |drums.kd|

TRACK | Voc | Drums | Piano |

=SCORE   | Voc | Drums   | Piano |
#
1  | a"  | drums.kd    | b"    |
2& | c   | drums.sn    | d     |
#
1  |     | =hu.a    |       |
`,
			`
=SCORE         | Drums   | Piano | Voc |
 # 4/4 @120.00
    1          | MN36::: | b"    | a"  |
    2&         | MN40::: | d     | c   |
 #
    1          | MN40::: |       |     |
    2          | MN36::: |       |     |
`,
		},
		{ // 2
			`
$include("testdata/includes/header")
$include("drumnote")

=rhythm |a|
#
1   |$_euclid(3,12,&)| 
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
=SCORE         | Drums | Piano | Voc |
 # 4/4 @120.00
    1          | f"    | b"    | a"  |
    1&         | e"    |       |     |
    2          | e"    |       |     |
    2&         | e"    | d     | c   |
    3          | f"    |       |     |
    3&         | e"    |       |     |
    4          | e"    |       |     |
    4&         | e"    |       |     |
 #
    1          | f"    |       |     |
    1&         | e"    |       |     |
    2          | e"    |       |     |
    2&         | e"    |       |     |
`,
		},
		{ // 3
			`
$include("testdata/includes/header")
$include("drumnote")

=rhythm |a|
#
1&  | $_euclid(3,12,&) | 
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
=SCORE         | Drums | Piano | Voc |
 # 4/4 @120.00
    1          |       | b"    | a"  |
    1&         | f"    |       |     |
    2          | e"    |       |     |
    2&         | e"    | d     | c   |
    3          | e"    |       |     |
    3&         | f"    |       |     |
    4          | e"    |       |     |
    4&         | e"    |       |     |
 #
    1          | e"    |       |     |
    1&         | f"    |       |     |
    2          | e"    |       |     |
    2&         | e"    |       |     |
    3          | e"    |       |     |
`,
		},
		{ // 4
			`
TRACK | Voc | Drums | Piano |

=SCORE   | Voc | Drums  | Piano |
# 3/4
1        | a"  | d    | b"    |
2&       | c   |      | d     |
$include("testdata/includes/score")
#
1        | d   |      |       | B
`,
			`
=SCORE         | Drums | Piano | Voc |
 # 3/4 @120.00
    1          | d     | b"    | a"  |
    2&         |       | d     | c   |
 # 4/4
    1          |       | f"    | e"  |
    2&         |       | d     | c   |
 # 3/4
    1          |       |       | d   |
`,
		},
		{ // 5
			`
TRACK | Voc | Drums | Piano |

=SCORE  | Voc | Drums | Piano |
$include("testdata/includes/score")
#
1       | d   |       |       | B
`,
			`
=SCORE         | Drums | Piano | Voc |
 # 4/4 @120.00
    1          |       | f"    | e"  |
    2&         |       | d     | c   |
 #
    1          |       |       | d   |
`,
		},
		{ // 6
			`
TRACK | Voc | Drums | Piano |

=SCORE | Voc | Drums   | Piano |
$include("testdata/includes/score")
$include("testdata/includes/score")
`,
			`
=SCORE         | Drums | Piano | Voc |
 # 4/4 @120.00
    1          |       | f"    | e"  |
    2&         |       | d     | c   |
 #
    1          |       | f"    | e"  |
    2&         |       | d     | c   |
`,
		},
		{ // 7
			`
TRACK | Voc | Drums | Piano |

=SCORE   | Voc | Drums   | Piano |
# 4/4
1        | a"  | d    | b"    |
2&       | c   |      | d     |
$include("testdata/includes/score")
$include("testdata/includes/score")
#
1        | d   |      |       | B

`,
			`
=SCORE         | Drums | Piano | Voc |
 # 4/4 @120.00
    1          | d     | b"    | a"  |
    2&         |       | d     | c   |
 #
    1          |       | f"    | e"  |
    2&         |       | d     | c   |
 #
    1          |       | f"    | e"  |
    2&         |       | d     | c   |
 #
    1          |       |       | d   |

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
=SCORE         | Voc |
 # 4/4 @120.00
    1&         | c   |
    2          | d   |

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
=SCORE         | Voc |
 # 4/4 @120.00
    1&         | c   |
    2          | d   |
 #
    1&         | c   |
    2          | d   |
 #
    1&         | c   |
    2          | d   |
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
=SCORE          | Voc |
 #A 4/4 @120.00
    1&          | c   |
    2           | d   |
 #
    1&          | c   |
    2           | d   |
 #
    1&          | c   |
    2           | d   |
 #B 3/4
    1           | C   |
 # 4/4
    1&          | c   |
    2           | d   |
 #
    1&          | c   |
    2           | d   |
 #
    1&          | c   |
    2           | d   |
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
=SCORE          | flute |
 #A 4/4 @120.00
    1           | a     |
 #
 #
 #
 #
    1           | a     |
 #
 #
 #
 # 5/6
    1           | B     |
 # 4/4
    1           | G     |
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
=SCORE          | left |
 #A 4/4 @120.00
    1           | a    |
 #B
    1           | E    |
 #C
    1           | B    |
 #
    1           | E    |
 #
    1           | g    |
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
=SCORE         | left |
 # 4/4 @120.00
    1          | a    |
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
=SCORE         | left  |
 # 4/4 @120.00
    1          | d     |
 #
    1          | *     |
    2          | (c e) |
    4          | (g a) |
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
=SCORE         | left |
 # 4/4 @120.00
    1          | d    |
    1          | e    |
`,
		},
		{ // 16
			`
TRACK | left |
 Ch   | 1    |

=SCORE | left |
# \ionian^d#
1      | ^1   |
2      | ^2   |
`,
			`
=SCORE         | left |
 # 4/4 @120.00
    1          | d#   |
    2          | f    |
`,
		},
		{ // 17
			`
TRACK | left |
 Ch   | 1    |

=SCORE | left |
#
1      | ^1   |
2      | ^2   |
#  \ionian^d#
1      | ^1   |
2      | ^2   |
#  \ionian^f#
1      | ^1   |
2      | ^2   |
`,
			`
=SCORE         | left |
 # 4/4 @120.00
    1          | c'   |
    2          | d'   |
 #
    1          | d#   |
    2          | f    |
 #
    1          | f#   |
    2          | g#   |
`,
		},
		{ // 18
			`
TRACK | left |
 Ch   | 1    |

=SCORE | left |
#
1      | ^1   |
2      | ^2   |
#  \ionian^d#
1      | ^1   |
2      | \ionian^g |
2      | ^2   |
#  \ionian^f#
1      | ^1   |
2      | ^2   |
`,
			`
=SCORE         | left |
 # 4/4 @120.00
    1          | c'   |
    2          | d'   |
 #
    1          | d#   |
    2          | a    |
 #
    1          | g    |
    2          | a    |
`,
		},
		{ // 19
			`
TRACK | left |
 Ch   | 1    |

=patt | a |
#
1      | ^1   |
2      | \ionian^g |
2      | ^2   |

=SCORE | left |
#
1      | ^1   |
2      | ^2   |
#  \ionian^d#
1      | =patt  |
#  \ionian^f#
1      | ^1   |
2      | ^2   |
`,
			`
=SCORE         | left |
 # 4/4 @120.00
    1          | c'   |
    2          | d'   |
 #
    1          | d#   |
    2          | a    |
 #
    1          | g    |
    2          | a    |
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
1      | ^1   |
2      | \myscale^a" |
2      | ^2   |

=SCORE | left |
#
1      | ^1   |
2      | ^2   |
#  \myscale^d#
1      | =patt  |
#  \myscale^d"
1      | ^1   |
2      | ^2   |
`,
			`
=SCORE         | left |
 # 4/4 @120.00
    1          | c'   |
    2          | d'   |
 #
    1          | d#   |
    2          | c"'  |
 #
    1          | a"   |
    2          | c"'  |
`,
		},
		{ // 21
			`
TRACK | left | right |
 Ch   | 1    | 2     |

=patt | |
#
1    |^1 |
2    |^2 |
#
1    |^1 |
2    |^2 |
#
1    |^1 |
2    |^2 |


=SCORE | left | right |
# \minor^b'
1      | ^2   | =patt |
2      | ^1 | |
#
1      |=patt |  | 
*2
#
1      | ^2   | ^1  |
2      | ^1   | ^1  |
`,
			`
=SCORE         | left | right |
 # 4/4 @120.00
    1          | c#"  | b'    |
    2          | b'   | c#"   |
 #
    1          | b'   | b'    |
    2          | c#"  | c#"   |
 #
    1          | b'   | b'    |
    2          | c#"  | c#"   |
 #
    1          | b'   |       |
    2          | c#"  |       |
 #
    1          | c#"  | b'    |
    2          | b'   | b'    |
`,
		},
		{ // 22
			`
TRACK | left | right |
 Ch   | 1    | 2     |

=patt | |
#
1    |^1 |
2    |^2 |
#
1    |^3 |
2    |^3 |


=SCORE | right   |
#  \major^c"
1      | =patt%3 |
*3
#
1      | =patt%2 |
*4
`,
			`
=SCORE         | right |
 # 4/4 @120.00
    1          | c"    |
    2          | d"    |
 #
    1          | e"    |
    2          | e"    |
 #
    1          | c"    |
    2          | d"    |
 #
    1          | e"    |
    2          | e"    |
 #
    1          | c"    |
    2          | d"    |
 #
    1          | e"    |
    2          | e"    |
 #
    1          | c"    |
    2          | d"    |
 #
    1          | e"    |
    2          | e"    |
 #
`,
		},
		{ // 23
			`
TRACK | x | y |
 Ch   | 1 |   |

=patt | 2    | 1   | 0  |
#
 1    |      | =.2 | ^1 |
 4&   | ^4   |     |    |
#
 1    | ^3   |     | ^2 |
 2    | ^2   |     |    |

=SCORE | x          | y          |
#  \major^c"
 1     |            | =patt.1%3  |
 4&    | =!patt.1%3 |            |
*5
`,
			`
=SCORE         | x  | y  |
 # 4/4 @120.00
    4&         | f" | f" |
 #
    1          | e" | e" |
    2          | d" | d" |
 #
    4&         | f" | f" |
 #
    1          | e" | e" |
    2          | d" | d" |
 #
    4&         | f" | f" |
 #
    1          | e" | e" |
    2          | d" | d" |
`,
		},
		{ // 24
			`
TRACK | x | y |
 Ch   | 1 |   |

=patt | 2    | 
#
 1    |      |
 4&   | ^4   |     
#
 1    | ^3   |   
 2    | ^2   |     

=SCORE | x            | y           |
#  \major^c"
 1     |              | =patt.2[1:] |
 4&    | =!patt.2[1:] |             |
*1
`,
			`
=SCORE         | x  | y  |
 # 4/4 @120.00
    1          |    | e" |
    2          |    | d" |
    4&         | e" |    |
 #
    1&         | d" |    |
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
=SCORE              | x  | y  |
 #start 4/4 @120.00
    1               |    | c" |
    4&              | d" |    |
 #
    1               |    | c" |
    4&              | d" |    |
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
=SCORE         | x  |
 # 4/4 @120.00
    1          | c' |
    2          | d' |
    3          | e' |
    4          | f' |
 #
    1          | c' |
    2          | g' |
    2&         | g" |
    3          | e' |
    4          | f' |
 #
    1          | c' |
    2          | d' |
    3          | e' |
    4          | f' |
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
=SCORE         | x  |
 # 4/4 @120.00
    1          | c' |
    2          | d' |
    3          | e' |
    4          | f' |
 #
    1          | c' |
    2          | g' |
    2&         | g" |
    4          | f' |
 #
    1          | c' |
    2          | d' |
    3          | e' |
    4          | f' |
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
=SCORE         | x  |
 # 4/4 @120.00
    1          | c' |
    2          | d' |
    3          | e' |
    4          | f' |
 #
    1          | g' |
    1&         | g" |
    2          | d' |
    3          | e' |
    4          | f' |
 #
    1          | c' |
    2          | d' |
    3          | e' |
    4          | f' |
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
=SCORE         | x  |
 # 4/4 @120.00
    1          | g' |
    1&         | g" |
    2          | d' |
    3          | e' |
    4          | f' |
 #
    1          | c' |
    2          | d' |
    3          | e' |
    4          | f' |
 #
    1          | c' |
    2          | d' |
    3          | e' |
    4          | f' |
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
=SCORE         | x  |
 # 4/4 @120.00
    1          | c' |
    2          | g' |
    3          | e' |
    4          | f' |
 #
    1          | c' |
    2          | d' |
    3          | e' |
    4          | f' |
 #
    1          | c' |
    2          | d' |
    3          | e' |
    4          | f' |
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
=SCORE         | x  |
 # 4/4 @120.00
    1          | c' |
    2          | *  |
    3          | e' |
    4          | f' |
 #
    1          | c' |
    2          | d' |
    3          | e' |
    4          | f' |
 #
    1          | c' |
    2          | d' |
    3          | e' |
    4          | f' |
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
=SCORE         | x  |
 # 4/4 @120.00
    1          | g' |
    2          | d' |
    3          | e' |
    4          | f' |
 #
    1          | c' |
    2          | d' |
    3          | e' |
    4          | f' |
 #
    1          | c' |
    2          | d' |
    3          | e' |
    4          | f' |
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
=SCORE         | x  |
 # 4/4 @120.00
    1          | *  |
    2          | d' |
    3          | e' |
    4          | f' |
 #
    1          | c' |
    2          | d' |
    3          | e' |
    4          | f' |
 #
    1          | c' |
    2          | d' |
    3          | e' |
    4          | f' |
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

short |
ba    | (b' A)


=SCORE | x  | 
#
 1     | =long.p%2| 
       | /ba |
#
#
 1     | | 
`,
			`
=SCORE         | x      |
 # 4/4 @120.00
    1          | (b' A) |
    2          | d'     |
    3          | e'     |
    4          | f'     |
 #
    1          | c'     |
    2          | d'     |
    3          | e'     |
    4          | f'     |
 #
    1          | c'     |
    2          | d'     |
    3          | e'     |
    4          | f'     |
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

short |
ba    | (b' A)


=SCORE | x  | 
#
 1     | =long.p%2| 
 2     | /ba |
#
#
 1     | | 
`,
			`
=SCORE         | x      |
 # 4/4 @120.00
    1          | c'     |
    2          | (b' A) |
    3          | e'     |
    4          | f'     |
 #
    1          | c'     |
    2          | d'     |
    3          | e'     |
    4          | f'     |
 #
    1          | c'     |
    2          | d'     |
    3          | e'     |
    4          | f'     |
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
1     | .1. |

short |
ba    | (b' A)


=SCORE | x  | 
#
 1     | =long| 
#
#
 1     | .2. |
 2     | /ba |
#
  
`,
			`
=SCORE         | x      |
 # 4/4 @120.00
    1          | c'     |
    2          | d'     |
    3          | e'     |
    4          | f'     |
 #
    1          | c'     |
    2          | d'     |
    3          | e'     |
    4          | f'     |
 #
    1          | c'     |
    2          | (b' A) |
    3          | e'     |
    4          | f'     |
 #
    1          | c'     |
    2          | d'     |
    3          | e'     |
    4          | f'     |
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
=SCORE         | x |
 # 4/4 @120.00
    1          | a |
    2          | b |
    3          | a |
    4          | c |
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

chords |
test   | (a b c)

=SCORE | x  | 
#
 1     | =patt(test...)| 
  
`,
			`
=SCORE         | x |
 # 4/4 @120.00
    1          | a |
    2          | b |
    3          | a |
    4          | c |
 `,
		},
		{ // 39
			`
TRACK | x | 
 Ch   | 1 |   


=SCORE | x  | 
# \minor^e
 1     | ^1 |
 2     | ^2 |
 3     | ^3 | 
  
`,
			`
=SCORE         | x  |
 # 4/4 @120.00
    1          | e  |
    2          | f# |
    3          | g  |
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
=SCORE         | x  |
 # 4/4 @120.00
    1          | c' |
    2          | d' |
    3          | c' |
    4          | e' |
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
=SCORE         | x  |
 # 4/4 @120.00
    1          | a  |
    2          | b  |
    3          | c' |
    4          | d' |
 #
    1          | e' |
    2          | f' |
    3          | g' |
    4          | a  |  
   `,
		},
		{ // 42
			`
$include("drumnote")

=rhythm |a|
#
1   |$_euclid(3,12,&)| 
*3

TRACK | Drums |

=SCORE |  Drums                      |
#
1      |  =rhythm(f",drums.sn) |
2&     |                             |
#
1      |                             |
`,
			`
=SCORE         | Drums   |
 # 4/4 @120.00
    1          | f"      |
    1&         | MN40::: |
    2          | MN40::: |
    2&         | MN40::: |
    3          | f"      |
    3&         | MN40::: |
    4          | MN40::: |
    4&         | MN40::: |
 #
    1          | f"      |
    1&         | MN40::: |
    2          | MN40::: |
    2&         | MN40::: |
`,
		},
		{ // 43
			`
$include("drumnote")

=rhythm |a|
#
1   |$_euclid(3,12,&)| 
*3

TRACK | Drums |

=SCORE |  Drums                      |
#
1      |  =rhythm((f" g),drums.sn) |
2&     |                             |
#
1      |                             |
`,
			`
=SCORE         | Drums   |
 # 4/4 @120.00
    1          | (f" g)  |
    1&         | MN40::: |
    2          | MN40::: |
    2&         | MN40::: |
    3          | (f" g)  |
    3&         | MN40::: |
    4          | MN40::: |
    4&         | MN40::: |
 #
    1          | (f" g)  |
    1&         | MN40::: |
    2          | MN40::: |
    2&         | MN40::: |
`,
		},
		{ // 44
			`
$include("drumnote")

=rhythm |a|
#
1   |$_euclid(3,12,&)| 
*3

TRACK | Drums |

=SCORE |  Drums                      |
#
1      |  =rhythm(drums.kd,(drums.sn drums.kd)) |
2&     |                             |
#
1      |                             |
`,
			`
=SCORE         | Drums             |
 # 4/4 @120.00
    1          | MN36:::           |
    1&         | (MN40::: MN36:::) |
    2          | (MN40::: MN36:::) |
    2&         | (MN40::: MN36:::) |
    3          | MN36:::           |
    3&         | (MN40::: MN36:::) |
    4          | (MN40::: MN36:::) |
    4&         | (MN40::: MN36:::) |
 #
    1          | MN36:::           |
    1&         | (MN40::: MN36:::) |
    2          | (MN40::: MN36:::) |
    2&         | (MN40::: MN36:::) |
`,
		},
		{ // 45
			`
$include("drumnote")

=rhythm |a|
#
1   |$_euclid(3,12,&)| 
*3

TRACK | Drums |

=SCORE |  Drums                      |
#
1      |  =rhythm(drums.kd,{a,b}) |
2&     |                             |
#
1      |                             |
`,
			`
=SCORE         | Drums   |
 # 4/4 @120.00
    1          | MN36::: |
    1&         | {a,b}   |
    2          | {a,b}   |
    2&         | {a,b}   |
    3          | MN36::: |
    3&         | {a,b}   |
    4          | {a,b}   |
    4&         | {a,b}   |
 #
    1          | MN36::: |
    1&         | {a,b}   |
    2          | {a,b}   |
    2&         | {a,b}   |
`,
		},
		{ // 46
			`
$include("drumnote")

=rhythm |a|
#
1   |$_euclid(3,12,&)| 
*3

TRACK | Drums |

=SCORE |  Drums                      |
#
1      |  =rhythm(drums.kd,{drums.kd,b}) |
2&     |                             |
#
1      |                             |
`,
			`
=SCORE         | Drums       |
 # 4/4 @120.00
    1          | MN36:::     |
    1&         | {MN36:::,b} |
    2          | {MN36:::,b} |
    2&         | {MN36:::,b} |
    3          | MN36:::     |
    3&         | {MN36:::,b} |
    4          | {MN36:::,b} |
    4&         | {MN36:::,b} |
 #
    1          | MN36:::     |
    1&         | {MN36:::,b} |
    2          | {MN36:::,b} |
    2&         | {MN36:::,b} |
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
=SCORE         | Vocals   |
 # 4/4 @120.00
    1          | (a "a")  |
    1          | (b> "b") |
    2          | (c "c")  |
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
=SCORE         | Drums | Vocals   |
 # 4/4 @120.00
    1          |       | (a "a")  |
    1          |       | (b> "b") |
    2          |       | (c "c")  |
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
=SCORE         | Drums | Vocals   |
 # 4/4 @120.00
    1          |       | (a "a")  |
    1          |       | (b> "b") |
    2          |       | (c "c")  |
`,
		},
		{ // 50
			`
=melody | Verse1 |
 #
  1     | a^2      |
  2     | a^-2     |

TRACK   | Vocals |
 Channel| 1      |

=SCORE  | Vocals |
 #
  1     |   =melody    |
`,
			`
=SCORE         | Vocals |
 # 4/4 @120.00
    1          | a^2    |
    2          | a^-2   |
`,
		},
		{ // 51
			`
      
$include("testdata/includes/lyrics")

=melody | Verse1 |
 #
  1     | a      |
  2     | c      |

TRACK   | Vocals |
 Channel| 1      |

=SCORE | Vocals         |
 #
  1    | =melody@lyrics.Verse |
`,
			`
=SCORE         | Vocals    |
 # 4/4 @120.00
    1          | (a "Hi-") |
    2          | (c "ho")  |
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
$include("testdata/includes/score2")
`,
			`
=SCORE          | Drums | Piano | Voc       |
 # 4/4 @100.00
    1           |       |       | (a "Hi-") |
    2           |       |       | (c "ho")  |
 #C 3/4 @134.00
    1           |       | f"    | e"        |
    2&          |       | d     | c         |
 # 6/4 @100.00
    4&          |       | f     | g'        |
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

	// sketch.DEBUG = true

	for i, test := range tests {

		// fmt.Printf("############ %v ###############\n", i)

		/*
		   if i != 52 {
		     continue
		   }
		*/

		if i == 36 {
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

		got := "=SCORE" + strings.Split(strings.TrimSpace(result), "=SCORE")[1]
		wanted := strings.TrimSpace(test.expected)

		if got != wanted {
			t.Errorf("[%v] score\n%s\n\nunrolled gives \n%s\n\nbut this was expected:\n%s\n%q\nvs\n%q\n", i, test.input, got, wanted, got, wanted)
			//      t.Errorf("[%v] score\n%s\n\nunrolled gives \n%q\n\nbut this was expected:\n%q\n", i, test.input, got, wanted)
		}
	}
}

func TestInclude2(t *testing.T) {
	//t.Skip()
	tests := []struct {
		input    string
		expected string
	}{
		{ // 0
			`
$include("drumnote")

TRACK | Voc | Drums | Piano |

=SCORE | Voc | Drums | Piano |
#
1      | a"  | drums.kd    | b"    |
2&     | c   | drums.sn    | d     |
#
1      |     | drums.ho    |       |
`,
			`
$include("drumnote")

TRACK    | Voc | Drums | Piano |

=SCORE   | Voc | Drums    | Piano |
 #
    1    | a"  | drums.kd | b"    |
    2&   | c   | drums.sn | d     |
 #
    1    |     | drums.ho |       |

`,
		},
		{ // 1
			`
$include("testdata/includes/header")
$include("drumnote")

=hu | a        |
#
1   | drums.sn |
2   | drums.kd |

TRACK | Voc | Drums | Piano |

=SCORE   | Voc | Drums | Piano |
#
 1       | a"  | drums.kd    | b"    |
 2&      | c   | drums.sn    | d     |
#
 1       |     | =hu.a       |       |

`,
			`
$include("testdata/includes/header")
$include("drumnote")

=hu      | a        |
 #
    1    | drums.sn |
    2    | drums.kd |

TRACK    | Voc | Drums | Piano |

=SCORE   | Voc | Drums    | Piano |
 #
    1    | a"  | drums.kd | b"    |
    2&   | c   | drums.sn | d     |
 #
    1    |     | =hu.a    |       |

`,
		},
		{ // 2
			`
TRACK | Voc | Drums | Piano |

=SCORE | Voc | Drums | Piano |
# 4/4
1  | a"  | d    | b"    |
2& | c   |      | d     |
$include("testdata/includes/score")
#
1  | d   |      |       |

`,
			`
TRACK    | Voc | Drums | Piano |

=SCORE   | Voc | Drums | Piano |
 # 4/4
    1    | a"  | d     | b"    |
    2&   | c   |       | d     |
 $include("testdata/includes/score")
 #
    1    | d   |       |       |

`,
		},
		{ // 3
			`
TRACK | Voc | Drums | Piano |

=SCORE   | Voc | Drums | Piano |
# 4/4
1  | a"  | d    | b"    |
2& | c   |      | d     |
$include("testdata/includes/score")
$include("testdata/includes/score")
#
1  | d   |      |       |

`,
			`
TRACK    | Voc | Drums | Piano |

=SCORE   | Voc | Drums | Piano |
 # 4/4
    1    | a"  | d     | b"    |
    2&   | c   |       | d     |
 $include("testdata/includes/score")
 $include("testdata/includes/score")
 #
    1    | d   |       |       |
`,
		},
	}

	//  score.DEBUG = true

	for i, test := range tests {

		/*
		   if i > 1 {
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
			t.Errorf("[%v] could not format score: %s\n%s\n", i, err.Error(), test.input)
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
