package muskel_test

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"gitlab.com/gomidi/midi/midimessage/channel"
	"gitlab.com/gomidi/midi/midimessage/meta"
	"gitlab.com/gomidi/midi/smf/smfwriter"
	"gitlab.com/gomidi/muskel"
	"gitlab.com/gomidi/muskel/score"
	"gitlab.com/gomidi/muskel/smf"
	"gitlab.com/gomidi/muskel/unroller"
)

func TestSMFWriter(t *testing.T) {

	score.DEBUG = false

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
=
   |<p>|
Ch |1  |
1  |c" |
2  |d  |
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
=
   |<p>|
Ch |1  |
1  |{c,e,g#}2 |
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
=
   |<p>|
Ch |1  |
1  |c |
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
=
   |<p>|
Ch |1  |
1  |{c,e,g#}5 |
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
=
   |<p>|
Ch |1  |
1  |c |

1  |:|
2  |*|
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
=
   |<p>|
Ch |1  |
1  |{c,e,g#}6 |

1  |:|
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
=
   |<p>|
Ch |1  |
Delay|1/32|
1  |{c,e,g#}6 |

1  |:|
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
=
   |<p>|
Ch |1  |
Delay|-1/32|
1  | |

1  |{c,e,g#}6 |

1  |:|
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
=
   |<p>|
Ch |1  |
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
=
   |<p>|
Ch |1  |
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
=
   |<p>|
Ch |1  |
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
=
   |<p>|
Ch |1  |
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
=
   |<p>|
Ch |1  |
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
=
   |<p>|
Ch |1  |
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
=
   |<p>|
Ch |1  |
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
=
   |<p>|
Ch |1  |
1  |MN48 |
2  |MN50< |
3  |MN52 |
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
			// 16
			`
=
   |<p>|
Ch |1  |
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
=
   |<p>|
Ch |1  |
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
=
   |<p>|
Ch |1  |
1  | |

1  |{c,e,g#}6> |

1  |:|
2  |*|
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
=
   |<p>|
Ch |1  |
1  | |

1  |{c,e,g#}6< |

1  |:|
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
templ: 1{c,e,g#}2
=
   |<p>|
Ch |1  |
1  |templ |
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
ta: c
tb: e
tc: g#
=
   |<p>|
Ch |1  |
1  |{ta,tb,tc}2 |
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
ta: MN48
tb: MN52
tc: MN56
=
   |<p>|
Ch |1  |
1  |{ta,tb,tc}2 |
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
ta: MN48=
tb: MN52=
tc: MN56=
=
   |<p>|
Ch |1  |
1  |{ta+,tb-,tc--}2 |
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
=
   |<p>|
Ch |1  |
1  |c: |
			`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 48 dyn =
[90] channel.NoteOff channel 0 key 48
`,
		},
		{
			// 25
			`
=
   |<p>|
Ch |1  |
1  |{c:,e:,g#:}2 |
			`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 48 dyn =
[90] channel.NoteOff channel 0 key 48
[230] channel.NoteOn channel 0 key 52 dyn =
[90] channel.NoteOff channel 0 key 52
[230] channel.NoteOn channel 0 key 56 dyn =
[90] channel.NoteOff channel 0 key 56
[230] meta.Undefined type:  0
`,
		},
		{
			// 26
			`
=
   |<p>|
Ch |1  |
1  |{c,e,g#}2 |
2  |a: |
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
[90] channel.NoteOff channel 0 key 57
`,
		},
		{
			// 27
			`
=
   |<p>|
Ch |1  |
1  |{c:,e:,g#:}2 |
2  |a: |
			`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 48 dyn =
[90] channel.NoteOff channel 0 key 48
[230] channel.NoteOn channel 0 key 52 dyn =
[90] channel.NoteOff channel 0 key 52
[230] channel.NoteOn channel 0 key 56 dyn =
[90] channel.NoteOff channel 0 key 56
[230] meta.Undefined type:  0
[0] channel.NoteOn channel 0 key 57 dyn =
[90] channel.NoteOff channel 0 key 57
`,
		},
		{
			// 28
			`
=
   |<p>|
Ch |1  |
1  | d |
2  |{c,e,g#}3 |
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
=
   |<p>|
Ch |1  |
1  | d |
4  |{c,e,g#}5 |

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
=
   |<p>|
Ch |1  |
1  | d |`,
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
=
   |<p>|
Ch |1  |
1  | d |`,
			2,
			false,
			false,
			`
[0] channel.NoteOn channel 0 key 50 dyn =
[3840] channel.NoteOff channel 0 key 50


`,
		},
	}

	for i, test := range tests {

		/*
			if i != 30 {
				continue
			}
		*/
		sc, err := muskel.Parse(strings.NewReader(test.input), fmt.Sprintf("test-%v", i))

		if err != nil {
			t.Errorf("[%v] can't parse score: %s", i, err)
			continue
		}

		usc, err := unroller.Unroll(sc)

		if err != nil {
			t.Errorf("[%v] can't unroll score: %s", i, err)
			continue
		}

		var lg logger
		lg.wantedTrack = test.track
		lg.includeMeta = test.includeMeta
		lg.includeControlChange = test.includeControlChange
		err = smf.WriteSMFTo(usc, ioutil.Discard, "*", smfwriter.Debug(&lg))

		if err != nil {
			t.Errorf("[%v] can't write SMF: %s", i, err)
			continue
		}

		//got := strings.Replace(lg.bf.String(), preStr, "", 1)
		//got := strings.TrimSpace(strings.Replace(strings.TrimSpace(lg.bf.String()), preStr, "", 1))
		got := strings.TrimSpace(lg.bf.String())
		expected := strings.TrimSpace(test.expected)

		if got != expected {
			t.Errorf("[%v]\ngot: \n%s\nexpected: \n%s\n", i, got, expected)
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

	if _, isMeta := vals[1].(meta.Message); isMeta && !l.includeMeta {
		return
	}

	if _, isCC := vals[1].(channel.ControlChange); isCC && !l.includeControlChange {
		return
	}

	fmt.Fprintf(&l.bf, "[%v] %s\n", vals[0], l.format(vals[1]))

}
