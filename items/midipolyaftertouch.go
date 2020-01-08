package items

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"gitlab.com/gomidi/midi/midimessage/channel"
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

func (v MIDIPolyAftertouch) WriteMIDI(wr SMFWriter) (addedNotes []uint8) {
	ki := uint8(int8(v.Key) + MIDITrack.Track.MIDITranspose)

	if MIDITrack.polyaftertouchGlide.active && v.Key == MIDITrack.polyaftertouchGlide.key {
		distance := int64(math.Round(float64(wr.Position()+uint64(wr.Delta())-MIDITrack.polyaftertouchGlide.startPosition) / float64(MIDITrack.GlideResolution(wr))))
		diff := int64(v.Value) - int64(MIDITrack.polyaftertouchGlide.startValue)

		MIDITrack.polyaftertouchGlide.glideFunc(wr, distance, diff, func(vl float64) {
			vll := math.Round(vl + float64(MIDITrack.polyaftertouchGlide.startValue))
			if vll > 127 {
				vll = 127
			}

			if vll < 0 {
				vll = 0
			}

			wr.PolyAftertouch(MIDITrack.polyaftertouchGlide.key, uint8(vll))
		})
		MIDITrack.polyaftertouchGlide.active = false
	}

	wr.PolyAftertouch(ki, v.Value)

	if len(v.Tilde) > 0 {
		MIDITrack.polyaftertouchGlide.key = ki
		MIDITrack.polyaftertouchGlide.startPosition = wr.Position()
		MIDITrack.polyaftertouchGlide.startValue = v.Value
		MIDITrack.polyaftertouchGlide.active = true
		MIDITrack.polyaftertouchGlide.glideFunc = linearGlide
		if v.Tilde == "~~" {
			MIDITrack.polyaftertouchGlide.glideFunc = exponentialGlide
		}
		wr.BackupTimeline()
	}

	if v.Dotted != "" && v.Value != 0 {
		wr.Plan(0, dottedLengths[v.Dotted][0], dottedLengths[v.Dotted][1], channel.Channel(wr.Channel()).PolyAftertouch(ki, 0))
	}

	return addedNotes
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
