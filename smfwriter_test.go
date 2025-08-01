package muskel_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"gitlab.com/golang-utils/fs/filesystems/mockfs"
	"gitlab.com/golang-utils/fs/path"
	"gitlab.com/gomidi/midi/v2"
	smfmidi "gitlab.com/gomidi/midi/v2/smf"
	"gitlab.com/gomidi/muskel"
	"gitlab.com/gomidi/muskel/file"
	"gitlab.com/gomidi/muskel/items"
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
|TRACK |p  |
|Ch     |1  |

|=SCORE |p  |
|#
|1      |c" |
|2      |d  |

			`,
			1,
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
|TRACK |p |
|Ch      |1 |

|=SCORE |p        |
|#
|1      |{c,e,g#} |
|2      |*        |

			`,
			1,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 48 dyn =
[320] channel.NoteOff channel 0 key 48
[0] channel.NoteOn channel 0 key 52 dyn =
[320] channel.NoteOff channel 0 key 52
[0] channel.NoteOn channel 0 key 56 dyn =
[320] channel.NoteOff channel 0 key 56
`,
		},
		{
			// 2
			`
|TRACK |p  |
|Ch      |1  |

|=SCORE |p |
|#
|1      |c |

			`,
			1,
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
|TRACK |p  |
|Ch      |1  |

|=SCORE |p        |
|#
|1      |{c,e,g#} |

			`,
			1,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 48 dyn =
[1280] channel.NoteOff channel 0 key 48
[0] channel.NoteOn channel 0 key 52 dyn =
[1280] channel.NoteOff channel 0 key 52
[0] channel.NoteOn channel 0 key 56 dyn =
[1280] channel.NoteOff channel 0 key 56
`,
		},
		{
			// 4
			`
|TRACK |p|
|Ch      |1|

|=SCORE |p |
|#
|1      |c |
|#
|1      |  |
|2      |* |

			`,
			1,
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
|TRACK |p|
|Ch      |1|

|=SCORE |p |
|#
|1  |{c,e,g#}|
|#
|1  ||
|2  |*|

			`,
			1,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 48 dyn =
[1600] channel.NoteOff channel 0 key 48
[0] channel.NoteOn channel 0 key 52 dyn =
[1600] channel.NoteOff channel 0 key 52
[0] channel.NoteOn channel 0 key 56 dyn =
[1600] channel.NoteOff channel 0 key 56
`,
		},
		{
			// 6
			`
|TRACK |p|
|Ch      |1|
|Delay|1/32|

|=SCORE |p |
|#
|1  |{c,e,g#} |
|#
|1  ||
|2  |*|

			`,
			1,
			false,
			false,
			`
[120] channel.NoteOn channel 0 key 48 dyn =
[1600] channel.NoteOff channel 0 key 48
[0] channel.NoteOn channel 0 key 52 dyn =
[1600] channel.NoteOff channel 0 key 52
[0] channel.NoteOn channel 0 key 56 dyn =
[1600] channel.NoteOff channel 0 key 56
`,
		},
		{
			// 7
			`
|TRACK |p|
|Ch      |1|
|Delay|-1/32|

|=SCORE |p |
|#
|1  | |
|#
|1  |{c,e,g#} |
|#
|1  ||
|2  |*|

			`,
			1,
			false,
			false,
			`
[3720] channel.NoteOn channel 0 key 48 dyn =
[1600] channel.NoteOff channel 0 key 48
[0] channel.NoteOn channel 0 key 52 dyn =
[1600] channel.NoteOff channel 0 key 52
[0] channel.NoteOn channel 0 key 56 dyn =
[1600] channel.NoteOff channel 0 key 56
`,
		},
		{
			// 8
			`
|TRACK |p|
|Ch      |1|

|=SCORE |p |
|#
|1  |c |
|2  |d |
|3  |e |
|4  |* |

			`,
			1,
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
|TRACK |p|
|Ch      |1|

|=SCORE |p |
|#
|1  |c> |
|2  |d |
|3  |e |
|4  |* |

			`,
			1,
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
|TRACK |p|
|Ch      |1|

|=SCORE |p |
|#
|1  |MN48> |
|2  |MN50 |
|3  |MN52 |
|4  |* |

			`,
			1,
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
|TRACK |p|
|Ch      |1|

|=SCORE |p |
|#
|1  |c |
|2  |d> |
|3  |e |
|4  |* |
			`,
			1,
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
|TRACK |p|
|Ch      |1|

|=SCORE |p |
|#
|1  |MN48 |
|2  |MN50> |
|3  |MN52 |
|4  |* |

			`,
			1,
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
|TRACK |p|
|Ch      |1|

|=SCORE |p |
|#
|1  |c |
|2  |d |
|3  |e> |
|4  |* |

			`,
			1,
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
|TRACK |p|
|Ch      |1|

|=SCORE |p |
|#
|1  |c |
|2  |d< |
|3  |e |
|4  |* |
			`,
			1,
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
|TRACK |p|
|Ch      |1|

|.cc |
|sc  | MN48+

|=SCORE |p |
|#
|1  |.cc.sc |
|2  |MN50< |
|3  |MN52 |
|4  |* |

			`,
			1,
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
|TRACK |p|
|Ch      |1|

|=SCORE |p |
|#
|1  |c |
|2  |d |
|3  |e< |
|4  |* |

			`,
			1,
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
|TRACK |p|
|Ch      |1|

|=SCORE |p |
|#
|1  |c< |
|2  |d |
|3  |e |
|4  |* |

			`,
			1,
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
|TRACK |p        |
|Ch      |1        |

|=SCORE  |p        |
|#
|1       |         |
|#
|1       |{c,e,g#}>|
|#
|1       |         |
|2       |*        |
			`,
			1,
			false,
			false,
			`
[3840] channel.NoteOn channel 0 key 48 dyn =
[1600] channel.NoteOff channel 0 key 48
[0] channel.NoteOn channel 0 key 52 dyn =
[1600] channel.NoteOff channel 0 key 52
[0] channel.NoteOn channel 0 key 56 dyn =
[1600] channel.NoteOff channel 0 key 56
`,
		},
		{
			// 19
			`
|TRACK  |p|
|Ch |1  |

|=SCORE |p|
|#
|1  | |
|#
|1  |{c,e,g#}< |
|#
|1  ||
|2  |*|
			`,
			1,
			false,
			false,
			`
[3840] channel.NoteOn channel 0 key 48 dyn =
[1600] channel.NoteOff channel 0 key 48
[0] channel.NoteOn channel 0 key 52 dyn =
[1600] channel.NoteOff channel 0 key 52
[0] channel.NoteOn channel 0 key 56 dyn =
[1600] channel.NoteOff channel 0 key 56
`,
		},
		{
			// 20
			`
|=templ |p|
|#
|1      |{c,e,g#}|
|2      |*|

|TRACK |p|
|Ch |1  |

|=SCORE |p|
|#
|1  |=templ |
			`,
			1,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 48 dyn =
[320] channel.NoteOff channel 0 key 48
[0] channel.NoteOn channel 0 key 52 dyn =
[320] channel.NoteOff channel 0 key 52
[0] channel.NoteOn channel 0 key 56 dyn =
[320] channel.NoteOff channel 0 key 56
`,
		},
		{
			// 21
			`

|.tt |
|a|c
|b|e
|c|g#

|TRACK |p|
|Ch |1  |

|=SCORE |p|
|#
|1  |{.tt.a,.tt.b,.tt.c}|
|2  |*|
			`,
			1,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 48 dyn =
[320] channel.NoteOff channel 0 key 48
[0] channel.NoteOn channel 0 key 52 dyn =
[320] channel.NoteOff channel 0 key 52
[0] channel.NoteOn channel 0 key 56 dyn =
[320] channel.NoteOff channel 0 key 56
`,
		},
		{
			// 22
			`
|.dr |
|ta | MN48
|tb | MN52
|tc | MN56

|TRACK |p|
|Ch |1  |

|=SCORE |p|
|#
|1  |{.dr.ta,.dr.tb,.dr.tc} |
|2  |*|
			`,
			1,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 48 dyn =
[320] channel.NoteOff channel 0 key 48
[0] channel.NoteOn channel 0 key 52 dyn =
[320] channel.NoteOff channel 0 key 52
[0] channel.NoteOn channel 0 key 56 dyn =
[320] channel.NoteOff channel 0 key 56
`,
		},
		{
			// 23
			`
|.dr |
|ta | MN48
|tb | MN52
|tc | MN56


|TRACK |p|
|Ch |1  |

|=SCORE |p|
|#
|1  |{.dr.ta+,.dr.tb-,.dr.tc--} |
|2  |*|
			`,
			1,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 48 dyn +
[320] channel.NoteOff channel 0 key 48
[0] channel.NoteOn channel 0 key 52 dyn -
[320] channel.NoteOff channel 0 key 52
[0] channel.NoteOn channel 0 key 56 dyn --
[320] channel.NoteOff channel 0 key 56
`,
		},
		{
			// 24
			`
|TRACK |p|
|Ch |1  |

|=SCORE |p|
|#
|1  |c::: |

			`,
			1,
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
|TRACK |p|
|Ch |1  |

|=SCORE |p|
|#
|1  |{c:::,e:::,g#:::} |
|2  |* |

			`,
			1,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 48 dyn =
[60] channel.NoteOff channel 0 key 48
[260] channel.NoteOn channel 0 key 52 dyn =
[60] channel.NoteOff channel 0 key 52
[260] channel.NoteOn channel 0 key 56 dyn =
[60] channel.NoteOff channel 0 key 56
`,
		},
		{
			// 26
			`			
|TRACK |p|
|Ch |1  |

|=SCORE |p|
|#
|1  |{c,e,g#} |
|2  |a::: |
			`,
			1,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 48 dyn =
[320] channel.NoteOff channel 0 key 48
[0] channel.NoteOn channel 0 key 52 dyn =
[320] channel.NoteOff channel 0 key 52
[0] channel.NoteOn channel 0 key 56 dyn =
[320] channel.NoteOff channel 0 key 56
[0] channel.NoteOn channel 0 key 57 dyn =
[60] channel.NoteOff channel 0 key 57
`,
		},
		{
			// 27
			`
|TRACK |p|
|Ch |1  |

|=SCORE |p|
|#
|1  |{c:::,e:::,g#:::} |
|2  |a::: |
			`,
			1,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 48 dyn =
[60] channel.NoteOff channel 0 key 48
[260] channel.NoteOn channel 0 key 52 dyn =
[60] channel.NoteOff channel 0 key 52
[260] channel.NoteOn channel 0 key 56 dyn =
[60] channel.NoteOff channel 0 key 56
[0] channel.NoteOn channel 0 key 57 dyn =
[60] channel.NoteOff channel 0 key 57
`,
		},
		{
			// 28
			`
|TRACK |p|
|Ch |1  |

|=SCORE |p|
|#
|1  | d |
|2  |{c,e,g#} |
|3  |a |
|4 |*|
			`,
			1,
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
[0] channel.NoteOn channel 0 key 57 dyn =
[960] channel.NoteOff channel 0 key 57
`,
		},
		{
			// 29
			`
|TRACK |p|
|Ch |1  |

|=SCORE |p|
|#
|1  | d |
|4  |{c,e,g#} |
|#
|1  |a |
|2 |*|
			`,
			1,
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
[0] channel.NoteOn channel 0 key 57 dyn =
[960] channel.NoteOff channel 0 key 57
`,
		},
		{
			// 30
			`
|TRACK |p|
|Ch |1  |

|=SCORE |p|
|#
|1  | d |

`,
			1,
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
|TRACK |p|
|Ch |1  |

|=SCORE |p|
|#
|1  | d |

`,
			1,
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
|TRACK |p|
|Ch |1  |

|=SCORE |p|
|#
|1  |c |
|#
|1  |* |
|#
|1  |e |

			`,
			1,
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
|TRACK |p|
|Ch |1  |

|=SCORE |p|
|#
|1  |c |
|#
|1  | * |
|#
|1  |e |

			`,
			1,
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
|TRACK |p|
|Ch |1  |

|=SCORE |p|
|#
|1  |c |
|#
|1  | |
|#
|1  |e |

			`,
			1,
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
|TRACK |p|
|Ch |1  |

|=SCORE |p|
|#
|1  |a' |
|#
|1  |* |
|#
|1  |b' |
			`,
			1,
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
|TRACK |p|
|Ch |1  |

|=SCORE |p|
|#
|1  |a' |
|#
|1  | * |
|#
|1  |b' |

			`,
			1,
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
|TRACK |p|
|Ch |1  |

|=SCORE |p|
|#
|1  | (a' b') |
|#
|1  | * |
|#
|1  | (b' c") |
			`,
			1,
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
|TRACK |p|
|Ch |1  |

|=SCORE |p|
|#
|1  | (a' b') |
|#
|1  |  |
|#
|1  | (b' c") |
			`,
			1,
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
|TRACK |p|
|Ch |1  |

|=SCORE |p|
|#
|1  | (a' b'_) |
|#
|1  |  |
|#
|1  | (_b' c") |
			`,
			1,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 69 dyn =
[0] channel.NoteOn channel 0 key 71 dyn =
[7680] channel.NoteOff channel 0 key 71
[0] channel.NoteOff channel 0 key 69
[0] channel.NoteOn channel 0 key 72 dyn =
[3840] channel.NoteOff channel 0 key 72
`,
		},
		{
			// 40
			`
|TRACK |p|
|Ch |1  |

|=SCORE |p|
|#
|1  |a' |
|1& |b_|
|#
|1  |  |
|1& |_b|
|#
|1  |c" |
			`,
			1,
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
|TRACK |p|
|Ch |1  |

|=SCORE |p|
|#
|1  |a' |
|1& |b'_|
|#
|1  |  |
|1& | (_b' :)|
|#
|1  |c" |
			`,
			1,
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
|TRACK |p|
|Ch |1  |

|=SCORE |p|
|#
|1  | (a' b'_) |
|#
|1  | _b' |
			`,
			1,
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
|TRACK |p|
|Ch |1  |

|=SCORE |p|
|#
|1  | (a' b'_ c"_) |
|#
|1  | (_b' _c") |
			`,
			1,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 69 dyn =
[0] channel.NoteOn channel 0 key 71 dyn =
[0] channel.NoteOn channel 0 key 72 dyn =
[3840] channel.NoteOff channel 0 key 71
[0] channel.NoteOff channel 0 key 72
[3840] channel.NoteOff channel 0 key 69
`,
		},
		{
			// 44
			`
|TRACK |p|
|Ch |1  |

|=SCORE |p|
|#
|2  | (a' "hiho") |
			`,
			1,
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
|TRACK |p|
|Ch |1  |

|=SCORE |p|
|#
|2  |"hiho" |
			`,
			1,
			false,
			false,
			`
[960] meta.Lyric: "hiho"
`,
		},
		{
			// 46
			`
|TRACK |p|
|Ch |1  |

|=SCORE |p|
|#
|1  |c:: |
			`,
			1,
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
|TRACK |p|
|Ch |1  |

|=SCORE |p|
|#
|1  |c: |
			`,
			1,
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
|TRACK |p|
|Ch |1  |

|=SCORE |p|
|#
|1  |AT(23):: |
			`,
			1,
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
|TRACK |p|
|Ch |1  |

|=SCORE |p|
|#
|1  |AT(23): |
			`,
			1,
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
|=cc  |x|
|#
|1    |c<|
|2    |* |

|TRACK |p|
|Ch |1  |

|=SCORE |p|
|#
|2  |=cc |
|3  |*  |
			`,
			1,
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
|=cc  |x|
|#
|1    |c<|
|2    |* |

|TRACK |p|
|Ch |1  |

|=SCORE |p|
|#
|2  |=cc> |
|3  |*  |
			`,
			1,
			false,
			false,
			`
[960] channel.NoteOn channel 0 key 48 dyn =
[960] channel.NoteOff channel 0 key 48
`,
		},
		{
			// 52
			`
|=cc  |x|
|#
|1    |c|
|2    |* |

|TRACK |p|
|Ch |1  |

|=SCORE |p|
|#
|2  |=cc< |
|3  |*  |
			`,
			1,
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
|TRACK |p|
|Ch |1  |

|=SCORE |p|
|#
|1  | d+ |`,
			1,
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
|TRACK |p|
|Ch |1  |

|=SCORE |p|
|#
|1  | d++ |`,
			1,
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
|TRACK |p|
|Ch |1  |

|=SCORE |p|
|#
|1  | d+++ |`,
			1,
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
|TRACK |p|
|Ch |1  |

|=SCORE |p|
|#
|1  | d- |`,
			1,
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
|TRACK |p|
|Ch |1  |

|=SCORE |p|
|#
|1  | d-- |`,
			1,
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
|TRACK |p|
|Ch |1  |
|VelScale|step:30|

|=SCORE |p|
|#
|1  | d+ |`,
			1,
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
|TRACK |p|
|Ch |1  |
|VelScale|step:30|

|=SCORE |p|
|#
|1  | d++ |

`,
			1,
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
|TRACK |p|
|Ch |1  |
|VelScale|step:30|

|=SCORE |p|
|#
|1  | d- |
`,
			1,
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
|TRACK |p|
|Ch |1  |
|VelScale|step:30|

|=SCORE |p|
|#
|1  | d-- |
`,
			1,
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
|TRACK |p|
|Ch |1  |
|VelScale|min:30|

|=SCORE |p|
|#
|1  | d---- |
`,
			1,
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
|TRACK |p|
|Ch |1  |
|VelScale|max:100|

|=SCORE |p|
|#
|1  | d++++ |
`,
			1,
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
|TRACK |p|
|Ch |1  |
|VelScale|center:78|

|=SCORE |p|
|#
|1  | d |
`,
			1,
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
|TRACK   | git | Vocals |
|Ch      | 2   | 1      |
|PbRange | 2   | 2      |

|=SCORE  | git      | Vocals |
|#
|1       | =.Vocals |        |
|#
|1       |       | f++       |   
|2       |        |    g-    |
|#
|1       |        |      c'  |
`,
			2,
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
|TRACK |Vocals|git|
|Ch      |1       |2    |
|PbRange |2       |2    |

|=SCORE  |Vocals|git|
|#
|1   |        | =.Vocals|
|    |        | /d   |
|#
|1   | f++    |     |   
|2   | g-     |     |
|#
|1   |        | /c   |
|2   |        | /PB(20) |
`,
			2,
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
|TRACK |p|
|Ch |1  |

|=SCORE |p|
|#A
|1  |c" |
|2  |d  |
|[A]
			`,
			1,
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
|=rhythm | x |
|#
|1       | $euclid(3,8,&) |

|TRACK |p|
|Ch |1  |

|=SCORE |p|
|#
|1  | =rhythm(c"+,d) |
|#
|1  |*|
			`,
			1,
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
|=rhythm |x|
|#
|1 | $euclid(3,8,1) |
|*2

|TRACK |p|
|Ch |1  |

|=SCORE |p|
|#
|1  |=rhythm(c"+,d) |
|#
|1 | |
|#
|1  |*|
			`,
			1,
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
|TRACK |p|
|Ch |1  |
|VelScale|center:78|

|=SCORE |p|
|#
|1  | d |
|2  | % |
|3  | * |
`,
			1,
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
|TRACK |p|
|Ch |1  |
|VelScale|center:78|

|=SCORE |p|
|#
|1  | d |
|2  | * |
|3  | % |
|4  | * |
`,
			1,
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
|TRACK |p|
|Ch |1  |
|VelScale|center:78|

|=SCORE |p|
|#
|1  | d |
|3  | e |
|4  | * |
|#
|1  | ./. |
`,
			1,
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
|TRACK |p|
|Ch |1  |
|VelScale|center:78|

|=SCORE |p|
|#
|2  | g |
|#
|1  | d |
|3  | e |
|4  | * |
|#
|1  | .2. |
|*3
|#
|1  | c |
|2  | * |
`,
			1,
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
|TRACK |p|
|Ch |1  |
|VelScale|center:78|

|=SCORE |p|
|#
|2  | g |
|#
|1  | d |
|3  | e |
|4  | * |
|#
|1  | .2. |
|*5
|#
|1  | c |
|2  | * |
`,
			1,
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
|TRACK |p|
|Ch |1  |
|VelScale|center:78|

|=SCORE |p|
|#
|2  | g |
|#
|1  | d |
|3  | e |
|4  | * |
|#
|1  | .2. |
|*4
|#
|1  | c |
|2  | * |
`,
			1,
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
|TRACK |p|
|Ch |1  |
|VelScale|center:78|

|=SCORE |p|
|#
|2  | g |
|#
|1  | d |
|3  | e |
|4  | * |
|#
|1  | .2. |
|*5
`,
			1,
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
|TRACK |p|
|Ch |1  |
|VelScale|center:78|

|=patt |a|
|#
|2  | g |
|#
|1  | d |
|3  | e |
|4  | * |
|#
|1  | .2. |
|#
|1  |     |

|=SCORE |p|
|#
|1  | =patt.a%2 |
|*7
`,
			1,
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
|TRACK |p|
|Ch |1  |

|=SCORE |p|
|#A
|1  |c" |
|2  |d  |
|[A]
|[A]
			`,
			1,
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
|TRACK |p|
|Ch |1  |

|=SCORE |p|
|#A
|1  |c" |
|2  |d  |
|[A]
|#
|1  | c' |
|2 |* |
|[A]
			`,
			1,
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
|TRACK |p|
|Ch |1  |

|=SCORE |p|
|#A
|1  |c" |
|3  |*|
|#B
|1  |f"  |
|3 |*|
|[A]
|#
|1  | c' |
|3 |* |
|[A]
|[B]
			`,
			1,
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
|TRACK |p|
|Ch |1  |

|=temp |a|
|#A
|1  |c" | 
|3  |*|
|#B
|1  |f"  | 
|3 |*|
|[A]
|#
|1  | c' |
|3 |* |
|[A]
|[B]

|=SCORE |p|
|#
|1    |=temp.a|
|*6
			`,
			1,
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
|TRACK  | Voice |
|Channel | 3     |

|@voice  |        
|Hi      | V1 
|ho-ha   |        
|hu      |

|=melody | Verse |
|#A
|1       | c"    |
|3       | *     |
|#B
|1       | f"    | 
|4       | *     |
|[A]
|[B]

|=SCORE  | Voice                     |
|# 4/4 @120
|1       | =melody.Verse@voice.V1[0:2] |
|*3

			`,
			1,
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
|TRACK  | Voice |
|Channel | 3     |

|@voice   |        
|Hi      | V1 
|ho-ha   |        
|hu      |

|=melody | Verse |
|#A
|1       | c"    | 
|3       | *     |
|#B
|1       | f"    | 
|4       | *     |
|[A]
|[B]

|=SCORE  | Voice                     |
|# 4/4 @120
|1       | =melody.Verse[2:]@voice.V1[0:2] |
|*3

			`,
			1,
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
|TRACK  | Voice |
|Channel | 3     |

|=SCORE  | Voice                     |
|# 4/4 @120
|1       | e_ |
|1       | f  |
|3       | *  |
|4       | _e |

			`,
			1,
			false,
			false,
			`
[0] channel.NoteOn channel 2 key 53 dyn =
[0] channel.NoteOn channel 2 key 52 dyn =
[1920] channel.NoteOff channel 2 key 53
[960] channel.NoteOff channel 2 key 52
`,
		},
		{
			// 85
			`
|TRACK  | Voice |
|Channel | 3     |

|=SCORE  | Voice                     |
|# 4/4 @120
|1       | e_ |
|        | f  |
|3       | *  |
|4       | _e |

			`,
			1,
			false,
			false,
			`
[0] channel.NoteOn channel 2 key 53 dyn =
[0] channel.NoteOn channel 2 key 52 dyn =
[1920] channel.NoteOff channel 2 key 53
[960] channel.NoteOff channel 2 key 52
`,
		},
		{
			// 86
			`
|TRACK  | Voice |
|Channel | 3     |

|=SCORE  | Voice                     |
|# 4/4 @120
|1       | 1 |
|2       | 2  |
|3       | * |

			`,
			1,
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
|TRACK  | Voice |
|Channel | 3     |

|=SCORE  | Voice                     |
|# 4/4 @120 \major^d#'
|1       | 1 |
|2       | 2  |
|3       | * |

			`,
			1,
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
|TRACK |p  |
|Ch     |1  |

|=SCORE |p  |
|#
|1      |c" |
|1&     |:  |
|2      |d  |

			`,
			1,
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
|TRACK |p  |
|Ch     |1  |

|=patt | |
|#
|1      |c" |
|2      |d  |

|=SCORE |p  |
|#
|1      |=patt |
|2      |/:  |
|3      |d  |

			`,
			1,
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
|TRACK |p  |
|Ch     |1  |

|=SCORE |p  |
|#
|1      |c"+- |
|2      |1----+++  |
|3      |d+++-  |

			`,
			1,
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
|TRACK |p |
|Ch      |1 |

|=SCORE |p        |
|#
|1      |{c,e,g#,a} |
|3      |*        |

			`,
			1,
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
`,
		},
		{
			// 92
			`
|TRACK |p |
|Ch      |1 |

|=SCORE |p        |
|#
|1      |{c,e,g#,{a,b}} |
|3      |*        |

			`,
			1,
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
`,
		},
		{
			// 93
			`
|TRACK |p |
|Ch      |1 |

|=SCORE |p        |
|#
|1      |{c,e,{a,b},g#} |
|3      |*        |

			`,
			1,
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
[0] channel.NoteOn channel 0 key 56 dyn =
[480] channel.NoteOff channel 0 key 56
`,
		},
		{
			// 94
			`
|TRACK |p |
|Ch      |1 |

|=SCORE |p        |
|#
|1      |{c,e,{a,b,c#,d},g#} |
|3      |*        |

			`,
			1,
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
[0] channel.NoteOn channel 0 key 56 dyn =
[480] channel.NoteOff channel 0 key 56
`,
		},
		{
			// 95
			`
|TRACK |p |
|Ch      |1 |

|=SCORE |p        |
|#
|1      |{c,e,{c,e,c,e},c} |
|3      |*        |

			`,
			1,
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
[0] channel.NoteOn channel 0 key 48 dyn =
[480] channel.NoteOff channel 0 key 48
`,
		},
		{
			// 96
			`
|TRACK |p |
|Ch      |1 |

|=SCORE |p        |
|#
|1      |{c,e,{c,e,c,{g,b}},c} |
|3      |*        |

			`,
			1,
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
[0] channel.NoteOn channel 0 key 48 dyn =
[480] channel.NoteOff channel 0 key 48
`,
		},
		{
			// 97
			`
|TRACK |p |
|Ch      |1 |

|.sh |
|xx | c
|yy | e
|zz | g
|pp | b

|=SCORE |p        |
|#
|1      |{.sh.xx,.sh.yy,{.sh.xx,.sh.yy,.sh.xx,{.sh.zz,.sh.pp}},.sh.xx} |
|3      |*        |

			`,
			1,
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
[0] channel.NoteOn channel 0 key 48 dyn =
[480] channel.NoteOff channel 0 key 48
`,
		},
		{
			// 98
			`
|TRACK |p |
|Ch      |1 |

|.sh |
|xx | c
|yy | e
|zz | g
|pp | b

|=SCORE |p        |
|#
|1      |{.sh.xx,.sh.yy,{.sh.xx,:,.sh.xx,{.sh.zz,.sh.pp}},.sh.xx} |
|3      |*        |

			`,
			1,
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
[0] channel.NoteOn channel 0 key 48 dyn =
[480] channel.NoteOff channel 0 key 48
`,
		},
		{
			// 99
			`
|=melody | Verse1 |
| #
|  1     | a      |
|  1     | b>     |
|  2     | c      |

|TRACK   | Vocals |
| Channel| 1      |

|@lyr  |
| a b c| V1

|=SCORE | Vocals         |
| #
|  1    | =melody@lyr.V1 |

			`,
			1,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 57 dyn =
[0] meta.Lyric: "a"
[30] channel.NoteOff channel 0 key 57
[0] channel.NoteOn channel 0 key 59 dyn =
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
|=melody | Verse1 |
| #
|  1     | a      |
|  1     | b>     |
|  2     | c      |

|TRACK   | Vocals | Drums |
| Channel| 1      |       |

|@lyr  |
| a b c| V1

|=SCORE | Vocals         |
| #
|  1    | =melody@lyr.V1 |

			`,
			1,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 57 dyn =
[0] meta.Lyric: "a"
[30] channel.NoteOff channel 0 key 57
[0] channel.NoteOn channel 0 key 59 dyn =
[0] meta.Lyric: "b"
[930] channel.NoteOff channel 0 key 59
[0] channel.NoteOn channel 0 key 48 dyn =
[0] meta.Lyric: "c"
[2880] channel.NoteOff channel 0 key 48
`,
		},
		{
			// 101
			`

|TRACK   | Vocals | 
| Channel| 1      |      

|=SCORE | Vocals |
| #
|  1    | 1^1--     |
|  2    | 1^-1++     |

			`,
			1,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 62 dyn --
[960] channel.NoteOff channel 0 key 62
[0] channel.NoteOn channel 0 key 59 dyn ++
[2880] channel.NoteOff channel 0 key 59
`,
		},
		{
			// 102
			`

|TRACK   | Vocals | 
| Channel| 1      |      

|=SCORE | Vocals |
| #
|  1    | 1^2--     |
|  2    | 1^-1++     |

			`,
			1,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 64 dyn --
[960] channel.NoteOff channel 0 key 64
[0] channel.NoteOn channel 0 key 59 dyn ++
[2880] channel.NoteOff channel 0 key 59
`,
		},
		{
			// 103
			`

|TRACK   | Vocals | 
| Channel| 1      |      

|=SCORE | Vocals |
| #
|  1    | PC(2)  |
|  2    | PC(3)  |

			`,
			1,
			false,
			false,
			`
[0] channel.ProgramChange channel 0 program 1
[960] channel.ProgramChange channel 0 program 2
`,
		},
		{
			// 104
			`
'cc

|TRACK   | Vocals | 
| Channel| 1      |      

|=SCORE | Vocals |
| #
|  2    | =!reset(2,3) |

			`,
			1,
			false,
			true,
			`
[0] channel.ControlChange channel 0 controller 101 ("Registered Parameter (MSB)") value 0
[0] channel.ControlChange channel 0 controller 100 ("Registered Parameter (LSB)") value 0
[0] channel.ControlChange channel 0 controller 6 ("Data Entry (MSB)") value 2
[0] channel.ControlChange channel 0 controller 38 ("Data Entry (LSB)") value 0
[0] channel.ControlChange channel 0 controller 101 ("Registered Parameter (MSB)") value 127
[0] channel.ControlChange channel 0 controller 100 ("Registered Parameter (LSB)") value 127
[960] channel.ControlChange channel 0 controller 0 ("Bank Select (MSB)") value 2
[240] channel.ProgramChange channel 0 program 2
[120] channel.ControlChange channel 0 controller 121 ("All Controllers Off") value 0
[120] channel.ControlChange channel 0 controller 7 ("Volume (MSB)") value 100
[0] channel.ControlChange channel 0 controller 11 ("Expression (MSB)") value 127
[0] channel.ControlChange channel 0 controller 64 ("Hold Pedal (on/off)") value 0
[0] channel.ControlChange channel 0 controller 10 ("Pan position (MSB)") value 64
[240] channel.ControlChange channel 0 controller 101 ("Registered Parameter (MSB)") value 0
[120] channel.ControlChange channel 0 controller 100 ("Registered Parameter (LSB)") value 0
[120] channel.ControlChange channel 0 controller 6 ("Data Entry (MSB)") value 2
[120] channel.ControlChange channel 0 controller 38 ("Data Entry (LSB)") value 0
[120] channel.ControlChange channel 0 controller 101 ("Registered Parameter (MSB)") value 127
[120] channel.ControlChange channel 0 controller 100 ("Registered Parameter (LSB)") value 127
`,
		},
		{
			// 105
			`
|TRACK |p|
|Ch      |1|

|=SCORE |p |
|#
|1  |(MN46 MN48>) |
|2  |MN50 |
|3  |MN52 |
|4  |* |

			`,
			1,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 46 dyn =
[30] channel.NoteOn channel 0 key 48 dyn =
[930] channel.NoteOff channel 0 key 46
[0] channel.NoteOff channel 0 key 48
[0] channel.NoteOn channel 0 key 50 dyn =
[960] channel.NoteOff channel 0 key 50
[0] channel.NoteOn channel 0 key 52 dyn =
[960] channel.NoteOff channel 0 key 52
`,
		},
		{
			// 106
			`
|TRACK |p|
|Ch      |1|

|=SCORE |p |
|#
|1  |(MN48> MN46) |
|2  |MN50 |
|3  |MN52 |
|4  |* |

			`,
			1,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 46 dyn =
[30] channel.NoteOn channel 0 key 48 dyn =
[930] channel.NoteOff channel 0 key 46
[0] channel.NoteOff channel 0 key 48
[0] channel.NoteOn channel 0 key 50 dyn =
[960] channel.NoteOff channel 0 key 50
[0] channel.NoteOn channel 0 key 52 dyn =
[960] channel.NoteOff channel 0 key 52
`,
		},
		{
			// 107
			`
|TRACK |p|
|Ch      |1|

|=SCORE |p |
|#
|1  |{MN48>,MN46} |
|2  |MN50 |
|3  |MN52 |
|4  |* |

			`,
			1,
			false,
			false,
			`
[30] channel.NoteOn channel 0 key 48 dyn =
[450] channel.NoteOff channel 0 key 48
[0] channel.NoteOn channel 0 key 46 dyn =
[480] channel.NoteOff channel 0 key 46
[0] channel.NoteOn channel 0 key 50 dyn =
[960] channel.NoteOff channel 0 key 50
[0] channel.NoteOn channel 0 key 52 dyn =
[960] channel.NoteOff channel 0 key 52
`,
		},
		{
			// 108
			`
|TRACK |p|
|Ch      |1|

|=SCORE |p |
|#
|1 | |
|#
|1  |{MN48,MN46<} |
|2  |MN50 |
|3  |MN52 |
|4  |* |

			`,
			1,
			false,
			false,
			`
[3840] channel.NoteOn channel 0 key 48 dyn =
[450] channel.NoteOff channel 0 key 48
[0] channel.NoteOn channel 0 key 46 dyn =
[510] channel.NoteOff channel 0 key 46
[0] channel.NoteOn channel 0 key 50 dyn =
[960] channel.NoteOff channel 0 key 50
[0] channel.NoteOn channel 0 key 52 dyn =
[960] channel.NoteOff channel 0 key 52
`,
		},
		{
			// 109
			`
|TRACK |p|
|Ch      |1|

|=SCORE |p |
|#
|1  |CC(35,10)~ |
|2  |CC(35,20) |
|3  |MN52 |
|4  |* |

			`,
			1,
			false,
			true,
			`
[0] channel.ControlChange channel 0 controller 101 ("Registered Parameter (MSB)") value 0
[0] channel.ControlChange channel 0 controller 100 ("Registered Parameter (LSB)") value 0
[0] channel.ControlChange channel 0 controller 6 ("Data Entry (MSB)") value 2
[0] channel.ControlChange channel 0 controller 38 ("Data Entry (LSB)") value 0
[0] channel.ControlChange channel 0 controller 101 ("Registered Parameter (MSB)") value 127
[0] channel.ControlChange channel 0 controller 100 ("Registered Parameter (LSB)") value 127
[0] channel.ControlChange channel 0 controller 35 value 10
[60] channel.ControlChange channel 0 controller 35 value 11
[60] channel.ControlChange channel 0 controller 35 value 11
[60] channel.ControlChange channel 0 controller 35 value 12
[60] channel.ControlChange channel 0 controller 35 value 13
[60] channel.ControlChange channel 0 controller 35 value 13
[60] channel.ControlChange channel 0 controller 35 value 14
[60] channel.ControlChange channel 0 controller 35 value 14
[60] channel.ControlChange channel 0 controller 35 value 15
[60] channel.ControlChange channel 0 controller 35 value 16
[60] channel.ControlChange channel 0 controller 35 value 16
[60] channel.ControlChange channel 0 controller 35 value 17
[60] channel.ControlChange channel 0 controller 35 value 18
[60] channel.ControlChange channel 0 controller 35 value 18
[60] channel.ControlChange channel 0 controller 35 value 19
[60] channel.ControlChange channel 0 controller 35 value 19
[60] channel.ControlChange channel 0 controller 35 value 20
[0] channel.ControlChange channel 0 controller 35 value 20
[960] channel.NoteOn channel 0 key 52 dyn =
[960] channel.NoteOff channel 0 key 52
`,
		},
		{
			// 110
			`
|TRACK |p|
|Ch      |1|

|=SCORE |p |
|#
|1  |c~ |
|2  |d |
|3  |MN52 |
|4  |* |

			`,
			1,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 48 dyn =
[60] channel.Pitchbend channel 0 value 512 absValue 0
[60] channel.Pitchbend channel 0 value 1024 absValue 0
[60] channel.Pitchbend channel 0 value 1536 absValue 0
[60] channel.Pitchbend channel 0 value 2048 absValue 0
[60] channel.Pitchbend channel 0 value 2560 absValue 0
[60] channel.Pitchbend channel 0 value 3072 absValue 0
[60] channel.Pitchbend channel 0 value 3584 absValue 0
[60] channel.Pitchbend channel 0 value 4096 absValue 0
[60] channel.Pitchbend channel 0 value 4608 absValue 0
[60] channel.Pitchbend channel 0 value 5120 absValue 0
[60] channel.Pitchbend channel 0 value 5632 absValue 0
[60] channel.Pitchbend channel 0 value 6144 absValue 0
[60] channel.Pitchbend channel 0 value 6656 absValue 0
[60] channel.Pitchbend channel 0 value 7168 absValue 0
[60] channel.Pitchbend channel 0 value 7680 absValue 0
[60] channel.Pitchbend channel 0 value 8191 absValue 0
[960] channel.NoteOff channel 0 key 48
[0] channel.Pitchbend channel 0 value 0 absValue 0
[0] channel.NoteOn channel 0 key 52 dyn =
[960] channel.NoteOff channel 0 key 52
`,
		},
		{
			// 111
			`
|TRACK |p|
|Ch      |1|

|=SCORE |p |
|#
|    1    | d#+++                        |
|    1&,  | (d#+++ g+ a#++ d#'+++)  |
|    1&,; | *                             |
|    2&   | (d#+++ g+ a#++ d#'+++)  |
|    3;   | *                             |
|    3&   |                               |
|    3&,  | d+++                         |
|    4&   | (d+++ f+ b++ d'+++)     |
|    4&;  | *                             |
			`,
			1,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 51 dyn +++
[720] channel.NoteOff channel 0 key 51
[0] channel.NoteOn channel 0 key 51 dyn +++
[0] channel.NoteOn channel 0 key 55 dyn +
[0] channel.NoteOn channel 0 key 58 dyn ++
[0] channel.NoteOn channel 0 key 63 dyn +++
[120] channel.NoteOff channel 0 key 51
[0] channel.NoteOff channel 0 key 55
[0] channel.NoteOff channel 0 key 58
[0] channel.NoteOff channel 0 key 63
[600] channel.NoteOn channel 0 key 51 dyn +++
[0] channel.NoteOn channel 0 key 55 dyn +
[0] channel.NoteOn channel 0 key 58 dyn ++
[0] channel.NoteOn channel 0 key 63 dyn +++
[600] channel.NoteOff channel 0 key 51
[0] channel.NoteOff channel 0 key 55
[0] channel.NoteOff channel 0 key 58
[0] channel.NoteOff channel 0 key 63
[600] channel.NoteOn channel 0 key 50 dyn +++
[720] channel.NoteOff channel 0 key 50
[0] channel.NoteOn channel 0 key 50 dyn +++
[0] channel.NoteOn channel 0 key 53 dyn +
[0] channel.NoteOn channel 0 key 59 dyn ++
[0] channel.NoteOn channel 0 key 62 dyn +++
[120] channel.NoteOff channel 0 key 50
[0] channel.NoteOff channel 0 key 53
[0] channel.NoteOff channel 0 key 59
[0] channel.NoteOff channel 0 key 62
`,
		},
		{
			// 112
			`
|TRACK |p|
|Ch      |1|

|=SCORE |p |
|#
|    1    | d#+++                    |
|    1&,  | (d#+++ g+< a#++< d#'+++) |
|    1&,; | *                        |
|    2&   | (d#+++ g+< a#++< d#'+++) |
|    3;   | *                        |
|    3&,  | d+++                     |
|    4&   | (d+++ f+< b++ d'+++)    |
|    4&;  | *                        |
			`,
			1,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 51 dyn +++
[690] channel.NoteOff channel 0 key 51
[0] channel.NoteOn channel 0 key 55 dyn +
[0] channel.NoteOn channel 0 key 58 dyn ++
[30] channel.NoteOn channel 0 key 51 dyn +++
[0] channel.NoteOn channel 0 key 63 dyn +++
[120] channel.NoteOff channel 0 key 51
[0] channel.NoteOff channel 0 key 55
[0] channel.NoteOff channel 0 key 58
[0] channel.NoteOff channel 0 key 63
[570] channel.NoteOn channel 0 key 55 dyn +
[0] channel.NoteOn channel 0 key 58 dyn ++
[30] channel.NoteOn channel 0 key 51 dyn +++
[0] channel.NoteOn channel 0 key 63 dyn +++
[600] channel.NoteOff channel 0 key 51
[0] channel.NoteOff channel 0 key 55
[0] channel.NoteOff channel 0 key 58
[0] channel.NoteOff channel 0 key 63
[600] channel.NoteOn channel 0 key 50 dyn +++
[690] channel.NoteOff channel 0 key 50
[0] channel.NoteOn channel 0 key 53 dyn +
[30] channel.NoteOn channel 0 key 50 dyn +++
[0] channel.NoteOn channel 0 key 59 dyn ++
[0] channel.NoteOn channel 0 key 62 dyn +++
[120] channel.NoteOff channel 0 key 50
[0] channel.NoteOff channel 0 key 53
[0] channel.NoteOff channel 0 key 59
[0] channel.NoteOff channel 0 key 62
`,
		},
		{
			// 113
			`
|TRACK |p|
|Ch      |1|

|=SCORE |p |
|# @110
|    1    | d#                    |
|# @130
|  1 | e |
			`,
			0,
			true,
			false,
			`
[0] meta.Copyright: ""
[0] meta.TrackSequenceName: "BPM"
[0] meta.Tempo BPM: 110.00
[3840] meta.Tempo BPM: 130.00
`,
		},
		{
			// 114
			`
|TRACK |p|
|Ch      |1|

|=SCORE |p |
|# @110
|    1    | d#                    |
|    2 @130 | |
|# @140
|  1 | e |
			`,
			0,
			true,
			false,
			`
[0] meta.Copyright: ""
[0] meta.TrackSequenceName: "BPM"
[0] meta.Tempo BPM: 110.00
[960] meta.Tempo BPM: 130.00
[2880] meta.Tempo BPM: 140.00
`,
		},
		{
			// 115
			`
|TRACK |p|
|Ch      |1|

|=SCORE |p |
|# @110
|    1    | d#                    |
|    2 @130 | |
|    3 @135 ||
|# @140
|  1 | e |
			`,
			0,
			true,
			false,
			`
[0] meta.Copyright: ""
[0] meta.TrackSequenceName: "BPM"
[0] meta.Tempo BPM: 110.00
[960] meta.Tempo BPM: 130.00
[960] meta.Tempo BPM: 135.00
[1920] meta.Tempo BPM: 140.00
`,
		},
		{
			// 116
			`
|TRACK |p|
|Ch      |1|

|=SCORE |p |
|#
|    1    | c                    |
|    1;   | d                    |
|    1,    | e                    |
|    1,;    | f                    |
|    1&    | g                    |
|   1&;    | a                    |
|    1&,    | b                    |
|    1&,;    | c'                    |
|     2    | c"                    |
|    2;   | d"                    |
|    2,    | e"                    |
|    2,;    | f"                    |
|    2&    | g"                    |
|    2&;    | a"                    |
|    2&,    | b"                    |
|    2&,;    | c"'                    |
|    3  | * |

			`,
			1,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 48 dyn =
[120] channel.NoteOff channel 0 key 48
[0] channel.NoteOn channel 0 key 50 dyn =
[120] channel.NoteOff channel 0 key 50
[0] channel.NoteOn channel 0 key 52 dyn =
[120] channel.NoteOff channel 0 key 52
[0] channel.NoteOn channel 0 key 53 dyn =
[120] channel.NoteOff channel 0 key 53
[0] channel.NoteOn channel 0 key 55 dyn =
[120] channel.NoteOff channel 0 key 55
[0] channel.NoteOn channel 0 key 57 dyn =
[120] channel.NoteOff channel 0 key 57
[0] channel.NoteOn channel 0 key 59 dyn =
[120] channel.NoteOff channel 0 key 59
[0] channel.NoteOn channel 0 key 60 dyn =
[120] channel.NoteOff channel 0 key 60
[0] channel.NoteOn channel 0 key 72 dyn =
[120] channel.NoteOff channel 0 key 72
[0] channel.NoteOn channel 0 key 74 dyn =
[120] channel.NoteOff channel 0 key 74
[0] channel.NoteOn channel 0 key 76 dyn =
[120] channel.NoteOff channel 0 key 76
[0] channel.NoteOn channel 0 key 77 dyn =
[120] channel.NoteOff channel 0 key 77
[0] channel.NoteOn channel 0 key 79 dyn =
[120] channel.NoteOff channel 0 key 79
[0] channel.NoteOn channel 0 key 81 dyn =
[120] channel.NoteOff channel 0 key 81
[0] channel.NoteOn channel 0 key 83 dyn =
[120] channel.NoteOff channel 0 key 83
[0] channel.NoteOn channel 0 key 84 dyn =
[120] channel.NoteOff channel 0 key 84
`,
		},
		{
			// 117
			`
|TRACK |p|
|Ch      |1|

|=SCORE |p |
|#
|    1    | c                    |
|    1.125   | d                    |
|    1.25    | e                    |
|    1.375    | f                    |
|    1.5    | g                    |
|    1.625    | a                    |
|    1.75    | b                    |
|    1.875    | c'                    |
|     2    | c"                    |
|    2.125   | d"                    |
|    2.25    | e"                    |
|    2.375   | f"                    |
|    2.5    | g"                    |
|    2.625    | a"                    |
|    2.75    | b"                    |
|    2.875    | c"'                    |
|    3  | * |

			`,
			1,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 48 dyn =
[120] channel.NoteOff channel 0 key 48
[0] channel.NoteOn channel 0 key 50 dyn =
[120] channel.NoteOff channel 0 key 50
[0] channel.NoteOn channel 0 key 52 dyn =
[120] channel.NoteOff channel 0 key 52
[0] channel.NoteOn channel 0 key 53 dyn =
[120] channel.NoteOff channel 0 key 53
[0] channel.NoteOn channel 0 key 55 dyn =
[120] channel.NoteOff channel 0 key 55
[0] channel.NoteOn channel 0 key 57 dyn =
[120] channel.NoteOff channel 0 key 57
[0] channel.NoteOn channel 0 key 59 dyn =
[120] channel.NoteOff channel 0 key 59
[0] channel.NoteOn channel 0 key 60 dyn =
[120] channel.NoteOff channel 0 key 60
[0] channel.NoteOn channel 0 key 72 dyn =
[120] channel.NoteOff channel 0 key 72
[0] channel.NoteOn channel 0 key 74 dyn =
[120] channel.NoteOff channel 0 key 74
[0] channel.NoteOn channel 0 key 76 dyn =
[120] channel.NoteOff channel 0 key 76
[0] channel.NoteOn channel 0 key 77 dyn =
[120] channel.NoteOff channel 0 key 77
[0] channel.NoteOn channel 0 key 79 dyn =
[120] channel.NoteOff channel 0 key 79
[0] channel.NoteOn channel 0 key 81 dyn =
[120] channel.NoteOff channel 0 key 81
[0] channel.NoteOn channel 0 key 83 dyn =
[120] channel.NoteOff channel 0 key 83
[0] channel.NoteOn channel 0 key 84 dyn =
[120] channel.NoteOff channel 0 key 84
`,
		},
		{
			// 118
			`
|TRACK |p|
|Ch      |1|

|=SCORE |p |
|#
|    1    | d#+                    |
|  2      | eb-                    |
|3 | * |
			`,
			1,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 51 dyn +
[960] channel.NoteOff channel 0 key 51
[0] channel.NoteOn channel 0 key 51 dyn -
[960] channel.NoteOff channel 0 key 51
`,
		},
		{
			// 119
			`
'drumnote
			
|TRACK  | dr |
|Ch     | 1  |

|=test     | 1                | 2   |
|1         | $euclid(3,8,0.5) | =.1(.drums.kd++,.drums.sn+) |
|#
|1         |                   |...       |
|#
|#

|=SCORE | dr     |
| 1     | =test.2 |
| *3
            `,
			1,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 36 dyn ++
[60] channel.NoteOff channel 0 key 36
[420] channel.NoteOn channel 0 key 40 dyn +
[60] channel.NoteOff channel 0 key 40
[420] channel.NoteOn channel 0 key 40 dyn +
[60] channel.NoteOff channel 0 key 40
[420] channel.NoteOn channel 0 key 36 dyn ++
[60] channel.NoteOff channel 0 key 36
[420] channel.NoteOn channel 0 key 40 dyn +
[60] channel.NoteOff channel 0 key 40
[420] channel.NoteOn channel 0 key 40 dyn +
[60] channel.NoteOff channel 0 key 40
[420] channel.NoteOn channel 0 key 36 dyn ++
[60] channel.NoteOff channel 0 key 36
[420] channel.NoteOn channel 0 key 40 dyn +
[60] channel.NoteOff channel 0 key 40
[420] channel.NoteOn channel 0 key 36 dyn ++
[60] channel.NoteOff channel 0 key 36
[420] channel.NoteOn channel 0 key 40 dyn +
[60] channel.NoteOff channel 0 key 40
[420] channel.NoteOn channel 0 key 40 dyn +
[60] channel.NoteOff channel 0 key 40
[420] channel.NoteOn channel 0 key 36 dyn ++
[60] channel.NoteOff channel 0 key 36
[420] channel.NoteOn channel 0 key 40 dyn +
[60] channel.NoteOff channel 0 key 40
[420] channel.NoteOn channel 0 key 40 dyn +
[60] channel.NoteOff channel 0 key 40
[420] channel.NoteOn channel 0 key 36 dyn ++
[60] channel.NoteOff channel 0 key 36
[420] channel.NoteOn channel 0 key 40 dyn +
[60] channel.NoteOff channel 0 key 40
[420] channel.NoteOn channel 0 key 36 dyn ++
[60] channel.NoteOff channel 0 key 36
[420] channel.NoteOn channel 0 key 40 dyn +
[60] channel.NoteOff channel 0 key 40
[420] channel.NoteOn channel 0 key 40 dyn +
[60] channel.NoteOff channel 0 key 40
[420] channel.NoteOn channel 0 key 36 dyn ++
[60] channel.NoteOff channel 0 key 36
[420] channel.NoteOn channel 0 key 40 dyn +
[60] channel.NoteOff channel 0 key 40
[420] channel.NoteOn channel 0 key 40 dyn +
[60] channel.NoteOff channel 0 key 40
[420] channel.NoteOn channel 0 key 36 dyn ++
[60] channel.NoteOff channel 0 key 36
[420] channel.NoteOn channel 0 key 40 dyn +
[60] channel.NoteOff channel 0 key 40
[420] channel.NoteOn channel 0 key 36 dyn ++
[60] channel.NoteOff channel 0 key 36
[420] channel.NoteOn channel 0 key 40 dyn +
[60] channel.NoteOff channel 0 key 40
[420] channel.NoteOn channel 0 key 40 dyn +
[60] channel.NoteOff channel 0 key 40
[420] channel.NoteOn channel 0 key 36 dyn ++
[60] channel.NoteOff channel 0 key 36
[420] channel.NoteOn channel 0 key 40 dyn +
[60] channel.NoteOff channel 0 key 40
[420] channel.NoteOn channel 0 key 40 dyn +
[60] channel.NoteOff channel 0 key 40
[420] channel.NoteOn channel 0 key 36 dyn ++
[60] channel.NoteOff channel 0 key 36
[420] channel.NoteOn channel 0 key 40 dyn +
[60] channel.NoteOff channel 0 key 40
`,
		},
		{
			// 120
			`
|TRACK    | a | 
|Ch       | 1 | 
 
|=patt    | a    | 
|    1    | c++  | 
|    4    | D~~  |
|    4&   | E+   | 
|    4&,  | *    |

|=SCORE   | a          | 
| #
|    1    | =patt.a%2 | 
| #
|    1&   | *         |

            `,
			1,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 48 dyn ++
[2880] channel.NoteOff channel 0 key 48
[0] channel.NoteOn channel 0 key 38 dyn =
[60] channel.Pitchbend channel 0 value 128 absValue 0
[60] channel.Pitchbend channel 0 value 512 absValue 0
[60] channel.Pitchbend channel 0 value 1152 absValue 0
[60] channel.Pitchbend channel 0 value 2048 absValue 0
[60] channel.Pitchbend channel 0 value 3200 absValue 0
[60] channel.Pitchbend channel 0 value 4608 absValue 0
[60] channel.Pitchbend channel 0 value 6272 absValue 0
[60] channel.Pitchbend channel 0 value 8191 absValue 0
[240] channel.NoteOff channel 0 key 38
[240] channel.Pitchbend channel 0 value 0 absValue 0
[0] channel.NoteOn channel 0 key 48 dyn ++
[480] channel.NoteOff channel 0 key 48
`,
		},
		{ // 121
			`
|TRACK  | voc | git |
|Ch     | 1   | 2   |

|=patt | 1  |
|#
|  1   | c  |
|#
|  1   | d  |
|#
|  1   | e  |
|#
|  1   | f  |
|#
|  1   | g  |
|#
|  1   | b  |
|#
|  1   | c# |
|#
|  1   | d# |

|=SCORE | voc     | git        |
|#A
| 1     | =patt.1 | =patt.1%10 |
|*3
|#
| 1     | .4.     |            |
|*3
|[A]
  `,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 1 key 48 dyn =
[3840] channel.NoteOff channel 1 key 48
[0] channel.NoteOn channel 1 key 50 dyn =
[3840] channel.NoteOff channel 1 key 50
[0] channel.NoteOn channel 1 key 52 dyn =
[3840] channel.NoteOff channel 1 key 52
[0] channel.NoteOn channel 1 key 53 dyn =
[3840] channel.NoteOff channel 1 key 53
[0] channel.NoteOn channel 1 key 55 dyn =
[3840] channel.NoteOff channel 1 key 55
[0] channel.NoteOn channel 1 key 59 dyn =
[3840] channel.NoteOff channel 1 key 59
[0] channel.NoteOn channel 1 key 49 dyn =
[3840] channel.NoteOff channel 1 key 49
[0] channel.NoteOn channel 1 key 51 dyn =
[3840] channel.NoteOff channel 1 key 51
[0] channel.NoteOn channel 1 key 48 dyn =
[3840] channel.NoteOff channel 1 key 48
[0] channel.NoteOn channel 1 key 50 dyn =
[3840] channel.NoteOff channel 1 key 50
[0] channel.NoteOn channel 1 key 52 dyn =
[3840] channel.NoteOff channel 1 key 52
[0] channel.NoteOn channel 1 key 53 dyn =
[3840] channel.NoteOff channel 1 key 53
[0] channel.NoteOn channel 1 key 55 dyn =
[3840] channel.NoteOff channel 1 key 55
[0] channel.NoteOn channel 1 key 59 dyn =
[3840] channel.NoteOff channel 1 key 59
[0] channel.NoteOn channel 1 key 49 dyn =
[3840] channel.NoteOff channel 1 key 49
[0] channel.NoteOn channel 1 key 51 dyn =
[3840] channel.NoteOff channel 1 key 51
`,
		},
		{ // 122
			`
|TRACK  | voc | git |
|Ch     | 2   | 1   |

|=patt | 1  |
|#
|  1   | c  |
|#
|  1   | d  |
|#
|  1   | e  |
|#
|  1   | f  |
|#
|  1   | g  |
|#
|  1   | b  |
|#
|  1   | c# |
|#
|  1   | d# |

|=SCORE | voc     | git        |
|#A
| 1     | =patt.1 | =patt.1%10 |
|*3
|#
| 1     | .4.     |            |
|*3
|[A]
  `,
			1,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 48 dyn =
[3840] channel.NoteOff channel 0 key 48
[0] channel.NoteOn channel 0 key 50 dyn =
[3840] channel.NoteOff channel 0 key 50
[0] channel.NoteOn channel 0 key 52 dyn =
[3840] channel.NoteOff channel 0 key 52
[0] channel.NoteOn channel 0 key 53 dyn =
[3840] channel.NoteOff channel 0 key 53
[0] channel.NoteOn channel 0 key 55 dyn =
[3840] channel.NoteOff channel 0 key 55
[0] channel.NoteOn channel 0 key 59 dyn =
[3840] channel.NoteOff channel 0 key 59
[0] channel.NoteOn channel 0 key 49 dyn =
[3840] channel.NoteOff channel 0 key 49
[0] channel.NoteOn channel 0 key 51 dyn =
[3840] channel.NoteOff channel 0 key 51
[0] channel.NoteOn channel 0 key 48 dyn =
[3840] channel.NoteOff channel 0 key 48
[0] channel.NoteOn channel 0 key 50 dyn =
[3840] channel.NoteOff channel 0 key 50
[0] channel.NoteOn channel 0 key 52 dyn =
[3840] channel.NoteOff channel 0 key 52
[0] channel.NoteOn channel 0 key 53 dyn =
[3840] channel.NoteOff channel 0 key 53
[0] channel.NoteOn channel 0 key 55 dyn =
[3840] channel.NoteOff channel 0 key 55
[0] channel.NoteOn channel 0 key 59 dyn =
[3840] channel.NoteOff channel 0 key 59
[0] channel.NoteOn channel 0 key 49 dyn =
[3840] channel.NoteOff channel 0 key 49
[0] channel.NoteOn channel 0 key 51 dyn =
[3840] channel.NoteOff channel 0 key 51
`,
		},
		{ // 123
			`
|TRACK  | voc |
|Ch     | 1   |

|=SCORE | voc  |
|#
|  1   | 1  |
|#
|  1   | 2  |
|#
|  1   | 1' |
|#
|  1   | 2"  |
|#
|  1   | -1  |
|#
|  1   | -1'  |
|#
|  1   | -2" |
`,
			1,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 60 dyn =
[3840] channel.NoteOff channel 0 key 60
[0] channel.NoteOn channel 0 key 62 dyn =
[3840] channel.NoteOff channel 0 key 62
[0] channel.NoteOn channel 0 key 72 dyn =
[3840] channel.NoteOff channel 0 key 72
[0] channel.NoteOn channel 0 key 86 dyn =
[3840] channel.NoteOff channel 0 key 86
[0] channel.NoteOn channel 0 key 59 dyn =
[3840] channel.NoteOff channel 0 key 59
[0] channel.NoteOn channel 0 key 47 dyn =
[3840] channel.NoteOff channel 0 key 47
[0] channel.NoteOn channel 0 key 33 dyn =
[3840] channel.NoteOff channel 0 key 33
`,
		},
		{ // 124
			`
|TRACK  | voc |
|Ch     | 1   |

|=SCORE | voc  |
|# @110~
|# @140
`,
			0,
			true,
			false,
			`
[0] meta.Copyright: ""
[0] meta.TrackSequenceName: "BPM"
[0] meta.Tempo BPM: 110.00
[60] meta.Tempo BPM: 110.47
[60] meta.Tempo BPM: 110.94
[60] meta.Tempo BPM: 111.41
[60] meta.Tempo BPM: 111.87
[60] meta.Tempo BPM: 112.34
[60] meta.Tempo BPM: 112.81
[60] meta.Tempo BPM: 113.28
[60] meta.Tempo BPM: 113.75
[60] meta.Tempo BPM: 114.22
[60] meta.Tempo BPM: 114.69
[60] meta.Tempo BPM: 115.16
[60] meta.Tempo BPM: 115.62
[60] meta.Tempo BPM: 116.09
[60] meta.Tempo BPM: 116.56
[60] meta.Tempo BPM: 117.03
[60] meta.Tempo BPM: 117.50
[60] meta.Tempo BPM: 117.97
[60] meta.Tempo BPM: 118.44
[60] meta.Tempo BPM: 118.91
[60] meta.Tempo BPM: 119.37
[60] meta.Tempo BPM: 119.84
[60] meta.Tempo BPM: 120.31
[60] meta.Tempo BPM: 120.78
[60] meta.Tempo BPM: 121.25
[60] meta.Tempo BPM: 121.72
[60] meta.Tempo BPM: 122.19
[60] meta.Tempo BPM: 122.66
[60] meta.Tempo BPM: 123.12
[60] meta.Tempo BPM: 123.59
[60] meta.Tempo BPM: 124.06
[60] meta.Tempo BPM: 124.53
[60] meta.Tempo BPM: 125.00
[60] meta.Tempo BPM: 125.47
[60] meta.Tempo BPM: 125.94
[60] meta.Tempo BPM: 126.41
[60] meta.Tempo BPM: 126.88
[60] meta.Tempo BPM: 127.34
[60] meta.Tempo BPM: 127.81
[60] meta.Tempo BPM: 128.28
[60] meta.Tempo BPM: 128.75
[60] meta.Tempo BPM: 129.22
[60] meta.Tempo BPM: 129.69
[60] meta.Tempo BPM: 130.16
[60] meta.Tempo BPM: 130.63
[60] meta.Tempo BPM: 131.09
[60] meta.Tempo BPM: 131.56
[60] meta.Tempo BPM: 132.03
[60] meta.Tempo BPM: 132.50
[60] meta.Tempo BPM: 132.97
[60] meta.Tempo BPM: 133.44
[60] meta.Tempo BPM: 133.91
[60] meta.Tempo BPM: 134.37
[60] meta.Tempo BPM: 134.84
[60] meta.Tempo BPM: 135.31
[60] meta.Tempo BPM: 135.78
[60] meta.Tempo BPM: 136.25
[60] meta.Tempo BPM: 136.72
[60] meta.Tempo BPM: 137.19
[60] meta.Tempo BPM: 137.66
[60] meta.Tempo BPM: 138.13
[60] meta.Tempo BPM: 138.59
[60] meta.Tempo BPM: 139.06
[60] meta.Tempo BPM: 139.53
[60] meta.Tempo BPM: 140.00
[0] meta.Tempo BPM: 140.00
`,
		},
		{ // 125
			`
|TRACK  | voc |
|Ch     | 1   |

|=SCORE | voc  |
|# @110~~
|# @140
`,
			0,
			true,
			false,
			`
[0] meta.Copyright: ""
[0] meta.TrackSequenceName: "BPM"
[0] meta.Tempo BPM: 110.00
[60] meta.Tempo BPM: 110.01
[60] meta.Tempo BPM: 110.03
[60] meta.Tempo BPM: 110.07
[60] meta.Tempo BPM: 110.12
[60] meta.Tempo BPM: 110.18
[60] meta.Tempo BPM: 110.26
[60] meta.Tempo BPM: 110.36
[60] meta.Tempo BPM: 110.47
[60] meta.Tempo BPM: 110.59
[60] meta.Tempo BPM: 110.73
[60] meta.Tempo BPM: 110.89
[60] meta.Tempo BPM: 111.05
[60] meta.Tempo BPM: 111.24
[60] meta.Tempo BPM: 111.44
[60] meta.Tempo BPM: 111.65
[60] meta.Tempo BPM: 111.87
[60] meta.Tempo BPM: 112.12
[60] meta.Tempo BPM: 112.37
[60] meta.Tempo BPM: 112.64
[60] meta.Tempo BPM: 112.93
[60] meta.Tempo BPM: 113.23
[60] meta.Tempo BPM: 113.54
[60] meta.Tempo BPM: 113.87
[60] meta.Tempo BPM: 114.22
[60] meta.Tempo BPM: 114.58
[60] meta.Tempo BPM: 114.95
[60] meta.Tempo BPM: 115.34
[60] meta.Tempo BPM: 115.74
[60] meta.Tempo BPM: 116.16
[60] meta.Tempo BPM: 116.59
[60] meta.Tempo BPM: 117.04
[60] meta.Tempo BPM: 117.50
[60] meta.Tempo BPM: 117.98
[60] meta.Tempo BPM: 118.47
[60] meta.Tempo BPM: 118.97
[60] meta.Tempo BPM: 119.49
[60] meta.Tempo BPM: 120.03
[60] meta.Tempo BPM: 120.58
[60] meta.Tempo BPM: 121.14
[60] meta.Tempo BPM: 121.72
[60] meta.Tempo BPM: 122.31
[60] meta.Tempo BPM: 122.92
[60] meta.Tempo BPM: 123.54
[60] meta.Tempo BPM: 124.18
[60] meta.Tempo BPM: 124.83
[60] meta.Tempo BPM: 125.50
[60] meta.Tempo BPM: 126.18
[60] meta.Tempo BPM: 126.88
[60] meta.Tempo BPM: 127.59
[60] meta.Tempo BPM: 128.31
[60] meta.Tempo BPM: 129.05
[60] meta.Tempo BPM: 129.80
[60] meta.Tempo BPM: 130.57
[60] meta.Tempo BPM: 131.36
[60] meta.Tempo BPM: 132.16
[60] meta.Tempo BPM: 132.97
[60] meta.Tempo BPM: 133.80
[60] meta.Tempo BPM: 134.64
[60] meta.Tempo BPM: 135.50
[60] meta.Tempo BPM: 136.37
[60] meta.Tempo BPM: 137.25
[60] meta.Tempo BPM: 138.15
[60] meta.Tempo BPM: 139.07
[60] meta.Tempo BPM: 140.00
[0] meta.Tempo BPM: 140.00
`,
		},
		/*


			TRACK    | bass | dr  |
			 Channel | 1    | 10  |
			 Volume  |      | 127 |

			$$include("drumnote")
			$$embed(".drums.*")

			=patt    | bass | drums  |
			    1    | E++  | .kd+++ |
			    2    |      | .sn+   |
			    3    |      | .kd+   |
			    3,;  | *    |        |
			    3&   | B'++ |        |
			    4    | D~~  | .sn+   |
			    4&   | E+   |        |
			    4&,  | *    |        |

			=SCORE   | bass          | dr             |
			 #
			    1    | =patt.bass%10 | =patt.drums%10 |
			 *9

		*/

		/*
					{
						// 113
						`
			TRACK |p|
			Ch      |1|

			=SCORE |p |
			#
			    1    | d#+++:                        |
			    1&   |                               |
			    1&.  | (d#+++ g+::< a#++::< d#'+++)  |
			    1&.; | *                             |
			    2    |                               |
			    2&   | (d#+++ g+::< a#++::< d#'+++)  |
			    3    |                               |
			    3;   | *                             |
			    3&   |                               |
			    3&.  | d+++:                         |
			    4    |                               |
			    4&   | (d+++ f+::< b++::< d'+++)     |
			    4&;  | *                             |
						`,
						2,
						false,
						false,
						`
			`,
					},
		*/
		/*
		  1                                      | c++     |           |           | d#+++:                        | MN36+++:::  | MN46+++:::  |           |        |
		    1&                                     |         |           |           |                               |             | MN42+++:::  |           |        |
		    1&.                                    | D#:     |           |           | (d#+++ g+::< a#++::< d#'+++)  | MN40+++:::  |             |           |        |
		    1&.;                                   |         |           |           | *                             |             |             |           |        |
		    2                                      |         |           |           |                               | MN36++:::   | MN42+++:::  |           |        |
		    2&                                     | c+      |           |           | (d#+++ g+::< a#++::< d#'+++)  | MN40+++:::  | MN46+++:::  |           |        |
		    3                                      | c++     |           |           |                               | MN36+++:::  | MN42+++:::  |           |        |
		    3;                                     |         |           |           | *                             |             |             |           |        |
		    3&                                     |         |           |           |                               |             | MN42+++:::  |           |        |
		    3&.                                    | D#:     |           |           | d+++:                         | MN36++:::   |             |           |        |
		    4                                      |         |           |           |                               | MN40+++:::< | MN46+++:::  |           |        |
		    4&                                     | c+      |           |           | (d+++ f+::< b++::< d'+++)     | MN36++:::   | MN42+++:::  |           |        |
		    4&;                                    |         |           |           | *                             |             |             |           |        |
		*/

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

	skip := map[int]bool{
		66:  true, // TODO: fix pitchbend absolute value bug
		83:  true,
		110: true, // TODO: fix pitchbend absolute value bug
		120: true, // TODO: fix pitchbend absolute value bug
	}

	//sketch.DEBUG = true

	for i, test := range tests {

		/*
			if i > 128 {
				continue
			}
		*/

		/*
				if i == 108 || i == 112 {
					continue
				}

			if i > 0 {
				continue
			}
		*/

		if skip[i] {
			continue
		}

		/*
			if i != 52 {
				continue
			}
		*/
		/*
		 */
		//sketch.DEBUG = true

		//fmt.Println("##############")

		//fmt.Printf("--        test-%v  -----\n", i)
		loc := path.MustFileFromSystem(fmt.Sprintf("test-%v", i))
		fsys, err := mockfs.New(loc.Dir())

		if err != nil {
			panic(err.Error())
		}

		err = muskel.WritePredefinedTemplates(fsys)

		if err != nil {
			panic(err.Error())
		}

		var opts []score.Option

		opts = append(opts, score.FS(fsys))

		sc := score.New(path.Relative(path.Name(loc)), nil, opts...)
		f := file.FromReader(strings.NewReader(test.input), sc)

		sc.Files[fmt.Sprintf("test-%v", i)] = f

		err = f.Parse()
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
		//err = smf.WriteSMFTo(sc, ioutil.Discard, "*", smfwriter.Debug(&lg))

		var bf bytes.Buffer

		//	err = smf.WriteSMFTo(sc, ioutil.Discard, "*", smf.WithLogger(nil))
		err = smf.WriteSMFTo(sc, &bf, "*", smf.WithLogger(nil))

		if err != nil {
			t.Errorf("[%v] can't write SMF: %s", i, err)
			continue
		}

		tr := 1

		str, err := lg.helperSMFToString(bf.Bytes())

		if err != nil {
			t.Errorf("[%v] can't write SMF, track %v: %s", i, tr, err)
			continue
		}

		//got := strings.TrimSpace(lg.bf.String())
		got := strings.TrimSpace(str)
		expected := strings.TrimSpace(test.expected)

		if got != expected {
			t.Errorf("[%v]\n%s\ngot: \n%s\nexpected: \n%s\n", i, test.input, got, expected)
		}

	}

	items.DEBUG = false
}

type logger struct {
	bf                   strings.Builder
	wantedTrack          int
	currentTrack         int
	includeMeta          bool
	includeControlChange bool
}

func (lg logger) helperSMFToString(smf []byte) (string, error) {
	mf, err := smfmidi.ReadFrom(bytes.NewReader(smf))

	if err != nil {
		return "", err
	}

	if len(mf.Tracks)-1 < lg.wantedTrack {
		return "", fmt.Errorf("track %v not found, only had %v tracks", lg.wantedTrack, len(mf.Tracks))
	}

	var bf bytes.Buffer

	tr := mf.Tracks[lg.wantedTrack]

	for _, ev := range tr {
		var ch, byte1, byte2 uint8
		var text string
		var pbrel int16
		var pbabs uint16

		switch {
		case ev.Message.GetNoteStart(&ch, &byte1, &byte2):
			// channel.NoteOn channel 0 key 72 dyn =
			fmt.Fprintf(&bf, "[%v] channel.NoteOn channel %v key %v dyn %s\n", ev.Delta, ch, byte1, _velocityToDynamic(byte2))
		case ev.Message.GetNoteEnd(&ch, &byte1):
			// channel.NoteOff channel 0 key 72
			fmt.Fprintf(&bf, "[%v] channel.NoteOff channel %v key %v\n", ev.Delta, ch, byte1)
		case ev.Message.GetMetaLyric(&text):
			//[0] meta.Lyric: "hiho"
			fmt.Fprintf(&bf, "[%v] meta.Lyric: %q\n", ev.Delta, text)
		case ev.Message.GetAfterTouch(&ch, &byte1):
			//[0] channel.Aftertouch channel 0 pressure 23
			fmt.Fprintf(&bf, "[%v] channel.Aftertouch channel %v pressure %v\n", ev.Delta, ch, byte1)
		case ev.Message.GetPitchBend(&ch, &pbrel, &pbabs):
			// [960] channel.Pitchbend channel 1 value 20 absValue 0
			fmt.Fprintf(&bf, "[%v] channel.Pitchbend channel %v value %v absValue %v\n", ev.Delta, ch, pbrel, pbabs)
		case ev.Message.GetProgramChange(&ch, &byte1):
			fmt.Fprintf(&bf, "[%v] channel.ProgramChange channel %v program %v\n", ev.Delta, ch, byte1)
			// [0] channel.ProgramChange channel 0 program 1
		case ev.Message.GetControlChange(&ch, &byte1, &byte2):
			if !lg.includeControlChange {
				continue
			}

			ccname := midi.ControlChangeName[byte1]

			if ccname != "" {
				ccname = fmt.Sprintf(" (%q)", ccname)
			}

			//		channel.ControlChange channel 0 controller 101 ("Registered Parameter (MSB)") value 0
			fmt.Fprintf(&bf, "[%v] channel.ControlChange channel %v controller %v%s value %v\n", ev.Delta, ch, byte1, ccname, byte2)
		default:
			if ev.Message.IsMeta() && !lg.includeMeta {
				continue
			}

			var text string
			var tempo float64
			switch {
			case ev.Message.GetMetaCopyright(&text):
				fmt.Fprintf(&bf, "[%v] meta.Copyright: %q\n", ev.Delta, text)
			case ev.Message.GetMetaTrackName(&text):
				fmt.Fprintf(&bf, "[%v] meta.TrackSequenceName: %q\n", ev.Delta, text)
			case ev.Message.GetMetaTempo(&tempo):
				fmt.Fprintf(&bf, "[%v] meta.Tempo BPM: %0.2f\n", ev.Delta, tempo)
			}

		}
	}

	return bf.String(), nil
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

func (l *logger) format(msg smfmidi.Message) string {
	var channel, val1, val2 uint8
	switch {
	case msg.GetNoteStart(&channel, &val1, &val2):
		dyn := _velocityToDynamic(val2)
		return fmt.Sprintf("channel.NoteOn channel %v key %v dyn %s", channel, val1, dyn)
	case msg.GetNoteEnd(&channel, &val1):
		return fmt.Sprintf("channel.NoteOff channel %v key %v", channel, val1)
	default:
		return msg.String()
	}

	/*
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
	*/
}

func (l *logger) Printf(format string, vals ...interface{}) {
	if len(vals) != 2 {
		return
	}

	msg, ok := vals[1].(smfmidi.Message)

	if !ok {
		return
	}

	if msg.Is(smfmidi.MetaEndOfTrackMsg) {
		l.currentTrack++
		if l.includeMeta {
			fmt.Fprintf(&l.bf, "[%v] %s\n", vals[0], l.format(msg))
		}
		return
	}

	if l.currentTrack != l.wantedTrack {
		return
	}

	//if _, isMetaUndef := vals[1].(meta.Undefined); isMetaUndef {
	if msg.Is(smfmidi.MetaUndefinedMsg) {
		fmt.Fprintf(&l.bf, "[%v] %s\n", vals[0], l.format(msg))
	}

	var text string
	if msg.GetMetaLyric(&text) {
		//if msg.Is(smfmidi.MetaL{
		//if _, isLyric := vals[1].(meta.Lyric); isLyric {
		fmt.Fprintf(&l.bf, "[%v] %s\n", vals[0], l.format(msg))
	}

	if msg.IsMeta() {
		return
	}

	/*
		if _, isMeta := vals[1].(meta.Message); isMeta && !l.includeMeta {
			return
		}
	*/

	var channel, val1, val2 uint8
	if msg.GetControlChange(&channel, &val1, &val2) {
		//if _, isCC := vals[1].(channel.ControlChange); isCC && !l.includeControlChange {
		return
	}

	fmt.Fprintf(&l.bf, "[%v] %s\n", vals[0], l.format(msg))

}
