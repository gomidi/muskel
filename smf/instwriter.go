package smf

import (
	"fmt"
	"math"
	"math/rand"

	"gitlab.com/gomidi/midi"
	"gitlab.com/gomidi/midi/cc"
	"gitlab.com/gomidi/midi/mid"
	"gitlab.com/gomidi/midi/midimessage/channel"
	"gitlab.com/gomidi/midi/smf"
	"gitlab.com/gomidi/muskel/items"
	"gitlab.com/gomidi/muskel/score"
)

type instrSMFWriter struct {
	instrNo               int
	firstDeltaSet         bool
	delayTicks            int32
	deltaCached           int32
	noteDelay             int32
	noteDelayCompensation int32
	tempoBPM              float64
	resolution            smf.MetricTicks

	smf.Writer
}

func (i *instrSMFWriter) setDelay(num, denom int) {
	// fmt.Printf("#0 setting delay to %v/%v\n", num, denom)
	if num == 0 {
		i.delayTicks = 0
		// fmt.Printf("#1 setting delay to %v\n", i.delayTicks)
		return
	}
	i.delayTicks = int32(math.Round(float64(i.resolution.Ticks4th()) * 4 * float64(num) / float64(denom)))
	// fmt.Printf("#2 setting delay to %v\n", i.delayTicks)
}

func calcNoteDelay(resolution smf.MetricTicks) (delay int32) {
	return int32(math.Round(float64(resolution.Ticks4th()) * 4 / 128))
}

func (i *instrSMFWriter) setlaidBack() {
	// fmt.Println("setlaidBack")
	i.noteDelay = calcNoteDelay(i.resolution)
	i.deltaCached += i.noteDelay
}

func (i *instrSMFWriter) setAhead() {
	// fmt.Println("setAhead")
	i.noteDelay = calcNoteDelay(i.resolution) * (-1)
	i.deltaCached += i.noteDelay
}

func (i *instrSMFWriter) setStraight() {
	// fmt.Println("setStraight")
	i.noteDelay = 0
}

func (i *instrSMFWriter) Write(msg midi.Message) error {
	if i.instrNo < 0 {
		return i.Writer.Write(msg)
	}

	delta := i.deltaCached

	switch msg.(type) {
	case channel.NoteOn, channel.NoteOff, channel.NoteOffVelocity:
		// don't return but continue in this function
		// fmt.Printf("deltaCached is: %v noteDelayCompensation is: %v\n", i.deltaCached, i.noteDelayCompensation)
	default:
		if delta < 0 {
			delta = 0
		}
		i.Writer.SetDelta(uint32(delta))
		i.deltaCached = 0
		return i.Writer.Write(msg)
	}

	if !i.firstDeltaSet {
		delta += i.delayTicks
		i.firstDeltaSet = true
	}

	if delta > 0 && i.noteDelayCompensation != 0 {
		delta += i.noteDelayCompensation
		i.noteDelayCompensation = 0
	}

	if delta < 0 {
		delta = 0
	}

	// fmt.Printf("setting real delta to %v before writing %s\n", delta, msg)
	i.Writer.SetDelta(uint32(delta))
	i.deltaCached = 0
	if i.noteDelay != 0 {
		// remove noteDelayCompensation when we have a negative noteDelay at the start of the piece
		if delta == 0 {
			// fmt.Printf("reset noteDelayCompensation caused by delta == 0\n")
			i.noteDelayCompensation = 0
		} else {
			i.noteDelayCompensation -= i.noteDelay
			// fmt.Printf("setting noteDelayCompensation to %v\n", i.noteDelayCompensation)
		}
		i.noteDelay = 0
	}
	return i.Writer.Write(msg)
}

func (i *instrSMFWriter) SetDelta(ticks uint32) {
	if i.instrNo < 0 {
		i.Writer.SetDelta(ticks)
		return
	}
	t := int32(ticks)

	i.deltaCached = t

	/*
		fmt.Printf("SetDelta(%v) noteDelay: %v, noteDelayComp: %v\n", ticks, i.noteDelay, i.noteDelayCompensation)
		//panic(ticks)

		if t > 0 && i.noteDelayCompensation != 0 {
			t += i.noteDelayCompensation
			i.noteDelayCompensation = 0
		}

		if i.noteDelay != 0 {
			t += i.noteDelay
			i.noteDelayCompensation -= i.noteDelay
			i.noteDelay = 0
		}

		if t < 0 {
			t = 0
		}

		i.Writer.SetDelta(uint32(t))
	*/
}

func calcAdd(distance int64, diff float64) float64 {
	return diff / float64(distance)
}

func calcAdd2(distance int64, diff float64) float64 {
	return diff / (float64(distance) * float64(distance))
}

func halfTonesToPitchbend(halftones float64, _range uint8) int16 {
	res := int16(math.Round(halftones * 8192.0 / float64(_range)))
	if res < -8191 {
		return -8191
	}

	if res > 8192 {
		return 8192
	}

	return res
}

func linearGlissando(wr *mid.SMFWriter, distance int64, noteDiff int64, callback func(val float64)) {

	// pitch0      = f(0)        =  note0
	// pitch       = f(step)     =  note0 + (noteTarget-note0)/distance * step
	// pitchTarget = f(distance) =  note0 + (noteTarget-note0)/distance * distance = note0 + noteTarget - note0 = noteTarget
	// y             f(x)        =  a     +      m                      * x
	// m                         =  (noteTarget-note0)/distance

	m := calcAdd(distance, float64(noteDiff))
	//fmt.Printf("linearGlissando: m = %0.5f\n", m)
	wr.RestoreTimeline()
	//var pb int16

	for step := int64(1); step <= distance; step++ {
		//iw.wr.Forward(0, 1, 32)
		wr.Forward(0, 1, 64)
		callback(m * float64(step))
		//pb = halfTonesToPitchbend(m*float64(step), iw.instr.PitchbendRange)
		//iw.wr.Pitchbend(pb)
	}

	//iw.prevPitchbend = pb
}

func exponentialGlissando(wr *mid.SMFWriter, distance int64, noteDiff int64, callback func(val float64)) {
	// y             f(x)        =  a     +      m                       * x²
	// pitch0      = f(0)        =  note0
	// pitch       = f(step)     =  note0 + (noteTarget-note0)/distance² * step²
	// pitchTarget = f(distance) =  note0 + (noteTarget-note0)/distance² * distance² = note0 + noteTarget - note0 = noteTarget
	// m                         =  (noteTarget-note0)/distance²

	m := calcAdd2(distance, float64(noteDiff))
	wr.RestoreTimeline()
	//var pb int16

	for step := int64(1); step <= distance; step++ {
		//iw.wr.Forward(0, 1, 32)
		wr.Forward(0, 1, 64)
		callback(m * float64(step) * float64(step))
		//pb = halfTonesToPitchbend(m*float64(step)*float64(step), iw.instr.PitchbendRange)
		//iw.wr.Pitchbend(pb)
	}

	//iw.prevPitchbend = pb
}

var dottedLength = [2]uint32{3, 128}

//var dottedLength = [2]uint32{1, 64}

func (iw *instWriter) writeItem(item interface{}, stopNotes func()) (addedNotes []uint8) {
	if item == nil {
		return nil
	}
	switch v := item.(type) {
	case items.GlissStart:
		iw.startGlissandoNote = iw.prevKey
		delete(iw.noteOns, iw.startGlissandoNote)
		stopNotes()
		iw.wr.Pitchbend(0)
		iw.startGlissando = iw.wr.Position()
		iw.inGlissando = true
		iw.wr.BackupTimeline()
		iw.glissandoFunc = linearGlissando
		if v == items.GlissandoExponential {
			iw.glissandoFunc = exponentialGlissando
		}

		addedNotes = append(addedNotes, iw.startGlissandoNote)
	case items.Note:
		vel := v.Velocity
		if vel < 0 {
			vel = iw.prevVel
		}

		iw.prevVel = vel

		vl := uint8(vel + int8(rand.Intn(4)))

		// fmt.Printf("%#v\n", v)

		var n uint8
		n = v.ToMIDI()

		switch v.PosShift {
		case 0:
			iw.p.iw.setStraight()
		case 1:
			iw.p.iw.setlaidBack()
		case -1:
			iw.p.iw.setAhead()
		}

		//	}

		// only add MIDITranspose to Notes not to MIDINotes
		n = uint8(int8(n) + iw.instr.MIDITranspose)
		iw.prevKey = n

		if iw.inGlissando {
			/*
				1. get the distance to the previous note
				2. calc the diff in semitones
				3. make a line of 32ths in steps of the factor
			*/
			//distance := int64(math.Round(float64(iw.wr.Position()+uint64(iw.wr.Delta())-iw.startGlissando) / float64(iw.wr.Ticks32th())))
			distance := int64(math.Round(float64(iw.wr.Position()+uint64(iw.wr.Delta())-iw.startGlissando) / float64(iw.wr.Ticks64th())))
			noteDiff := int64(n) - int64(iw.startGlissandoNote)

			var pb int16

			iw.glissandoFunc(iw.wr, distance, noteDiff, func(vl float64) {
				pb = halfTonesToPitchbend(vl, iw.instr.PitchbendRange)
				iw.wr.Pitchbend(pb)
			})

			iw.prevPitchbend = pb
			/*
				addStep := calc32thsAdd(distance, noteDiff)
				iw.wr.RestoreTimeline()

				for step := int64(1); step <= distance; step++ {
					iw.wr.Forward(0, 1, 32)
					iw.wr.Pitchbend(halfTonesToPitchbend(float64(step)*addStep, iw.instr.PitchbendRange))
				}
			*/

			delete(iw.noteOns, iw.startGlissandoNote)
		}
		stopNotes()

		//fmt.Printf("NoteOn %v\n", n)
		if !iw.inGlissando {
			if iw.startGlissandoNote != 0 {
				iw.startGlissandoNote = 0
				iw.wr.Pitchbend(0)
				iw.prevPitchbend = 0
			}
			iw.wr.NoteOn(n, vl)
			if v.Dotted {
				iw.wr.Plan(0, dottedLength[0], dottedLength[1], iw.wr.Channel.NoteOff(n))
			} else {
				addedNotes = append(addedNotes, n)
			}
		}
		if iw.inGlissando {
			//  set pitchbend back to 0
			iw.inGlissando = false
			if v.Dotted {
				iw.wr.Pitchbend(iw.prevPitchbend)
				iw.wr.Plan(0, dottedLength[0], dottedLength[1], iw.wr.Channel.NoteOff(iw.startGlissandoNote))
				//iw.wr.NoteOff(iw.startGlissandoNote)
			} else {
				addedNotes = append(addedNotes, iw.startGlissandoNote)
			}
		}

		if v.GlissandoStart {
			//iw.startGlissandoDelta = iw.wr.DeltaTime()
			iw.startGlissando = iw.wr.Position()
			iw.startGlissandoNote = n
			iw.inGlissando = true
			iw.glissandoFunc = linearGlissando
			if v.GlissandoExp {
				iw.glissandoFunc = exponentialGlissando
			}
			iw.wr.BackupTimeline()
		}
	case items.MIDINote:
		switch v.PosShift {
		case 0:
			iw.p.iw.setStraight()
		case 1:
			iw.p.iw.setlaidBack()
		case -1:
			iw.p.iw.setAhead()
		}

		iw.inGlissando = false
		stopNotes()
		vel := v.Velocity

		if vel < 0 {
			vel = iw.prevVel
		}

		iw.prevVel = vel

		vl := uint8(vel + int8(rand.Intn(4)))
		n := uint8(v.Note)
		//fmt.Printf("MIDINoteOn %v\n", n)
		if iw.startGlissandoNote != 0 {
			iw.startGlissandoNote = 0
			iw.wr.Pitchbend(0)
		}
		iw.wr.NoteOn(n, vl)
		if v.Dotted {
			iw.wr.Plan(0, dottedLength[0], dottedLength[1], iw.wr.Channel.NoteOff(n))
		} else {
			addedNotes = append(addedNotes, n)
		}
	case items.MIDICC:
		//fmt.Printf("MIDICC %v, %v\n", v[0], v[1])

		if iw.inCCGlissando && v.Controller == iw.CCGlissandoController {
			//distance := int64(math.Round(float64(iw.wr.Position()+uint64(iw.wr.Delta())-iw.startCCGlissando) / float64(iw.wr.Ticks32th())))
			distance := int64(math.Round(float64(iw.wr.Position()+uint64(iw.wr.Delta())-iw.startCCGlissando) / float64(iw.wr.Ticks64th())))
			diff := int64(v.Value) - int64(iw.startCCGlissandoValue)

			iw.CCglissandoFunc(iw.wr, distance, diff, func(vl float64) {
				vll := math.Round(vl + float64(iw.startCCGlissandoValue))
				if vll > 127 {
					vll = 127
				}

				if vll < 0 {
					vll = 0
				}

				iw.wr.ControlChange(iw.CCGlissandoController, uint8(vll))
			})
			iw.inCCGlissando = false
		}

		iw.wr.ControlChange(v.Controller, v.Value)

		if len(v.Tilde) > 0 {
			iw.CCGlissandoController = v.Controller
			iw.startCCGlissando = iw.wr.Position()
			iw.startCCGlissandoValue = v.Value
			iw.inCCGlissando = true
			iw.CCglissandoFunc = linearGlissando
			if v.Tilde == "~~" {
				iw.CCglissandoFunc = exponentialGlissando
			}
			iw.wr.BackupTimeline()
		}

		if v.Dotted && v.Value > 0 {
			iw.wr.Plan(0, dottedLength[0], dottedLength[1], iw.wr.Channel.ControlChange(v.Controller, 0))
		}

	case items.MIDIPitchbend:
		if iw.inPBGlissando {
			//distance := int64(math.Round(float64(iw.wr.Position()+uint64(iw.wr.Delta())-iw.startPBGlissando) / float64(iw.wr.Ticks32th())))
			distance := int64(math.Round(float64(iw.wr.Position()+uint64(iw.wr.Delta())-iw.startPBGlissando) / float64(iw.wr.Ticks64th())))
			diff := int64(v.Value+8191) - int64(iw.startPBGlissandoValue+8191)
			// fmt.Printf("%v - %v = pb diff: %v\n", int8(v.value+8191), int8(iw.startPBGlissandoValue+8191), diff)

			iw.PBglissandoFunc(iw.wr, distance, diff, func(vl float64) {
				vll := math.Round(vl + float64(iw.startPBGlissandoValue))
				// fmt.Printf("gliss raw: %0.2f\n", vll)
				if vll > 8192 {
					vll = 8192
				}

				if vll < -8191 {
					vll = -8191
				}

				// fmt.Printf("gliss normalized: %v\n", int16(vll))

				iw.wr.Pitchbend(int16(vll))
			})
			iw.inPBGlissando = false
		}
		//fmt.Printf("MIDIPitchbend %v, \n", int16(v))
		iw.wr.Pitchbend(v.Value)

		if len(v.Tilde) > 0 {
			iw.startPBGlissando = iw.wr.Position()
			iw.startPBGlissandoValue = v.Value
			iw.inPBGlissando = true
			iw.PBglissandoFunc = linearGlissando
			if v.Tilde == "~~" {
				iw.PBglissandoFunc = exponentialGlissando
			}
			iw.wr.BackupTimeline()
		}

		if v.Dotted && v.Value != 0 {
			iw.wr.Plan(0, dottedLength[0], dottedLength[1], iw.wr.Channel.Pitchbend(0))
		}
	case items.MIDIPolyAftertouch:
		//fmt.Printf("MIDIPolyAftertouch %v, %v\n", v[0], v[1])
		ki := uint8(int8(v.Key) + iw.instr.MIDITranspose)

		if iw.inPATGlissando && v.Key == iw.PATGlissandoKey {
			//distance := int64(math.Round(float64(iw.wr.Position()+uint64(iw.wr.Delta())-iw.startPATGlissando) / float64(iw.wr.Ticks32th())))
			distance := int64(math.Round(float64(iw.wr.Position()+uint64(iw.wr.Delta())-iw.startPATGlissando) / float64(iw.wr.Ticks64th())))
			diff := int64(v.Value) - int64(iw.startPATGlissandoValue)

			iw.PATglissandoFunc(iw.wr, distance, diff, func(vl float64) {
				vll := math.Round(vl + float64(iw.startPATGlissandoValue))
				if vll > 127 {
					vll = 127
				}

				if vll < 0 {
					vll = 0
				}

				iw.wr.PolyAftertouch(iw.PATGlissandoKey, uint8(vll))
			})
			iw.inPATGlissando = false
		}

		iw.wr.PolyAftertouch(ki, v.Value)

		if len(v.Tilde) > 0 {
			iw.PATGlissandoKey = ki
			iw.startPATGlissando = iw.wr.Position()
			iw.startPATGlissandoValue = v.Value
			iw.inPATGlissando = true
			iw.PATglissandoFunc = linearGlissando
			if v.Tilde == "~~" {
				iw.PATglissandoFunc = exponentialGlissando
			}
			iw.wr.BackupTimeline()
		}

		if v.Dotted && v.Value != 0 {
			iw.wr.Plan(0, dottedLength[0], dottedLength[1], iw.wr.Channel.PolyAftertouch(ki, 0))
		}
	case items.MIDIAftertouch:
		//fmt.Printf("Aftertouch %v, \n", uint8(v))
		if iw.inATGlissando {
			//distance := int64(math.Round(float64(iw.wr.Position()+uint64(iw.wr.Delta())-iw.startCCGlissando) / float64(iw.wr.Ticks32th())))
			distance := int64(math.Round(float64(iw.wr.Position()+uint64(iw.wr.Delta())-iw.startATGlissando) / float64(iw.wr.Ticks64th())))
			diff := int64(v.Value) - int64(iw.startATGlissandoValue)

			iw.ATglissandoFunc(iw.wr, distance, diff, func(vl float64) {
				vll := math.Round(vl + float64(iw.startATGlissandoValue))
				if vll > 127 {
					vll = 127
				}

				if vll < 0 {
					vll = 0
				}

				iw.wr.Aftertouch(uint8(vll))
			})
			iw.inATGlissando = false
		}

		iw.wr.Aftertouch(v.Value)

		if len(v.Tilde) > 0 {
			iw.startATGlissando = iw.wr.Position()
			iw.startATGlissandoValue = v.Value
			iw.inATGlissando = true
			iw.ATglissandoFunc = linearGlissando
			if v.Tilde == "~~" {
				iw.ATglissandoFunc = exponentialGlissando
			}
			iw.wr.BackupTimeline()
		}

		if v.Dotted && v.Value > 0 {
			iw.wr.Plan(0, dottedLength[0], dottedLength[1], iw.wr.Channel.Aftertouch(0))
		}
	case items.OSCMessage:
		// TODO implement OSC
	case items.NTuple:
		// definition of a tuple
		// we need the complete time length over which the tuple is spread
		// it is not that easy, since we need to define the noteoff of the
		// last element. we probably also might want to cut of things as
		// the next event appears
		// on the good side: this should allow us also to define lengths
		// like the conventional western standard notation
		// on the other hand: it could be a nice effect to have the last
		// note playing until the next event appears (we always could set
		// a rest if needed) and for percussions it would not matter
		// (same problem as with holding percussive notes, maybe we'll want
		// to have a quick note on the playing notes when they'll want to end
		// and carry that over through the processing
		// maybe even an additional parameter could be put into Forward, like
		// "go Forward these steps but while you are on the way, ensure the
		// following notes will end on these times: ..."
		// then we could predefine 32th lengths for drum notes and the needed
		// length for ntuples. but then there would be no way to hold the
		// last note of a ntuple.
		// the total length of the ntuple could be specified by the ending position, e.g.
		// {a',e'',f}3&
		// would mean: the total duration is from the current position until 3&

		iw.inGlissando = false

		//length := uint32(v.EndPos - iw.lastNum32th)
		length := uint32(v.EndPos - iw.currentDistanceToStartOfBarIn32th)
		//fmt.Printf("NTuple endpos: %v lastnum32th: %v length: %v\n", v.EndPos, iw.lastNum32th, length)

		switch v.PosShift {
		case 0:
			iw.p.iw.setStraight()
		case 1:
			iw.p.iw.setlaidBack()
		case -1:
			iw.p.iw.setAhead()
		}

		var delta uint32

		for _, it := range v.Items {
			if item != items.Hold {
				if delta > 0 {
					iw.wr.Forward(0, delta, uint32(len(v.Items))*32)
				}
				added := iw.writeItem(it, stopNotes)
				delta = 0

				for _, addn := range added {
					iw.noteOns[addn] = true
				}
			}
			delta += length
		}
		iw.wr.Forward(0, delta, uint32(len(v.Items))*32)

		switch v.PosShift {
		case 0:
			iw.p.iw.noteDelayCompensation = 0
		case 1:
			iw.p.iw.noteDelayCompensation = calcNoteDelay(iw.p.iw.resolution) * (-1)
		case -1:
			iw.p.iw.noteDelayCompensation = calcNoteDelay(iw.p.iw.resolution)
		}
		stopNotes()

		iw.lastNum32th = iw.currentDistanceToStartOfBarIn32th + uint(length)

		iw.wr.Undefined(0, nil)
		// iw.p.iw.Writer

		/*

			switch vv := v.Items[len(v.Items)-1].(type) {
			case items.MIDINote:
				if vv.Dotted {
					// 3, 128
					iw.lastNum128th -= uint(length)*4 - 3
				}
			case items.Note:
				if vv.Dotted {
					iw.lastNum128th -= uint(length)*4 - 3
				}
			}
		*/

	case items.MultiItem:
		stopNotes()

		for _, it := range v {
			// ignore the returned velocities
			added := iw.writeItem(it, func() {})
			addedNotes = append(addedNotes, added...)
		}
		return addedNotes

	case items.Lyric:
		stopNotes()
		iw.wr.Lyric(string(v))
	default:
		if item == items.Hold {
			return nil
		}

		if item == items.Rest {
			iw.inGlissando = false
			stopNotes()
			iw.p.iw.setStraight()
		} else {
			panic(fmt.Sprintf("unsupported item: %#v", v))
		}
	}

	return addedNotes
}

type instWriter struct {
	p                  *writer
	wr                 *mid.SMFWriter
	instr              *score.Track
	emptyBars          uint32
	lastItem           interface{}
	prevVel            int8
	noteOns            map[uint8]bool
	repeatFrom         int
	repeatingBars      int
	currentlyRepeating int
	totalBeginning     bool
	insideRepitition   bool
	lastNum32th        uint
	inGlissando        bool
	startGlissando     uint64
	startGlissandoNote uint8
	prevKey            uint8
	glissandoFunc      func(wr *mid.SMFWriter, distance int64, noteDiff int64, callback func(val float64))
	prevPitchbend      int16

	CCGlissandoController uint8
	startCCGlissando      uint64
	startCCGlissandoValue uint8
	inCCGlissando         bool
	CCglissandoFunc       func(wr *mid.SMFWriter, distance int64, noteDiff int64, callback func(val float64))

	PATGlissandoKey        uint8
	startPATGlissando      uint64
	startPATGlissandoValue uint8
	inPATGlissando         bool
	PATglissandoFunc       func(wr *mid.SMFWriter, distance int64, noteDiff int64, callback func(val float64))

	startATGlissando      uint64
	startATGlissandoValue uint8
	inATGlissando         bool
	ATglissandoFunc       func(wr *mid.SMFWriter, distance int64, noteDiff int64, callback func(val float64))

	startPBGlissando      uint64
	startPBGlissandoValue int16
	inPBGlissando         bool
	PBglissandoFunc       func(wr *mid.SMFWriter, distance int64, noteDiff int64, callback func(val float64))

	currentDistanceToStartOfBarIn32th uint

	lastBarNo int

	// currentScale        *Scale
}

func newTrackSMFWriter(p *writer, wr *mid.SMFWriter, instr *score.Track) *instWriter {
	return &instWriter{
		p:              p,
		wr:             wr,
		instr:          instr,
		noteOns:        map[uint8]bool{},
		totalBeginning: true,
	}
}

func (iw *instWriter) writeIntro() {
	//fmt.Printf("Channel: %v\n", iw.instr.MIDIChannel)
	iw.wr.SetChannel(uint8(iw.instr.MIDIChannel))
	//fmt.Printf("Track: %#v\n", iw.instr.Name)
	iw.wr.Program(iw.instr.Name)
	iw.wr.TrackSequenceName(iw.instr.Name)
	iw.wr.Instrument(iw.instr.Name)

	if iw.instr.MIDIBank >= 0 {
		//fmt.Printf("Bank: %v\n", iw.instr.MIDIBank)
		iw.wr.ControlChange(cc.BankSelectMSB, uint8(iw.instr.MIDIBank))
	}

	if iw.instr.MIDIProgram >= 0 {
		//fmt.Printf("ProgramChange: %v\n", iw.instr.MIDIProgram)
		iw.wr.ProgramChange(uint8(iw.instr.MIDIProgram))
	}

	if iw.instr.MIDIVolume >= 0 {
		//fmt.Printf("Volume: %v\n", iw.instr.MIDIVolume)
		iw.wr.ControlChange(cc.VolumeMSB, uint8(iw.instr.MIDIVolume))
	}

	pitchBendRange := uint8(2)
	if iw.instr.PitchbendRange > 0 {
		pitchBendRange = iw.instr.PitchbendRange
	}

	iw.wr.PitchBendSensitivityRPN(pitchBendRange, 0)
}

func (iw *instWriter) writeUnrolled() {
	var lastEv *score.Event

	iw.prevVel = 63
	iw.lastBarNo = 0

	stopNotes := func() {
		for nt, isOn := range iw.noteOns {
			if isOn {
				//fmt.Printf("NoteOff %v\n", nt)
				iw.wr.NoteOff(nt)
				delete(iw.noteOns, nt)
			}
		}
	}

	isHolding := false

	for _, ev := range iw.instr.Unrolled {

		if ev.Item == nil {
			continue
		}

		diffBars := ev.BarNo
		if lastEv != nil {
			diffBars -= iw.lastBarNo
		}

		if ev.Item == items.Rest && len(iw.noteOns) == 0 {
			continue
		}

		iw.currentDistanceToStartOfBarIn32th = ev.DistanceToStartOfBarIn32th

		if ev.Item != items.Hold {
			if lastEv != nil && iw.lastBarNo == ev.BarNo {
				isHolding = false
				//fmt.Printf("FORWARD(0, %v-%v, 32)\n", ev.DistanceToStartOfBarIn32th, iw.lastNum32)
				iw.wr.Forward(0, uint32(ev.DistanceToStartOfBarIn32th-iw.lastNum32th), 32)
			} else {
				if !isHolding && len(iw.noteOns) > 0 {
					iw.wr.Forward(uint32(diffBars), 0, 32)
					stopNotes()
					//fmt.Printf("FORWARD(0, %v, 32)\n", ev.DistanceToStartOfBarIn32th)
					iw.wr.Forward(0, uint32(ev.DistanceToStartOfBarIn32th), 32)
				} else {
					//fmt.Printf("FORWARD(%v, %v, 32)\n", diffBars, ev.DistanceToStartOfBarIn32th)
					iw.wr.Forward(uint32(diffBars), uint32(ev.DistanceToStartOfBarIn32th), 32)
				}
				isHolding = false

			}

			// iw.scaleAt(ev.BarNo)

			var addedNotes []uint8
			// fmt.Printf("writing item: %#v\n", ev.Item)
			addedNotes = iw.writeItem(ev.Item, stopNotes)

			for _, nt := range addedNotes {
				iw.noteOns[nt] = true
			}
			iw.lastItem = ev.Item
			lastEv = ev
			iw.lastBarNo = lastEv.BarNo
			switch iw.lastItem.(type) {
			case items.NTuple:
				// iw.lastNum32 is set by the ntuple itself
				//if iw.lastNum32th > iw.
				barLen := uint(iw.p.score.Bars[ev.BarNo].Length32th())
				// fmt.Printf("post NTuple: iw.lastNum32th: %v, barlen: %v\n", iw.lastNum32th, barLen)
				if iw.lastNum32th >= barLen {
					iw.lastNum32th -= barLen
					iw.lastBarNo += 1
				}
			default:
				iw.lastNum32th = lastEv.DistanceToStartOfBarIn32th
			}

			// fmt.Printf("setting iw.lastNum32th to %v\n", iw.lastNum32th)
		} else {
			iw.wr.FinishPlanned()
			isHolding = true
			iw.lastNum32th = 0
		}
	}

	// fmt.Printf("len bars: %v\n", len(iw.p.score.Bars))
	lastBar := iw.p.score.Bars[len(iw.p.score.Bars)-1]
	barDiff := lastBar.BarNo
	if lastEv != nil {
		barDiff -= iw.lastBarNo
	}

	// fmt.Printf("barDiff: %v iw.lastBarNo: %v\n", barDiff, iw.lastBarNo)
	if barDiff > 0 {
		iw.wr.Forward(uint32(barDiff), 0, 32)
	} else {
		iw.wr.Forward(1, 0, 32)
	}
	iw.wr.FinishPlanned()
	stopNotes()
	iw.p.writeEndOfTrack(lastBar.BarNo)
}

func (iw *instWriter) writeTrack() {

	iw.writeIntro()
	iw.prevVel = 60
	iw.emptyBars = 0
	iw.noteOns = map[uint8]bool{}
	iw.lastItem = nil

	iw.writeUnrolled()
}
