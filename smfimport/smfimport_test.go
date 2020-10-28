package smfimport

import (
	"bytes"
	"testing"

	"gitlab.com/gomidi/midi/gm"
	"gitlab.com/gomidi/midi/smf"
	"gitlab.com/gomidi/midi/writer"
)

func TestImport(t *testing.T) {
	var mid bytes.Buffer
	wr := writer.NewSMF(&mid, 4)
	ticks := wr.Header().TimeFormat.(smf.MetricTicks)
	_ = ticks

	writer.Meter(wr, 3, 4)
	writer.TempoBPM(wr, 110.0)
	wr.SetDelta(ticks.Ticks4th() * 3)
	writer.Meter(wr, 4, 4)
	wr.SetDelta(ticks.Ticks4th() * 4)
	writer.TempoBPM(wr, 130)
	wr.SetDelta(ticks.Ticks4th() * 4)
	writer.Meter(wr, 3, 4)
	writer.TempoBPM(wr, 130)
	//wr.SetDelta(ticks.Ticks4th() * 3)
	//wr.SetDelta(ticks.Ticks4th() * 6)
	//writer.Meter(wr, 3, 4)
	writer.EndOfTrack(wr)

	writer.TrackSequenceName(wr, "piano")
	wr.SetChannel(1)
	writer.ProgramChange(wr, 2)
	//writer.ControlChange(wr, cc.VolumeMSB, 90)
	wr.SetDelta(ticks.Ticks8th())
	writer.NoteOn(wr, 60, 100)
	wr.SetDelta(ticks.Ticks8th())
	writer.NoteOn(wr, 74, 80)
	wr.SetDelta(ticks.Ticks8th())
	writer.NoteOff(wr, 60)
	wr.SetDelta(ticks.Ticks8th() * 3)
	writer.NoteOff(wr, 74)
	writer.EndOfTrack(wr)

	writer.TrackSequenceName(wr, "vocal")
	wr.SetChannel(2)
	writer.ProgramChange(wr, 20)
	//writer.ControlChange(wr, cc.VolumeMSB, 80)
	wr.SetDelta(ticks.Ticks4th() * 3)
	writer.NoteOn(wr, 65, 120)
	wr.SetDelta(ticks.Ticks4th() * 4)
	writer.NoteOn(wr, 43, 40)
	writer.NoteOff(wr, 65)
	wr.SetDelta(ticks.Ticks4th() * 5)
	writer.NoteOff(wr, 43)
	writer.NoteOn(wr, 69, 100)
	//wr.SetDelta(ticks.Ticks4th() * 5)
	wr.SetDelta(ticks.Ticks4th())
	writer.NoteOff(wr, 69)
	//wr.SetDelta(ticks.Ticks8th() * 7)
	writer.EndOfTrack(wr)

	writer.TrackSequenceName(wr, "drums")
	wr.SetChannel(9)
	writer.NoteOn(wr, gm.DrumKey_AcousticBassDrum.Key(), 120)
	wr.SetDelta(ticks.Ticks8th())
	writer.NoteOff(wr, gm.DrumKey_AcousticBassDrum.Key())
	writer.NoteOn(wr, gm.DrumKey_AcousticSnare.Key(), 40)
	wr.SetDelta(ticks.Ticks4th())
	writer.NoteOff(wr, gm.DrumKey_AcousticSnare.Key())
	writer.EndOfTrack(wr)

	im := New("testfile", &mid)
	var out bytes.Buffer

	im.WriteMsklTo(&out, MonoTracks(2), DrumTracks(3))

	expected := `| TRACK           | drums--9                                 | piano--1                                 | vocal--2                                 |
| --------------- | ---------------------------------------- | ---------------------------------------- | ---------------------------------------- |
|  File           |                                          |                                          |                                          |
|  Channel        | 10                                       | 2                                        | 3                                        |
|  Program        |                                          |                                          |                                          |
|  Bank           |                                          |                                          |                                          |
|  Transpose      |                                          |                                          |                                          |
|  Volume         |                                          |                                          |                                          |
|  Delay          |                                          |                                          |                                          |
|  PitchbendRange | 2                                        | 2                                        | 2                                        |
|  VelocityScale  | min:1 max:127 random:4 step:15 center:63 | min:1 max:127 random:4 step:15 center:63 | min:1 max:127 random:4 step:15 center:63 |
|  Ambitus        |                                          |                                          |                                          |
|  Import         |                                          |                                          |                                          |

| =SCORE         | drums--9   | piano--1 | vocal--2 |
| -------------- | ---------- | -------- | -------- |
|  # 3/4 @110.00 |            |          |          |
|     1          | MN35++++:: | PC(2)    | PC(20)   |
|     1&         | MN38--::   | c'++_    |          |
|     2          |            | d"+_     |          |
|     2&         |            | _c'      |          |
|  # 4/4         |            |          |          |
|     1          |            | _d"      | f'++++   |
|  # @130.00     |            |          |          |
|     1          |            |          | G--      |
|  # 3/4         |            |          |          |
|     2          |            |          | a'++     |
|     3          |            |          | *        |

`

	if out.String() != expected {
		t.Errorf("got\n%s\nexpected\n%s\n", out.String(), expected)
	}

}
