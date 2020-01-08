package items

import (
	"fmt"
	"math/rand"
	"regexp"
	"strconv"

	"gitlab.com/gomidi/midi/midimessage/channel"
)

type MIDINote struct {
	Note int8
	//Velocity int8
	Dynamic  string
	Dotted   string
	PosShift int // 0 = no, 1 = laidback, -1 = ahead of time
	NoteOn   bool
	NoteOff  bool
}

var _ Item = &MIDINote{}

func (m MIDINote) String() string {
	//vel := velocityToDynamic(m.Velocity)
	dotted := ""
	if m.Dotted != "" {
		dotted = m.Dotted
	}

	posShift := ""

	if m.PosShift == 1 {
		posShift = ">"
	}

	if m.PosShift == -1 {
		posShift = "<"
	}

	s := fmt.Sprintf("MN%v%s%s%s", m.Note, m.Dynamic, dotted, posShift)

	if m.NoteOn {
		return s + "_"
	}

	if m.NoteOff {
		return "_" + s
	}

	return s
}

func (mn MIDINote) Dup() Item {
	return &mn
}

func (v MIDINote) WriteMIDI(wr SMFWriter) (addedNotes []uint8) {
	switch v.PosShift {
	case 0:
		MIDITrack.SetStraight()
	case 1:
		MIDITrack.SetLaidBack()
	case -1:
		MIDITrack.SetAhead()
	}

	MIDITrack.noteGlide.active = false
	if !v.NoteOff && !v.NoteOn {
		MIDITrack.StopNotes(wr)
	}

	vscale := MIDITrack.VelocityScale
	vel := DynamicToVelocity(v.Dynamic, int8(vscale[0]), int8(vscale[1]), int8(vscale[2]), int8(vscale[3]), int8(vscale[4]))

	if vel < 0 {
		vel = MIDITrack.PrevVel
	}

	MIDITrack.PrevVel = vel

	vl := uint8(vel + int8(rand.Intn(int(vscale[2]))))
	n := uint8(v.Note)
	//fmt.Printf("MIDINoteOn %v\n", n)
	if MIDITrack.noteGlide.startNote != 0 {
		MIDITrack.noteGlide.startNote = 0
		wr.Pitchbend(0)
	}

	if !v.NoteOff {
		MIDITrack.writeNoteOn(wr, n, vl)
	}

	if !v.NoteOff && !v.NoteOn {
		if v.Dotted != "" {
			// TODO what about planned noteoffs? how to track them in filternotes????
			wr.Plan(0, dottedLengths[v.Dotted][0], dottedLengths[v.Dotted][1], channel.Channel(wr.Channel()).NoteOff(n))
		} else {
			addedNotes = append(addedNotes, n)
		}
	}

	if v.NoteOff {
		MIDITrack.writeNoteOff(wr, n)
		//iw.wr.NoteOff(n)
	}

	return addedNotes
}

var regexMIDINote = regexp.MustCompile("^([0-9]{1,3})([-+" + regexp.QuoteMeta("=") + "]*)(:{0,3})([<>]{0,1})$")

func (nt *MIDINote) Parse(data string, posIn32th uint) (err error) {

	nt.NoteOn, nt.NoteOff, data = stripNoteOnOff(data)

	mt := regexMIDINote.FindStringSubmatch(data)
	if len(mt) < 2 {
		return fmt.Errorf("not a valid MIDINote: %q", data)
	}
	//	fmt.Printf("parse MIDInotes: %#v\n", mt)
	i, err := strconv.Atoi(mt[1])
	if err != nil {
		return fmt.Errorf("not a valid MIDINote: %q: %s", data, err.Error())
	}
	if i < 1 && i > 128 {
		return fmt.Errorf("not a valid MIDINote: %q: must be >= 1 and <= 128", data)
	}
	nt.Note = int8(i)
	//nt.Velocity = DynamicToVelocity(mt[2])
	nt.Dynamic = mt[2]
	nt.Dotted = mt[3]
	//nt.Dotted = mt[3] == ":"

	if mt[4] == ">" {
		nt.PosShift = 1
	}

	if mt[4] == "<" {
		nt.PosShift = -1
	}
	return nil
}
