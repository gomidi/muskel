package items

import (
	"fmt"
	"strconv"
	"strings"
)

type MIDIAftertouch struct {
	Value  uint8
	Dotted string
	Tilde  string
}

var _ Item = &MIDIAftertouch{}

func (m MIDIAftertouch) Dup() Item {
	return &m
}

func (m MIDIAftertouch) String() string {
	var bf strings.Builder

	bf.WriteString(fmt.Sprintf("AT(%v)", m.Value))

	if m.Dotted != "" {
		bf.WriteString(m.Dotted)
	}

	if m.Tilde != "" {
		bf.WriteString(m.Tilde)
	}

	return bf.String()
}

func (at *MIDIAftertouch) Parse(data string, posIn32th uint) (err error) {
	if idx := strings.Index(data, "~"); idx > -1 {
		at.Tilde = data[idx:]
		data = data[:idx]
	}

	if idx := strings.Index(data, ":"); idx > -1 {
		at.Dotted = data[idx:]
		data = data[:idx]
	}

	dt := strings.Trim(data, "()")
	var val int
	val, err = strconv.Atoi(dt)

	if err != nil {
		return
	}

	at.Value = uint8(val)
	return
}
