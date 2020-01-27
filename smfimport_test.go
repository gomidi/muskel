package muskel_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"gitlab.com/gomidi/midi/smf/smfwriter"
	"gitlab.com/gomidi/muskel/file"
	"gitlab.com/gomidi/muskel/score"
	"gitlab.com/gomidi/muskel/smf"
	"gitlab.com/gomidi/muskel/smfimport"
)

func TestSMFImport(t *testing.T) {

	var tests = []struct {
		msklIn               string
		midiOut              string
		msklOut              string
		track                int
		includeMeta          bool
		includeControlChange bool
	}{
		{ // 0
			`
TRACK | bass | git |
Channel | 9   | 8   |

=SCORE | bass | git |
#
1      | c    | d   |
#
1      | *    |     |
2      |      | *   |
`,
			`
[0] channel.NoteOn channel 8 key 48 dyn =
[3840] channel.NoteOff channel 8 key 48
`,
			`
TRACK           | bass--8                                   | git--7                                    |
 File           |                                           |                                           |
 Channel        | 9                                         | 8                                         |
 Program        | 1                                         | 1                                         |
 Bank           | 1                                         | 1                                         |
 Transpose      |                                           |                                           |
 Volume         | 1                                         | 1                                         |
 Delay          |                                           |                                           |
 PitchbendRange |                                           |                                           |
 VelocityScale  | min: 0 max: 0 random: 0 step: 0 center: 0 | min: 0 max: 0 random: 0 step: 0 center: 0 |

=SCORE     | bass--8 | git--7 |
 # @120.00
    1      | c=_     | d=_    |
 #
    1      | _c      |        |
    2      |         | _d     |

`,
			2,
			false,
			false,
		},
		{ // 1
			`
TRACK | bass | git |
Channel | 9   | 8   |

=SCORE | bass | git |
#
1      | c    | d   |
# 3/4
1      | *    |     |
2      |      | *   |
`,
			`
[0] channel.NoteOn channel 8 key 48 dyn =
[3840] channel.NoteOff channel 8 key 48
`,
			`
TRACK           | bass--8                                   | git--7                                    |
 File           |                                           |                                           |
 Channel        | 9                                         | 8                                         |
 Program        | 1                                         | 1                                         |
 Bank           | 1                                         | 1                                         |
 Transpose      |                                           |                                           |
 Volume         | 1                                         | 1                                         |
 Delay          |                                           |                                           |
 PitchbendRange |                                           |                                           |
 VelocityScale  | min: 0 max: 0 random: 0 step: 0 center: 0 | min: 0 max: 0 random: 0 step: 0 center: 0 |

=SCORE     | bass--8 | git--7 |
 # @120.00
    1      | c=_     | d=_    |
 # 3/4
    1      | _c      |        |
    2      |         | _d     |

`,
			2,
			false,
			false,
		},
		{ // 2
			`
TRACK | bass | git |
Channel | 9   | 8   |

=SCORE | bass | git |
# 3/4 @110
1      | c    | d   |
# 4/4 @200
1      | *    |     |
2      |      | *   |
`,
			`
[0] channel.NoteOn channel 8 key 48 dyn =
[2880] channel.NoteOff channel 8 key 48
`,
			`
TRACK           | bass--8                                   | git--7                                    |
 File           |                                           |                                           |
 Channel        | 9                                         | 8                                         |
 Program        | 1                                         | 1                                         |
 Bank           | 1                                         | 1                                         |
 Transpose      |                                           |                                           |
 Volume         | 1                                         | 1                                         |
 Delay          |                                           |                                           |
 PitchbendRange |                                           |                                           |
 VelocityScale  | min: 0 max: 0 random: 0 step: 0 center: 0 | min: 0 max: 0 random: 0 step: 0 center: 0 |

=SCORE         | bass--8 | git--7 |
 # 3/4 @110.00
    1          | c=_     | d=_    |
 # 4/4 @200.00
    1          | _c      |        |
    2          |         | _d     |

`,
			2,
			false,
			false,
		},
		{ // 3
			`
TRACK | bass | git |
Channel | 9   | 8   |

=SCORE | bass | git |
# 3/4 @110
1      | c    | d   |
2 @115 | *    |     |
# 4/4 @200
2      |      | *   |
`,
			`
[0] channel.NoteOn channel 8 key 48 dyn =
[960] channel.NoteOff channel 8 key 48
`,
			`
TRACK           | bass--8                                   | git--7                                    |
 File           |                                           |                                           |
 Channel        | 9                                         | 8                                         |
 Program        | 1                                         | 1                                         |
 Bank           | 1                                         | 1                                         |
 Transpose      |                                           |                                           |
 Volume         | 1                                         | 1                                         |
 Delay          |                                           |                                           |
 PitchbendRange |                                           |                                           |
 VelocityScale  | min: 0 max: 0 random: 0 step: 0 center: 0 | min: 0 max: 0 random: 0 step: 0 center: 0 |

=SCORE         | bass--8 | git--7 |
 # 3/4 @110.00
    1          | c=_     | d=_    |
    2 @115.00  |         |        |
    2          | _c      |        |
 # 4/4 @200.00
    2          |         | _d     |

`,
			2,
			false,
			false,
		},
	}

	for i, test := range tests {

		sc := score.New(fmt.Sprintf("test-%v", i), nil)
		f := file.FromReader(strings.NewReader(test.msklIn), sc)

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

		var midibf bytes.Buffer

		var lg logger
		lg.wantedTrack = test.track
		lg.includeMeta = test.includeMeta
		lg.includeControlChange = test.includeControlChange
		err = smf.WriteSMFTo(sc, &midibf, "*", smfwriter.Debug(&lg))

		if err != nil {
			t.Errorf("[%v] can't write SMF: %s", i, err)
			continue
		}

		//fmt.Printf("%v\n", midibf.Bytes())

		got := strings.TrimSpace(lg.bf.String())
		expected := strings.TrimSpace(test.midiOut)

		if got != expected {
			t.Errorf("[%v]\n%s\ngot midiout: \n%s\nexpected: \n%s\n", i, test.msklIn, got, expected)
		}

		cv := smfimport.New("convert-main", bytes.NewReader(midibf.Bytes()))
		cv.IgnoreCC = !test.includeControlChange

		var bf bytes.Buffer

		err = cv.WriteMsklTo(&bf)

		if err != nil {
			t.Errorf("[%v] could not write imported score: %s\n%s\n", i, err.Error(), test.midiOut)
			continue
		}

		result := bf.String()

		//    fmt.Println(result)

		//res := strings.Split(result, "=\n")
		//exp := strings.Split(test.expected, "=\n")

		//got = "=SCORE" + strings.Split(strings.TrimSpace(result), "=SCORE")[1]
		got = strings.TrimSpace(result)
		wanted := strings.TrimSpace(test.msklOut)

		if got != wanted {
			t.Errorf("[%v] score unrolled gives \n%s\n\nbut this was expected:\n%s\n%q\nvs\n%q\n", i, got, wanted, got, wanted)
			//      t.Errorf("[%v] score\n%s\n\nunrolled gives \n%q\n\nbut this was expected:\n%q\n", i, test.input, got, wanted)
		}
	}
}