package items

import (
	"fmt"
	"strings"
)

var _ Item = &OSCMessage{}

type OSCMessage struct {
	Path string
	Args []interface{}
}

func (*OSCMessage) Parse(data string, posIn32th uint) (err error) {
	panic("not implemented")
}

var _ Item = &OSCMessage{}

func (m OSCMessage) Dup() Item {
	return &m
}

func (m OSCMessage) String() string {
	var bf strings.Builder

	bf.WriteString(fmt.Sprintf("OSC(%s", m.Path))

	for _, arg := range m.Args {
		bf.WriteString(fmt.Sprintf(",%v", arg))
	}

	bf.WriteString(")")

	return bf.String()
}

func (v OSCMessage) WriteMIDI(wr SMFWriter) (addedNotes []uint8) {
	return nil
}
