package smfimport

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

type Importer struct {
	score        *score.Score
	rd           *mid.Reader
	src          io.Reader
	cols         map[colsKey][]positionedMsg // key: trackno and midi channel
	tracknames   map[int16]string
	trackMetaMsg map[int16][]positionedMsg
	tempos       []positionedMsg
	timeSigns    []positionedMsg
	markers      []positionedMsg
	ticks4th     uint32
	IgnoreCC     bool
	lastTick     uint64
}

func New(fname string, src io.Reader, opts ...mid.ReaderOption) *Importer {
	opts = append(opts, mid.NoLogger())
	return &Importer{
		rd:           mid.NewReader(opts...),
		src:          src,
		score:        score.New(fname, nil),
		cols:         map[colsKey][]positionedMsg{}, // key: trackno and midi channel
		tracknames:   map[int16]string{},
		trackMetaMsg: map[int16][]positionedMsg{},
	}
}

func (c *Importer) WriteMsklTo(wr io.Writer) error {
	err := c.readSMF()
	if err != nil {
		return err
	}
	c.setTracks()
	c.setBars()
	c.setMarkers()
	c.setTempos()
	c.score.IsUnrolled = true
	return c.score.WriteUnrolled(wr)
}

func (c *Importer) readSMF() error {
	c.rd.Msg.Each = c.registerMsg
	innerrd := smfreader.New(c.src)
	err := c.rd.ReadSMFFrom(innerrd)
	if err != nil {
		return fmt.Errorf("could not read smf: %v", err)
	}
	tf := innerrd.Header().TimeFormat
	//fmt.Printf("%v\n", tf)

	if ticks, isMetric := tf.(smf.MetricTicks); isMetric {
		c.ticks4th = ticks.Ticks4th()
		//fmt.Printf("c.ticks4th: %v\n", c.ticks4th)
		return nil
	}
	return fmt.Errorf("does not support SMPTE time format at the moment")

}

func (c *Importer) addBar(b *sketch.Bar) {
	//fmt.Printf("adding bar #%v %s\n", b.No, b)
	c.score.Bars = append(c.score.Bars, b)
}

func (c *Importer) addMissingBars(until uint64, lastBar *sketch.Bar) (no int) {
	//lastTick
	//lastPos := c.ticksTo32ths(c.lastTick)
	no = lastBar.No

	pos := c.ticksTo32ths(until)
	diff := pos - lastBar.Position

	if diff > uint(lastBar.Length32th()) {
		num := diff / uint(lastBar.Length32th())

		for n := uint(1); n < num; n++ {
			no++
			b := sketch.NewBar()
			b.No = no
			b.TempoChange = lastBar.TempoChange
			b.TimeSig = lastBar.TimeSig
			b.Position = lastBar.Position + n*uint(lastBar.Length32th())
			//fmt.Printf("add missing bar: #%v\n", no)
			c.addBar(b)
		}
	}
	return
}

func (c *Importer) setBars() {
	var no = 1

	var lastBar = sketch.NewBar()
	lastBar.No = no
	lastBar.Position = 0
	lastBar.TimeSig = [2]uint8{4, 4}
	lastBar.TimeSigChange = [2]uint8{4, 4}
	lastBar.TempoChange = 120

	if len(c.timeSigns) == 0 {
		//fmt.Printf("no time signature; adding #%v\n", lastBar.No)
		c.addBar(lastBar)
		c.addMissingBars(c.lastTick, lastBar)
		return
	}

	var sorted = sortPositionedMsg(c.timeSigns)

	sort.Sort(sorted)

	//var currentTimeSig [2]uint8

	first := sorted[0]

	if first.absPos == 0 {
		ts := first.msg.(meta.TimeSig)
		lastBar.TimeSig = [2]uint8{ts.Numerator, ts.Denominator}
		lastBar.TimeSigChange = [2]uint8{ts.Numerator, ts.Denominator}
		sorted = sorted[1:]
	}

	//fmt.Printf("add first bar; adding #%v\n", lastBar.No)
	c.addBar(lastBar)

	for _, ts := range sorted {
		//_ = ts
		no = c.addMissingBars(ts.absPos, lastBar)
		pos := c.ticksTo32ths(ts.absPos)
		/*
			diff := pos - lastBar.Position
			if diff > uint(lastBar.Length32th()) {
				no++
				num := diff / uint(lastBar.Length32th())

				for n := uint(0); n < num; n++ {
					b := sketch.NewBar()
					b.No = no
					b.TimeSig = lastBar.TimeSig
					b.Position = lastBar.Position + n*uint(lastBar.Length32th())
					c.addBar(b)
				}
			}
		*/

		no++
		tts := ts.msg.(meta.TimeSig)
		b := sketch.NewBar()
		b.No = no
		b.Position = pos
		b.TimeSig = [2]uint8{tts.Numerator, tts.Denominator}
		b.TimeSigChange = [2]uint8{tts.Numerator, tts.Denominator}
		b.TempoChange = lastBar.TempoChange
		lastBar = b
		//fmt.Printf("add time sig change bar; adding #%v\n", lastBar.No)
		c.addBar(lastBar)
	}

	c.addMissingBars(c.lastTick, lastBar)
}

func (c *Importer) msgToEvent(m positionedMsg) *items.Event {
	var ev items.Event
	ev.Position = c.ticksTo32ths(m.absPos)
	comp := ev.AbsPosTicks(c.ticks4th, 0)
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
		if c.IgnoreCC {
			return nil
		}
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
	default:
		return nil
	}

	//fmt.Printf("converted %v @ %v to %v @ %v\n", m.msg, m.absPos, ev.String(), ev.Position)

	return &ev
}

func (c *Importer) setTracks() {
	for k, msgs := range c.cols {
		kk := fmt.Sprintf("%s--%v", c.tracknames[k.trackNo], k.channel)
		var trk track.Track
		trk.Name = kk
		trk.MIDIChannel = int8(k.channel)
		c.score.AddTrack(&trk)

		for _, m := range msgs {
			ev := c.msgToEvent(m)
			if ev != nil {
				//fmt.Printf("[%v] adding event %q to %q\n", ev.Position, ev.String(), kk)
				c.score.Unrolled[kk] = append(c.score.Unrolled[kk], ev)
			}
		}
	}

}

func (c *Importer) registerMsg(pos *mid.Position, msg midi.Message) {
	if pos.AbsoluteTicks > c.lastTick {
		c.lastTick = pos.AbsoluteTicks
	}

	if msg == meta.EndOfTrack {
		// ignore?
		return
	}

	switch v := msg.(type) {
	case meta.Instrument:
		c.tracknames[pos.Track] = v.Text()
	case meta.TrackSequenceName:
		c.tracknames[pos.Track] = v.Text()
	case meta.Tempo:
		c.tempos = append(c.tempos, positionedMsg{pos.AbsoluteTicks, msg})
	case meta.TimeSig:
		c.timeSigns = append(c.timeSigns, positionedMsg{pos.AbsoluteTicks, msg})
	case meta.Marker:
		c.markers = append(c.markers, positionedMsg{pos.AbsoluteTicks, msg})
	case meta.Cuepoint:
		c.markers = append(c.markers, positionedMsg{pos.AbsoluteTicks, msg})
	default:
		if mt, isMeta := msg.(meta.Message); isMeta {
			mmsg := c.trackMetaMsg[pos.Track]
			mmsg = append(mmsg, positionedMsg{pos.AbsoluteTicks, mt})
			c.trackMetaMsg[pos.Track] = mmsg
			return
		}

		if cm, isChannel := msg.(channel.Message); isChannel {
			cl := c.cols[colsKey{pos.Track, cm.Channel()}]
			cl = append(cl, positionedMsg{pos.AbsoluteTicks, cm})
			c.cols[colsKey{pos.Track, cm.Channel()}] = cl
			return
		}

		// rest is ignored, dunno if there is any use of it
		return
	}
}

func (c *Importer) setMarkers() {
	var sortedMarkers = sortPositionedMsg(c.markers)
	for _, mrk := range sortedMarkers {
		b := c.getBarOf(c.ticksTo32ths(mrk.absPos))
		switch v := mrk.msg.(type) {
		case meta.Marker:
			b.Part = v.Text()
		case meta.Cuepoint:
			b.Part = v.Text()
		default:
			panic("must not happen")
		}
	}
}

func (c *Importer) getBarOf(pos uint) *sketch.Bar {
	bidx := c.score.GetBarIdxOf(pos)
	if bidx >= 0 {
		return c.score.Bars[bidx]
	}
	return nil
}

func (c *Importer) setTempos() {
	var sortedTempo = sortPositionedMsg(c.tempos)
	for _, tc := range sortedTempo {
		pos := c.ticksTo32ths(tc.absPos)
		b := c.getBarOf(pos)
		if b == nil {
			//fmt.Printf("could not find bar of pos %v\n", pos)
			continue
		}
		if b.Position == pos {
			b.TempoChange = tc.msg.(meta.Tempo).FractionalBPM()
		} else {
			relPos := pos - b.Position
			//fmt.Printf("setting InnerTempoChanges[%v] to %0.2f\n", relPos, tc.msg.(meta.Tempo).FractionalBPM())
			b.InnerTempoChanges[relPos] = tc.msg.(meta.Tempo).FractionalBPM()
		}
	}

}

func (c *Importer) ticksTo32ths(ticks uint64) uint {
	if ticks == 0 {
		return 0
	}

	//fmt.Printf("ticks4th: %v ticks: %v\n", c.ticks4th, ticks)
	return uint(ticks / uint64(c.ticks4th/8))
}
