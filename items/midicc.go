package items

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"gitlab.com/gomidi/midi/midimessage/channel"
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

func (v MIDICC) WriteMIDI(wr SMFWriter) (addedNotes []uint8) {

	//fmt.Printf("MIDICC %v, %v\n", v[0], v[1])

	if MIDITrack.ccGlide.active && v.Controller == MIDITrack.ccGlide.controller {
		//distance := int64(math.Round(float64(iw.wr.Position()+uint64(iw.wr.Delta())-iw.startCCGlissando) / float64(iw.wr.Ticks32th())))
		distance := int64(math.Round(float64(wr.Position()+uint64(wr.Delta())-MIDITrack.ccGlide.startPosition) / float64(MIDITrack.GlideResolution(wr))))
		diff := int64(v.Value) - int64(MIDITrack.ccGlide.startValue)

		MIDITrack.ccGlide.glideFunc(wr, distance, diff, func(vl float64) {
			vll := math.Round(vl + float64(MIDITrack.ccGlide.startValue))
			if vll > 127 {
				vll = 127
			}

			if vll < 0 {
				vll = 0
			}

			wr.ControlChange(MIDITrack.ccGlide.controller, uint8(vll))
		})
		MIDITrack.ccGlide.active = false
	}

	wr.ControlChange(v.Controller, v.Value)

	if len(v.Tilde) > 0 {
		MIDITrack.ccGlide.controller = v.Controller
		MIDITrack.ccGlide.startPosition = wr.Position()
		MIDITrack.ccGlide.startValue = v.Value
		MIDITrack.ccGlide.active = true
		MIDITrack.ccGlide.glideFunc = linearGlide
		if v.Tilde == "~~" {
			MIDITrack.ccGlide.glideFunc = exponentialGlide
		}
		wr.BackupTimeline()
	}

	if v.Dotted != "" && v.Value > 0 {
		wr.Plan(0, dottedLengths[v.Dotted][0], dottedLengths[v.Dotted][1], channel.Channel(wr.Channel()).ControlChange(v.Controller, 0))
	}
	return addedNotes
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
