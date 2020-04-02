package items

import (
	"fmt"
	"strconv"
	"strings"
)

type MIDIProgramChange struct {
	Value uint8
}

var _ Item = &MIDIProgramChange{}

func (m MIDIProgramChange) Dup() Item {
	return &m
}

func (m MIDIProgramChange) String() string {
	return fmt.Sprintf("PC(%v)", m.Value)
}

func (at *MIDIProgramChange) Parse(data string, posIn32th uint) (err error) {
	data = data[2:]
	dt := strings.Trim(data, "()")
	var val int
	val, err = strconv.Atoi(dt)

	if err != nil {
		return
	}

	at.Value = uint8(val - 1)
	return
}
