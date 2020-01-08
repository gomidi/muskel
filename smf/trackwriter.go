package smf

import (
	"gitlab.com/gomidi/midi/cc"
	"gitlab.com/gomidi/muskel/items"
	"gitlab.com/gomidi/muskel/sketch"
	"gitlab.com/gomidi/muskel/track"
)

type TrackWriter struct {
	track *track.Track
	*smfWriter
	events             []*sketch.Event
	emptyBars          uint32
	lastItem           interface{}
	repeatFrom         int
	repeatingBars      int
	currentlyRepeating int
	totalBeginning     bool
	insideRepitition   bool
	lastBarNo          int
	// currentScale        *Scale
}

func newTrackWriter(wr *smfWriter, events []*sketch.Event, track *track.Track) *TrackWriter {
	return &TrackWriter{
		smfWriter:      wr,
		events:         events,
		totalBeginning: true,
		track:          track,
	}
}

func (iw *TrackWriter) lastTick() uint64 {
	return iw.smfWriter.delay.lastTick
}

func (iw *TrackWriter) writeIntro() {
	//fmt.Printf("Channel: %v\n", iw.instr.MIDIChannel)

	iw.SetChannel(uint8(iw.track.MIDIChannel))
	//fmt.Printf("Track: %#v\n", iw.instr.Name)
	iw.Program(iw.track.Name)
	iw.TrackSequenceName(iw.track.Name)
	iw.Instrument(iw.track.Name)

	if iw.track.MIDIBank >= 0 {
		//fmt.Printf("Bank: %v\n", iw.instr.MIDIBank)
		iw.ControlChange(cc.BankSelectMSB, uint8(iw.track.MIDIBank))
	}

	if iw.track.MIDIProgram >= 0 {
		//fmt.Printf("ProgramChange: %v\n", iw.instr.MIDIProgram)
		iw.ProgramChange(uint8(iw.track.MIDIProgram))
	}

	if iw.track.MIDIVolume >= 0 {
		//fmt.Printf("Volume: %v\n", iw.instr.MIDIVolume)
		iw.ControlChange(cc.VolumeMSB, uint8(iw.track.MIDIVolume))
	}

	pitchBendRange := uint8(2)
	if iw.track.PitchbendRange > 0 {
		pitchBendRange = iw.track.PitchbendRange
	}

	iw.PitchBendSensitivityRPN(pitchBendRange, 0)
}

func (iw *TrackWriter) forward(bars, n32th uint32) {
	if bars == 0 && n32th == 0 {
		return
	}
	// fmt.Printf("forward: %v %v\n", bars, n32th)
	iw.Forward(bars, n32th, 32)
}

func (tw *TrackWriter) barOfEvent(ev *sketch.Event) (bar *sketch.Bar) {
	if len(tw.smfWriter.score.Bars) == 0 {
		panic("no bars")
	}

	idx := -1

	for _, b := range tw.smfWriter.score.Bars {
		//fmt.Printf("b.Position: %v ev.Position: %v\n", b.Position, ev.Position)
		if b.Position > ev.Position {
			break
		}
		idx++
	}

	if idx < 0 {
		idx = 0
	}

	return tw.smfWriter.score.Bars[idx]
}

func (iw *TrackWriter) writeUnrolled() {
	var lastEv *sketch.Event

	items.MIDITrack.Reset()

	items.MIDITrack.Track.MIDITranspose = iw.track.MIDITranspose
	items.MIDITrack.Track.PitchbendRange = iw.track.PitchbendRange
	items.MIDITrack.VelocityScale = iw.track.VelocityScale
	items.MIDITrack.WriterWrapper = iw.smfWriter.delay
	iw.lastBarNo = 0

	isHolding := true

	//fmt.Printf("TrackWriter#writeUnrolled: %v\n", iw.events)

	//for _, ev := range iw.track.Unrolled {
	for _, ev := range iw.events {

		bar := iw.barOfEvent(ev)
		barNo := bar.No
		//fmt.Printf("Bar: %v at %v event: %s at: %v\n", barNo, bar.Position, ev, ev.Position)

		diffBars := barNo //ev.BarNo
		if lastEv != nil {
			diffBars -= iw.lastBarNo
		}

		distanceToStartOfBarIn32th := ev.Position - bar.Position

		// fmt.Printf("ev.BarNo: %v; diffbars: %v; ev.Item: %T\n", ev.BarNo, diffBars, ev.Item)

		if ev.Item == nil || ev.Item == items.Hold {
			continue
		}

		if items.MIDITrack.CurrentNTuple != nil /* && ev.Item != items.Hold */ {
			endPos := distanceToStartOfBarIn32th //   ev.DistanceToStartOfBarIn32th
			if diffBars > 0 {
				endPos += uint(iw.smfWriter.score.Bars[iw.lastBarNo].Length32th())
			}

			addNotes := items.MIDITrack.WritecurrentNTuple(iw, endPos)

			for _, nt := range addNotes {
				items.MIDITrack.NoteOns[nt] = true
			}

			barLen := uint(iw.smfWriter.score.Bars[iw.lastBarNo].Length32th())
			// fmt.Printf("post NTuple: iw.lastNum32th: %v, barlen: %v\n", iw.lastNum32th, barLen)
			if items.MIDITrack.LastNum32th >= barLen {
				items.MIDITrack.LastNum32th -= barLen
				iw.lastBarNo += 1
			}

		}

		if ev.Item == items.Rest && (len(items.MIDITrack.NoteOns) == 0 || (diffBars > 0 && !isHolding)) {
			continue
		}

		items.MIDITrack.CurrentDistanceToStartOfBarIn32th = distanceToStartOfBarIn32th //  ev.DistanceToStartOfBarIn32th

		mi, isMi := ev.Item.(*items.MultiItem)

		if isMi && mi.IsHold() {
			isHolding = true
			items.MIDITrack.LastNum32th = 0
		}

		if lastEv != nil && iw.lastBarNo == barNo /*ev.BarNo*/ {
			//if DEBUG {
			//fmt.Printf("[1] FORWARD(0, %v-%v, 32), lastitem: %v\n", ev.DistanceToStartOfBarIn32th, items.MIDITrack.LastNum32th, iw.lastItem)
			//}
			//iw.wr.Forward(0, uint32(ev.DistanceToStartOfBarIn32th-items.MIDITrack.LastNum32th), 32)
			iw.forward(0, uint32(distanceToStartOfBarIn32th-items.MIDITrack.LastNum32th))
		} else {
			// fmt.Printf("isHolding: %v, len(noteOns): %v\n", isHolding, len(iw.noteOns))
			if !isHolding && len(items.MIDITrack.NoteOns) > 0 {
				//fmt.Printf("[4] FORWARD(1, 0, 32)\n")
				//iw.wr.Forward(1, 0, 32)
				iw.forward(1, 0)
				items.MIDITrack.StopNotes(iw)
				//fmt.Printf("[2] FORWARD(%v, %v, 32)\n", uint32(diffBars)-1, ev.DistanceToStartOfBarIn32th)
				//iw.wr.Forward(uint32(diffBars)-1, uint32(ev.DistanceToStartOfBarIn32th), 32)
				iw.forward(uint32(diffBars)-1, uint32(distanceToStartOfBarIn32th))
			} else {
				//fmt.Printf("[3] FORWARD(%v, %v, 32)\n", diffBars, ev.DistanceToStartOfBarIn32th)
				//iw.wr.Forward(uint32(diffBars), uint32(ev.DistanceToStartOfBarIn32th), 32)
				iw.forward(uint32(diffBars), uint32(distanceToStartOfBarIn32th))
			}

		}

		// iw.scaleAt(ev.BarNo)

		var addedNotes []uint8
		//fmt.Printf("writing item: %#v\n", ev.Item)

		addedNotes = items.MIDITrack.WriteItem(iw, ev.Item) // iw.writeItem(ev.Item)
		//		addedNotes = iw.writeItem(ev.Item)

		for _, nt := range addedNotes {
			items.MIDITrack.NoteOns[nt] = true
		}

		lastEv = ev
		iw.lastBarNo = barNo
		items.MIDITrack.LastNum32th = distanceToStartOfBarIn32th

		if !isMi || mi.IsHold() {
			isHolding = true
		}

		if !isMi || !mi.IsHold() {
			iw.lastItem = ev.Item
		}

		// fmt.Printf("setting iw.lastNum32th to %v\n", iw.lastNum32th)

	}

	if items.MIDITrack.CurrentNTuple != nil {
		endPos := uint(iw.smfWriter.score.Bars[iw.lastBarNo].Length32th())

		addNotes := items.MIDITrack.WritecurrentNTuple(iw, endPos)

		for _, nt := range addNotes {
			items.MIDITrack.NoteOns[nt] = true
		}

		barLen := uint(iw.smfWriter.score.Bars[iw.lastBarNo].Length32th())
		// fmt.Printf("post NTuple: iw.lastNum32th: %v, barlen: %v\n", iw.lastNum32th, barLen)
		if items.MIDITrack.LastNum32th >= barLen {
			items.MIDITrack.LastNum32th -= barLen
			iw.lastBarNo += 1
		}

	}

	// fmt.Printf("len bars: %v\n", len(iw.p.score.Bars))
	//lastBar := iw.p.score.Bars[len(iw.p.score.Bars)-1]
	lastBar := iw.smfWriter.score.Bars[len(iw.smfWriter.score.Bars)-1]
	barDiff := lastBar.No
	if lastEv != nil {
		barDiff -= iw.lastBarNo
	}

	// fmt.Printf("barDiff: %v iw.lastBarNo: %v\n", barDiff, iw.lastBarNo)
	if barDiff > 0 {
		//iw.wr.Forward(uint32(barDiff), 0, 32)
		//fmt.Printf("[5] FORWARD(%v, %v, 32)\n", uint32(barDiff), 0)
		iw.forward(uint32(barDiff), 0)
	} else {
		//iw.wr.Forward(1, 0, 32)
		//fmt.Printf("[6] FORWARD(%v, %v, 32)\n", 1, 0)
		iw.forward(1, 0)
	}
	iw.finishplanned()
	items.MIDITrack.StopNotes(iw.smfWriter)
	iw.smfWriter.writeEndOfTrack(lastBar.No)
}

func (iw *TrackWriter) finishplanned() {
	// fmt.Println("finish planned")
	iw.FinishPlanned()
}

func (iw *TrackWriter) writeTrack() {

	iw.writeIntro()

	iw.emptyBars = 0
	iw.lastItem = nil

	iw.writeUnrolled()
}
