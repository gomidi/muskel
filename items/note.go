package items

import (
	"bytes"
	"fmt"
	"math"
	"math/rand"
	"regexp"
	"strconv"
	"strings"

	"gitlab.com/gomidi/midi/midimessage/channel"
)

// TODO translate augmentation symbols like ^ and ° to pitchbend messages

type Note struct {
	Letter    string
	Octave    int
	Augmenter string
	//Velocity       int8
	Dynamic        string
	Transposition  int8
	GlissandoStart bool
	GlissandoExp   bool
	Dotted         string
	ScaleNote      int8
	PosShift       int // 0 = no, 1 = laidback, -1 = ahead of time
	NoteOn         bool
	NoteOff        bool
}

var _ Item = &Note{}

func (note Note) ToMIDI() (midinote_ uint8) {
	midinote := 48 // c

	if note.Octave > 0 {
		midinote += 12 * (note.Octave - 1)
	}

	if note.Octave < 0 {
		midinote += 12 * note.Octave
	}

	switch strings.ToLower(note.Letter) {
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

	switch note.Augmenter {
	case "":
	case "#":
		midinote += 1
		//case "^":
		//case "°":
	}

	midinote += int(note.Transposition)

	if midinote > 127 {
		midinote = 127
	}

	if midinote < 0 {
		midinote = 0
	}

	return uint8(midinote)

}

func (n Note) String() string {
	var bf strings.Builder

	if n.ScaleNote == 0 {
		letter := n.Letter
		if n.Octave < 0 {
			letter = strings.ToUpper(letter)
		}

		bf.WriteString(letter)
		bf.WriteString(n.Augmenter)
		if n.Transposition != 0 {
			bf.WriteString(fmt.Sprintf("^%v", n.Transposition))
		}

		switch n.Octave {
		case -5, 5:
			bf.WriteString("\"\"")
		case -4, 4:
			bf.WriteString("\"'")
		case -3, 3:
			bf.WriteString("\"")
		case -2, 2:
			bf.WriteString("'")
		case 1:
			// do nothing
		case 0:
			panic("invalid octave")
		case -1:
			// do nothing
		case 6:
			bf.WriteString("\"\"'")
		case 7:
			bf.WriteString("\"\"\"")
		case 8:
			bf.WriteString("\"\"\"'")
		case 9:
			bf.WriteString("\"\"\"\"")
		case 10:
			bf.WriteString("\"\"\"\"'")
		// case 10:
		// bf.WriteString("\"\"\"\"\"")
		default:
			panic(fmt.Sprintf("invalid octave: %v", n.Octave))
		}

	} else {
		bf.WriteString(fmt.Sprintf("^%v", n.ScaleNote))
		if n.Transposition != 0 {
			bf.WriteString(fmt.Sprintf("^%v", n.Transposition))
		}
	}

	//bf.WriteString(velocityToDynamic(n.Velocity))
	bf.WriteString(n.Dynamic)

	if n.Dotted != "" {
		bf.WriteString(n.Dotted)
	}

	if n.GlissandoStart {
		bf.WriteString("~")
	}

	if n.GlissandoExp {
		bf.WriteString("~")
	}

	if n.PosShift == 1 {
		bf.WriteString(">")
	}

	if n.PosShift == -1 {
		bf.WriteString("<")
	}

	if n.NoteOn {
		return bf.String() + "_"
	}

	if n.NoteOff {
		return "_" + bf.String()
	}

	return bf.String()
}

func (n Note) Dup() Item {
	return &n
}

func parseTransposition(data string) (trans int8, rest string, err error) {
	var numbf bytes.Buffer
	var restbf bytes.Buffer
	var numComplete bool
	var minusPossible = true

	for _, char := range data {
		switch char {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			minusPossible = false
			if !numComplete {
				numbf.WriteRune(char)
			} else {
				restbf.WriteRune(char)
			}
		case '-':
			if !minusPossible {
				numComplete = true
			}
			if !numComplete {
				numbf.WriteRune(char)
			} else {
				restbf.WriteRune(char)
			}
		default:
			if !numComplete {
				numComplete = true
			}
			restbf.WriteRune(char)
		}
	}

	tr, err := strconv.Atoi(numbf.String())
	if err != nil {
		return 0, "", fmt.Errorf("can't parse number of transposition in %q", data)
	}
	return int8(tr), restbf.String(), nil
}

func (nt *Note) Parse(data string, posIn32th uint) (err error) {
	original := data

	nt.NoteOn, nt.NoteOff, data = stripNoteOnOff(data)

	switch data[:1] {
	case "a", "b", "c", "d", "e", "f", "g":
		nt.Letter = data[:1]
		nt.Octave = 1
	case "A", "B", "C", "D", "E", "F", "G":
		nt.Letter = strings.ToLower(data[:1])
		nt.Octave = -1
	default:
		err = fmt.Errorf("invalid note: %#v I", original)
		return
	}

	if len(data) > 1 {
		data = data[1:]
	} else {
		data = ""
	}

	if len(data) > 0 && data[0] == '#' {
		nt.Augmenter = "#"
		if len(data) > 1 {
			data = data[1:]
		} else {
			data = ""
		}
	}

	if idx := strings.Index(data, "^"); idx >= 0 {
		if len(data) < 2 {
			err = fmt.Errorf("invalid note: %#v II", original)
		}
		nt.Transposition, data, err = parseTransposition(data[1:])
		if err != nil {
			err = fmt.Errorf("invalid note: %#v: %s III", original, err)
			return
		}
	}

	//fmt.Printf("note data: %q\n", data)

	var dynamic string
	var gliss int

	for _, l := range data {
		switch l {
		case ':':
			nt.Dotted += ":"
		case '+':
			dynamic += "+"
		case '-':
			dynamic += "-"
		case '=':
			dynamic = "="
		case '~':
			gliss++
		case '>':
			nt.PosShift = 1
		case '<':
			nt.PosShift = -1
		//case '#':
		//	nt.Augmenter += "#"
		//case '^':
		//nt.Augmenter += "^"
		//case '°':
		//nt.Augmenter += "°"
		case '"':
			if nt.Octave > 0 {
				nt.Octave += 2
			}

			if nt.Octave < 0 {
				nt.Octave -= 2
			}
		case '\'':
			if nt.Octave > 0 {
				nt.Octave += 1
			}

			if nt.Octave < 0 {
				nt.Octave -= 1
			}
		default:
			err = fmt.Errorf("invalid note: %#v IV", original)
			return
		}
	}

	if gliss > 0 {
		nt.GlissandoStart = true
	}

	if gliss > 1 {
		nt.GlissandoExp = true
	}

	//nt.Velocity = DynamicToVelocity(dynamic)
	nt.Dynamic = dynamic

	return
}

func (v Note) WriteMIDI(wr SMFWriter) (addedNotes []uint8) {
	vscale := MIDITrack.VelocityScale
	vel := DynamicToVelocity(v.Dynamic, int8(vscale[0]), int8(vscale[1]), int8(vscale[2]), int8(vscale[3]), int8(vscale[4]))

	if vel < 0 {
		vel = MIDITrack.PrevVel
	}

	MIDITrack.PrevVel = vel

	vl := uint8(vel + int8(rand.Intn(int(vscale[2]))))

	// fmt.Printf("%#v\n", v)

	var n uint8
	n = v.ToMIDI()

	switch v.PosShift {
	case 0:
		MIDITrack.SetStraight()
	case 1:
		MIDITrack.SetLaidBack()
	case -1:
		MIDITrack.SetAhead()
	}

	//	}

	// only add MIDITranspose to Notes not to MIDINotes
	n = uint8(int8(n) + MIDITrack.Track.MIDITranspose)
	MIDITrack.PrevKey = n

	if MIDITrack.noteGlide.active {
		/*
			1. get the distance to the previous note
			2. calc the diff in semitones
			3. make a line of 32ths in steps of the factor
		*/
		//distance := int64(math.Round(float64(iw.wr.Position()+uint64(iw.wr.Delta())-iw.startGlissando) / float64(iw.wr.Ticks32th())))
		distance := int64(math.Round(float64(wr.Position()+uint64(wr.Delta())-MIDITrack.noteGlide.startPosition) / float64(MIDITrack.GlideResolution(wr))))
		noteDiff := int64(n) - int64(MIDITrack.noteGlide.startNote)

		var pb int16

		MIDITrack.noteGlide.glideFunc(wr, distance, noteDiff, func(vl float64) {
			pb = HalfTonesToPitchbend(vl, MIDITrack.Track.PitchbendRange)
			wr.Pitchbend(pb)
		})

		MIDITrack.PrevPitchbend = pb
		/*
			addStep := calc32thsAdd(distance, noteDiff)
			iw.wr.RestoreTimeline()

			for step := int64(1); step <= distance; step++ {
				iw.wr.Forward(0, 1, 32)
				iw.wr.Pitchbend(halfTonesToPitchbend(float64(step)*addStep, iw.instr.PitchbendRange))
			}
		*/

		delete(MIDITrack.NoteOns, MIDITrack.noteGlide.startNote)
	}

	if !v.NoteOn && !v.NoteOff {
		MIDITrack.StopNotes(wr)
	}

	//fmt.Printf("NoteOn %v\n", n)
	if !MIDITrack.noteGlide.active {
		if MIDITrack.noteGlide.startNote != 0 {
			MIDITrack.noteGlide.startNote = 0
			wr.Pitchbend(0)
			MIDITrack.PrevPitchbend = 0
		}

		if !v.NoteOff {
			MIDITrack.writeNoteOn(wr, n, vl)
			// iw.wr.NoteOn(n, vl)
		}

		if !v.NoteOn && !v.NoteOff {
			if v.Dotted != "" {
				// TODO what about planned noteoffs? how to track them in filternotes????
				wr.Plan(0, dottedLengths[v.Dotted][0], dottedLengths[v.Dotted][1], channel.Channel(wr.Channel()).NoteOff(n))
			} else {
				addedNotes = append(addedNotes, n)
			}
		}
	}

	if v.NoteOff {
		MIDITrack.writeNoteOff(wr, n)
		// iw.wr.NoteOff(n)
	}

	if MIDITrack.noteGlide.active && !v.NoteOff && !v.NoteOn {
		//  set pitchbend back to 0
		MIDITrack.noteGlide.active = false
		if v.Dotted != "" {
			wr.Pitchbend(MIDITrack.PrevPitchbend)
			// TODO what about planned noteoffs? how to track them in filternotes????
			wr.Plan(0, dottedLengths[v.Dotted][0], dottedLengths[v.Dotted][1], channel.Channel(wr.Channel()).NoteOff(MIDITrack.noteGlide.startNote))
			//iw.wr.NoteOff(iw.startGlissandoNote)
		} else {
			addedNotes = append(addedNotes, MIDITrack.noteGlide.startNote)
		}
	}

	if v.GlissandoStart && !v.NoteOff && !v.NoteOn {
		//iw.startGlissandoDelta = iw.wr.DeltaTime()
		MIDITrack.noteGlide.startPosition = wr.Position()
		MIDITrack.noteGlide.startNote = n
		MIDITrack.noteGlide.active = true
		MIDITrack.noteGlide.glideFunc = linearGlide
		if v.GlissandoExp {
			MIDITrack.noteGlide.glideFunc = exponentialGlide
		}
		wr.BackupTimeline()
	}

	MIDITrack.SetStraight()

	return addedNotes
}

var regScaleNote = regexp.MustCompile("^(-){0,1}([0-9]+)(.*)$")

func (nt *Note) parseScale(data string) (err error) {
	original := data

	nt.NoteOn, nt.NoteOff, data = stripNoteOnOff(data)

	var transp int8

	if idx := strings.Index(data, "^"); idx > 0 {
		if len(data[idx:]) < 2 {
			err = fmt.Errorf("invalid scale note: %#v I", original)
		}
		pre := data[:idx]
		transp, data, err = parseTransposition(data[idx+1:])
		if err != nil {
			err = fmt.Errorf("invalid scale note: %#v: %s II", original, err)
			return
		}

		data = pre + data
	}

	mt := regScaleNote.FindStringSubmatch(data)
	//fmt.Printf("scale note %q match: %#v\n", data, mt)
	var i int
	i, err = strconv.Atoi(mt[1] + mt[2])

	if err != nil {
		err = fmt.Errorf("invalid scale note: %#v III", data)
		return
	}

	nt.ScaleNote = int8(i)
	switch {
	case transp == 0:
	case nt.ScaleNote > 0 && transp > 0:
		nt.ScaleNote += transp
	case nt.ScaleNote < 0 && transp < 0:
		nt.ScaleNote += transp
	case nt.ScaleNote == 0:
		err = fmt.Errorf("invalid scale note: %#v III", data)
		return
	case nt.ScaleNote+transp == 0:
		if transp > 0 {
			nt.ScaleNote = 1
		} else {
			nt.ScaleNote = -1
		}
	case nt.ScaleNote < 0 && transp > 0:
		nt.ScaleNote = nt.ScaleNote + transp
		if nt.ScaleNote >= 0 {
			nt.ScaleNote += 1
		}
	}

	var dynamic string

	for _, l := range mt[3] {
		switch l {
		case ':':
			nt.Dotted += ":"
		case '+':
			dynamic += "+"
		case '-':
			dynamic += "-"
		case '=':
			dynamic = "="
		case '~':
			nt.GlissandoStart = true
		case '#':
			nt.Augmenter += "#"
		//case '^':
		//nt.Augmenter += "^"
		//case '°':
		//nt.Augmenter += "°"
		case '>':
			nt.PosShift = 1
		case '<':
			nt.PosShift = -1
		/*
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
		*/
		default:
			err = fmt.Errorf("invalid scale note: %#v IV", data)
			return
		}
	}

	//nt.Velocity = DynamicToVelocity(dynamic)
	nt.Dynamic = dynamic

	return
}
