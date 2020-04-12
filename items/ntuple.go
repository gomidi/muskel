package items

import (
	"bytes"
	"fmt"
	"strings"
)

type NTuple struct {
	Events      []*Event
	LengthTicks uint
}

func (m NTuple) Dup() Item {
	return &m
}

func (m NTuple) String() string {
	var bf strings.Builder

	bf.WriteString("{")

	for i, it := range m.Events {
		if i > 0 {
			bf.WriteRune(',')
		}
		bf.WriteString(it.String())
	}

	bf.WriteString("}")

	return bf.String()
}

func (ntp *NTuple) split(data string) (res []string) {
	var embed int
	var bf bytes.Buffer
	for _, char := range data {
		switch char {
		case '{':
			embed++
			bf.WriteRune(char)
		case '}':
			embed--
			bf.WriteRune(char)
		case ',':
			if embed > 0 {
				bf.WriteRune(char)
			} else {
				res = append(res, bf.String())
				bf.Reset()
			}
		default:
			bf.WriteRune(char)
		}
	}

	res = append(res, bf.String())
	return

}

var _ UnrollGetter = &NTuple{}

func (p *NTuple) GetES(c Columner, ev *Event, start, endPos uint) ([]*EventStream, error) {
	// TODO look inside each item and replace random things and templatecalls if there
	newItm, err := ReplaceNtupleTokens(c, p)
	if err != nil {
		return nil, err
	}

	var nev = ev.Dup()
	nev.Position = start
	nev.PosShift = 0
	nev.Item = newItm
	posEv := NewEventStream(nev.Position, endPos, false, nev)
	return []*EventStream{posEv}, nil
}

func (ntp *NTuple) SetDynamic(dyn string) {
	for _, ev := range ntp.Events {
		_ = ev
		panic("TODO")
	}
	// TODO resolv and then set it
	//nt.Dynamic = AddDynamic(nt.Dynamic, dyn)
}

// ntuple has the form {c,e,d}>
func (ntp *NTuple) Parse(data string, posIn32th uint) (err error) {
	data = data[1:]
	orig := data

	idx := strings.LastIndex(data, "}")

	if idx <= 0 {
		return fmt.Errorf("invalid n-tuple: %#v", "{"+orig)
	}

	//data = strings.TrimSpace(strings.TrimLeft(data[idx:], "}"))

	/*
		if strings.HasSuffix(pos, ">") {
			ntp.PosShift = 1
			pos = strings.TrimRight(pos, ">")
		}

		if strings.HasSuffix(pos, "<") {
			ntp.PosShift = -1
			pos = strings.TrimRight(pos, "<")
		}
	*/
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

	data = strings.TrimSpace(data[:idx])

	d := ntp.split(data)

	if len(d) < 2 {
		err = fmt.Errorf("empty or single item n-tuples are not allowed: %#v", "{"+orig)
		return
	}

	// ntp.EndPos = endPos

	for ppos, dat := range d {
		var ev Event
		err := ev.Parse(dat)

		//itt, err := parseItem(ntp.Parser, it, posIn32th)
		if err != nil {
			return fmt.Errorf("ERROR invalid n-tuple at position %v: %#v: %s", ppos, dat, err)
		}

		/*
			if pc, isPC := itt.(*TemplateCall); isPC {
				itt = pc.FirstItem()
			}
		*/

		ntp.Events = append(ntp.Events, &ev)

	}
	return nil
}
