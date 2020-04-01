package items

import (
	"fmt"
	"strings"
)

var eventDebug bool

type Event struct {
	Item     Item
	Position uint
	PosShift int // 0 = no, +n = laidback by n 64th, -n = ahead of time by n 64th
}

func (e *Event) _parsePosShift(data string) (rest string, wasShift bool) {
	if len(data) == 0 {
		return "", false
	}
	switch data[len(data)-1] {
	case '>':
		e.PosShift += 1
		return data[:len(data)-1], true
	case '<':
		e.PosShift -= 1
		return data[:len(data)-1], true
	default:
		return data, false
	}
}

func (e *Event) posShiftString() string {
	if e.PosShift == 0 {
		return ""
	}

	var bd strings.Builder

	if e.PosShift < 0 {
		for i := e.PosShift; i < 0; i++ {
			bd.WriteString("<")
		}
		return bd.String()
	}

	for i := e.PosShift; i > 0; i-- {
		bd.WriteString(">")
	}
	return bd.String()
}

func (e *Event) String() string {
	if e.Item == nil {
		if eventDebug {
			return Pos32thToString(e.Position) + "[nil]" + e.posShiftString()
		}
		return ""
	}

	if eventDebug {
		return Pos32thToString(e.Position) + e.Item.String() + e.posShiftString()
	}
	return e.Item.String() + e.posShiftString()
}

func (e *Event) parsePosShift(data string) string {
	var loop = true

	for loop {
		data, loop = e._parsePosShift(data)
	}
	return data
}

func (e *Event) Parse(data string) error {
	var parser Parser

	//fmt.Printf("parsing: %q\n", data)
	data = e.parsePosShift(data)

	it, err := parser.ParseItem(data, 0)

	if err != nil {
		return err
	}

	if cl, is := it.(*Call); is {
		cl.PosShift = e.PosShift
		e.PosShift = 0
	}

	e.Item = it
	//fmt.Printf("result: %s posshift: %v\n", e.Item, e.PosShift)
	return nil
}

func (e *Event) Dup() *Event {
	return &Event{
		Item:     e.Item.Dup(),
		Position: e.Position,
		PosShift: e.PosShift,
	}
}

// AbsPosTicks returns the absolute Position in ticks
func (e *Event) AbsPosTicks(ticks4th uint32, trackDelay int) uint {
	pos := uint(ticks4th) / 8 * e.Position
	if e.PosShift == 0 {
		res := int(pos) + trackDelay

		if res < 0 {
			res = 0
		}
		return uint(res)
	}
	shift := (e.PosShift * int(ticks4th) / 32)

	//fmt.Printf("AbsPosTicks pos: %v shift: %v trackDelay: %v\n", pos, shift, trackDelay)
	res := int(pos) + shift + trackDelay

	if res < 0 {
		res = 0
	}

	return uint(res)
}

type EventSequencer interface {
	Get() EventSequence
}

/*
- slicing mit []
- dynamik change mit + und -
- überschreiben mit /
- diatonische transposition mit ^
- chromatische transposition mit # und b
- octavierungen mit " und '
- startsynchronisierung mit !
- zufallswahrscheinlichkeit mit xx%
- zufällige auswahlmit %()
- microtimingverschiebung mit > und <
*/

type EventSequence []*Event

type EventSequenceMiddleWare []func(in EventSequence) EventSequence

// first function must retrieve the eventsequence (receives nil)
func (mw EventSequenceMiddleWare) Get() (es EventSequence) {
	if len(mw) == 0 {
		return nil
	}

	for _, fn := range mw {
		es = fn(es)
	}

	return
}

type EventSequenceGetter struct {
	mw       EventSequenceMiddleWare
	original string
}

func (e *EventSequenceGetter) Get() *EventSequenceItem {
	es := e.mw.Get()
	return &EventSequenceItem{
		seq:      es,
		original: e.original,
	}
}

func (e *EventSequenceGetter) String() string {
	return e.original
}

func (e *EventSequenceGetter) Parse(data string, posIn32th uint) error {
	panic("TODO: implement")
}

func (e EventSequenceGetter) Dup() Item {
	return &e
}

type EventSequenceItem struct {
	seq      EventSequence
	original string
}

func (e *EventSequenceItem) String() string {
	return e.original
}

func (e EventSequenceItem) Dup() Item {
	return &e
}

func (e *EventSequenceItem) Parse(data string, posIn32th uint) error {
	panic("TODO: implement")
}

// these two appear inside the eventsstream
var _ Item = &EventSequenceGetter{}
var _ Item = &EventSequenceItem{}

type Params []interface{}

func (s EventSequence) Call(params Params) EventSequence {
	panic("TODO implement")
}

func (s EventSequence) Slice(from, to int) EventSequence {
	panic("TODO implement")
}

func (s EventSequence) Override(over EventSequence) EventSequence {
	panic("TODO implement")
}

func (s EventSequence) DiatonicTranspose(steps int8) EventSequence {
	panic("TODO implement")
}

func (s EventSequence) Transpose(halftones int8) EventSequence {
	panic("TODO implement")
}

func (s EventSequence) TransposeOctave(octave int8) EventSequence {
	panic("TODO implement")
}

func (s EventSequence) StartSync(position int64) EventSequence {
	panic("TODO implement")
}

func (s EventSequence) MircoTimingShift(shift int8) EventSequence {
	panic("TODO implement")
}

func (s EventSequence) ChangeDynamic(diff int8) EventSequence {
	panic("TODO implement")
}

func (s EventSequence) Len() int {
	return len(s)
}

func (s EventSequence) Swap(a, b int) {
	s[a], s[b] = s[b], s[a]
}

var typeSortEventPriority = map[string]int{
	// all others are 0
	"*items.Scale": 20,
}

func (s EventSequence) Less(a, b int) bool {
	if s[a].Position == s[b].Position {
		typeA := fmt.Sprintf("%T", s[a].Item)
		typeB := fmt.Sprintf("%T", s[b].Item)
		prioA := typeSortEventPriority[typeA]
		prioB := typeSortEventPriority[typeB]

		//fmt.Printf("types: %q vs %q prios: %v vs %v\n", typeA, typeB, prioA, prioB)
		if typeA != typeB {
			return prioA > prioB
		}

		switch v := s[a].Item.(type) {
		case *MIDICC:
			other := s[b].Item.(*MIDICC)
			return v.Controller < other.Controller
		default:
			return false
		}

	}

	//fmt.Printf("types: %T vs %T pos: %v vs %v\n", e[a].message, e[b].message, e[a].position, e[b].position)
	return s[a].Position < s[b].Position
}
