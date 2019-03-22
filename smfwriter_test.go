package muskel

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"gitlab.com/gomidi/midi/midimessage/channel"
	"gitlab.com/gomidi/midi/midimessage/meta"
	"gitlab.com/gomidi/midi/smf/smfwriter"
)

func TestSMFWriter(t *testing.T) {

	DEBUG = false

	var tests = []struct {
		input                string
		track                int
		includeMeta          bool
		includeControlChange bool
		expected             string
	}{
		{
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
`,
		},
		{
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
`,
		},
		{
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
`,
		},
	}

	for i, test := range tests {
		sc, err := Parse(strings.NewReader(test.input), fmt.Sprintf("test-%v", i))

		if err != nil {
			t.Errorf("[%v] can't parse score: %s", i, err)
			continue
		}

		usc, err := sc.Unroll()

		if err != nil {
			t.Errorf("[%v] can't unroll score: %s", i, err)
			continue
		}

		var lg logger
		lg.wantedTrack = test.track
		lg.includeMeta = test.includeMeta
		lg.includeControlChange = test.includeControlChange
		err = usc.writeSMFTo(ioutil.Discard, "*", smfwriter.Debug(&lg))

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
		return
	}

	if len(vals) != 2 {
		return
	}

	if l.currentTrack != l.wantedTrack {
		return
	}

	if _, isMeta := vals[1].(meta.Message); isMeta && !l.includeMeta {
		return
	}

	if _, isCC := vals[1].(channel.ControlChange); isCC && !l.includeControlChange {
		return
	}

	fmt.Fprintf(&l.bf, "[%v] %s\n", vals[0], l.format(vals[1]))

}
