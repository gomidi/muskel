package items

import (
	"fmt"
	"strconv"
	"strings"
)

type MIDIPolyAftertouch struct {
	Key    uint8
	Value  uint8
	Dotted string
	Tilde  string
}

var _ Item = &MIDIPolyAftertouch{}

func (m MIDIPolyAftertouch) Dup() Item {
	return &m
}

func (m MIDIPolyAftertouch) String() string {
	var bf strings.Builder

	bf.WriteString(fmt.Sprintf("PT(%v,%v)", m.Key, m.Value))

	if m.Dotted != "" {
		bf.WriteString(m.Dotted)
	}

	if m.Tilde != "" {
		bf.WriteString(m.Tilde)
	}

	return bf.String()
}

func (pt *MIDIPolyAftertouch) Parse(data string, posIn32th uint) (err error) {
	if idx := strings.Index(data, "~"); idx > -1 {
		pt.Tilde = data[idx:]
		data = data[:idx]
	}

	if idx := strings.Index(data, ":"); idx > -1 {
		pt.Dotted = data[idx:]
		data = data[:idx]
	}

	dt := strings.Trim(data, "()")
	d := strings.Split(dt, ",")
	if len(d) != 2 {
		err = fmt.Errorf("invalid format for MIDI PT: %#v, must be PT(nt,mm) where nt is the note and mm is the value", data)
		return
	}

	var nt = &Note{}
	err = nt.Parse(d[0], posIn32th)
	if err != nil {
		err = fmt.Errorf("invalid format for MIDI PT can't parse note: %s", err.Error())
		return
	}

	var val int
	val, err = strconv.Atoi(d[1])

	if err != nil {
		err = fmt.Errorf("invalid format for MIDI PT can't parse value: %s", err.Error())
		return
	}

	pt.Key = nt.ToMIDI()
	pt.Value = uint8(val)

	return
}
