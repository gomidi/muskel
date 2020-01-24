package items

import (
	"fmt"
	"strconv"
	"strings"
)

type MIDICC struct {
	Controller uint8
	Value      uint8
	Dotted     string
	Tilde      string
}

var _ Item = &MIDICC{}

func (m MIDICC) Dup() Item {
	return &m
}

func (m MIDICC) String() string {
	var bf strings.Builder

	bf.WriteString(fmt.Sprintf("CC(%v,%v)", m.Controller, m.Value))

	if m.Dotted != "" {
		bf.WriteString(m.Dotted)
	}

	if m.Tilde != "" {
		bf.WriteString(m.Tilde)
	}

	return bf.String()
}

func (cc *MIDICC) Parse(data string, posIn32th uint) (err error) {
	var ccval int = -1

	if len(data) > 2 {
		if strings.ToLower(data[0:2]) == "on" {
			ccval = 127
			data = data[2:]
		}

		if strings.ToLower(data[0:3]) == "off" {
			ccval = 0
			data = data[3:]
		}
	}

	if idx := strings.Index(data, "~"); idx > -1 {
		cc.Tilde = data[idx:]
		data = data[:idx]
	}

	if idx := strings.Index(data, ":"); idx > -1 {
		cc.Dotted = data[idx:]
		data = data[:idx]
	}

	dt := strings.Trim(data, "()")
	d := strings.Split(dt, ",")
	if ccval == -1 && len(d) != 2 {
		err = fmt.Errorf("invalid format for MIDI CC: %#v, must be CC(nn,mm) where nn is the controller number and mm is the controler value", data)
		return
	}

	var ccnum int
	ccnum, err = strconv.Atoi(d[0])

	if err != nil {
		return
	}

	if ccval > -1 {
		cc.Controller = uint8(ccnum)
		cc.Value = uint8(ccval)
		return
	}

	ccval, err = strconv.Atoi(d[1])

	if err != nil {
		return
	}

	cc.Controller = uint8(ccnum)
	cc.Value = uint8(ccval)

	return
}
