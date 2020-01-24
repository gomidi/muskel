package smf

import (
	"fmt"

	"gitlab.com/gomidi/midi"
	"gitlab.com/gomidi/midi/midimessage/channel"
)

type event struct {
	position  uint
	monitor   bool
	stopNotes bool
	message   midi.Message
}

type events []*event

func (e events) Swap(a, b int) {
	e[a], e[b] = e[b], e[a]
}

func (e *events) add(ev *event) {
	*e = append(*e, ev)
}

func (e events) Len() int {
	return len(e)
}

var typeSortPriority = map[string]int{
	// meta messages / all others are 0
	"meta.Copyright":         20,
	"meta.TrackSequenceName": 15,
	"meta.Instrument":        10,
	"meta.TimeSig":           8,
	"meta.Tempo":             7,
	"channel.NoteOff":        -1,
	"channel.ControlChange":  -2,
	"channel.ProgramChange":  -3,
	"channel.Pitchbend":      -4,
	"channel.NoteOn":         -10,
	"meta.Lyric":             -15,
}

func (e events) Less(a, b int) bool {
	if e[a].position == e[b].position {
		typeA := fmt.Sprintf("%T", e[a].message)
		typeB := fmt.Sprintf("%T", e[b].message)
		prioA := typeSortPriority[typeA]
		prioB := typeSortPriority[typeB]

		//fmt.Printf("types: %q vs %q prios: %v vs %v\n", typeA, typeB, prioA, prioB)
		if typeA != typeB {
			return prioA > prioB
		}

		if e[a].stopNotes != e[b].stopNotes {
			return e[a].stopNotes
		}

		switch v := e[a].message.(type) {
		case channel.NoteOn:
			other := e[b].message.(channel.NoteOn)
			return v.Key() < other.Key()
		case channel.NoteOff:
			other := e[b].message.(channel.NoteOff)
			return v.Key() < other.Key()
		case channel.PolyAftertouch:
			other := e[b].message.(channel.PolyAftertouch)
			return v.Key() < other.Key()
		default:
			return false
		}

	}

	//fmt.Printf("types: %T vs %T pos: %v vs %v\n", e[a].message, e[b].message, e[a].position, e[b].position)
	return e[a].position < e[b].position
}
