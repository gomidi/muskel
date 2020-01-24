package items

import (
	"fmt"
	"regexp"
	"strconv"
)

type MIDINote struct {
	Note     int8
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
