package items

import (
	"strings"
)

type MultiItem struct {
	Items    []Item
	Parser   *Parser
	Exploded bool // exploded means: =patt((c d b)...) will become  =patt(c,d,b)
}

func (m *MultiItem) Parse(data string, posIn32th uint) (err error) {
	d := strings.Split(data, " ")
	for _, dd := range d {
		it, err := parseItem(m.Parser, strings.TrimSpace(dd), posIn32th)

		if err != nil {
			return err
		}

		m.Items = append(m.Items, it)
	}

	return nil
}

var _ Item = &MultiItem{}

func (m MultiItem) Dup() Item {
	return &m
}

func (m MultiItem) String() string {
	var bd strings.Builder
	bd.WriteString("(")
	for i, it := range m.Items {
		if it != nil {
			if i > 0 {
				bd.WriteString(" ")
			}
			bd.WriteString(it.String())
		}
	}

	bd.WriteString(")")

	return bd.String()
}

func (m MultiItem) IsHold() bool {
	var hasStoppingNote bool
	var hasHold bool

	for _, val := range m.Items {
		switch v := val.(type) {
		case *Note:
			if !v.NoteOn && !v.NoteOff {
				hasStoppingNote = true
			}
		case *MIDINote:
			if !v.NoteOn && !v.NoteOff {
				hasStoppingNote = true
			}
		case hold:
			hasHold = true
		}
	}

	return hasHold && !hasStoppingNote
}

func (v MultiItem) WriteMIDI(wr SMFWriter) (addedNotes []uint8) {
	if !v.IsHold() {
		MIDITrack.StopNotes(wr)
	}

	cachedNoteOns := MIDITrack.NoteOns

	MIDITrack.NoteOns = map[uint8]bool{}

	for _, it := range v.Items {
		if it != Hold {
			// ignore the returned velocities
			added := MIDITrack.WriteItem(wr, it)
			addedNotes = append(addedNotes, added...)
		}
	}

	MIDITrack.NoteOns = cachedNoteOns
	return addedNotes
}
