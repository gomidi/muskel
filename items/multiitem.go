package items

import (
	"fmt"
	"strings"
)

type MultiItem struct {
	Events []*Event
	itemGroupModifier
}

func (m *MultiItem) SetDynamic(dyn string) {
	for _, ev := range m.Events {
		_ = ev
		panic("TODO")
	}
	// TODO resolv and then set it
	//nt.Dynamic = AddDynamic(nt.Dynamic, dyn)
}

var _ UnrollGetter = &MultiItem{}

func (p *MultiItem) GetES(c Columner, ev *Event, start, endPos uint) (x []*EventStream, err error) {
	// TODO look inside each item and replace random things and templatecalls if there
	var posEv *EventStream
	var newItsm = make([]*Event, len(p.Events))
	for itidx, itm := range p.Events {
		switch vv := itm.Item.(type) {
		case *Pattern:
			posEv, err = vv.GetEventStream(c, start, endPos)
			if err != nil {
				return nil, err
			}
			newItsm[itidx] = posEv.Events[0]
		case *Token:
			posEv, err = vv.GetEventStream(c, start, endPos)
			if err != nil {
				return nil, err
			}
			newItsm[itidx] = posEv.Events[0]
		default:
			newItsm[itidx] = itm
		}
	}
	p.Events = newItsm
	var nev = ev.Dup()
	nev.Position = start
	nev.Item = p.Dup()
	posEv = NewEventStream(false, nev.Position, endPos, false, nev)
	return []*EventStream{posEv}, nil
}

func (m *MultiItem) Parse(data string, posIn32th uint) (err error) {
	idx := strings.Index(data, "(")
	endIdx := strings.LastIndex(data, ")")

	if idx < 0 || endIdx < 0 || endIdx <= idx {
		return fmt.Errorf("invalid multiItem: %q", data)
	}

	m.itemGroupModifier.Reset()

	innerData := data[idx+1 : endIdx]

	if len(data) > endIdx {
		rest := m.itemGroupModifier.parseExploded(data[endIdx+1:])
		err = m.itemGroupModifier.parse(rest)
		if err != nil {
			return fmt.Errorf("error in multiitem %q, item group modifier: %s", data, err)
		}
	}

	//fmt.Printf("slice: %v\n", m.Slice)

	d := strings.Split(innerData, " ")
	for _, dd := range d {
		var ev Event

		err = ev.Parse(strings.TrimSpace(dd))

		if err != nil {
			return fmt.Errorf("error in multiitem %q, item %q: %s", data, dd, err)
		}

		m.Events = append(m.Events, &ev)
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
	for i, it := range m.Events {
		if it != nil {
			if i > 0 {
				bd.WriteString(" ")
			}
			bd.WriteString(it.String())
		}
	}

	bd.WriteString(")")

	return bd.String() + m.itemGroupModifier.string()
}

func (m MultiItem) IsHold() bool {
	var hasStoppingNote bool
	var hasHold bool

	for _, val := range m.Events {
		switch v := val.Item.(type) {
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

type sortMultiItem []*Event

func (s sortMultiItem) Len() int {
	return len(s)
}

func (s sortMultiItem) Swap(a, b int) {
	s[a], s[b] = s[b], s[a]
}

func (s sortMultiItem) Less(a, b int) bool {
	if s[a].Position == s[b].Position {
		return s[a].PosShift < s[b].PosShift
	}
	return s[a].Position < s[b].Position
}
