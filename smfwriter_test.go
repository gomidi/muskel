package muskel_test

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"gitlab.com/gomidi/midi/midimessage/channel"
	"gitlab.com/gomidi/midi/midimessage/meta"
	"gitlab.com/gomidi/midi/smf/smfwriter"
	"gitlab.com/gomidi/muskel/file"
	"gitlab.com/gomidi/muskel/score"
	"gitlab.com/gomidi/muskel/smf"
)

func TestSMFWriter(t *testing.T) {
	//t.Skip()
	//	score.DEBUG = false

	var tests = []struct {
		input                string
		track                int
		includeMeta          bool
		includeControlChange bool
		expected             string
	}{
		{
			// 0
			`
TRACK |p  |
Ch     |1  |

=SCORE |p  |
#
1      |c" |
2      |d  |

			`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 72 dyn =
[960] channel.NoteOff channel 0 key 72
[0] channel.NoteOn channel 0 key 50 dyn =
[2880] channel.NoteOff channel 0 key 50
`,
		},
		{
			// 1
			`
TRACK |p |
Ch      |1 |

=SCORE |p        |
#
1      |{c,e,g#} |
2      |*        |

			`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 48 dyn =
[320] channel.NoteOff channel 0 key 48
[0] channel.NoteOn channel 0 key 52 dyn =
[320] channel.NoteOff channel 0 key 52
[0] channel.NoteOn channel 0 key 56 dyn =
[320] channel.NoteOff channel 0 key 56
[0] meta.Undefined type:  0
`,
		},
		{
			// 2
			`
TRACK |p  |
Ch      |1  |

=SCORE |p |
#
1      |c |

			`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 48 dyn =
[3840] channel.NoteOff channel 0 key 48
`,
		},
		{
			// 3
			`
TRACK |p  |
Ch      |1  |

=SCORE |p        |
#
1      |{c,e,g#} |

			`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 48 dyn =
[1280] channel.NoteOff channel 0 key 48
[0] channel.NoteOn channel 0 key 52 dyn =
[1280] channel.NoteOff channel 0 key 52
[0] channel.NoteOn channel 0 key 56 dyn =
[1280] channel.NoteOff channel 0 key 56
[0] meta.Undefined type:  0
`,
		},
		{
			// 4
			`
TRACK |p|
Ch      |1|

=SCORE |p |
#
1      |c |
#
1      |  |
2      |* |

			`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 48 dyn =
[4800] channel.NoteOff channel 0 key 48
`,
		},
		{
			// 5
			`
TRACK |p|
Ch      |1|

=SCORE |p |
#
1  |{c,e,g#}|
#
1  ||
2  |*|

			`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 48 dyn =
[1600] channel.NoteOff channel 0 key 48
[0] channel.NoteOn channel 0 key 52 dyn =
[1600] channel.NoteOff channel 0 key 52
[0] channel.NoteOn channel 0 key 56 dyn =
[1600] channel.NoteOff channel 0 key 56
[0] meta.Undefined type:  0
`,
		},
		{
			// 6
			`
TRACK |p|
Ch      |1|
Delay|1/32|

=SCORE |p |
#
1  |{c,e,g#} |
#
1  ||
2  |*|

			`,
			2,
			false,
			false,
			`
[120] channel.NoteOn channel 0 key 48 dyn =
[1600] channel.NoteOff channel 0 key 48
[0] channel.NoteOn channel 0 key 52 dyn =
[1600] channel.NoteOff channel 0 key 52
[0] channel.NoteOn channel 0 key 56 dyn =
[1600] channel.NoteOff channel 0 key 56
[0] meta.Undefined type:  0
`,
		},
		{
			// 7
			`
TRACK |p|
Ch      |1|
Delay|-1/32|

=SCORE |p |
#
1  | |
#
1  |{c,e,g#} |
#
1  ||
2  |*|

			`,
			2,
			false,
			false,
			`
[3720] channel.NoteOn channel 0 key 48 dyn =
[1600] channel.NoteOff channel 0 key 48
[0] channel.NoteOn channel 0 key 52 dyn =
[1600] channel.NoteOff channel 0 key 52
[0] channel.NoteOn channel 0 key 56 dyn =
[1600] channel.NoteOff channel 0 key 56
[0] meta.Undefined type:  0
`,
		},
		{
			// 8
			`
TRACK |p|
Ch      |1|

=SCORE |p |
#
1  |c |
2  |d |
3  |e |
4  |* |

			`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 48 dyn =
[960] channel.NoteOff channel 0 key 48
[0] channel.NoteOn channel 0 key 50 dyn =
[960] channel.NoteOff channel 0 key 50
[0] channel.NoteOn channel 0 key 52 dyn =
[960] channel.NoteOff channel 0 key 52
`,
		},
		{
			// 9
			`
TRACK |p|
Ch      |1|

=SCORE |p |
#
1  |c> |
2  |d |
3  |e |
4  |* |

			`,
			2,
			false,
			false,
			`
[30] channel.NoteOn channel 0 key 48 dyn =
[930] channel.NoteOff channel 0 key 48
[0] channel.NoteOn channel 0 key 50 dyn =
[960] channel.NoteOff channel 0 key 50
[0] channel.NoteOn channel 0 key 52 dyn =
[960] channel.NoteOff channel 0 key 52
`,
		},
		{
			// 10
			`
TRACK |p|
Ch      |1|

=SCORE |p |
#
1  |MN48> |
2  |MN50 |
3  |MN52 |
4  |* |

			`,
			2,
			false,
			false,
			`
[30] channel.NoteOn channel 0 key 48 dyn =
[930] channel.NoteOff channel 0 key 48
[0] channel.NoteOn channel 0 key 50 dyn =
[960] channel.NoteOff channel 0 key 50
[0] channel.NoteOn channel 0 key 52 dyn =
[960] channel.NoteOff channel 0 key 52
`,
		},
		{
			// 11
			`
TRACK |p|
Ch      |1|

=SCORE |p |
#
1  |c |
2  |d> |
3  |e |
4  |* |
			`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 48 dyn =
[990] channel.NoteOff channel 0 key 48
[0] channel.NoteOn channel 0 key 50 dyn =
[930] channel.NoteOff channel 0 key 50
[0] channel.NoteOn channel 0 key 52 dyn =
[960] channel.NoteOff channel 0 key 52
`,
		},
		{
			// 12
			`
TRACK |p|
Ch      |1|

=SCORE |p |
#
1  |MN48 |
2  |MN50> |
3  |MN52 |
4  |* |

			`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 48 dyn =
[990] channel.NoteOff channel 0 key 48
[0] channel.NoteOn channel 0 key 50 dyn =
[930] channel.NoteOff channel 0 key 50
[0] channel.NoteOn channel 0 key 52 dyn =
[960] channel.NoteOff channel 0 key 52
`,
		},
		{
			// 13
			`
TRACK |p|
Ch      |1|

=SCORE |p |
#
1  |c |
2  |d |
3  |e> |
4  |* |

			`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 48 dyn =
[960] channel.NoteOff channel 0 key 48
[0] channel.NoteOn channel 0 key 50 dyn =
[990] channel.NoteOff channel 0 key 50
[0] channel.NoteOn channel 0 key 52 dyn =
[930] channel.NoteOff channel 0 key 52
`,
		},
		{
			// 14
			`
TRACK |p|
Ch      |1|

=SCORE |p |
#
1  |c |
2  |d< |
3  |e |
4  |* |
			`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 48 dyn =
[930] channel.NoteOff channel 0 key 48
[0] channel.NoteOn channel 0 key 50 dyn =
[990] channel.NoteOff channel 0 key 50
[0] channel.NoteOn channel 0 key 52 dyn =
[960] channel.NoteOff channel 0 key 52
`,
		},
		{
			// 15
			`
TRACK |p|
Ch      |1|

cc |
sc | MN48+

=SCORE |p |
#
1  |cc.sc |
2  |MN50< |
3  |MN52 |
4  |* |

			`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 48 dyn +
[930] channel.NoteOff channel 0 key 48
[0] channel.NoteOn channel 0 key 50 dyn =
[990] channel.NoteOff channel 0 key 50
[0] channel.NoteOn channel 0 key 52 dyn =
[960] channel.NoteOff channel 0 key 52
`,
		},
		{
			// 16
			`
TRACK |p|
Ch      |1|

=SCORE |p |
#
1  |c |
2  |d |
3  |e< |
4  |* |

			`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 48 dyn =
[960] channel.NoteOff channel 0 key 48
[0] channel.NoteOn channel 0 key 50 dyn =
[930] channel.NoteOff channel 0 key 50
[0] channel.NoteOn channel 0 key 52 dyn =
[990] channel.NoteOff channel 0 key 52
`,
		},
		{
			// 17
			`
TRACK |p|
Ch      |1|

=SCORE |p |
#
1  |c< |
2  |d |
3  |e |
4  |* |

			`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 48 dyn =
[960] channel.NoteOff channel 0 key 48
[0] channel.NoteOn channel 0 key 50 dyn =
[960] channel.NoteOff channel 0 key 50
[0] channel.NoteOn channel 0 key 52 dyn =
[960] channel.NoteOff channel 0 key 52
`,
		},
		{
			// 18
			`
TRACK |p        |
Ch      |1        |

=SCORE  |p        |
#
1       |         |
#
1       |{c,e,g#}>|
#
1       |         |
2       |*        |
			`,
			2,
			false,
			false,
			`
[3870] channel.NoteOn channel 0 key 48 dyn =
[1600] channel.NoteOff channel 0 key 48
[0] channel.NoteOn channel 0 key 52 dyn =
[1600] channel.NoteOff channel 0 key 52
[0] channel.NoteOn channel 0 key 56 dyn =
[1570] channel.NoteOff channel 0 key 56
[0] meta.Undefined type:  0
`,
		},
		{
			// 19
			`
TRACK  |p|
Ch |1  |

=SCORE |p|
#
1  | |
#
1  |{c,e,g#}< |
#
1  ||
2  |*|
			`,
			2,
			false,
			false,
			`
[3810] channel.NoteOn channel 0 key 48 dyn =
[1600] channel.NoteOff channel 0 key 48
[0] channel.NoteOn channel 0 key 52 dyn =
[1600] channel.NoteOff channel 0 key 52
[0] channel.NoteOn channel 0 key 56 dyn =
[1630] channel.NoteOff channel 0 key 56
[0] meta.Undefined type:  0
`,
		},
		{
			// 20
			`
=templ |p|
#
1      |{c,e,g#}|
2      |*|

TRACK |p|
Ch |1  |

=SCORE |p|
#
1  |=templ |
			`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 48 dyn =
[320] channel.NoteOff channel 0 key 48
[0] channel.NoteOn channel 0 key 52 dyn =
[320] channel.NoteOff channel 0 key 52
[0] channel.NoteOn channel 0 key 56 dyn =
[320] channel.NoteOff channel 0 key 56
[0] meta.Undefined type:  0
`,
		},
		{
			// 21
			`

tt |
a|c
b|e
c|g#

TRACK |p|
Ch |1  |

=SCORE |p|
#
1  |{tt.a,tt.b,tt.c}|
2  |*|
			`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 48 dyn =
[320] channel.NoteOff channel 0 key 48
[0] channel.NoteOn channel 0 key 52 dyn =
[320] channel.NoteOff channel 0 key 52
[0] channel.NoteOn channel 0 key 56 dyn =
[320] channel.NoteOff channel 0 key 56
[0] meta.Undefined type:  0
`,
		},
		{
			// 22
			`
dr |
ta | MN48
tb | MN52
tc | MN56

TRACK |p|
Ch |1  |

=SCORE |p|
#
1  |{dr.ta,dr.tb,dr.tc} |
2  |*|
			`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 48 dyn =
[320] channel.NoteOff channel 0 key 48
[0] channel.NoteOn channel 0 key 52 dyn =
[320] channel.NoteOff channel 0 key 52
[0] channel.NoteOn channel 0 key 56 dyn =
[320] channel.NoteOff channel 0 key 56
[0] meta.Undefined type:  0
`,
		},
		{
			// 23
			`
dr |
ta | MN48
tb | MN52
tc | MN56


TRACK |p|
Ch |1  |

=SCORE |p|
#
1  |{dr.ta+,dr.tb-,dr.tc--} |
2  |*|
			`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 48 dyn +
[320] channel.NoteOff channel 0 key 48
[0] channel.NoteOn channel 0 key 52 dyn -
[320] channel.NoteOff channel 0 key 52
[0] channel.NoteOn channel 0 key 56 dyn --
[320] channel.NoteOff channel 0 key 56
[0] meta.Undefined type:  0
`,
		},
		{
			// 24
			`
TRACK |p|
Ch |1  |

=SCORE |p|
#
1  |c::: |

			`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 48 dyn =
[60] channel.NoteOff channel 0 key 48
`,
		},
		{
			// 25
			`
TRACK |p|
Ch |1  |

=SCORE |p|
#
1  |{c:::,e:::,g#:::} |
2  |* |

			`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 48 dyn =
[60] channel.NoteOff channel 0 key 48
[260] channel.NoteOn channel 0 key 52 dyn =
[60] channel.NoteOff channel 0 key 52
[260] channel.NoteOn channel 0 key 56 dyn =
[60] channel.NoteOff channel 0 key 56
[260] meta.Undefined type:  0
`,
		},
		{
			// 26
			`			
TRACK |p|
Ch |1  |

=SCORE |p|
#
1  |{c,e,g#} |
2  |a::: |
			`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 48 dyn =
[320] channel.NoteOff channel 0 key 48
[0] channel.NoteOn channel 0 key 52 dyn =
[320] channel.NoteOff channel 0 key 52
[0] channel.NoteOn channel 0 key 56 dyn =
[320] channel.NoteOff channel 0 key 56
[0] meta.Undefined type:  0
[0] channel.NoteOn channel 0 key 57 dyn =
[60] channel.NoteOff channel 0 key 57
`,
		},
		{
			// 27
			`
TRACK |p|
Ch |1  |

=SCORE |p|
#
1  |{c:::,e:::,g#:::} |
2  |a::: |
			`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 48 dyn =
[60] channel.NoteOff channel 0 key 48
[260] channel.NoteOn channel 0 key 52 dyn =
[60] channel.NoteOff channel 0 key 52
[260] channel.NoteOn channel 0 key 56 dyn =
[60] channel.NoteOff channel 0 key 56
[260] meta.Undefined type:  0
[0] channel.NoteOn channel 0 key 57 dyn =
[60] channel.NoteOff channel 0 key 57
`,
		},
		{
			// 28
			`
TRACK |p|
Ch |1  |

=SCORE |p|
#
1  | d |
2  |{c,e,g#} |
3  |a |
4 |*|
			`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 50 dyn =
[960] channel.NoteOff channel 0 key 50
[0] channel.NoteOn channel 0 key 48 dyn =
[320] channel.NoteOff channel 0 key 48
[0] channel.NoteOn channel 0 key 52 dyn =
[320] channel.NoteOff channel 0 key 52
[0] channel.NoteOn channel 0 key 56 dyn =
[320] channel.NoteOff channel 0 key 56
[0] meta.Undefined type:  0
[0] channel.NoteOn channel 0 key 57 dyn =
[960] channel.NoteOff channel 0 key 57
`,
		},
		{
			// 29
			`
TRACK |p|
Ch |1  |

=SCORE |p|
#
1  | d |
4  |{c,e,g#} |
#
1  |a |
2 |*|
			`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 50 dyn =
[2880] channel.NoteOff channel 0 key 50
[0] channel.NoteOn channel 0 key 48 dyn =
[320] channel.NoteOff channel 0 key 48
[0] channel.NoteOn channel 0 key 52 dyn =
[320] channel.NoteOff channel 0 key 52
[0] channel.NoteOn channel 0 key 56 dyn =
[320] channel.NoteOff channel 0 key 56
[0] meta.Undefined type:  0
[0] channel.NoteOn channel 0 key 57 dyn =
[960] channel.NoteOff channel 0 key 57
`,
		},
		{
			// 30
			`
TRACK |p|
Ch |1  |

=SCORE |p|
#
1  | d |

`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 50 dyn =
[3840] channel.NoteOff channel 0 key 50
`,
		},
		{
			// 31
			`
TRACK |p|
Ch |1  |

=SCORE |p|
#
1  | d |

`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 50 dyn =
[3840] channel.NoteOff channel 0 key 50


`,
		},
		{
			// 32
			`
TRACK |p|
Ch |1  |

=SCORE |p|
#
1  |c |
#
1  |* |
#
1  |e |

			`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 48 dyn =
[3840] channel.NoteOff channel 0 key 48
[3840] channel.NoteOn channel 0 key 52 dyn =
[3840] channel.NoteOff channel 0 key 52
`,
		},
		{
			// 33
			`
TRACK |p|
Ch |1  |

=SCORE |p|
#
1  |c |
#
1  | * |
#
1  |e |

			`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 48 dyn =
[3840] channel.NoteOff channel 0 key 48
[3840] channel.NoteOn channel 0 key 52 dyn =
[3840] channel.NoteOff channel 0 key 52
`,
		},
		{
			// 34
			`
TRACK |p|
Ch |1  |

=SCORE |p|
#
1  |c |
#
1  | |
#
1  |e |

			`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 48 dyn =
[7680] channel.NoteOff channel 0 key 48
[0] channel.NoteOn channel 0 key 52 dyn =
[3840] channel.NoteOff channel 0 key 52
`,
		},
		{
			// 35
			`
TRACK |p|
Ch |1  |

=SCORE |p|
#
1  |a' |
#
1  |* |
#
1  |b' |
			`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 69 dyn =
[3840] channel.NoteOff channel 0 key 69
[3840] channel.NoteOn channel 0 key 71 dyn =
[3840] channel.NoteOff channel 0 key 71
`,
		},
		{
			// 36
			`
TRACK |p|
Ch |1  |

=SCORE |p|
#
1  |a' |
#
1  | * |
#
1  |b' |

			`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 69 dyn =
[3840] channel.NoteOff channel 0 key 69
[3840] channel.NoteOn channel 0 key 71 dyn =
[3840] channel.NoteOff channel 0 key 71
`,
		},
		{
			// 37
			`
TRACK |p|
Ch |1  |

=SCORE |p|
#
1  | (a' b') |
#
1  | * |
#
1  | (b' c") |
			`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 69 dyn =
[0] channel.NoteOn channel 0 key 71 dyn =
[3840] channel.NoteOff channel 0 key 69
[0] channel.NoteOff channel 0 key 71
[3840] channel.NoteOn channel 0 key 71 dyn =
[0] channel.NoteOn channel 0 key 72 dyn =
[3840] channel.NoteOff channel 0 key 71
[0] channel.NoteOff channel 0 key 72
`,
		},
		{
			// 38
			`
TRACK |p|
Ch |1  |

=SCORE |p|
#
1  | (a' b') |
#
1  |  |
#
1  | (b' c") |
			`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 69 dyn =
[0] channel.NoteOn channel 0 key 71 dyn =
[7680] channel.NoteOff channel 0 key 69
[0] channel.NoteOff channel 0 key 71
[0] channel.NoteOn channel 0 key 71 dyn =
[0] channel.NoteOn channel 0 key 72 dyn =
[3840] channel.NoteOff channel 0 key 71
[0] channel.NoteOff channel 0 key 72
`,
		},
		{
			// 39
			`
TRACK |p|
Ch |1  |

=SCORE |p|
#
1  | (a' b'_) |
#
1  |  |
#
1  | (_b' c") |
			`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 69 dyn =
[0] channel.NoteOn channel 0 key 71 dyn =
[7680] channel.NoteOff channel 0 key 69
[0] channel.NoteOff channel 0 key 71
[0] channel.NoteOn channel 0 key 72 dyn =
[3840] channel.NoteOff channel 0 key 72
`,
		},
		{
			// 40
			`
TRACK |p|
Ch |1  |

=SCORE |p|
#
1  |a' |
1& |b_|
#
1  |  |
1& |_b|
#
1  |c" |
			`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 69 dyn =
[480] channel.NoteOn channel 0 key 59 dyn =
[3840] channel.NoteOff channel 0 key 59
[3360] channel.NoteOff channel 0 key 69
[0] channel.NoteOn channel 0 key 72 dyn =
[3840] channel.NoteOff channel 0 key 72
`,
		},
		{
			// 41
			`
TRACK |p|
Ch |1  |

=SCORE |p|
#
1  |a' |
1& |b'_|
#
1  |  |
1& | (_b' :)|
#
1  |c" |
			`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 69 dyn =
[480] channel.NoteOn channel 0 key 71 dyn =
[3840] channel.NoteOff channel 0 key 71
[3360] channel.NoteOff channel 0 key 69
[0] channel.NoteOn channel 0 key 72 dyn =
[3840] channel.NoteOff channel 0 key 72
`,
		},
		{
			// 42
			`
TRACK |p|
Ch |1  |

=SCORE |p|
#
1  | (a' b'_) |
#
1  | _b' |
			`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 69 dyn =
[0] channel.NoteOn channel 0 key 71 dyn =
[3840] channel.NoteOff channel 0 key 71
[3840] channel.NoteOff channel 0 key 69
`,
		},
		{
			// 43
			`
TRACK |p|
Ch |1  |

=SCORE |p|
#
1  | (a' b'_ c"_) |
#
1  | (_b' _c") |
			`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 69 dyn =
[0] channel.NoteOn channel 0 key 71 dyn =
[0] channel.NoteOn channel 0 key 72 dyn =
[3840] channel.NoteOff channel 0 key 69
[0] channel.NoteOff channel 0 key 71
[0] channel.NoteOff channel 0 key 72
`,
		},
		{
			// 44
			`
TRACK |p|
Ch |1  |

=SCORE |p|
#
2  | (a' "hiho") |
			`,
			2,
			false,
			false,
			`
[960] channel.NoteOn channel 0 key 69 dyn =
[0] meta.Lyric: "hiho"
[2880] channel.NoteOff channel 0 key 69
`,
		},
		{
			// 45
			`
TRACK |p|
Ch |1  |

=SCORE |p|
#
2  |"hiho" |
			`,
			2,
			false,
			false,
			`
[960] meta.Lyric: "hiho"
`,
		},
		{
			// 46
			`
TRACK |p|
Ch |1  |

=SCORE |p|
#
1  |c:: |
			`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 48 dyn =
[120] channel.NoteOff channel 0 key 48
`,
		},
		{
			// 47
			`
TRACK |p|
Ch |1  |

=SCORE |p|
#
1  |c: |
			`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 48 dyn =
[240] channel.NoteOff channel 0 key 48
`,
		},
		{
			// 48
			`
TRACK |p|
Ch |1  |

=SCORE |p|
#
1  |AT(23):: |
			`,
			2,
			false,
			false,
			`
[0] channel.Aftertouch channel 0 pressure 23
[120] channel.Aftertouch channel 0 pressure 0
`,
		},
		{
			// 49
			`
TRACK |p|
Ch |1  |

=SCORE |p|
#
1  |AT(23): |
			`,
			2,
			false,
			false,
			`
[0] channel.Aftertouch channel 0 pressure 23
[240] channel.Aftertouch channel 0 pressure 0
`,
		},
		{
			// 50
			`
=cc  |x|
#
1    |c<|
2    |* |

TRACK |p|
Ch |1  |

=SCORE |p|
#
2  |=cc |
3  |*  |
			`,
			2,
			false,
			false,
			`
[930] channel.NoteOn channel 0 key 48 dyn =
[990] channel.NoteOff channel 0 key 48
`,
		},
		{
			// 51
			`
=cc  |x|
#
1    |c<|
2    |* |

TRACK |p|
Ch |1  |

=SCORE |p|
#
2  |=cc> |
3  |*  |
			`,
			2,
			false,
			false,
			`
[930] channel.NoteOn channel 0 key 48 dyn =
[990] channel.NoteOff channel 0 key 48
`,
		},
		{
			// 52
			`
=cc  |x|
#
1    |c|
2    |* |

TRACK |p|
Ch |1  |

=SCORE |p|
#
2  |=cc< |
3  |*  |
			`,
			2,
			false,
			false,
			`
[930] channel.NoteOn channel 0 key 48 dyn =
[990] channel.NoteOff channel 0 key 48
`,
		},
		{
			// 53
			`
TRACK |p|
Ch |1  |

=SCORE |p|
#
1  | d+ |`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 50 dyn +
[3840] channel.NoteOff channel 0 key 50


`,
		},
		{
			// 54
			`
TRACK |p|
Ch |1  |

=SCORE |p|
#
1  | d++ |`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 50 dyn ++
[3840] channel.NoteOff channel 0 key 50


`,
		},
		{
			// 55
			`
TRACK |p|
Ch |1  |

=SCORE |p|
#
1  | d+++ |`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 50 dyn +++
[3840] channel.NoteOff channel 0 key 50


`,
		},
		{
			// 56
			`
TRACK |p|
Ch |1  |

=SCORE |p|
#
1  | d- |`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 50 dyn -
[3840] channel.NoteOff channel 0 key 50


`,
		},
		{
			// 57
			`
TRACK |p|
Ch |1  |

=SCORE |p|
#
1  | d-- |`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 50 dyn --
[3840] channel.NoteOff channel 0 key 50


`,
		},
		{
			// 58
			`
TRACK |p|
Ch |1  |
VelScale|step:30|

=SCORE |p|
#
1  | d+ |`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 50 dyn ++
[3840] channel.NoteOff channel 0 key 50


`,
		},
		{
			// 59
			`
TRACK |p|
Ch |1  |
VelScale|step:30|

=SCORE |p|
#
1  | d++ |

`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 50 dyn ++++
[3840] channel.NoteOff channel 0 key 50


`,
		},
		{
			// 60
			`
TRACK |p|
Ch |1  |
VelScale|step:30|

=SCORE |p|
#
1  | d- |
`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 50 dyn --
[3840] channel.NoteOff channel 0 key 50


`,
		},
		{
			// 61
			`
TRACK |p|
Ch |1  |
VelScale|step:30|

=SCORE |p|
#
1  | d-- |
`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 50 dyn ----
[3840] channel.NoteOff channel 0 key 50


`,
		},
		{
			// 62
			`
TRACK |p|
Ch |1  |
VelScale|min:30|

=SCORE |p|
#
1  | d---- |
`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 50 dyn --
[3840] channel.NoteOff channel 0 key 50


`,
		},
		{
			// 63
			`
TRACK |p|
Ch |1  |
VelScale|max:100|

=SCORE |p|
#
1  | d++++ |
`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 50 dyn ++
[3840] channel.NoteOff channel 0 key 50


`,
		},
		{
			// 64
			`
TRACK |p|
Ch |1  |
VelScale|center:78|

=SCORE |p|
#
1  | d |
`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 50 dyn +
[3840] channel.NoteOff channel 0 key 50


`,
		},
		{
			// 65
			`
TRACK   | git | Vocals |
Ch      | 2   | 1      |
PbRange | 2   | 2      |

=SCORE  | git      | Vocals |
#
1       | =.Vocals |        |
#
1       |       | f++       |   
2       |        |    g-    |
#
1       |        |      c'  |
`,
			3,
			false,
			false,
			`
[3840] channel.NoteOn channel 1 key 53 dyn ++
[960] channel.NoteOff channel 1 key 53
[0] channel.NoteOn channel 1 key 55 dyn -
[2880] channel.NoteOff channel 1 key 55
[0] channel.NoteOn channel 1 key 60 dyn =
[3840] channel.NoteOff channel 1 key 60
`,
		},
		{
			// 66
			`
TRACK |Vocals|git|
Ch      |1       |2    |
PbRange |2       |2    |

=SCORE  |Vocals|git|
#
1   |        | =.Vocals|
    |        | /d   |
#
1   | f++    |     |   
2   | g-     |     |
#
1   |        | /c   |
2   |        | /PB(20) |
`,
			3,
			false,
			false,
			`
[0] channel.NoteOn channel 1 key 50 dyn =
[3840] channel.NoteOff channel 1 key 50
[0] channel.NoteOn channel 1 key 53 dyn ++
[960] channel.NoteOff channel 1 key 53
[0] channel.NoteOn channel 1 key 55 dyn -
[2880] channel.NoteOff channel 1 key 55
[0] channel.NoteOn channel 1 key 48 dyn =
[960] channel.Pitchbend channel 1 value 20 absValue 0
[2880] channel.NoteOff channel 1 key 48
`,
		},
		{
			// 67
			`
TRACK |p|
Ch |1  |

=SCORE |p|
#A
1  |c" |
2  |d  |
[A]
			`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 72 dyn =
[960] channel.NoteOff channel 0 key 72
[0] channel.NoteOn channel 0 key 50 dyn =
[2880] channel.NoteOff channel 0 key 50
[0] channel.NoteOn channel 0 key 72 dyn =
[960] channel.NoteOff channel 0 key 72
[0] channel.NoteOn channel 0 key 50 dyn =
[2880] channel.NoteOff channel 0 key 50
`,
		},
		{
			// 68
			`
=rhythm | x |
#
1       | $_euclid(3,8,&) |

TRACK |p|
Ch |1  |

=SCORE |p|
#
1  | =rhythm(c"+,d) |
#
1  |*|
			`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 72 dyn +
[480] channel.NoteOff channel 0 key 72
[0] channel.NoteOn channel 0 key 50 dyn =
[480] channel.NoteOff channel 0 key 50
[0] channel.NoteOn channel 0 key 50 dyn =
[480] channel.NoteOff channel 0 key 50
[0] channel.NoteOn channel 0 key 72 dyn +
[480] channel.NoteOff channel 0 key 72
[0] channel.NoteOn channel 0 key 50 dyn =
[480] channel.NoteOff channel 0 key 50
[0] channel.NoteOn channel 0 key 50 dyn =
[480] channel.NoteOff channel 0 key 50
[0] channel.NoteOn channel 0 key 72 dyn +
[480] channel.NoteOff channel 0 key 72
[0] channel.NoteOn channel 0 key 50 dyn =
[480] channel.NoteOff channel 0 key 50
`,
		},
		{
			// 69
			`
=rhythm |x|
#
1 | $_euclid(3,8,1) |
*2

TRACK |p|
Ch |1  |

=SCORE |p|
#
1  |=rhythm(c"+,d) |
#
1 | |
#
1  |*|
			`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 72 dyn +
[960] channel.NoteOff channel 0 key 72
[0] channel.NoteOn channel 0 key 50 dyn =
[960] channel.NoteOff channel 0 key 50
[0] channel.NoteOn channel 0 key 50 dyn =
[960] channel.NoteOff channel 0 key 50
[0] channel.NoteOn channel 0 key 72 dyn +
[960] channel.NoteOff channel 0 key 72
[0] channel.NoteOn channel 0 key 50 dyn =
[960] channel.NoteOff channel 0 key 50
[0] channel.NoteOn channel 0 key 50 dyn =
[960] channel.NoteOff channel 0 key 50
[0] channel.NoteOn channel 0 key 72 dyn +
[960] channel.NoteOff channel 0 key 72
[0] channel.NoteOn channel 0 key 50 dyn =
[960] channel.NoteOff channel 0 key 50
`,
		},
		{
			// 70
			`
TRACK |p|
Ch |1  |
VelScale|center:78|

=SCORE |p|
#
1  | d |
2  | % |
3  | * |
`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 50 dyn +
[960] channel.NoteOff channel 0 key 50
[0] channel.NoteOn channel 0 key 50 dyn +
[960] channel.NoteOff channel 0 key 50
`,
		},
		{
			// 71
			`
TRACK |p|
Ch |1  |
VelScale|center:78|

=SCORE |p|
#
1  | d |
2  | * |
3  | % |
4  | * |
`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 50 dyn +
[960] channel.NoteOff channel 0 key 50
[960] channel.NoteOn channel 0 key 50 dyn +
[960] channel.NoteOff channel 0 key 50
`,
		},
		{
			// 72
			`
TRACK |p|
Ch |1  |
VelScale|center:78|

=SCORE |p|
#
1  | d |
3  | e |
4  | * |
#
1  | ./. |
`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 50 dyn +
[1920] channel.NoteOff channel 0 key 50
[0] channel.NoteOn channel 0 key 52 dyn +
[960] channel.NoteOff channel 0 key 52
[960] channel.NoteOn channel 0 key 50 dyn +
[1920] channel.NoteOff channel 0 key 50
[0] channel.NoteOn channel 0 key 52 dyn +
[960] channel.NoteOff channel 0 key 52
`,
		},
		{
			// 73
			`
TRACK |p|
Ch |1  |
VelScale|center:78|

=SCORE |p|
#
2  | g |
#
1  | d |
3  | e |
4  | * |
#
1  | .2. |
*3
#
1  | c |
2  | * |
`,
			2,
			false,
			false,
			`
[960] channel.NoteOn channel 0 key 55 dyn +
[2880] channel.NoteOff channel 0 key 55
[0] channel.NoteOn channel 0 key 50 dyn +
[1920] channel.NoteOff channel 0 key 50
[0] channel.NoteOn channel 0 key 52 dyn +
[960] channel.NoteOff channel 0 key 52
[1920] channel.NoteOn channel 0 key 55 dyn +
[2880] channel.NoteOff channel 0 key 55
[0] channel.NoteOn channel 0 key 50 dyn +
[1920] channel.NoteOff channel 0 key 50
[0] channel.NoteOn channel 0 key 52 dyn +
[960] channel.NoteOff channel 0 key 52
[1920] channel.NoteOn channel 0 key 55 dyn +
[2880] channel.NoteOff channel 0 key 55
[0] channel.NoteOn channel 0 key 50 dyn +
[1920] channel.NoteOff channel 0 key 50
[0] channel.NoteOn channel 0 key 52 dyn +
[960] channel.NoteOff channel 0 key 52
[960] channel.NoteOn channel 0 key 48 dyn +
[960] channel.NoteOff channel 0 key 48
`,
		},
		{
			// 74
			`
TRACK |p|
Ch |1  |
VelScale|center:78|

=SCORE |p|
#
2  | g |
#
1  | d |
3  | e |
4  | * |
#
1  | .2. |
*5
#
1  | c |
2  | * |
`,
			2,
			false,
			false,
			`
[960] channel.NoteOn channel 0 key 55 dyn +
[2880] channel.NoteOff channel 0 key 55
[0] channel.NoteOn channel 0 key 50 dyn +
[1920] channel.NoteOff channel 0 key 50
[0] channel.NoteOn channel 0 key 52 dyn +
[960] channel.NoteOff channel 0 key 52
[1920] channel.NoteOn channel 0 key 55 dyn +
[2880] channel.NoteOff channel 0 key 55
[0] channel.NoteOn channel 0 key 50 dyn +
[1920] channel.NoteOff channel 0 key 50
[0] channel.NoteOn channel 0 key 52 dyn +
[960] channel.NoteOff channel 0 key 52
[1920] channel.NoteOn channel 0 key 55 dyn +
[2880] channel.NoteOff channel 0 key 55
[0] channel.NoteOn channel 0 key 50 dyn +
[1920] channel.NoteOff channel 0 key 50
[0] channel.NoteOn channel 0 key 52 dyn +
[960] channel.NoteOff channel 0 key 52
[1920] channel.NoteOn channel 0 key 55 dyn +
[2880] channel.NoteOff channel 0 key 55
[0] channel.NoteOn channel 0 key 50 dyn +
[1920] channel.NoteOff channel 0 key 50
[0] channel.NoteOn channel 0 key 52 dyn +
[960] channel.NoteOff channel 0 key 52
[960] channel.NoteOn channel 0 key 48 dyn +
[960] channel.NoteOff channel 0 key 48
`,
		},
		{
			// 75
			`
TRACK |p|
Ch |1  |
VelScale|center:78|

=SCORE |p|
#
2  | g |
#
1  | d |
3  | e |
4  | * |
#
1  | .2. |
*4
#
1  | c |
2  | * |
`,
			2,
			false,
			false,
			`
[960] channel.NoteOn channel 0 key 55 dyn +
[2880] channel.NoteOff channel 0 key 55
[0] channel.NoteOn channel 0 key 50 dyn +
[1920] channel.NoteOff channel 0 key 50
[0] channel.NoteOn channel 0 key 52 dyn +
[960] channel.NoteOff channel 0 key 52
[1920] channel.NoteOn channel 0 key 55 dyn +
[2880] channel.NoteOff channel 0 key 55
[0] channel.NoteOn channel 0 key 50 dyn +
[1920] channel.NoteOff channel 0 key 50
[0] channel.NoteOn channel 0 key 52 dyn +
[960] channel.NoteOff channel 0 key 52
[1920] channel.NoteOn channel 0 key 55 dyn +
[2880] channel.NoteOff channel 0 key 55
[0] channel.NoteOn channel 0 key 50 dyn +
[1920] channel.NoteOff channel 0 key 50
[0] channel.NoteOn channel 0 key 52 dyn +
[960] channel.NoteOff channel 0 key 52
[1920] channel.NoteOn channel 0 key 55 dyn +
[2880] channel.NoteOff channel 0 key 55
[0] channel.NoteOn channel 0 key 48 dyn +
[960] channel.NoteOff channel 0 key 48
`,
		},
		{
			// 76
			`
TRACK |p|
Ch |1  |
VelScale|center:78|

=SCORE |p|
#
2  | g |
#
1  | d |
3  | e |
4  | * |
#
1  | .2. |
*5
`,
			2,
			false,
			false,
			`
[960] channel.NoteOn channel 0 key 55 dyn +
[2880] channel.NoteOff channel 0 key 55
[0] channel.NoteOn channel 0 key 50 dyn +
[1920] channel.NoteOff channel 0 key 50
[0] channel.NoteOn channel 0 key 52 dyn +
[960] channel.NoteOff channel 0 key 52
[1920] channel.NoteOn channel 0 key 55 dyn +
[2880] channel.NoteOff channel 0 key 55
[0] channel.NoteOn channel 0 key 50 dyn +
[1920] channel.NoteOff channel 0 key 50
[0] channel.NoteOn channel 0 key 52 dyn +
[960] channel.NoteOff channel 0 key 52
[1920] channel.NoteOn channel 0 key 55 dyn +
[2880] channel.NoteOff channel 0 key 55
[0] channel.NoteOn channel 0 key 50 dyn +
[1920] channel.NoteOff channel 0 key 50
[0] channel.NoteOn channel 0 key 52 dyn +
[960] channel.NoteOff channel 0 key 52
[1920] channel.NoteOn channel 0 key 55 dyn +
[2880] channel.NoteOff channel 0 key 55
[0] channel.NoteOn channel 0 key 50 dyn +
[1920] channel.NoteOff channel 0 key 50
[0] channel.NoteOn channel 0 key 52 dyn +
[960] channel.NoteOff channel 0 key 52
`,
		},
		{
			// 77
			`
TRACK |p|
Ch |1  |
VelScale|center:78|

=patt |a|
#
2  | g |
#
1  | d |
3  | e |
4  | * |
#
1  | .2. |
#
1  |     |

=SCORE |p|
#
1  | =patt.a%2 |
*7
`,
			2,
			false,
			false,
			`
[960] channel.NoteOn channel 0 key 55 dyn +
[2880] channel.NoteOff channel 0 key 55
[0] channel.NoteOn channel 0 key 50 dyn +
[1920] channel.NoteOff channel 0 key 50
[0] channel.NoteOn channel 0 key 52 dyn +
[960] channel.NoteOff channel 0 key 52
[1920] channel.NoteOn channel 0 key 55 dyn +
[2880] channel.NoteOff channel 0 key 55
[0] channel.NoteOn channel 0 key 50 dyn +
[1920] channel.NoteOff channel 0 key 50
[0] channel.NoteOn channel 0 key 52 dyn +
[960] channel.NoteOff channel 0 key 52
[1920] channel.NoteOn channel 0 key 55 dyn +
[2880] channel.NoteOff channel 0 key 55
[0] channel.NoteOn channel 0 key 50 dyn +
[1920] channel.NoteOff channel 0 key 50
[0] channel.NoteOn channel 0 key 52 dyn +
[960] channel.NoteOff channel 0 key 52
[1920] channel.NoteOn channel 0 key 55 dyn +
[2880] channel.NoteOff channel 0 key 55
[0] channel.NoteOn channel 0 key 50 dyn +
[1920] channel.NoteOff channel 0 key 50
[0] channel.NoteOn channel 0 key 52 dyn +
[960] channel.NoteOff channel 0 key 52
`,
		},
		{
			// 78
			`
TRACK |p|
Ch |1  |

=SCORE |p|
#A
1  |c" |
2  |d  |
[A]
[A]
			`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 72 dyn =
[960] channel.NoteOff channel 0 key 72
[0] channel.NoteOn channel 0 key 50 dyn =
[2880] channel.NoteOff channel 0 key 50
[0] channel.NoteOn channel 0 key 72 dyn =
[960] channel.NoteOff channel 0 key 72
[0] channel.NoteOn channel 0 key 50 dyn =
[2880] channel.NoteOff channel 0 key 50
[0] channel.NoteOn channel 0 key 72 dyn =
[960] channel.NoteOff channel 0 key 72
[0] channel.NoteOn channel 0 key 50 dyn =
[2880] channel.NoteOff channel 0 key 50
`,
		},
		{
			// 79
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
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 72 dyn =
[960] channel.NoteOff channel 0 key 72
[0] channel.NoteOn channel 0 key 50 dyn =
[2880] channel.NoteOff channel 0 key 50
[0] channel.NoteOn channel 0 key 72 dyn =
[960] channel.NoteOff channel 0 key 72
[0] channel.NoteOn channel 0 key 50 dyn =
[2880] channel.NoteOff channel 0 key 50
[0] channel.NoteOn channel 0 key 60 dyn =
[960] channel.NoteOff channel 0 key 60
[2880] channel.NoteOn channel 0 key 72 dyn =
[960] channel.NoteOff channel 0 key 72
[0] channel.NoteOn channel 0 key 50 dyn =
[2880] channel.NoteOff channel 0 key 50
`,
		},
		{
			// 80
			`
TRACK |p|
Ch |1  |

=SCORE |p|
#A
1  |c" |
3  |*|
#B
1  |f"  |
3 |*|
[A]
#
1  | c' |
3 |* |
[A]
[B]
			`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 72 dyn =
[1920] channel.NoteOff channel 0 key 72
[1920] channel.NoteOn channel 0 key 77 dyn =
[1920] channel.NoteOff channel 0 key 77
[1920] channel.NoteOn channel 0 key 72 dyn =
[1920] channel.NoteOff channel 0 key 72
[1920] channel.NoteOn channel 0 key 60 dyn =
[1920] channel.NoteOff channel 0 key 60
[1920] channel.NoteOn channel 0 key 72 dyn =
[1920] channel.NoteOff channel 0 key 72
[1920] channel.NoteOn channel 0 key 77 dyn =
[1920] channel.NoteOff channel 0 key 77
`,
		},
		{
			// 81
			`
TRACK |p|
Ch |1  |

=temp |a|
#A
1  |c" | 
3  |*|
#B
1  |f"  | 
3 |*|
[A]
#
1  | c' |
3 |* |
[A]
[B]

=SCORE |p|
#
1    |=temp.a|
*6
			`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 72 dyn =
[1920] channel.NoteOff channel 0 key 72
[1920] channel.NoteOn channel 0 key 77 dyn =
[1920] channel.NoteOff channel 0 key 77
[1920] channel.NoteOn channel 0 key 72 dyn =
[1920] channel.NoteOff channel 0 key 72
[1920] channel.NoteOn channel 0 key 60 dyn =
[1920] channel.NoteOff channel 0 key 60
[1920] channel.NoteOn channel 0 key 72 dyn =
[1920] channel.NoteOff channel 0 key 72
[1920] channel.NoteOn channel 0 key 77 dyn =
[1920] channel.NoteOff channel 0 key 77
`,
		},
		{
			// 82
			`
TRACK  | Voice |
Channel | 3     |

@voice  |        
Hi      | V1 
ho-ha   |        
hu      |

=melody | Verse |
#A
1       | c"    |
3       | *     |
#B
1       | f"    | 
4       | *     |
[A]
[B]

=SCORE  | Voice                     |
# 4/4 @120
1       | =melody.Verse@voice.V1[0:2] |
*3

			`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 2 key 72 dyn =
[0] meta.Lyric: "Hi"
[1920] channel.NoteOff channel 2 key 72
[1920] channel.NoteOn channel 2 key 77 dyn =
[0] meta.Lyric: "ho-"
[2880] channel.NoteOff channel 2 key 77
[960] channel.NoteOn channel 2 key 72 dyn =
[0] meta.Lyric: "ha"
[1920] channel.NoteOff channel 2 key 72
[1920] channel.NoteOn channel 2 key 77 dyn =
[0] meta.Lyric: "Hi"
[2880] channel.NoteOff channel 2 key 77
`,
		},
		{
			// 83
			`
TRACK  | Voice |
Channel | 3     |

@voice   |        
Hi      | V1 
ho-ha   |        
hu      |

=melody | Verse |
#A
1       | c"    | 
3       | *     |
#B
1       | f"    | 
4       | *     |
[A]
[B]

=SCORE  | Voice                     |
# 4/4 @120
1       | =melody.Verse[2:]@voice.V1[0:2] |
*3

			`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 2 key 77 dyn =
[0] meta.Lyric: "Hi"
[2880] channel.NoteOff channel 2 key 77
[960] channel.NoteOn channel 2 key 72 dyn =
[0] meta.Lyric: "ho-"
[1920] channel.NoteOff channel 2 key 72
[1920] channel.NoteOn channel 2 key 77 dyn =
[0] meta.Lyric: "ha"
[2880] channel.NoteOff channel 2 key 77
`,
		},
		{
			// 84
			`
TRACK  | Voice |
Channel | 3     |

=SCORE  | Voice                     |
# 4/4 @120
1       | e_ |
1       | f  |
3       | *  |
4       | _e |

			`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 2 key 52 dyn =
[0] channel.NoteOn channel 2 key 53 dyn =
[1920] channel.NoteOff channel 2 key 53
[960] channel.NoteOff channel 2 key 52
`,
		},
		{
			// 85
			`
TRACK  | Voice |
Channel | 3     |

=SCORE  | Voice                     |
# 4/4 @120
1       | e_ |
        | f  |
3       | *  |
4       | _e |

			`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 2 key 52 dyn =
[0] channel.NoteOn channel 2 key 53 dyn =
[1920] channel.NoteOff channel 2 key 53
[960] channel.NoteOff channel 2 key 52
`,
		},
		{
			// 86
			`
TRACK  | Voice |
Channel | 3     |

=SCORE  | Voice                     |
# 4/4 @120
1       | ^1 |
2       | ^2  |
3       | * |

			`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 2 key 60 dyn =
[960] channel.NoteOff channel 2 key 60
[0] channel.NoteOn channel 2 key 62 dyn =
[960] channel.NoteOff channel 2 key 62
`,
		},
		{
			// 87
			`
TRACK  | Voice |
Channel | 3     |

=SCORE  | Voice                     |
# 4/4 @120 \ionian^d#'
1       | ^1 |
2       | ^2  |
3       | * |

			`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 2 key 63 dyn =
[960] channel.NoteOff channel 2 key 63
[0] channel.NoteOn channel 2 key 65 dyn =
[960] channel.NoteOff channel 2 key 65
`,
		},
		{
			// 88
			`
TRACK |p  |
Ch     |1  |

=SCORE |p  |
#
1      |c" |
1&     |:  |
2      |d  |

			`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 72 dyn =
[960] channel.NoteOff channel 0 key 72
[0] channel.NoteOn channel 0 key 50 dyn =
[2880] channel.NoteOff channel 0 key 50
`,
		},
		{
			// 89
			`
TRACK |p  |
Ch     |1  |

=patt | |
#
1      |c" |
2      |d  |

=SCORE |p  |
#
1      |=patt |
2      |/:  |
3      |d  |

			`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 72 dyn =
[1920] channel.NoteOff channel 0 key 72
[0] channel.NoteOn channel 0 key 50 dyn =
[1920] channel.NoteOff channel 0 key 50
`,
		},
		{
			// 90
			`
TRACK |p  |
Ch     |1  |

=SCORE |p  |
#
1      |c"+- |
2      |^1----+++  |
3      |d+++-  |

			`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 72 dyn =
[960] channel.NoteOff channel 0 key 72
[0] channel.NoteOn channel 0 key 60 dyn -
[960] channel.NoteOff channel 0 key 60
[0] channel.NoteOn channel 0 key 50 dyn ++
[1920] channel.NoteOff channel 0 key 50
`,
		},
		{
			// 91
			`
TRACK |p |
Ch      |1 |

=SCORE |p        |
#
1      |{c,e,g#,a} |
3      |*        |

			`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 48 dyn =
[480] channel.NoteOff channel 0 key 48
[0] channel.NoteOn channel 0 key 52 dyn =
[480] channel.NoteOff channel 0 key 52
[0] channel.NoteOn channel 0 key 56 dyn =
[480] channel.NoteOff channel 0 key 56
[0] channel.NoteOn channel 0 key 57 dyn =
[480] channel.NoteOff channel 0 key 57
[0] meta.Undefined type:  0
`,
		},
		{
			// 92
			`
TRACK |p |
Ch      |1 |

=SCORE |p        |
#
1      |{c,e,g#,{a,b}} |
3      |*        |

			`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 48 dyn =
[480] channel.NoteOff channel 0 key 48
[0] channel.NoteOn channel 0 key 52 dyn =
[480] channel.NoteOff channel 0 key 52
[0] channel.NoteOn channel 0 key 56 dyn =
[480] channel.NoteOff channel 0 key 56
[0] channel.NoteOn channel 0 key 57 dyn =
[240] channel.NoteOff channel 0 key 57
[0] channel.NoteOn channel 0 key 59 dyn =
[240] channel.NoteOff channel 0 key 59
[0] meta.Undefined type:  0
[0] meta.Undefined type:  0
`,
		},
		{
			// 93
			`
TRACK |p |
Ch      |1 |

=SCORE |p        |
#
1      |{c,e,{a,b},g#} |
3      |*        |

			`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 48 dyn =
[480] channel.NoteOff channel 0 key 48
[0] channel.NoteOn channel 0 key 52 dyn =
[480] channel.NoteOff channel 0 key 52
[0] channel.NoteOn channel 0 key 57 dyn =
[240] channel.NoteOff channel 0 key 57
[0] channel.NoteOn channel 0 key 59 dyn =
[240] channel.NoteOff channel 0 key 59
[0] meta.Undefined type:  0
[0] channel.NoteOn channel 0 key 56 dyn =
[480] channel.NoteOff channel 0 key 56
[0] meta.Undefined type:  0
`,
		},
		{
			// 94
			`
TRACK |p |
Ch      |1 |

=SCORE |p        |
#
1      |{c,e,{a,b,c#,d},g#} |
3      |*        |

			`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 48 dyn =
[480] channel.NoteOff channel 0 key 48
[0] channel.NoteOn channel 0 key 52 dyn =
[480] channel.NoteOff channel 0 key 52
[0] channel.NoteOn channel 0 key 57 dyn =
[120] channel.NoteOff channel 0 key 57
[0] channel.NoteOn channel 0 key 59 dyn =
[120] channel.NoteOff channel 0 key 59
[0] channel.NoteOn channel 0 key 49 dyn =
[120] channel.NoteOff channel 0 key 49
[0] channel.NoteOn channel 0 key 50 dyn =
[120] channel.NoteOff channel 0 key 50
[0] meta.Undefined type:  0
[0] channel.NoteOn channel 0 key 56 dyn =
[480] channel.NoteOff channel 0 key 56
[0] meta.Undefined type:  0
`,
		},
		{
			// 95
			`
TRACK |p |
Ch      |1 |

=SCORE |p        |
#
1      |{c,e,{c,e,c,e},c} |
3      |*        |

			`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 48 dyn =
[480] channel.NoteOff channel 0 key 48
[0] channel.NoteOn channel 0 key 52 dyn =
[480] channel.NoteOff channel 0 key 52
[0] channel.NoteOn channel 0 key 48 dyn =
[120] channel.NoteOff channel 0 key 48
[0] channel.NoteOn channel 0 key 52 dyn =
[120] channel.NoteOff channel 0 key 52
[0] channel.NoteOn channel 0 key 48 dyn =
[120] channel.NoteOff channel 0 key 48
[0] channel.NoteOn channel 0 key 52 dyn =
[120] channel.NoteOff channel 0 key 52
[0] meta.Undefined type:  0
[0] channel.NoteOn channel 0 key 48 dyn =
[480] channel.NoteOff channel 0 key 48
[0] meta.Undefined type:  0
`,
		},
		{
			// 96
			`
TRACK |p |
Ch      |1 |

=SCORE |p        |
#
1      |{c,e,{c,e,c,{g,b}},c} |
3      |*        |

			`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 48 dyn =
[480] channel.NoteOff channel 0 key 48
[0] channel.NoteOn channel 0 key 52 dyn =
[480] channel.NoteOff channel 0 key 52
[0] channel.NoteOn channel 0 key 48 dyn =
[120] channel.NoteOff channel 0 key 48
[0] channel.NoteOn channel 0 key 52 dyn =
[120] channel.NoteOff channel 0 key 52
[0] channel.NoteOn channel 0 key 48 dyn =
[120] channel.NoteOff channel 0 key 48
[0] channel.NoteOn channel 0 key 55 dyn =
[60] channel.NoteOff channel 0 key 55
[0] channel.NoteOn channel 0 key 59 dyn =
[60] channel.NoteOff channel 0 key 59
[0] meta.Undefined type:  0
[0] meta.Undefined type:  0
[0] channel.NoteOn channel 0 key 48 dyn =
[480] channel.NoteOff channel 0 key 48
[0] meta.Undefined type:  0
`,
		},
		{
			// 97
			`
TRACK |p |
Ch      |1 |

sh |
xx | c
yy | e
zz | g
pp | b

=SCORE |p        |
#
1      |{sh.xx,sh.yy,{sh.xx,sh.yy,sh.xx,{sh.zz,sh.pp}},sh.xx} |
3      |*        |

			`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 48 dyn =
[480] channel.NoteOff channel 0 key 48
[0] channel.NoteOn channel 0 key 52 dyn =
[480] channel.NoteOff channel 0 key 52
[0] channel.NoteOn channel 0 key 48 dyn =
[120] channel.NoteOff channel 0 key 48
[0] channel.NoteOn channel 0 key 52 dyn =
[120] channel.NoteOff channel 0 key 52
[0] channel.NoteOn channel 0 key 48 dyn =
[120] channel.NoteOff channel 0 key 48
[0] channel.NoteOn channel 0 key 55 dyn =
[60] channel.NoteOff channel 0 key 55
[0] channel.NoteOn channel 0 key 59 dyn =
[60] channel.NoteOff channel 0 key 59
[0] meta.Undefined type:  0
[0] meta.Undefined type:  0
[0] channel.NoteOn channel 0 key 48 dyn =
[480] channel.NoteOff channel 0 key 48
[0] meta.Undefined type:  0
`,
		},
		{
			// 97
			`
TRACK |p |
Ch      |1 |

sh |
xx | c
yy | e
zz | g
pp | b

=SCORE |p        |
#
1      |{sh.xx,sh.yy,{sh.xx,:,sh.xx,{sh.zz,sh.pp}},sh.xx} |
3      |*        |

			`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 48 dyn =
[480] channel.NoteOff channel 0 key 48
[0] channel.NoteOn channel 0 key 52 dyn =
[480] channel.NoteOff channel 0 key 52
[0] channel.NoteOn channel 0 key 48 dyn =
[240] channel.NoteOff channel 0 key 48
[0] channel.NoteOn channel 0 key 48 dyn =
[120] channel.NoteOff channel 0 key 48
[0] channel.NoteOn channel 0 key 55 dyn =
[60] channel.NoteOff channel 0 key 55
[0] channel.NoteOn channel 0 key 59 dyn =
[60] channel.NoteOff channel 0 key 59
[0] meta.Undefined type:  0
[0] meta.Undefined type:  0
[0] channel.NoteOn channel 0 key 48 dyn =
[480] channel.NoteOff channel 0 key 48
[0] meta.Undefined type:  0
`,
		},
		{
			// 98
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
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 57 dyn =
[0] meta.Lyric: "a"
[0] channel.NoteOff channel 0 key 57
[30] channel.NoteOn channel 0 key 59 dyn =
[0] meta.Lyric: "b"
[930] channel.NoteOff channel 0 key 59
[0] channel.NoteOn channel 0 key 48 dyn =
[0] meta.Lyric: "c"
[2880] channel.NoteOff channel 0 key 48
`,
		},
		{
			// 99
			`
=melody | Verse1 |
 #
  1     | a      |
  1     | b>     |
  2     | c      |

TRACK   | Vocals | Drums |
 Channel| 1      |       |

@lyr  |
 a b c| V1

=SCORE | Vocals         |
 #
  1    | =melody@lyr.V1 |

			`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 57 dyn =
[0] meta.Lyric: "a"
[0] channel.NoteOff channel 0 key 57
[30] channel.NoteOn channel 0 key 59 dyn =
[0] meta.Lyric: "b"
[930] channel.NoteOff channel 0 key 59
[0] channel.NoteOn channel 0 key 48 dyn =
[0] meta.Lyric: "c"
[2880] channel.NoteOff channel 0 key 48
`,
		},
		{
			// 100
			`

TRACK   | Vocals | 
 Channel| 1      |      

=SCORE | Vocals |
 #
  1    | c^2'--     |
  2    | c^-2'++     |

			`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 62 dyn --
[960] channel.NoteOff channel 0 key 62
[0] channel.NoteOn channel 0 key 58 dyn ++
[2880] channel.NoteOff channel 0 key 58
`,
		},
		{
			// 101
			`

TRACK   | Vocals | 
 Channel| 1      |      

=SCORE | Vocals |
 #
  1    | ^1^2--     |
  2    | ^1^-1++     |

			`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 64 dyn --
[960] channel.NoteOff channel 0 key 64
[0] channel.NoteOn channel 0 key 59 dyn ++
[2880] channel.NoteOff channel 0 key 59
`,
		},
		/*
			{ // 47
						`
			=melody | Verse1 |
			 #
			  1     | a      |
			  1     | b      |
			  2     | c      |

			TRACK   | Vocals | Drums |
			 Channel| 1      |       |

			@lyr  |
			 a b c| V1

			=SCORE | Vocals         |
			 #
			  1    | =melody@lyr.V1 |
			`,
						`
			=SCORE | Vocals  |
			 # 4/4 @120.00
			  1    | (a "a") |
			  1    | (b "b") |
			  2    | (c "c") |
			`,
					},
		*/
	}

	for i, test := range tests {

		/*
			if i != 89 {
				continue
			}
		*/
		//sketch.DEBUG = true

		//fmt.Println("##############")

		//fmt.Printf("--        test-%v  -----\n", i)

		sc := score.New(fmt.Sprintf("test-%v", i), nil)
		f := file.FromReader(strings.NewReader(test.input), sc)

		sc.Files[fmt.Sprintf("test-%v", i)] = f

		err := f.Parse()
		if err != nil {
			t.Errorf("[%v] can't parse =SCORE %s", i, err)
			continue
		}

		err = sc.Unroll()
		if err != nil {
			t.Errorf("[%v] can't unroll =SCORE %s", i, err)
			continue
		}

		var lg logger
		lg.wantedTrack = test.track
		lg.includeMeta = test.includeMeta
		lg.includeControlChange = test.includeControlChange
		err = smf.WriteSMFTo(sc, ioutil.Discard, "*", smfwriter.Debug(&lg))

		if err != nil {
			t.Errorf("[%v] can't write SMF: %s", i, err)
			continue
		}

		got := strings.TrimSpace(lg.bf.String())
		expected := strings.TrimSpace(test.expected)

		if got != expected {
			t.Errorf("[%v]\n%s\ngot: \n%s\nexpected: \n%s\n", i, test.input, got, expected)
		}

	}
}

type logger struct {
	bf                   strings.Builder
	wantedTrack          int
	currentTrack         int
	includeMeta          bool
	includeControlChange bool
}

func _velocityToDynamic(vel uint8) (dyn string) {
	switch {
	case vel >= 119:
		return "++++"
	case vel >= 104:
		return "+++"
	case vel >= 89:
		return "++"
	case vel >= 74:
		return "+"
	case vel >= 59:
		return "="
	case vel >= 44:
		return "-"
	case vel >= 29:
		return "--"
	case vel >= 14:
		return "---"
	default:
		return "----"
	}
}

func (l *logger) format(val interface{}) string {
	switch v := val.(type) {
	case channel.NoteOn:
		dyn := _velocityToDynamic(v.Velocity())
		return fmt.Sprintf("channel.NoteOn channel %v key %v dyn %s", v.Channel(), v.Key(), dyn)
	case channel.NoteOff:
		return fmt.Sprintf("channel.NoteOff channel %v key %v", v.Channel(), v.Key())
	case channel.NoteOffVelocity:
		return fmt.Sprintf("channel.NoteOff channel %v key %v", v.Channel(), v.Key())
	default:
		return fmt.Sprintf("%s", val)
	}
}

func (l *logger) Printf(format string, vals ...interface{}) {
	if len(vals) > 1 && vals[1] == meta.EndOfTrack {
		l.currentTrack++
		if l.includeMeta {
			fmt.Fprintf(&l.bf, "[%v] %s\n", vals[0], l.format(vals[1]))
		}
		return
	}

	if len(vals) != 2 {
		return
	}

	if l.currentTrack != l.wantedTrack {
		return
	}

	if _, isMetaUndef := vals[1].(meta.Undefined); isMetaUndef {
		fmt.Fprintf(&l.bf, "[%v] %s\n", vals[0], l.format(vals[1]))
	}

	if _, isLyric := vals[1].(meta.Lyric); isLyric {
		fmt.Fprintf(&l.bf, "[%v] %s\n", vals[0], l.format(vals[1]))
	}

	if _, isMeta := vals[1].(meta.Message); isMeta && !l.includeMeta {
		return
	}

	if _, isCC := vals[1].(channel.ControlChange); isCC && !l.includeControlChange {
		return
	}

	fmt.Fprintf(&l.bf, "[%v] %s\n", vals[0], l.format(vals[1]))

}
