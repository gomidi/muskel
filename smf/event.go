package smf

import "gitlab.com/gomidi/midi/v2/smf"

type event struct {
	position  uint
	monitor   bool
	stopNotes bool
	message   smf.Message
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
	"meta.Marker":            -20,
	"meta.Cuepoint":          -25,
}

func (e events) Less(a, b int) bool {
	if e[a].position == e[b].position {
		msgA := e[a].message
		msgB := e[b].message

		typeA := msgA.Type().String()
		typeB := msgB.Type().String()
		prioA := typeSortPriority[typeA]
		prioB := typeSortPriority[typeB]

		//fmt.Printf("types: %q vs %q prios: %v vs %v\n", typeA, typeB, prioA, prioB)
		if typeA != typeB {
			return prioA > prioB
		}

		if e[a].stopNotes != e[b].stopNotes {
			return e[a].stopNotes
		}

		var keyA, keyB uint8
		var chanA, chanB uint8
		var velA, velB uint8

		if (msgA.GetNoteOn(&chanA, &keyA, &velA) || msgA.GetNoteOff(&chanA, &keyA, &velA) || msgA.GetPolyAfterTouch(&chanA, &keyA, &velA)) &&
			(msgB.GetNoteOn(&chanB, &keyB, &velB) || msgB.GetNoteOff(&chanB, &keyB, &velB) || msgB.GetPolyAfterTouch(&chanB, &keyB, &velB)) {
			return keyA < keyB
		}
		return false
	}

	//fmt.Printf("types: %T vs %T pos: %v vs %v\n", e[a].message, e[b].message, e[a].position, e[b].position)
	return e[a].position < e[b].position
}
