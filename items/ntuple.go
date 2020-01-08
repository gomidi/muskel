package items

import (
	"fmt"
	"strings"
	//	"gitlab.com/gomidi/midi/mid"
)

type NTuple struct {
	Items    []Item
	PosShift int // 0 = no, 1 = laidback, -1 = ahead of time
	Parser   *Parser
}

func (m NTuple) Dup() Item {
	return &m
}

func (m NTuple) String() string {
	var bf strings.Builder

	bf.WriteString("{")

	for i, it := range m.Items {
		if i > 0 {
			bf.WriteRune(',')
		}
		bf.WriteString(it.String())
	}

	bf.WriteString("}")

	if m.PosShift == 1 {
		bf.WriteString(">")
	}

	if m.PosShift == -1 {
		bf.WriteString("<")
	}

	return bf.String()
}

func (v *NTuple) WriteMIDI(wr SMFWriter) (addedNotes []uint8) {
	MIDITrack.CurrentNTuple = v
	return nil
}

func (v NTuple) writeMIDI(wr SMFWriter, endPos uint) (addedNotes []uint8) {
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

	MIDITrack.noteGlide.active = false

	//length := uint32(v.EndPos - iw.lastNum32th)
	//length := uint32(v.EndPos - iw.currentDistanceToStartOfBarIn32th)
	length := uint32(endPos - MIDITrack.CurrentDistanceToStartOfBarIn32th)
	//fmt.Printf("NTuple endpos: %v lastnum32th: %v length: %v\n", v.EndPos, iw.lastNum32th, length)

	switch v.PosShift {
	case 0:
		MIDITrack.SetStraight()
	case 1:
		MIDITrack.SetLaidBack()
	case -1:
		MIDITrack.SetAhead()
	}

	var delta uint32

	for _, it := range v.Items {
		if it != Hold {
			if delta > 0 {
				wr.Forward(0, delta, uint32(len(v.Items))*32)
			}
			added := MIDITrack.WriteItem(wr, it)
			delta = 0

			for _, addn := range added {
				MIDITrack.NoteOns[addn] = true
			}
		}
		delta += length
	}
	wr.Forward(0, delta, uint32(len(v.Items))*32)

	var comp int32
	switch v.PosShift {
	case 0:
		comp = 0
	case 1:
		comp = CalcNoteDelay(MIDITrack.ResolutionTicks()) * (-1)
	case -1:
		comp = CalcNoteDelay(MIDITrack.ResolutionTicks())
	}

	MIDITrack.SetNoteDelayCompensation(comp)

	MIDITrack.StopNotes(wr)

	MIDITrack.LastNum32th = MIDITrack.CurrentDistanceToStartOfBarIn32th + uint(length)

	wr.Undefined(0, nil)
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
	return addedNotes
}

// ntuple has the form {c,e,d}>
func (ntp *NTuple) Parse(data string, posIn32th uint) (err error) {

	orig := data

	dd := strings.Split(data, "}")

	if len(dd) != 2 {
		return fmt.Errorf("invalid n-tuple: %#v", "{"+orig)
	}

	pos := strings.TrimSpace(dd[1])

	if strings.HasSuffix(pos, ">") {
		ntp.PosShift = 1
		pos = strings.TrimRight(pos, ">")
	}

	if strings.HasSuffix(pos, "<") {
		ntp.PosShift = -1
		pos = strings.TrimRight(pos, "<")
	}

	// fmt.Printf("pos: %q postShift: %v\n", pos, ntp.posShift)

	/*
		if len(pos) == 0 {
			err = fmt.Errorf("invalid n-tuple %#v: position missing", "{"+orig)
			return
		}

		_, endPos, err := items.PositionTo32th("", pos)

		if err != nil {
			return nt, err
		}
	*/

	data = strings.TrimSpace(dd[0])

	d := strings.Split(data, ",")

	if len(d) < 2 {
		err = fmt.Errorf("empty or single item n-tuples are not allowed: %#v", "{"+orig)
		return
	}

	// ntp.EndPos = endPos

	for ppos, it := range d {
		itt, err := parseItem(ntp.Parser, it, posIn32th)
		if err != nil {
			return fmt.Errorf("ERROR invalid n-tuple at position %v: %#v: %s", ppos, it, err)
		}

		/*
			if pc, isPC := itt.(*TemplateCall); isPC {
				itt = pc.FirstItem()
			}
		*/

		if itt != nil {
			ntp.Items = append(ntp.Items, itt)
		}

	}
	return nil
}
