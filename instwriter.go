package muskel

import (
	"fmt"
	"math"
	"math/rand"

	"gitlab.com/gomidi/midi/cc"
	"gitlab.com/gomidi/midi/mid"
)

func calc32thsAdd(distance32th int64, notediff int8) float64 {
	return float64(notediff) / float64(distance32th)
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

func (iw *instWriter) writeItem(item interface{}, stopNotes func()) (addedNotes []uint8) {
	if item == nil {
		return nil
	}
	switch v := item.(type) {
	case Note:
		vel := v.velocity
		if vel < 0 {
			vel = iw.prevVel
		}

		iw.prevVel = vel

		vl := uint8(vel + int8(rand.Intn(4)))
		// only add MIDITranspose to Notes not to MIDINotes
		n := uint8(int8(v.toMIDI()) + iw.instr.MIDITranspose)

		if iw.inGlissando {
			/*
				1. get the distance to the previous note
				2. calc the diff in semitones
				3. make a line of 32ths in steps of the factor
			*/
			distance := int64(math.Round(float64(iw.wr.Position()+uint64(iw.wr.Delta())-iw.startGlissando) / float64(iw.wr.Ticks32th())))
			noteDiff := int8(n) - int8(iw.startGlissandoNote)
			addStep := calc32thsAdd(distance, noteDiff)
			iw.wr.RestoreTimeline()

			for step := int64(1); step <= distance; step++ {
				iw.wr.Forward(0, 1, 32)
				iw.wr.Pitchbend(halfTonesToPitchbend(float64(step)*addStep, iw.instr.PitchbendRange))
			}

			delete(iw.noteOns, iw.startGlissandoNote)
			addedNotes = append(addedNotes, iw.startGlissandoNote)
			_ = addStep
		}
		stopNotes()

		//fmt.Printf("NoteOn %v\n", n)
		if !iw.inGlissando {
			if iw.startGlissandoNote != 0 {
				iw.startGlissandoNote = 0
				iw.wr.Pitchbend(0)
			}
			iw.wr.NoteOn(n, vl)
			if v.dotted {
				iw.wr.Plan(0, 3, 128, iw.wr.Channel.NoteOff(n))
			} else {
				addedNotes = append(addedNotes, n)
			}
		}
		if iw.inGlissando {
			//  set pitchbend back to 0
			iw.inGlissando = false
		}

		if v.glissandoStart {
			//iw.startGlissandoDelta = iw.wr.DeltaTime()
			iw.startGlissando = iw.wr.Position()
			iw.startGlissandoNote = n
			iw.inGlissando = true
			iw.wr.BackupTimeline()
		}
	case MIDINote:
		iw.inGlissando = false
		stopNotes()
		vel := v.velocity

		if vel < 0 {
			vel = iw.prevVel
		}

		iw.prevVel = vel

		vl := uint8(vel + int8(rand.Intn(4)))
		n := uint8(v.note)
		//fmt.Printf("MIDINoteOn %v\n", n)
		if iw.startGlissandoNote != 0 {
			iw.startGlissandoNote = 0
			iw.wr.Pitchbend(0)
		}
		iw.wr.NoteOn(n, vl)
		if v.dotted {
			iw.wr.Plan(0, 3, 128, iw.wr.Channel.NoteOff(n))
		} else {
			addedNotes = append(addedNotes, n)
		}
	case MIDICC:
		//fmt.Printf("MIDICC %v, %v\n", v[0], v[1])
		iw.wr.ControlChange(v[0], v[1])
	case MIDIPitchbend:
		//fmt.Printf("MIDIPitchbend %v, \n", int16(v))
		iw.wr.Pitchbend(int16(v))
	case MIDIPolyAftertouch:
		//fmt.Printf("MIDIPolyAftertouch %v, %v\n", v[0], v[1])
		iw.wr.PolyAftertouch(v[0], v[1])
	case MIDIAftertouch:
		//fmt.Printf("Aftertouch %v, \n", uint8(v))
		iw.wr.Aftertouch(uint8(v))
	case OSCMessage:
		//	case *PatternCall:
		/*
			Was wir wollen:

			1. wir suchen die position des nächsten events
			2. dann kennen wir die entfernung vom pattern startpunkt zum nächsten event
			3. dann gehen wir durch die pattern-events takt für takt bis zur stelle, wo das nächste event kommt
			4. dann brechen wir das pattern ab.
		*/
		// TODO do something with the pc events
	/*
		barevts := v.Events

		_ = barevts
	*/
	case NTuple:
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

		length := uint32(v.endPos - iw.lastNum32)

		//fmt.Printf("total length: %v\n", length)

		//lengthPerItem := uint32(4.0 * math.Round(float64(length)/float64(len(v.items))))
		//lengthPerItem := uint32(math.Round(float64(length) / float64(len(v.items))))

		var delta uint32

		for _, it := range v.items {
			if item != Hold {
				if delta > 0 {
					//iw.wr.Forward(0, delta, 128)
					//iw.wr.Forward(0, delta, 32)
					//fmt.Printf("lengthPerItem: %v\n", float64(delta)/float64(uint32(len(v.items))*32))
					iw.wr.Forward(0, delta, uint32(len(v.items))*32)
				}
				added := iw.writeItem(it, stopNotes)
				delta = 0

				for _, addn := range added {
					iw.noteOns[addn] = true
				}
			}
			delta += length // lengthPerItem
		}
		iw.wr.Forward(0, delta, uint32(len(v.items))*32)
		stopNotes()
		iw.lastNum32 += uint(length)

	case MultiItem:
		stopNotes()

		for _, it := range v {
			// ignore the returned velocities
			added := iw.writeItem(it, func() {})
			addedNotes = append(addedNotes, added...)
		}
		return addedNotes

	case rest:
		//fmt.Printf("got REST\n")
		iw.inGlissando = false
		stopNotes()
	case hold:
		return nil
	case Lyric:
		stopNotes()
		iw.wr.Lyric(string(v))
	default:
		panic(fmt.Sprintf("unsupported item: %#v", v))
	}

	return addedNotes
}

type instWriter struct {
	p                   *SMFWriter
	wr                  *mid.SMFWriter
	instr               *Instrument
	emptyBars           uint32
	lastItem            interface{}
	prevVel             int8
	noteOns             map[uint8]bool
	repeatFrom          int
	repeatingBars       int
	currentlyRepeating  int
	totalBeginning      bool
	insideRepitition    bool
	lastNum32           uint
	inGlissando         bool
	startGlissando      uint64
	startGlissandoNote  uint8
	startGlissandoDelta uint32
}

func newInstrumentSMFWriter(p *SMFWriter, wr *mid.SMFWriter, instr *Instrument) *instWriter {
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

	iw.wr.PitchBendSensitivityRPN(iw.instr.PitchbendRange, 0)
}

func (iw *instWriter) writeUnrolled() {
	var lastEv *Event

	iw.prevVel = 63

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

	for _, ev := range iw.instr.unrolled {
		if ev.Item == nil {
			continue
		}
		diffBars := ev.BarNo
		if lastEv != nil {
			diffBars -= lastEv.BarNo
		}

		if ev.Item == Rest && len(iw.noteOns) == 0 {
			continue
		}

		if ev.Item != Hold {
			if lastEv != nil && lastEv.BarNo == ev.BarNo {
				isHolding = false
				//fmt.Printf("FORWARD(0, %v, 32)\n", ev.DistanceToStartOfBarIn32th-lastEv.DistanceToStartOfBarIn32th)
				iw.wr.Forward(0, uint32(ev.DistanceToStartOfBarIn32th-lastEv.DistanceToStartOfBarIn32th), 32)
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

			var addedNotes []uint8
			//fmt.Printf("writing item: %#v\n", ev.Item)
			addedNotes = iw.writeItem(ev.Item, stopNotes)

			for _, nt := range addedNotes {
				iw.noteOns[nt] = true
			}
			iw.lastItem = ev.Item
			lastEv = ev
		} else {
			iw.wr.FinishPlanned()
			isHolding = true
		}
	}

	iw.wr.Forward(1, 0, 32)
	iw.wr.FinishPlanned()
	stopNotes()
}

func (iw *instWriter) writeTrack() {

	iw.writeIntro()
	iw.prevVel = 60
	iw.emptyBars = 0
	iw.noteOns = map[uint8]bool{}
	iw.lastItem = nil

	iw.writeUnrolled()
}

type Instrument struct {
	Name           string
	MIDIChannel    int8
	MIDIProgram    int8
	MIDIVolume     int8
	MIDIBank       int8
	MIDITranspose  int8 // only transposes Notes not MIDINotes
	PitchbendRange uint8
	events         []BarEvents // in the order of bars
	colWidth       int
	unrolled       []*Event
}

func (i *Instrument) Dup() *Instrument {
	return &Instrument{
		Name:           i.Name,
		MIDIChannel:    i.MIDIChannel,
		MIDIProgram:    i.MIDIProgram,
		MIDIVolume:     i.MIDIVolume,
		MIDIBank:       i.MIDIBank,
		PitchbendRange: i.PitchbendRange,
		MIDITranspose:  i.MIDITranspose,
		colWidth:       i.colWidth, // ? set to 0??
		events:         i.events,
	}
}

func (i *Instrument) pad(s string) string {
	return pad(s, i.colWidth)
}

func (i *Instrument) calcColWidth(unrolled bool) {
	i.colWidth = len(i.Name)

	if unrolled {
		for _, ev := range i.unrolled {
			l := len(ev.originalData)
			if l > i.colWidth {
				i.colWidth = l
			}
		}
	} else {
		for _, be := range i.events {
			for _, ev := range be {
				l := len(ev.originalData)
				if l > i.colWidth {
					i.colWidth = l
				}
			}
		}
	}

	if i.colWidth < 2 {
		i.colWidth = 2
	}
}
