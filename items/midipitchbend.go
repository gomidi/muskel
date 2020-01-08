package items

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"gitlab.com/gomidi/midi/midimessage/channel"
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

func (v MIDIPitchbend) WriteMIDI(wr SMFWriter) (addedNotes []uint8) {
	if MIDITrack.pitchbendGlide.active {
		distance := int64(math.Round(float64(wr.Position()+uint64(wr.Delta())-MIDITrack.pitchbendGlide.startPosition) / float64(MIDITrack.GlideResolution(wr))))
		diff := int64(v.Value+8191) - int64(MIDITrack.pitchbendGlide.startValue+8191)
		// fmt.Printf("%v - %v = pb diff: %v\n", int8(v.value+8191), int8(iw.startPBGlissandoValue+8191), diff)

		MIDITrack.pitchbendGlide.glideFunc(wr, distance, diff, func(vl float64) {
			vll := math.Round(vl + float64(MIDITrack.pitchbendGlide.startValue))
			if vll > 8192 {
				vll = 8192
			}

			if vll < -8191 {
				vll = -8191
			}

			wr.Pitchbend(int16(vll))
		})
		MIDITrack.pitchbendGlide.active = false
	}
	//fmt.Printf("MIDIPitchbend %v, \n", int16(v))
	wr.Pitchbend(v.Value)

	if len(v.Tilde) > 0 {
		MIDITrack.pitchbendGlide.startPosition = wr.Position()
		MIDITrack.pitchbendGlide.startValue = v.Value
		MIDITrack.pitchbendGlide.active = true
		MIDITrack.pitchbendGlide.glideFunc = linearGlide
		if v.Tilde == "~~" {
			MIDITrack.pitchbendGlide.glideFunc = exponentialGlide
		}
		wr.BackupTimeline()
	}

	if v.Dotted != "" && v.Value != 0 {
		wr.Plan(0, dottedLengths[v.Dotted][0], dottedLengths[v.Dotted][1], channel.Channel(wr.Channel()).Pitchbend(0))
	}

	return addedNotes
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

	// fmt.Printf("parsing pitchbend %q: %v, normalized: %v\n", dt, val, int16(val))

	pb.Value = int16(val)
	return
}
