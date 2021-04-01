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
	//t.Skip()
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
|TRACK | bass | git |
|Channel | 9   | 8   |

|=SCORE | bass | git |
|#
|1      | c    | d   |
|#
|1      | *    |     |
|2      |      | *   |
`,
			`
[0] channel.NoteOn channel 8 key 48 dyn =
[3840] channel.NoteOff channel 8 key 48
`,
			`
| TRACK           | bass--8                                  | git--7                                   |
| --------------- | ---------------------------------------- | ---------------------------------------- |
|  File           |                                          |                                          |
|  Channel        | 9                                        | 8                                        |
|  Program        |                                          |                                          |
|  Bank           |                                          |                                          |
|  Transpose      |                                          |                                          |
|  Volume         |                                          |                                          |
|  Delay          |                                          |                                          |
|  PitchbendRange | 2                                        | 2                                        |
|  VelocityScale  | min:1 max:127 random:4 step:15 center:63 | min:1 max:127 random:4 step:15 center:63 |
|  Ambitus        |                                          |                                          |
|  Import         |                                          |                                          |

| =SCORE     | bass--8 | git--7 |
| ---------- | ------- | ------ |
|  # @120.00 |         |        |
|     1      | c_      | d_     |
|  #         |         |        |
|     1      | _c      |        |
|     2      |         | _d     |
`,
			2,
			false,
			false,
		},
		{ // 1
			`
|TRACK | bass | git | git2 |
|Channel | 9   | 8   | 7   |
|Import  |     |     | git |

|=SCORE | bass | git |
|#
|1      | c    | d   |
|# 3/4
|1      | *    |     |
|2      |      | *   |
`,
			`
[0] channel.NoteOn channel 8 key 48 dyn =
[3840] channel.NoteOff channel 8 key 48
`,
			`
| TRACK           | bass--8                                  | git--7                                   | git2--6                                  |
| --------------- | ---------------------------------------- | ---------------------------------------- | ---------------------------------------- |
|  File           |                                          |                                          |                                          |
|  Channel        | 9                                        | 8                                        | 7                                        |
|  Program        |                                          |                                          |                                          |
|  Bank           |                                          |                                          |                                          |
|  Transpose      |                                          |                                          |                                          |
|  Volume         |                                          |                                          |                                          |
|  Delay          |                                          |                                          |                                          |
|  PitchbendRange | 2                                        | 2                                        | 2                                        |
|  VelocityScale  | min:1 max:127 random:4 step:15 center:63 | min:1 max:127 random:4 step:15 center:63 | min:1 max:127 random:4 step:15 center:63 |
|  Ambitus        |                                          |                                          |                                          |
|  Import         |                                          |                                          |                                          |

| =SCORE     | bass--8 | git--7 | git2--6 |
| ---------- | ------- | ------ | ------- |
|  # @120.00 |         |        |         |
|     1      | c_      | d_     | d_      |
|  # 3/4     |         |        |         |
|     1      | _c      |        |         |
|     2      |         | _d     | _d      |
`,
			2,
			false,
			false,
		},
		{ // 2
			`
|TRACK | bass | git |
|Channel | 9   | 8   |

|=SCORE | bass | git |
|# 3/4 @110
|1      | c    | d   |
|# 4/4 @200
|1      | *    |     |
|2      |      | *   |
`,
			`
[0] channel.NoteOn channel 8 key 48 dyn =
[2880] channel.NoteOff channel 8 key 48
`,
			`
| TRACK           | bass--8                                  | git--7                                   |
| --------------- | ---------------------------------------- | ---------------------------------------- |
|  File           |                                          |                                          |
|  Channel        | 9                                        | 8                                        |
|  Program        |                                          |                                          |
|  Bank           |                                          |                                          |
|  Transpose      |                                          |                                          |
|  Volume         |                                          |                                          |
|  Delay          |                                          |                                          |
|  PitchbendRange | 2                                        | 2                                        |
|  VelocityScale  | min:1 max:127 random:4 step:15 center:63 | min:1 max:127 random:4 step:15 center:63 |
|  Ambitus        |                                          |                                          |
|  Import         |                                          |                                          |

| =SCORE         | bass--8 | git--7 |
| -------------- | ------- | ------ |
|  # 3/4 @110.00 |         |        |
|     1          | c_      | d_     |
|  # 4/4 @200.00 |         |        |
|     1          | _c      |        |
|     2          |         | _d     |
`,
			2,
			false,
			false,
		},
		{ // 3
			`
|TRACK | bass | git |
|Channel | 9   | 8   |

|=SCORE | bass | git |
|# 3/4 @110
|1      | c    | d   |
|2 @115 | *    |     |
|# 4/4 @200
|2      |      | *   |
`,
			`
[0] channel.NoteOn channel 8 key 48 dyn =
[960] channel.NoteOff channel 8 key 48
`,
			`
| TRACK           | bass--8                                  | git--7                                   |
| --------------- | ---------------------------------------- | ---------------------------------------- |
|  File           |                                          |                                          |
|  Channel        | 9                                        | 8                                        |
|  Program        |                                          |                                          |
|  Bank           |                                          |                                          |
|  Transpose      |                                          |                                          |
|  Volume         |                                          |                                          |
|  Delay          |                                          |                                          |
|  PitchbendRange | 2                                        | 2                                        |
|  VelocityScale  | min:1 max:127 random:4 step:15 center:63 | min:1 max:127 random:4 step:15 center:63 |
|  Ambitus        |                                          |                                          |
|  Import         |                                          |                                          |

| =SCORE         | bass--8 | git--7 |
| -------------- | ------- | ------ |
|  # 3/4 @110.00 |         |        |
|     1          | c_      | d_     |
|     2 @115.00  |         |        |
|     2          | _c      |        |
|  # 4/4 @200.00 |         |        |
|     2          |         | _d     |
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

		var cv *smfimport.Importer

		if test.includeControlChange {
			cv = smfimport.New("convert-main", bytes.NewReader(midibf.Bytes()))
		} else {
			cv = smfimport.New("convert-main", bytes.NewReader(midibf.Bytes()), smfimport.IgnoreCC())
		}

		var bf bytes.Buffer

		err = cv.WriteUnrolled(&bf)

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
