package smfconverter

import (
	"fmt"
	"io"
	"sort"

	"gitlab.com/gomidi/midi"
	"gitlab.com/gomidi/midi/mid"
	"gitlab.com/gomidi/midi/midimessage/channel"
	"gitlab.com/gomidi/midi/midimessage/meta"
	"gitlab.com/gomidi/midi/smf"
	"gitlab.com/gomidi/midi/smf/smfreader"
	"gitlab.com/gomidi/muskel/items"
	"gitlab.com/gomidi/muskel/score"
	"gitlab.com/gomidi/muskel/sketch"
	"gitlab.com/gomidi/muskel/track"
)

type positionedMsg struct {
	absPos uint64
	msg    midi.Message
}

/*
type positionedTempo struct {
	absPos uint64
	bpm    float64
}


type positionedTimeSign struct {
	absPos uint64
	num    uint8
	denom  uint8
}
*/

func ticksTo32ths(ticks4th uint32, ticks uint64) uint {
	return uint(uint64(ticks4th) * 8 / ticks)
}

type positionedMarkersAndCues struct {
	absPos uint64
	text   string
}

type colsKey struct {
	trackNo int16
	channel uint8
}

type sortPositionedMsg []positionedMsg

func (s sortPositionedMsg) Len() int {
	return len(s)
}

func (s sortPositionedMsg) Swap(a, b int) {
	s[a], s[b] = s[b], s[a]
}

func (s sortPositionedMsg) Less(a, b int) bool {
	return s[a].absPos < s[b].absPos
}

func makeBars(ticks4th uint32, timeSigns []positionedMsg) (bars []*sketch.Bar) {
	if len(timeSigns) == 0 {
		return nil
	}

	var sorted = sortPositionedMsg(timeSigns)

	sort.Sort(sorted)

	var no = 1

	//var currentTimeSig [2]uint8
	var lastBar sketch.Bar
	lastBar.No = no
	lastBar.Position = 0
	lastBar.TimeSig = [2]uint8{4 / 4}
	lastBar.TimeSigChange = [2]uint8{4 / 4}
	lastBar.TempoChange = 120

	first := sorted[0]

	if first.absPos == 0 {
		ts := first.msg.(meta.TimeSig)
		lastBar.TimeSig = [2]uint8{ts.Numerator / ts.Denominator}
		lastBar.TimeSigChange = [2]uint8{ts.Numerator / ts.Denominator}
		sorted = sorted[1:]
	}

	bars = append(bars, &lastBar)

	for _, ts := range sorted {
		//_ = ts
		pos := ticksTo32ths(ticks4th, ts.absPos)
		diff := pos - lastBar.Position
		if diff > uint(lastBar.Length32th()) {
			no++
			num := diff / uint(lastBar.Length32th())

			for n := uint(0); n < num; n++ {
				bars = append(bars, &sketch.Bar{No: no, TimeSig: lastBar.TimeSig, Position: lastBar.Position + n*uint(lastBar.Length32th())})
			}
		}

		no++
		tts := ts.msg.(meta.TimeSig)
		lastBar = sketch.Bar{No: no, Position: pos, TimeSig: [2]uint8{tts.Numerator, tts.Denominator}}
		bars = append(bars, &lastBar)
	}

	return bars
}

func Convert(src io.Reader, wr io.Writer, opts ...mid.ReaderOption) error {
	rd := mid.NewReader(opts...)

	//tracks := table.NewTracks(0, nil)

	// line-no 0 is not correct: we need the tracks table first, set the number after that
	//score := table.NewSketch("=SCORE", 0, nil)
	//_ = score
	sc := score.New("import.mid", nil)

	var cols = map[colsKey][]positionedMsg{} // key: trackno and midi channel
	var tracknames = map[int16]string{}
	var trackMetaMsg = map[int16][]positionedMsg{}
	var tempos []positionedMsg
	var timeSigns []positionedMsg
	var markers []positionedMsg

	_ = cols

	// TODO: find the bars and bar changes and the tempo changes
	// colled the messages by track and channel (each track-channel combination becomes a column)
	rd.Msg.Each = func(pos *mid.Position, msg midi.Message) {
		if msg == meta.EndOfTrack {
			// ignore?
			return
		}

		switch v := msg.(type) {
		case meta.Instrument:
			tracknames[pos.Track] = v.Text()
		case meta.TrackSequenceName:
			tracknames[pos.Track] = v.Text()
		case meta.Tempo:
			tempos = append(tempos, positionedMsg{pos.AbsoluteTicks, msg})
		case meta.TimeSig:
			timeSigns = append(timeSigns, positionedMsg{pos.AbsoluteTicks, msg})
		case meta.Marker:
			markers = append(markers, positionedMsg{pos.AbsoluteTicks, msg})
		case meta.Cuepoint:
			markers = append(markers, positionedMsg{pos.AbsoluteTicks, msg})
		default:
			if mt, isMeta := msg.(meta.Message); isMeta {
				mmsg := trackMetaMsg[pos.Track]
				mmsg = append(mmsg, positionedMsg{pos.AbsoluteTicks, mt})
				trackMetaMsg[pos.Track] = mmsg
				return
			}

			if cm, isChannel := msg.(channel.Message); isChannel {
				cl := cols[colsKey{pos.Track, cm.Channel()}]
				cl = append(cl, positionedMsg{pos.AbsoluteTicks, cm})
				cols[colsKey{pos.Track, cm.Channel()}] = cl
				return
			}

			// rest is ignored, dunno if there is any use of it
		}
	}

	innerrd := smfreader.New(src)

	err := rd.ReadSMFFrom(innerrd)

	if err != nil {
		return fmt.Errorf("could not read smf: %v", err)
	}

	ticks4th := innerrd.Header().TimeFormat.(smf.MetricTicks).Ticks4th()

	//tr := map[string][]positionedMsg{}

	for k, msgs := range cols {
		kk := fmt.Sprintf("%s--%v", tracknames[k.trackNo], k.channel)
		var trk track.Track
		trk.Name = kk
		trk.MIDIChannel = int8(k.channel)
		sc.AddTrack(&trk)

		for _, m := range msgs {
			var ev items.Event
			ev.Position = ticksTo32ths(ticks4th, m.absPos)
			// TODO find the difference to exacts ticks and set PosShift
			comp := ev.AbsPosTicks(ticks4th, 0)
			posDiff := int64(comp) - int64(m.absPos)
			switch {
			case posDiff < -15 || posDiff < 15:
				// do nothing
			case posDiff >= 15 && posDiff < 45:
				ev.PosShift = 1
			case posDiff >= 45 && posDiff < 75:
				ev.PosShift = 2
			case posDiff >= 75 && posDiff < 105:
				ev.PosShift = 3
			case posDiff <= -15 && posDiff > -45:
				ev.PosShift = -1
			case posDiff <= -45 && posDiff > -75:
				ev.PosShift = -2
			case posDiff <= -75 && posDiff > -105:
				ev.PosShift = -3
			}

			switch vv := m.msg.(type) {
			case channel.NoteOn:
				var it items.Note
				it.Letter, it.Augmenter, it.Octave = items.KeyToNote(vv.Key())
				it.Dynamic = items.VelocityToDynamic(vv.Velocity())
				it.NoteOn = true
				ev.Item = &it
			case channel.NoteOff:
				var it items.Note
				it.Letter, it.Augmenter, it.Octave = items.KeyToNote(vv.Key())
				it.NoteOff = true
				ev.Item = &it
			case channel.Aftertouch:
				var it items.MIDIAftertouch
				it.Value = vv.Pressure()
				ev.Item = &it
			case channel.Pitchbend:
				var it items.MIDIPitchbend
				it.Value = vv.Value()
				ev.Item = &it
			case channel.ControlChange:
				var it items.MIDICC
				it.Value = vv.Value()
				it.Controller = vv.Controller()
				ev.Item = &it
			case channel.ProgramChange:
				var it items.MIDIProgramChange
				it.Value = vv.Program()
				ev.Item = &it
			case channel.PolyAftertouch:
				var it items.MIDIPolyAftertouch
				it.Value = vv.Pressure()
				it.Key = vv.Key()
				ev.Item = &it
			}
			//tr[kk] = msgs
			sc.Unrolled[kk] = append(sc.Unrolled[kk], &ev)
		}
	}

	// TODO make bars out of the tempos and timesignatures
	// TODO add messages to score cols
	// TODO add tracks and convert track messages to tracks table
	//_ = tracks
	// TODO assign tempo changes

	sc.Bars = makeBars(ticks4th, timeSigns)

	var sortedMarkers = sortPositionedMsg(markers)

	for _, mrk := range sortedMarkers {
		bidx := sc.GetBarIdxOf(ticksTo32ths(ticks4th, mrk.absPos))

		switch v := mrk.msg.(type) {
		case meta.Marker:
			sc.Bars[bidx].Part = v.Text()
		case meta.Cuepoint:
			sc.Bars[bidx].Part = v.Text()
		default:
			panic("must not happen")
		}
	}

	var sortedTempo = sortPositionedMsg(tempos)

	for _, tc := range sortedTempo {
		pos := ticksTo32ths(ticks4th, tc.absPos)
		bidx := sc.GetBarIdxOf(pos)
		if sc.Bars[bidx].Position == pos {
			sc.Bars[bidx].TempoChange = tc.msg.(meta.Tempo).FractionalBPM()
		} else {
			relPos := pos - sc.Bars[bidx].Position
			sc.Bars[bidx].InnerTempoChanges[relPos] = tc.msg.(meta.Tempo).FractionalBPM()
		}
		// TODO in score make new lines for tempochanges within bars and without acompagning notes
	}

	return sc.WriteUnrolled(wr)

	// TODO sort all the tracked messages
}
