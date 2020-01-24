package items

import (
	"fmt"
	"strconv"
	"strings"
)

type MIDIPitchbend struct {
	Value  int16
	Dotted string
	Tilde  string
}

var _ Item = &MIDIPitchbend{}

func (m MIDIPitchbend) Dup() Item {
	return &m
}

func (m MIDIPitchbend) String() string {
	var bf strings.Builder

	bf.WriteString(fmt.Sprintf("PB(%v)", m.Value))

	if m.Dotted != "" {
		bf.WriteString(m.Dotted)
	}

	if m.Tilde != "" {
		bf.WriteString(m.Tilde)
	}

	return bf.String()
}

func (pb *MIDIPitchbend) Parse(data string, posIn32th uint) (err error) {

	if idx := strings.Index(data, "~"); idx > -1 {
		pb.Tilde = data[idx:]
		data = data[:idx]
	}

	if idx := strings.Index(data, ":"); idx > -1 {
		pb.Dotted = data[idx:]
		data = data[:idx]
	}

	dt := strings.Trim(data, "()")
	var val int
	val, err = strconv.Atoi(dt)

	if err != nil {
		return
	}

	pb.Value = int16(val)
	return
}
