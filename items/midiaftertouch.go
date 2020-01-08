package items

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"gitlab.com/gomidi/midi/midimessage/channel"
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

func (v MIDIAftertouch) WriteMIDI(wr SMFWriter) (addedNotes []uint8) {
	if MIDITrack.aftertouchGlide.active {
		distance := int64(math.Round(float64(wr.Position()+uint64(wr.Delta())-MIDITrack.aftertouchGlide.startPosition) / float64(MIDITrack.GlideResolution(wr))))
		diff := int64(v.Value) - int64(MIDITrack.aftertouchGlide.startValue)

		MIDITrack.aftertouchGlide.glideFunc(wr, distance, diff, func(vl float64) {
			vll := math.Round(vl + float64(MIDITrack.aftertouchGlide.startValue))
			if vll > 127 {
				vll = 127
			}

			if vll < 0 {
				vll = 0
			}

			wr.Aftertouch(uint8(vll))
		})
		MIDITrack.aftertouchGlide.active = false
	}

	wr.Aftertouch(v.Value)

	if len(v.Tilde) > 0 {
		MIDITrack.aftertouchGlide.startPosition = wr.Position()
		MIDITrack.aftertouchGlide.startValue = v.Value
		MIDITrack.aftertouchGlide.active = true
		MIDITrack.aftertouchGlide.glideFunc = linearGlide
		if v.Tilde == "~~" {
			MIDITrack.aftertouchGlide.glideFunc = exponentialGlide
		}
		wr.BackupTimeline()
	}

	if v.Dotted != "" && v.Value > 0 {
		wr.Plan(0, dottedLengths[v.Dotted][0], dottedLengths[v.Dotted][1], channel.Channel(wr.Channel()).Aftertouch(0))
	}

	return
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
