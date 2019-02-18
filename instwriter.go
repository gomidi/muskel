package muskel

import (
	"fmt"
	"math/rand"

	"gitlab.com/gomidi/midi/cc"
	"gitlab.com/gomidi/midi/gm"
	"gitlab.com/gomidi/midi/mid"
)

func (iw *instWriter) writeItem(item interface{}, stopNotes func()) (addedNotes []uint8) {
	switch v := item.(type) {
	case Note:
		stopNotes()
		vel := int8(-1)
		if v.dynamic != "" {
			vel = velocityFromDynamic(v.dynamic)
		}

		if vel < 0 {
			vel = iw.prevVel
		}

		iw.prevVel = vel

		vl := uint8(vel + int8(rand.Intn(4)))
		n := noteToMIDI(v)
		//fmt.Printf("NoteOn %v %v\n", n, vl)
		fmt.Printf("NoteOn %v\n", n)
		iw.wr.NoteOn(n, vl)
		addedNotes = append(addedNotes, n)
	case DrumNote:
		var note gm.DrumKey
		switch v.name {
		case "kd":
			note = gm.DrumKey_BassDrum1
		case "sn":
			note = gm.DrumKey_ElectricSnare
		case "ho":
			note = gm.DrumKey_OpenHiHat
		case "hc":
			note = gm.DrumKey_ClosedHiHat
		case "rd":
			note = gm.DrumKey_RideCymbal1
		case "tb":
			note = gm.DrumKey_Tambourine
		case "tl":
			note = gm.DrumKey_LowTom
		case "th":
			note = gm.DrumKey_HighTom
		case "tm":
			note = gm.DrumKey_HiMidTom
		case "sh":
			note = gm.DrumKey_CrashCymbal1
		}

		if note != 0 {
			stopNotes()

			vel := int8(-1)
			if v.dynamic != "" {
				vel = velocityFromDynamic(v.dynamic)
			}

			if vel < 0 {
				vel = iw.prevVel
			}

			iw.prevVel = vel
			vl := uint8(vel + int8(rand.Intn(4)))

			//fmt.Printf("Drums NoteOn %v %v\n", note.Key(), vl)
			fmt.Printf("Drums NoteOn %v\n", note.Key())
			iw.wr.NoteOn(note.Key(), vl)
			addedNotes = append(addedNotes, note.Key())
		}
	case MIDICC:
		fmt.Printf("MIDICC %v, %v\n", v[0], v[1])
		iw.wr.ControlChange(v[0], v[1])
	case MIDIPitchbend:
		fmt.Printf("MIDIPitchbend %v, \n", int16(v))
		iw.wr.Pitchbend(int16(v))
	case MIDIPolyAftertouch:
		fmt.Printf("MIDIPolyAftertouch %v, %v\n", v[0], v[1])
		iw.wr.PolyAftertouch(v[0], v[1])
	case MIDIAftertouch:
		fmt.Printf("Aftertouch %v, \n", uint8(v))
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

		length := uint32(v.endPos - iw.lastNum32)

		fmt.Printf("total length: %v\n", length)

		//lengthPerItem := uint32(4.0 * math.Round(float64(length)/float64(len(v.items))))
		//lengthPerItem := uint32(math.Round(float64(length) / float64(len(v.items))))

		var delta uint32

		for _, it := range v.items {
			if item != Hold {
				if delta > 0 {
					//iw.wr.Forward(0, delta, 128)
					//iw.wr.Forward(0, delta, 32)
					fmt.Printf("lengthPerItem: %v\n", float64(delta)/float64(uint32(len(v.items))*32))
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

	case Rest:
		fmt.Printf("got REST\n")
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
	p                  *SMFWriter
	wr                 *mid.SMFWriter
	instr              *Instrument
	emptyBars          uint32
	lastItem           interface{}
	prevVel            int8
	noteOns            map[uint8]bool
	repeatFrom         int
	repeatingBars      int
	currentlyRepeating int
	totalBeginning     bool
	insideRepitition   bool
	lastNum32          uint
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
	fmt.Printf("Channel: %v\n", iw.instr.MIDIChannel)
	iw.wr.SetChannel(uint8(iw.instr.MIDIChannel))
	fmt.Printf("Track: %#v\n", iw.instr.Name)
	iw.wr.Program(iw.instr.Name)
	iw.wr.Track(iw.instr.Name)
	iw.wr.Sequence(iw.instr.Name)

	if iw.instr.MIDIBank >= 0 {
		fmt.Printf("Bank: %v\n", iw.instr.MIDIBank)
		iw.wr.ControlChange(cc.BankSelectMSB, uint8(iw.instr.MIDIBank))
	}

	if iw.instr.MIDIProgram >= 0 {
		fmt.Printf("ProgramChange: %v\n", iw.instr.MIDIProgram)
		iw.wr.ProgramChange(uint8(iw.instr.MIDIProgram))
	}

	if iw.instr.MIDIVolume >= 0 {
		fmt.Printf("Volume: %v\n", iw.instr.MIDIVolume)
		//iw.wr.ProgramChange(uint8(iw.instr.MIDIProgram))
		iw.wr.ControlChange(cc.VolumeMSB, uint8(iw.instr.MIDIVolume))
	}
}

func (iw *instWriter) writeUnrolled() error {
	var lastEv *Event

	// there are no empty events any longer
	for _, ev := range iw.instr.unrolled {
		diffBars := ev.BarNo
		if lastEv != nil {
			diffBars -= lastEv.BarNo
		}

		if ev.Item != Hold {
			iw.wr.Forward(uint32(diffBars), uint32(ev.DistanceToStartOfBarIn32th), 32)
			stopNotes := func() {
				for nt, isOn := range iw.noteOns {
					if isOn {
						fmt.Printf("[3.5] NoteOff %v\n", nt)
						iw.wr.NoteOff(nt)
						iw.noteOns[nt] = false
					}
				}
			}

			var addedNotes []uint8

			addedNotes = iw.writeItem(ev.Item, stopNotes)

			for _, nt := range addedNotes {
				iw.noteOns[nt] = true
			}
			iw.lastItem = ev.Item
			lastEv = ev
		}
	}
	return nil
}

/*
func (iw *instWriter) writeBar(barNo int) error {
	if barNo < 0 {
		return fmt.Errorf("bars < 0 not allowed")
	}
	fmt.Printf("BAR: %v, IN REPETITION %v emptybars: %v\n", barNo, iw.insideRepitition, iw.emptyBars)
	b := iw.instr.events[barNo]
	cursorMoved := false

	var num32 uint8 = 0
	iw.lastNum32 = 0
	var barMove uint32 = 0

	if b.isEmpty() && iw.repeatingBars > 0 {
		//		rp := iw.repeatingBars
		//		rf := iw.repeatFrom
		iw.currentlyRepeating = (iw.currentlyRepeating + 1) % iw.repeatingBars
		//		cr := iw.currentlyRepeating
		iw.insideRepitition = true
		fmt.Printf("[1] REPEATING bar: %v\n", iw.repeatFrom+iw.currentlyRepeating)
		err := iw.writeBar(iw.repeatFrom + iw.currentlyRepeating)
		iw.insideRepitition = false
		//		iw.repeatingBars = rp
		//		iw.repeatFrom = rf
		//		iw.currentlyRepeating = cr
		return err
	} else {
		if !iw.insideRepitition {
			fmt.Printf("NO REPITITION\n")
			iw.repeatFrom = -1
			iw.repeatingBars = 0
			iw.currentlyRepeating = 0
		}
	}

	if !iw.insideRepitition {

		if nBars, untilNext := b.RepeatingBars(); nBars > 0 {
			if untilNext {
				iw.repeatFrom = barNo - nBars
				iw.currentlyRepeating = 0
				iw.repeatingBars = nBars
				iw.insideRepitition = true
				fmt.Printf("in barno %v [0] REPEATING bar: %v\n", barNo, iw.repeatFrom)
				err := iw.writeBar(iw.repeatFrom)
				iw.insideRepitition = false
				return err
			} else {
				iw.insideRepitition = true
				fmt.Printf("in barno %v [3] REPEATING bar: %v\n", barNo, barNo-1)
				err := iw.writeBar(barNo - 1)
				iw.insideRepitition = false
				return err
			}
		}
	}

	if !iw.totalBeginning {
		barMove = 1
		if !b.isHolding() {
			for nt, isOn := range iw.noteOns {
				if isOn {
					if !cursorMoved {
						fmt.Printf("[0] Forward(%v, 0, 0)\n", 1+iw.emptyBars)
						iw.wr.Forward(1+iw.emptyBars, 0, 0)
						cursorMoved = true
						iw.emptyBars = 0
					}
					fmt.Printf("[0] NoteOff %v\n", nt)
					iw.wr.NoteOff(nt)
					iw.noteOns[nt] = false
				}
			}
		}

	}

	iw.totalBeginning = false

	if b.isEmpty() {
		iw.emptyBars = iw.emptyBars + 1
		return nil
	}

	for barLine, ev := range b {
		if ev.Item == nil || ev.Item == Hold {
			continue
		}

		num32 = iw.p.bars[barNo].positions[barLine]

		diffNum := num32 - iw.lastNum32
		if diffNum > 0 {
			if !cursorMoved {
				cursorMoved = true
				fmt.Printf("[2] wr.Forward(%v, %v, 32)\n", iw.emptyBars+barMove, uint32(diffNum))
				iw.wr.Forward(iw.emptyBars+barMove, uint32(diffNum), 32)
			} else {
				fmt.Printf("[3] wr.Forward(0, %v, 32)\n", uint32(diffNum))
				iw.wr.Forward(0, uint32(diffNum), 32)
			}
		} else {
			if !cursorMoved {
				if iw.emptyBars > 0 {
					fmt.Printf("[4] Forward(%v, 0, 0)\n", iw.emptyBars)
					iw.wr.Forward(iw.emptyBars, 0, 0)
				} else {
					if barMove > 0 {
						fmt.Printf("[5] Forward(%v, 0, 0)\n", barMove)
						iw.wr.Forward(barMove, 0, 0)
					}
				}
				cursorMoved = true
			}
		}
		iw.emptyBars = 0
		iw.lastNum32 = num32
		var item interface{}

		switch it := ev.Item.(type) {
		case RepeatLastEvent:
			item = iw.lastItem
			// TODO maybe allow dynamic override
		default:
			_ = it
			item = ev.Item
		}

		stopNotes := func() {
			for nt, isOn := range iw.noteOns {
				if isOn {
					fmt.Printf("[3.5] NoteOff %v\n", nt)
					iw.wr.NoteOff(nt)
					iw.noteOns[nt] = false
				}
			}
		}

		var addedNotes []uint8

		addedNotes = iw.writeItem(item, stopNotes)
		iw.lastItem = item

		for _, nt := range addedNotes {
			iw.noteOns[nt] = true
		}

	}
	return nil
}
*/

func (iw *instWriter) writeTrack() error {

	iw.writeIntro()
	iw.prevVel = 60
	iw.emptyBars = 0
	iw.noteOns = map[uint8]bool{}
	iw.lastItem = nil

	iw.writeUnrolled()

	/*
		for barNo := 0; barNo < len(iw.instr.events); barNo++ {
			iw.writeBar(barNo)
		}
	*/

	var hasNoteOn bool

	for _, isOn := range iw.noteOns {
		if isOn {
			hasNoteOn = true
		}
	}

	if hasNoteOn {
		fmt.Printf("[6] Forward(%v, 0, 0)\n", iw.emptyBars+1)
		iw.wr.Forward(iw.emptyBars+1, 0, 0)
		for nt, isOn := range iw.noteOns {
			if isOn {
				fmt.Printf("[4] NoteOff %v\n", nt)
				iw.wr.NoteOff(nt)
				iw.noteOns[nt] = false
			}
		}
	}
	return nil
}

type Instrument struct {
	Name        string
	MIDIChannel int8
	MIDIProgram int8
	MIDIVolume  int8
	MIDIBank    int8
	events      []BarEvents // in the order of bars
	colWidth    int
	unrolled    []*Event
}

func (i *Instrument) Dup() *Instrument {
	return &Instrument{
		Name:        i.Name,
		MIDIChannel: i.MIDIChannel,
		MIDIProgram: i.MIDIProgram,
		MIDIVolume:  i.MIDIVolume,
		MIDIBank:    i.MIDIBank,
		colWidth:    i.colWidth, // ? set to 0??
	}
}

func (i *Instrument) pad(s string) string {
	return pad(s, i.colWidth)
}

func (i *Instrument) calcColWidth() {
	i.colWidth = len(i.Name)

	for _, be := range i.events {
		for _, ev := range be {
			l := len(ev.originalData)
			if l > i.colWidth {
				i.colWidth = l
			}
		}
	}

	if i.colWidth < 2 {
		i.colWidth = 2
	}
}
