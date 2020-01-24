package items

import (
	"bytes"
	"fmt"
	"strings"
	//	"gitlab.com/gomidi/midi/mid"
)

type NTuple struct {
	//Items       []Item
	Events      []*Event
	Parser      *Parser
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

// ntuple has the form {c,e,d}>
func (ntp *NTuple) Parse(data string, posIn32th uint) (err error) {

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
