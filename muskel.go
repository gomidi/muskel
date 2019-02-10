package muskel

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"strconv"
	"strings"

	"gitlab.com/gomidi/midi/cc"
	"gitlab.com/gomidi/midi/gm"
	"gitlab.com/gomidi/midi/mid"
	"gitlab.com/gomidi/midi/smf"
	"gitlab.com/gomidi/midi/smf/smfwriter"
)

// TODO see the following list

/*
+ patterns and definition
+ parts und sprünge
+ Glissandi!
+ accel. decel. cresc. descr. (wahrscheinlich mit Volume CC)
+ MIDI controls
- temperaments
- ggf. Chords
- translate augmentation symbols like ^ and ° to pitchbend messages
*/

type Piece struct {
	currentBar     int
	numInstruments int
	Instruments    []*Instrument
	rd             *bufio.Reader
	props          map[string]string
	temperament    map[string]string
	//patterns          map[string]string
	patterns          map[string]*PatternDefinition
	parts             map[string]int // part name to bar no
	jumps             map[int]string // bar no at the end of which the jump should be done mapped to the jump definition
	bars              []*Bar
	comments          map[int]string // line calculated from the start of the system
	headerComments    []string       // comments in the header, as they come
	inComment         bool
	inHeader          bool
	lineNo            int
	lineNosystemStart int
	writerSysLine     int
}

// pad pads the given string with empty space if it is less then length
func pad(s string, length int) string {
	if len(s) >= length {
		return s
	}

	diff := length - len(s)
	var bf bytes.Buffer

	bf.WriteString(s)

	for i := 0; i < diff; i++ {
		bf.WriteString(" ")
	}
	return bf.String()
}

type Note struct {
	letter         string
	octave         int
	augmenter      string
	dynamic        string
	glissandoStart bool
}

type DrumNote struct {
	name    string
	dynamic string
}

type MIDICC [2]uint8
type MIDIPitchbend int16
type MIDIPolyAftertouch [2]uint8
type MIDIAftertouch uint8

type OSCMessage struct {
	Path string
	Args []interface{}
}
type Pattern struct {
	name         string
	params       []string
	replacements []string
	slicer       string
	syncFirst    bool
}

type NTuple struct {
	endPos uint8 // 32th in bar from the beginning
	items  []Item
}
type RepeatLastEvent struct{}
type RepeatLastBar struct{}
type RepeatLastBarUntilChange struct{}
type RepeatLastNBarsUntilChange int
type Rest struct{}
type hold struct{}
type Lyric string

var Hold = hold{}

func (p *Piece) parseMIDICC(data string) (cc MIDICC, err error) {
	return
}

func (p *Piece) parseMIDIPitchbend(string) (pb MIDIPitchbend, err error) {
	return
}

func (p *Piece) parseMIDIAftertouch(string) (at MIDIAftertouch, err error) {
	return
}

func (p *Piece) parseMIDIPolyAftertouch(string) (pt MIDIPolyAftertouch, err error) {
	return
}

func (p *Piece) parseNote(data string) (item interface{}, err error) {
	if len(data) > 1 {

		switch data[:2] {
		case "kd", "sn", "ho", "hc", "rd", "tb", "tl", "tm", "th", "sh":
			var dn DrumNote
			dn.name = data[:2]
			for _, l := range data[2:] {
				switch l {
				case '+':
					dn.dynamic += "+"
				case '-':
					dn.dynamic += "-"
				case '*':
					dn.dynamic += "*"
				default:
					return nil, fmt.Errorf("invalid drum note: %#v", data)
				}
			}
			return dn, nil
		default:
		}
	}
	var nt Note

	switch data[:1] {
	case "a", "b", "c", "d", "e", "f", "g":
		nt.letter = data[:1]
		nt.octave = 1
	case "A", "B", "C", "D", "E", "F", "G":
		nt.letter = strings.ToLower(data[:1])
		nt.octave = -1
	default:
		return nil, fmt.Errorf("invalid note: %#v", data)
	}

	for _, l := range data[1:] {
		switch l {
		case '+':
			nt.dynamic += "+"
		case '-':
			nt.dynamic += "-"
		case '*':
			nt.dynamic += "*"
		case '~':
			nt.glissandoStart = true
		case '#':
			nt.augmenter = "#"
		case '^':
			nt.augmenter = "^"
		case '°':
			nt.augmenter = "°"
		case '"':
			if nt.octave > 0 {
				nt.octave += 2
			}

			if nt.octave < 0 {
				nt.octave -= 2
			}
		case '\'':
			if nt.octave > 0 {
				nt.octave += 1
			}

			if nt.octave < 0 {
				nt.octave -= 1
			}
		default:
			return nil, fmt.Errorf("invalid note: %#v", data)
		}
	}

	return nt, nil
}

/*
- notes, e.g. a'
- drumnotes e.g. sn
- motivs e.g. $a
- MIDI controller events e.g. PB200
- rests e.g. _
- samples with a timespan e.g. Sa
- samples with no timespan ("oneshot") e.g. Za
- repitition of last event e.g. %
- multiitems, consisting of several items happening at the same time. e.g. (a',PB200)
*/

func (p *Piece) parseOSC(data string) (om OSCMessage, err error) {
	return
}

func (p *Piece) parsePattern(data string) (*PatternCall, error) {
	var pc PatternCall
	err := pc.Parse(data)

	if err != nil {
		return nil, fmt.Errorf("could not parse pattern call: %s", err)
	}

	pd := p.GetPatternDefinition(pc.Name)

	if pd == nil {
		return nil, fmt.Errorf("could not find definition for pattern %s", pc.Name)
	}

	ppc := &pc

	pc.result, err = pd.Call(ppc, p)
	if err != nil {
		panic(fmt.Sprintf("could no call pattern %s with %q: %s", pc.Name, data, err))
	}

	// TODO get positions and Items out of the pc.result

	return ppc, err
}

// ntuple has the form {c,e,d}3&
// where 3& is the ending position that defines the total length
func (p *Piece) parseNTuple(data string) (nt NTuple, err error) {

	dd := strings.Split(data, "}")
	if len(dd) != 2 {
		return nt, fmt.Errorf("invalid n-tuple: %#v", "{"+data)
	}

	completedEndPos, endPos := positionTo32th("", dd[1])

	_ = completedEndPos

	data = strings.TrimSpace(dd[0])
	if data == "" {
		return
	}

	d := strings.Split(data, ",")

	var ntp NTuple
	ntp.endPos = endPos

	for pos, it := range d {
		itt, err := p.parseItem(it)
		if err != nil {
			return nt, fmt.Errorf("ERROR invalid n-tuple at position %v: %#v: %s", pos, it, err)
		}
		ntp.items = append(ntp.items, itt)
	}
	return ntp, nil
}

func (p *Piece) parseItem(data string) (interface{}, error) {
	data = strings.TrimSpace(data)
	switch len(data) {
	case 0:
		return nil, nil
	default:
		//fmt.Printf("data[0]: %#v\n", string(data[0]))
		switch data[0] {
		case '"':
			return Lyric(strings.Trim(data, `"`)), nil
		case '%':
			return RepeatLastEvent{}, nil
		case '=':
			return Hold, nil
		case '_':
			return Rest{}, nil
		case '{':
			return p.parseNTuple(data[1:])
		case '$':
			return p.parsePattern(data[1:])
		case 'S':
			return p.parseNote(data[1:])
		case 'Z':
			return p.parseNote(data[1:])
		case 'O':
			return p.parseOSC(data[1:])
		case '.':
			if data == "..." {
				return RepeatLastBarUntilChange{}, nil
			}
			if data == "./." {
				return RepeatLastBar{}, nil
			}
			if strings.Index("123456789", data[1:2]) != -1 && data[2] == '.' {
				n, _ := strconv.Atoi(data[1:2])
				//				fmt.Printf("RepeatLastNBarsUntilChange %v found\n", n)
				return RepeatLastNBarsUntilChange(n), nil
			}
		default:
			if len(data) > 1 {
				switch data[0:2] {
				case "CC":
					return p.parseMIDICC(data[2:])
				case "PB":
					return p.parseMIDIPitchbend(data[2:])
				case "AT":
					return p.parseMIDIAftertouch(data[2:])
				case "PT":
					return p.parseMIDIPolyAftertouch(data[2:])
				default:
				}
			}
			return p.parseNote(data)
		}
	}

	return nil, nil

}

type MultiItem []interface{}

func (p *Piece) parseData(data string) (interface{}, error) {
	if len(data) == 0 {
		return nil, nil
	}

	switch {
	case data[0] == '"':
		return Lyric(strings.Trim(data, `"`)), nil
	default:
		d := strings.Split(data, " ")
		switch len(d) {
		case 1:
			it, err := p.parseItem(d[0])
			if err != nil {
				return nil, err
			}
			return it, nil
		default:
			var m MultiItem

			for _, dd := range d {
				it, err := p.parseItem(strings.TrimSpace(dd))

				if err != nil {
					return nil, err
				}

				m = append(m, it)
			}
			return m, nil
		}

	}

	return nil, nil

}

// if returned vel is < 0, then it has not been set and
// is drived from the previous velocity
// always return such that rand.Intn(4) can be added or substracted
func velocityFromDynamic(dyn string) (vel int8) {
	vel = -1
	for _, rn := range dyn {
		switch rn {
		case '+':
			if vel == -1 {
				vel = 60
			}
			vel += 15
		case '-':
			if vel == -1 {
				vel = 60
			}
			vel -= 15
		case '*': // reset
			vel = 60
		}
	}

	if vel == -1 {
		return vel
	}

	if vel > 127 {
		vel = 123
	}

	if vel < 4 {
		vel = 4
	}

	return
}

func noteToMIDI(note Note) (midinote_ uint8) {
	/*
		rr := regNote.FindAllStringSubmatch(note, -1)
		if len(rr) < 1 {
			panic(fmt.Sprintf("invalid note: %v", note))
		}
		res := rr[0]
		//	fmt.Printf("res: %#v\n", res)

		if len(res) < 2 {
			panic(fmt.Sprintf("invalid note: %v", note))
		}

		letter := res[1]
	*/

	midinote := 48 // c

	if note.octave > 0 {
		midinote += 12 * (note.octave - 1)
	}

	if note.octave < 0 {
		midinote += 12 * note.octave
	}

	switch strings.ToLower(note.letter) {
	case "c":
		midinote += 0
	case "d":
		midinote += 2
	case "e":
		midinote += 4
	case "f":
		midinote += 5
	case "g":
		midinote += 7
	case "a":
		midinote += 9
	case "b":
		midinote += 11
	}

	switch note.augmenter {
	case "":
	case "#":
		midinote += 1
	case "^":
	case "°":
	}

	return uint8(midinote)

}

func isEmpty(b BarEvents) bool {
	if len(b) == 0 {
		return true
	}

	for _, e := range b {
		if e.Item != nil {
			return false
		}
	}

	return true
}

func (p *Piece) GetPatternDefinition(name string) *PatternDefinition {
	return p.patterns[name]
}

func (iw *instWriter) writeItem(item Item, stopNotes func()) (addedNotes []uint8) {
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
	case *PatternCall:
		// TODO do something with the pc events
		barevts := v.Events

		_ = barevts
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
		iw.lastNum32 += uint8(length)

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
	}

	return addedNotes
}

func (p *Piece) writeFirstTrack(wr *mid.SMFWriter) error {
	num := uint8(4)
	denom := uint8(4)

	for i, b := range p.bars {
		b.positions = make([]uint8, len(b.originalPositions))
		var oldPos string

		for pi, pos := range b.originalPositions {
			oldPos, b.positions[pi] = positionTo32th(oldPos, pos)
		}

		fmt.Printf("bar %v: %#v\n", i, b)

		if i > 0 {
			fmt.Printf("[firstrack] Forward(1, 0, 0)\n")
			wr.Forward(1, 0, 0)
			/*
				fmt.Printf("Forward(0, %v, %v)\n", uint32(num), uint32(denom))
				wr.Forward(0, uint32(num), uint32(denom))
			*/
		}

		if b.timeSigChange[0] > 0 {
			num = b.timeSigChange[0]
			denom = b.timeSigChange[1]
		}
		fmt.Printf("write meter %v/%v\n", num, denom)

		if b.tempoChange > 0 {
			wr.TempoBPM(b.tempoChange)
			fmt.Printf("write tempo change %v\n", b.tempoChange)
		}
		wr.Meter(num, denom)
		//wr.Forward(1, 0, 0)
	}

	return nil
}

type instWriter struct {
	p                  *Piece
	wr                 *mid.SMFWriter
	instr              *Instrument
	emptyBars          uint32
	lastItem           Item
	prevVel            int8
	noteOns            map[uint8]bool
	repeatFrom         int
	repeatingBars      int
	currentlyRepeating int
	totalBeginning     bool
	insideRepitition   bool
	lastNum32          uint8
}

func newInstrumentSMFWriter(p *Piece, wr *mid.SMFWriter, instr *Instrument) *instWriter {
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

func isHolding(be BarEvents) bool {
	if len(be) == 0 {
		return false
	}

	for _, ev := range be {
		_, is := ev.Item.(hold)
		if is {
			return true
		}
	}
	return false
}

func RepeatingBars(be BarEvents) (num int, untilNext bool) {
	if len(be) == 0 {
		return
	}

	for _, ev := range be {
		switch v := ev.Item.(type) {
		case RepeatLastBar:
			return 1, false
		case RepeatLastBarUntilChange:
			return 1, true
		case RepeatLastNBarsUntilChange:
			return int(v), true
		}
	}

	return
}

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

	if isEmpty(b) && iw.repeatingBars > 0 {
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

		if nBars, untilNext := RepeatingBars(b); nBars > 0 {
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
		if !isHolding(b) {
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

	if isEmpty(b) {
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
		var item Item

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

func (iw *instWriter) writeTrack() error {

	iw.writeIntro()
	iw.prevVel = 60
	iw.emptyBars = 0
	iw.noteOns = map[uint8]bool{}
	iw.lastItem = nil

	for barNo := 0; barNo < len(iw.instr.events); barNo++ {
		iw.writeBar(barNo)
	}

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

func (p *Piece) writeSMF(wr *mid.SMFWriter) error {
	err := p.writeFirstTrack(wr)

	if err != nil {
		return err
	}

	for _, instr := range p.Instruments {
		// ignore instruments without a MIDIchannel
		if instr.MIDIChannel < 0 {
			continue
		}

		fmt.Println("EOT")
		wr.EndOfTrack()
		iw := newInstrumentSMFWriter(p, wr, instr)
		err = iw.writeTrack()
		//		err = p.writeInstrumentTrack(wr, instr)
		if err != nil {
			return err
		}
	}

	fmt.Println("EOT")
	wr.EndOfTrack()
	return nil
}

func getQNNumberFromPos(pos string) (qnnumber int, rest string) {
	if len(pos) == 0 {
		panic("empty position is not valid")
	}
	qnnumber = -1
	rest = pos

	if len(pos) > 1 {
		if i2, err := strconv.Atoi(pos[0:2]); err == nil {
			qnnumber = i2
			rest = pos[2:]
		} else {
			if i1, err := strconv.Atoi(pos[0:1]); err == nil {
				qnnumber = i1
				rest = pos[1:]
			}
		}
	} else {
		if i1, err := strconv.Atoi(pos[0:1]); err == nil {
			qnnumber = i1
			rest = pos[1:]
		}
	}

	return
}

// lastPos must either be "", then pos must be complete
// (i.e. must start with a number) or lastPos must be complete
// then pos may be derived from it
func positionTo32th(lastPos, pos string) (completed string, num32th uint8) {

	number, rest := getQNNumberFromPos(pos)
	completed = pos

	if number == -1 {
		if lastPos == "" {
			panic("lastPos must be given, if pos is incomplete")
		}

		lastNum, lastRest := getQNNumberFromPos(lastPos)

		if lastNum < 1 {
			panic("lastPos must be given, if pos is incomplete")
		}

		number = lastNum
		rest = lastRest + rest
		completed = fmt.Sprintf("%v%s", number, rest)
	}

	num32th = uint8((number - 1) * 8)

	if rest == "" {
		return
	}

	switch rest {
	case ";":
		num32th += 1
	case ".":
		num32th += 2
	case ".;":
		num32th += 3
	case "&":
		num32th += 4
	case "&;":
		num32th += 5
	case "&.":
		num32th += 6
	case "&.;":
		num32th += 7
	default:
		panic("invalid rest: " + rest + " in " + pos)
	}

	return

}

func (p *Piece) WriteSMF(midifile string) error {

	numTracks := uint16(1) // first track is for tempo and time signatures

	for _, instr := range p.Instruments {
		if instr.MIDIChannel >= 0 {
			numTracks++
		}
	}

	return mid.NewSMFFile(midifile, numTracks, p.writeSMF,
		smfwriter.TimeFormat(smf.MetricTicks(960)),
		//		smfwriter.Debug(log.New(os.Stdout, "write MIDI", log.Lshortfile)),
	)
}

func (p *Piece) WriteFormatted(filepath string) (err error) {
	dir, err := ioutil.TempDir(".", "muskel-fmt")
	if err != nil {
		return err
	}

	base := path.Base(filepath)

	f, err := os.Create(path.Join(dir, base))
	if err != nil {
		return err
	}
	var bf bytes.Buffer
	p.writerSysLine = 0

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error: %v", r)
		}
		f.Close()
		os.RemoveAll(dir)
	}()

	p.format(&bf)

	_, err = bf.WriteTo(f)
	if err != nil {
		return
	}

	err = os.Rename(path.Join(dir, base), filepath)
	return
}

func (p *Piece) writeSystemLine(bf *bytes.Buffer, line string) {
	if comment, has := p.comments[p.writerSysLine]; has {
		p.writerSysLine++
		bf.WriteString(comment + "\n")
		p.writeSystemLine(bf, line)
		return
	}
	p.writerSysLine++
	bf.WriteString(line + "\n")
}

func (p *Piece) format(bf *bytes.Buffer) {
	//bf.WriteString("\n// info\n")

	//fmt.Fprintf(bf, "%s %s\n", pad("version:", 15), time.Now().Format(time.RFC822))

	for k, v := range p.props {
		//fmt.Fprintf(bf, "%s %s\n", pad(k+":", 15), v)
		fmt.Fprintf(bf, "%15s+ %s\n", k+":", v)
	}

	//bf.WriteString("\n// temperament\n")

	bf.WriteString("\n\n")

	for k, v := range p.temperament {
		fmt.Fprintf(bf, "/%s/%s\n", k, v)
	}

	bf.WriteString("\n\n")

	//bf.WriteString("\n// pattern definitions\n")

	for k, v := range p.patterns {
		//fmt.Fprintf(bf, "%s %s\n", pad("$"+k+":", 15), v)
		fmt.Fprintf(bf, "%15s+ %s\n", "$"+k+":", v)
	}

	bf.WriteString("\n\n")

	for _, comment := range p.headerComments {
		bf.WriteString(comment + "\n")
	}

	//fmt.Fprintf(bf, "\n\nBPM: %0.2f\n", p.tempoBPM)

	bf.WriteString("\n=\n")

	// 9 whitespace to first pipe
	l := "         |"

	for _, instr := range p.Instruments {
		instr.calcColWidth()
		//instrColWidths[i] = instr.colWidth
		l += fmt.Sprintf(" %s |", instr.pad(instr.Name))
	}

	p.writeSystemLine(bf, l)

	l = "Ch       |"

	for _, instr := range p.Instruments {
		if instr.MIDIChannel == -1 {
			l += fmt.Sprintf(" %v |", instr.pad("-"))
		} else {
			l += fmt.Sprintf(" %v |", instr.pad(fmt.Sprintf("%v", instr.MIDIChannel+1)))
		}
	}

	p.writeSystemLine(bf, l)

	l = "Bank     |"

	for _, instr := range p.Instruments {
		if instr.MIDIBank == -1 {
			l += fmt.Sprintf(" %v |", instr.pad("-"))
		} else {
			l += fmt.Sprintf(" %v |", instr.pad(fmt.Sprintf("%v", instr.MIDIBank+1)))
		}
	}

	p.writeSystemLine(bf, l)

	l = "Prog     |"

	for _, instr := range p.Instruments {
		if instr.MIDIProgram == -1 {
			l += fmt.Sprintf(" %v |", instr.pad("-"))
		} else {
			l += fmt.Sprintf(" %v |", instr.pad(fmt.Sprintf("%v", instr.MIDIProgram+1)))
		}
	}

	p.writeSystemLine(bf, l)

	l = "Vol      |"

	for _, instr := range p.Instruments {
		if instr.MIDIVolume == -1 {
			l += fmt.Sprintf(" %v |", instr.pad("-"))
		} else {
			l += fmt.Sprintf(" %v |", instr.pad(fmt.Sprintf("%v", instr.MIDIVolume+1)))
		}
	}

	p.writeSystemLine(bf, l)

	for i, bar := range p.bars {
		if bar.timeSigChange[0] != 0 {
			l = fmt.Sprintf("%v/%v", bar.timeSigChange[0], bar.timeSigChange[1])
		} else {
			l = ""
		}

		if bar.tempoChange > 0 {
			l += fmt.Sprintf("@%g", bar.tempoChange)
		}

		p.writeSystemLine(bf, l)

		for pi, pos := range bar.originalPositions {
			switch {
			case pos[0] == '&':
				//l = fmt.Sprintf("     %s |", pad(pos, 3))
				l = fmt.Sprintf("     %3s+ |")
			case strings.Index("123456789", pos[0:1]) != -1:
				//l = fmt.Sprintf("    %s |", pad(pos, 4))
				l = fmt.Sprintf("    %4s+ |")
			case pos[0] == '.':
				//l = fmt.Sprintf("     %s |", pad(pos, 3))
				l = fmt.Sprintf("     %3s+ |")
			case pos[0] == ';':
				//l = fmt.Sprintf("     %s |", pad(pos, 3))
				l = fmt.Sprintf("     %3s+ |")
			}

			for _, instr := range p.Instruments {
				be := instr.events[i]

				l += fmt.Sprintf(" %s |", instr.pad(be[pi].originalData))

			}

			if pi == 0 {
				for prt, br := range p.parts {
					if br == i {
						l += fmt.Sprintf(" %s", prt)
					}
				}
			}

			if pi == len(bar.originalPositions)-1 {
				for br, jmp := range p.jumps {
					if br == i {
						l += fmt.Sprintf(" %s", jmp)
					}
				}
			}

			p.writeSystemLine(bf, l)
		}

	}
	bf.WriteString("// END\n")

}

func (p *Piece) newBar(b *Bar) {
	p.bars = append(p.bars, b)
	p.currentBar++
}

type TimeSig [2]uint8

type Bar struct {
	tempoChange       float64
	timeSigChange     TimeSig
	positions         []uint8  // each position is number of 32th from start
	originalPositions []string // original positioning strings
}

type Item interface{}

type Event struct {
	Item         Item
	originalData string
}

type BarEvents []*Event

func NewPiece(r io.Reader) *Piece {
	p := &Piece{}
	p.currentBar = -1
	p.numInstruments = -1
	p.rd = bufio.NewReader(r)
	p.props = map[string]string{}
	p.temperament = map[string]string{}
	p.patterns = map[string]*PatternDefinition{}
	p.parts = map[string]int{}
	p.jumps = map[int]string{}
	p.comments = map[int]string{}
	return p
}

type Instrument struct {
	Name        string
	MIDIChannel int8
	MIDIProgram int8
	MIDIVolume  int8
	MIDIBank    int8
	events      []BarEvents // in the order of bars
	colWidth    int
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

func (p *Piece) parseSystemLine(line string) error {
	tabs := strings.Split(line, "|")

	//	parts := map[string]int{}

	switch len(tabs) {
	case 1:
		tab0 := strings.TrimSpace(tabs[0])
		// we have a bar-change, i.e. no pipes anywhere
		switch tab0 {
		case "":
			var b Bar
			p.newBar(&b)
			return nil
		// normal bar change without anything
		default:

			var b Bar

			if idx := strings.Index(tab0, "@"); idx >= 0 {
				bpm, err := strconv.ParseFloat(tab0[idx+1:], 64)
				if err != nil {
					return fmt.Errorf("error in tempo change %#v. must be a number", tab0[idx+1:])
				}
				b.tempoChange = bpm

				tab0 = strings.TrimSpace(tab0[:idx])
			}

			if tab0 == "" {
				// add bar
				p.newBar(&b)
				return nil
			}

			timeSig := strings.Split(tab0, "/")
			if len(timeSig) != 2 {
				return fmt.Errorf("error in time signature %#v. must be in format n/m where n and m are numbers > 0", tab0)
			}

			num, err := strconv.Atoi(timeSig[0])
			if err != nil {
				return fmt.Errorf("error in time signature %#v. must be in format n/m where n and m are numbers > 0", tab0)
			}

			denom, err := strconv.Atoi(timeSig[1])
			if err != nil {
				return fmt.Errorf("error in time signature %#v. must be in format n/m where n and m are numbers > 0", tab0)
			}

			b.timeSigChange[0] = uint8(num)
			b.timeSigChange[1] = uint8(denom)

			p.newBar(&b)
			return nil
		}
	case 0:
		panic("must not happen")
	default:
		if p.numInstruments == -1 {
			p.numInstruments = len(tabs) - 2
		}

		position := strings.TrimSpace(tabs[0])

		instrData := tabs[1 : len(tabs)-1]

		//fmt.Printf("instruments: [%v]: %v (should %v)\n", lineNo, len(tabs)-2, numInstruments)
		if missingTabs := p.numInstruments - (len(tabs) - 2); missingTabs > 0 {
			for mt := 0; mt < missingTabs; mt++ {
				instrData = append(instrData, "")
			}
		}
		/*
			if len(tabs)-2 != p.numInstruments {
				return fmt.Errorf("line %v has %v instruments, but must have %v", p.lineNo, len(tabs)-2, p.numInstruments)
			}
		*/

		for i, data := range instrData {
			var instr *Instrument
			if i > len(p.Instruments)-1 {
				instr = &Instrument{}
				instr.MIDIProgram = -1
				instr.MIDIChannel = -1
				instr.MIDIVolume = -1
				instr.MIDIBank = -1
				p.Instruments = append(p.Instruments, instr)
			} else {
				instr = p.Instruments[i]
			}

			data = strings.TrimSpace(data)

			switch strings.ToLower(position) {
			case "":
				if instr.Name == "" {
					instr.Name = data
				}
			case "ch", "channel":
				if instr.MIDIChannel == -1 && data != "-" {
					cc, errC := strconv.Atoi(data)
					if errC != nil {
						return fmt.Errorf("%#v is not a valid MIDI channel", data)
					}
					if cc > 16 || cc < 1 {
						return fmt.Errorf("%#v is not a valid MIDI channel", cc)
					}

					instr.MIDIChannel = int8(cc - 1)
				}
			case "prog", "program":
				if instr.MIDIProgram == -1 && data != "-" {
					pr, errP := strconv.Atoi(data)
					if errP != nil {
						return fmt.Errorf("%#v is not a valid MIDI program value", data)
					}
					if pr > 128 || pr < 1 {
						return fmt.Errorf("%#v is not a valid MIDI program value", pr)
					}
					instr.MIDIProgram = int8(pr - 1)
				}
			case "vol", "volume":
				if instr.MIDIVolume == -1 && data != "-" {
					pr, errP := strconv.Atoi(data)
					if errP != nil {
						return fmt.Errorf("%#v is not a valid MIDI volume value", data)
					}
					if pr > 128 || pr < 1 {
						return fmt.Errorf("%#v is not a valid MIDI volume value", pr)
					}
					instr.MIDIVolume = int8(pr - 1)
				}
			case "bank":
				if instr.MIDIBank == -1 && data != "-" {
					pr, errP := strconv.Atoi(data)
					if errP != nil {
						return fmt.Errorf("%#v is not a valid MIDI bank value", data)
					}
					if pr > 128 || pr < 1 {
						return fmt.Errorf("%#v is not a valid MIDI bank value", pr)
					}
					instr.MIDIBank = int8(pr - 1)
				}
			default:
				var barEvents BarEvents
				//fmt.Printf("currentBar: %v, len(instr.events)-1: %v\n", p.currentBar, len(instr.events)-1)
				if diff := (p.currentBar - (len(instr.events) - 1)); diff > 0 {
					for _d := 0; _d < diff; _d++ {
						instr.events = append(instr.events, barEvents)
					}
				}
				barEvents = instr.events[p.currentBar]
				ev := &Event{}
				//ev.originalPosition = position
				//ev.Position = 0 // TODO parse the position from position
				ev.originalData = data

				item, err := p.parseData(ev.originalData)

				if err != nil {
					return fmt.Errorf("column %v: %s", i+2, err)
				}

				//				fmt.Printf("item: %v\n", item)

				if item != nil {
					ev.Item = item
				}
				barEvents = append(barEvents, ev)
				instr.events[p.currentBar] = barEvents

				if i == 0 {
					p.bars[p.currentBar].originalPositions = append(p.bars[p.currentBar].originalPositions, position)
				}
			}
		}

		part := strings.TrimSpace(tabs[len(tabs)-1])

		switch len(part) {
		case 1:
			// part definition
			if _, hasPart := p.parts[part]; hasPart {
				return fmt.Errorf("part %#v already exists", part)
			} else {
				p.parts[part] = p.currentBar
			}

		case 0:
		// do nothing
		default:
			// probably a jump; check
			p.jumps[p.currentBar] = part
			//fmt.Printf("line: %v, position: %v part: %v\n", lineNo, position, part)
		}

	}

	return nil
}

func (p *Piece) parseHeader(lines string) error {
	ll := strings.Split(lines, "\n")
	//var props = map[string]string{}
	//var temperament = map[string]string{}
	//var patterns = map[string]string{}
	for _, l := range ll {
		if len(l) == 0 {
			continue
		}
		switch l[0] {
		case '/':
			keyVal := strings.Split(l[1:], "/")
			if len(keyVal) != 2 {
				return fmt.Errorf("invalid header line %#v, format must be /note/frequency", l)
			}

			p.temperament[strings.TrimSpace(keyVal[0])] = strings.TrimSpace(keyVal[1])
		case '$':
			var pd PatternDefinition
			err := pd.Parse(l[1:])
			if err != nil {
				return fmt.Errorf("invalid pattern definition line %#v: %s", l, err)
			}
			p.patterns[pd.Name] = &pd
		/*
			keyVal := strings.Split(l[1:], ":")
			if len(keyVal) != 2 {
				return fmt.Errorf("invalid header line %#v, format must be $name: definition", l)
			}



			p.patterns[strings.TrimSpace(keyVal[0])] = strings.TrimSpace(keyVal[1])
		*/
		default:
			keyVal := strings.Split(l, ":")
			if len(keyVal) != 2 {
				return fmt.Errorf("invalid header line %#v, format must be key: value", l)
			}

			key := strings.ToLower(strings.TrimSpace(keyVal[0]))
			val := strings.TrimSpace(keyVal[1])

			p.props[key] = val

		}

	}

	//fmt.Printf("properties: %#v\n", props)
	//fmt.Printf("temperament: %#v\n", temperament)
	//fmt.Printf("patterns: %#v\n", patterns)

	return nil
}

func (p *Piece) addSystemComment(relLine int, line []byte) {
	p.comments[relLine] = string(line)
}

func (p *Piece) addHeaderComment(line []byte) {
	p.headerComments = append(p.headerComments, string(line))
}

func (p *Piece) addComment(line []byte) {
	if p.inHeader {
		p.addHeaderComment(line)
		return
	}
	p.addSystemComment(p.lineNo-p.lineNosystemStart, line)
}

func (p *Piece) Parse() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error: %v", r)
		}
	}()
	err = p.parse()
	return
}

func (p *Piece) parse() (err error) {
	var line []byte
	var header bytes.Buffer

	p.inHeader = true
	p.inComment = false
	p.lineNo = 0
	p.lineNosystemStart = 0

	for {
		line, _, err = p.rd.ReadLine()

		if err != nil {
			break
		}

		p.lineNo++

		if len(line) > 0 && string(line[0]) == "=" {
			p.inHeader = false
			err := p.parseHeader(header.String())
			p.lineNosystemStart = p.lineNo + 1
			if err != nil {
				return err
			}
			continue
		}

		if len(line) > 1 {
			switch string(line[0:2]) {
			case "//":
				// comment: ignore
				p.addComment(line)
			case "/*":
				p.inComment = true
				p.addComment(line)
			case "*/":
				p.inComment = false
				p.addComment(line)
			default:
				if !p.inComment {
					//					fmt.Printf("line %v\n", p.lineNo)
					if p.inHeader {
						header.Write(line)
						header.WriteString("\n")
					} else {
						err = p.parseSystemLine(string(line))
						if err != nil {
							return fmt.Errorf("Error in line %v:\n%s\n", p.lineNo, err.Error())
						}
					}
				} else {
					p.addComment(line)
				}
			}
		} else {
			if p.inComment {
				p.addComment(line)
			} else {
				if !p.inHeader {
					err = p.parseSystemLine(string(line))
					if err != nil {
						return fmt.Errorf("Error in line %v:\n%s\n", p.lineNo, err.Error())
					}
				}
			}
		}

	}

	//	fmt.Printf("HEADER:\n%s\n", header.String())

	//fmt.Printf("instruments: %v\n", instrumentNames)
	//fmt.Printf("instruments midi channels: %v\n", instrumentMIDIChannels)
	//fmt.Printf("instruments midi programs: %v\n", instrumentMIDIPrograms)

	if err != io.EOF {
		return err
	}
	return nil
}
