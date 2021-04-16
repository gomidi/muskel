package smfimport

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"sort"
	"strings"

	"gitlab.com/gomidi/midi"
	"gitlab.com/gomidi/midi/reader"

	"gitlab.com/gomidi/midi/midimessage/channel"
	"gitlab.com/gomidi/midi/midimessage/meta"
	"gitlab.com/gomidi/midi/smf"
	"gitlab.com/gomidi/muskel/items"
	"gitlab.com/gomidi/muskel/score"
	"gitlab.com/gomidi/muskel/track"
	"gitlab.com/gomidi/quantizer"
)

type source struct {
	didRead bool
	io.Reader
	midiReader   *reader.Reader
	lastTick     uint64
	filename     string
	tracknames   map[int16]string
	tempos       []positionedMsg
	timeSigns    []positionedMsg
	markers      []positionedMsg
	trackMetaMsg map[int16][]positionedMsg
	cols         map[colsKey][]positionedMsg // key: trackno and midi channel
	ticks4th     uint32
}

type config struct {
	readerOpts      []func(*reader.Reader)
	ignoreCC        bool
	monoTracks      map[int16]bool
	keysTracks      map[int16]bool
	splitEventTypes bool
	quantize        bool
	detectDelay     bool
	noMicroTiming   bool
	noVelocity      bool
}

type destination struct {
	score     *score.Score
	noteCols  map[colsKey][]positionedMsg // key: trackno and midi channel
	lyricCols map[colsKey][]positionedMsg
	ccCols    map[colsKey][]positionedMsg
	dataCols  map[colsKey][]positionedMsg
}

type Importer struct {
	config
	source
	destination
	//*smfQuantizer
	quantizer *quantizer.Quantizer
}

type Option func(*Importer)

// SplitEventTypes creates separate tracks for the following types
// Types                                   | Track-Suffix
// midi notes, pitchbend                   |
// lyrics                                  | lyrics
// cc messages, aftertouch, polyaftertouch | controller
// other                                   | data
func SplitEventTypes() Option {
	return func(i *Importer) {
		i.splitEventTypes = true
	}
}

func IgnoreCC() Option {
	return func(i *Importer) {
		i.ignoreCC = true
	}
}

func MonoTracks(monoTracks ...int) Option {
	return func(i *Importer) {
		for _, t := range monoTracks {
			i.monoTracks[int16(t)] = true
		}
	}
}

func KeysTracks(keysTracks ...int) Option {
	return func(i *Importer) {
		for _, t := range keysTracks {
			i.keysTracks[int16(t)] = true
		}
	}
}

func ReaderOptions(opts ...func(*reader.Reader)) Option {
	return func(i *Importer) {
		i.readerOpts = opts
	}
}

func NoMicroTiming() Option {
	return func(i *Importer) {
		i.config.noMicroTiming = true
	}
}

func NoVelocity() Option {
	return func(i *Importer) {
		i.config.noVelocity = true
	}
}

func Quantize(opts ...func(*reader.Reader)) Option {
	return func(i *Importer) {
		i.config.quantize = true
	}
}

func DetectDelay(opts ...func(*reader.Reader)) Option {
	return func(i *Importer) {
		i.detectDelay = true
	}
}

func New(filename string, rd io.Reader, opts ...Option) *Importer {
	im := &Importer{}

	im.destination.noteCols = map[colsKey][]positionedMsg{}  // key: trackno and midi channel
	im.destination.ccCols = map[colsKey][]positionedMsg{}    // key: trackno and midi channel
	im.destination.lyricCols = map[colsKey][]positionedMsg{} // key: trackno and midi channel
	im.destination.dataCols = map[colsKey][]positionedMsg{}  // key: trackno and midi channel

	im.source.filename = filename
	im.source.cols = map[colsKey][]positionedMsg{} // key: trackno and midi channel
	im.source.tracknames = map[int16]string{}
	im.source.trackMetaMsg = map[int16][]positionedMsg{}

	im.config.monoTracks = map[int16]bool{}
	im.config.keysTracks = map[int16]bool{}

	for _, opt := range opts {
		opt(im)
	}

	if im.config.quantize {
		bt, err := ioutil.ReadAll(rd)

		if err != nil {
			return nil
		}

		in := bytes.NewReader(bt)

		var q_opts []quantizer.Option
		if im.detectDelay {
			q_opts = append(q_opts, quantizer.DetectDelay())
		}
		im.quantizer, err = quantizer.New(in, q_opts...)
		if err != nil {
			return nil
		}
		im.source.Reader = bytes.NewReader(bt)
	} else {
		im.source.Reader = rd
	}
	im.config.readerOpts = append(im.config.readerOpts, reader.NoLogger(), reader.Each(im.registerMsg), reader.SMFHeader(im.onHeader))
	im.source.midiReader = reader.New(im.readerOpts...)
	return im
}

func (c *Importer) WriteUnrolled(wr io.Writer) error {
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

func (c *Importer) WriteTracksAndScoreTable(trackswr, scorewr io.Writer) error {
	err := c.readSMF()
	if err != nil {
		return err
	}
	c.setTracks()
	c.setBars()
	c.setMarkers()
	c.setTempos()
	c.score.IsUnrolled = true
	return c.score.WriteTracksAndScoreTable(trackswr, scorewr)
}

func (c *Importer) mapCols() {
	if !c.config.splitEventTypes {
		c.destination.noteCols = c.source.cols
		return
	}

	// SplitEventTypes creates separate tracks for the following types
	// Types                                   | Track-Suffix
	// midi notes, pitchbend                   |
	// lyrics                                  | lyrics
	// cc messages, aftertouch, polyaftertouch | controller
	// other                                   | data
	for k, msgs := range c.source.cols {
		for _, msg := range msgs {
			switch msg.msg.(type) {
			case channel.NoteOn, channel.NoteOff, channel.NoteOffVelocity, channel.Pitchbend:
				c.destination.noteCols[k] = append(c.destination.noteCols[k], msg)
			case meta.Lyric:
				c.destination.lyricCols[k] = append(c.destination.lyricCols[k], msg)
			case channel.ControlChange, channel.Aftertouch, channel.PolyAftertouch:
				c.destination.ccCols[k] = append(c.destination.ccCols[k], msg)
			default:
				c.destination.dataCols[k] = append(c.destination.dataCols[k], msg)
			}
		}
	}
}

func (c *Importer) onHeader(hd smf.Header) {
	tf := hd.TimeFormat
	//fmt.Printf("%v\n", tf)

	if ticks, isMetric := tf.(smf.MetricTicks); isMetric {
		c.ticks4th = ticks.Ticks4th()
		//fmt.Printf("c.ticks4th: %v\n", c.ticks4th)
	}
}

func (c *Importer) readSMF() error {
	if c.didRead {
		return nil
	}

	c.didRead = true

	err := reader.ReadSMF(c.midiReader, c.Reader)
	if err != nil {
		return fmt.Errorf("could not read smf: %v", err)
	}
	if c.ticks4th == 0 {
		return fmt.Errorf("does not support SMPTE time format at the moment")
	}

	c.mapCols()
	return nil
}

func (c *Importer) addBar(b *items.Bar) {
	//fmt.Printf("adding bar #%v %s\n", b.No, b)
	c.score.Bars = append(c.score.Bars, b)
}

func (c *Importer) addMissingBars(until uint64, lastBar *items.Bar) (no int) {
	no = lastBar.No

	pos := c.ticksTo32ths(until)
	diff := pos - lastBar.Position
	//fmt.Printf("addMissingBars until: %v lastBar %v diff: %v lastbarLength: %v num: %v\n", until, lastBar.No, diff, lastBar.Length32th(), diff/lastBar.Length32th())

	num := diff / uint(lastBar.Length32th())

	if diff%uint(lastBar.Length32th()) != 0 {
		num++
	}

	for n := uint(1); n < num; n++ {
		no++
		b := items.NewBar()
		b.No = no
		b.TimeSig = lastBar.TimeSig
		b.Position = lastBar.Position + n*uint(lastBar.Length32th())
		c.addBar(b)
	}

	return
}

func (c *Importer) setBars() {
	var no = 1

	var lastBar = items.NewBar()
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
	first := sorted[0]

	if first.absPos == 0 {
		ts := first.msg.(meta.TimeSig)
		lastBar.TimeSig = [2]uint8{ts.Numerator, ts.Denominator}
		lastBar.TimeSigChange = [2]uint8{ts.Numerator, ts.Denominator}
		sorted = sorted[1:]
	}

	c.addBar(lastBar)

	for _, ts := range sorted {
		no = c.addMissingBars(ts.absPos, lastBar)
		pos := c.ticksTo32ths(ts.absPos)
		no++
		tts := ts.msg.(meta.TimeSig)
		b := items.NewBar()
		b.No = no
		b.Position = pos
		b.TimeSig = [2]uint8{tts.Numerator, tts.Denominator}
		b.TimeSigChange = [2]uint8{tts.Numerator, tts.Denominator}
		lastBar = b
		c.addBar(lastBar)
	}

	c.addMissingBars(c.lastTick, lastBar)
}

func (c *Importer) msgToEvent(m positionedMsg, isDrumTrack bool) *items.Event {
	var ev items.Event
	ev.Position = c.ticksTo32ths(m.absPos)
	comp := ev.AbsPosTicks(c.ticks4th, 0)

	if !c.noMicroTiming {
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

		if c.config.quantize {
			ev.PosShift = int(m.shift)
		}
	}

	switch vv := m.msg.(type) {
	case channel.NoteOn:
		if isDrumTrack {
			var it items.MIDINote
			it.Dotted = "::"
			it.Note = int8(vv.Key())
			if !c.noVelocity {
				it.Dynamic = items.VelocityToDynamic(vv.Velocity())
			}
			ev.Item = &it
		} else {
			var it items.Note
			it.Letter, it.Augmenter, it.Octave = items.KeyToNote(vv.Key())
			if !c.noVelocity {
				it.Dynamic = items.VelocityToDynamic(vv.Velocity())
			}
			it.NoteOn = true
			ev.Item = &it
		}
	case channel.NoteOff:
		if isDrumTrack {
			return nil
		} else {
			var it items.Note
			it.Letter, it.Augmenter, it.Octave = items.KeyToNote(vv.Key())
			it.NoteOff = true
			ev.Item = &it
			ev.PosShift = 0
		}
	case channel.Aftertouch:
		var it items.MIDIAftertouch
		it.Value = vv.Pressure()
		ev.Item = &it
	case channel.Pitchbend:
		var it items.MIDIPitchbend
		it.Value = vv.Value()
		ev.Item = &it
	case channel.ControlChange:
		if c.ignoreCC {
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

func (c *Importer) _setTracks(prefix string, cols map[colsKey][]positionedMsg) {
	for k, msgs := range cols {
		if len(msgs) == 0 {
			continue
		}
		kk := fmt.Sprintf(prefix+"%s--%v", c.tracknames[k.trackNo], k.channel)
		//var trk track.Track
		trk := track.New(kk)
		trk.Name = kk
		trk.MIDIChannel = int8(k.channel)

		// get delay for tracks
		if c.config.quantize && c.config.detectDelay {
			delay := c.quantizer.GetDelay(k.trackNo, k.channel)
			in128ths := c.ticksTo128ths(uint64(math.Abs(float64(delay))))
			//_ = delay
			_ = in128ths

			if in128ths > 0 {
				del := int(in128ths)
				if delay < 0 {
					del = del * (-1)
				}
				trk.Delay = [2]int{del, 128}
			}
		}

		//trk.MIDIBank = -1
		//trk.MIDIVolume = -1

		c.score.AddTrack(trk)

		isMono := c.monoTracks[k.trackNo]
		isDrum := c.keysTracks[k.trackNo]
		var lastRestPos uint
		var lastRestShift int
		var lastRest bool

		/*
			if isMono {
				fmt.Println("isMono: " + kk)
			}
		*/

		for _, m := range msgs {
			ev := c.msgToEvent(m, isDrum)
			if ev == nil {
				continue
			}
			if nt, ok := ev.Item.(*items.Note); ok && isMono {
				if nt.NoteOff {
					//ev.Item = items.Rest
					lastRestPos = ev.Position
					lastRestShift = ev.PosShift
					lastRest = true
					continue
				}

				if nt.NoteOn {
					nt.NoteOn = false
					ev.Item = nt

					if lastRest && lastRestPos != ev.Position {
						evpre := ev.Dup()
						evpre.Position = lastRestPos
						evpre.PosShift = lastRestShift
						evpre.Item = items.Rest
						c.score.Unrolled[kk] = append(c.score.Unrolled[kk], evpre)
					}

					lastRest = false
				}

			}

			c.score.Unrolled[kk] = append(c.score.Unrolled[kk], ev)
		}

		if lastRest {
			ev := &items.Event{}
			ev.Position = lastRestPos
			ev.PosShift = lastRestShift
			ev.Item = items.Rest
			c.score.Unrolled[kk] = append(c.score.Unrolled[kk], ev)
		}

		if isMono {
			c.score.RemoveObsoletRestsFromUnrolledTrack(kk)
		}
	}
}

func (c *Importer) setTracks() {
	c.destination.score = score.New(c.source.filename, nil, score.NoEmptyLines())

	c._setTracks("", c.destination.noteCols)

	if c.splitEventTypes {
		if len(c.destination.ccCols) > 0 {
			c._setTracks("cc-", c.destination.ccCols)
		}

		if len(c.destination.lyricCols) > 0 {
			c._setTracks("lyrics-", c.destination.lyricCols)
		}

		if len(c.destination.dataCols) > 0 {
			c._setTracks("data-", c.destination.dataCols)
		}
	}
}

func (c *Importer) registerMsg(pos *reader.Position, msg midi.Message) {
	if pos.AbsoluteTicks > c.lastTick {
		c.lastTick = pos.AbsoluteTicks
	}

	if msg == meta.EndOfTrack {
		return
	}

	switch v := msg.(type) {
	case meta.Instrument:
		c.tracknames[pos.Track] = strings.ReplaceAll(v.Text(), " ", "-")
	case meta.TrackSequenceName:
		c.tracknames[pos.Track] = strings.ReplaceAll(v.Text(), " ", "-")
	case meta.Tempo:
		c.tempos = append(c.tempos, positionedMsg{pos.AbsoluteTicks, msg, 0})
	case meta.TimeSig:
		c.timeSigns = append(c.timeSigns, positionedMsg{pos.AbsoluteTicks, msg, 0})
	case meta.Marker:
		c.markers = append(c.markers, positionedMsg{pos.AbsoluteTicks, msg, 0})
	case meta.Cuepoint:
		c.markers = append(c.markers, positionedMsg{pos.AbsoluteTicks, msg, 0})
	default:
		if mt, isMeta := msg.(meta.Message); isMeta {
			mmsg := c.trackMetaMsg[pos.Track]
			mmsg = append(mmsg, positionedMsg{pos.AbsoluteTicks, mt, 0})
			c.trackMetaMsg[pos.Track] = mmsg
			return
		}

		if cm, isChannel := msg.(channel.Message); isChannel {
			absPos := pos.AbsoluteTicks
			var shift int8
			if c.config.quantize && c.quantizer != nil {
				_absPos, _ := c.quantizer.Quantize(pos.Track, cm.Channel(), absPos)
				diff := _absPos - int64(absPos)
				in128ths := c.ticksTo128ths(uint64(math.Abs(float64(diff))))
				_ = in128ths
				switch in128ths {
				case 1, 2, 3:
					shift = int8(in128ths)
				default:
					if in128ths > 3 {
						shift = 3
					}
				}
				if diff < 0 {
					shift = shift * (-1)
				}
				// TODO: set the diff as <<< or >>> depending on length and minus sign
				absPos = uint64(_absPos)

			}
			cl := c.source.cols[colsKey{pos.Track, cm.Channel()}]
			cl = append(cl, positionedMsg{absPos, cm, shift})
			c.source.cols[colsKey{pos.Track, cm.Channel()}] = cl
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

func (c *Importer) getBarOf(pos uint) *items.Bar {
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

func (c *Importer) ticksTo64ths(ticks uint64) uint {
	if ticks == 0 {
		return 0
	}

	//fmt.Printf("ticks4th: %v ticks: %v\n", c.ticks4th, ticks)
	return uint(ticks / uint64(c.ticks4th/16))
}

func (c *Importer) ticksTo128ths(ticks uint64) uint {
	if ticks == 0 {
		return 0
	}

	//fmt.Printf("ticks4th: %v ticks: %v\n", c.ticks4th, ticks)
	return uint(ticks / uint64(c.ticks4th/32))
}
