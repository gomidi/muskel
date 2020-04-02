package items

import "fmt"

type EventSequencer interface {
	Get() EventSequence
}

type EventSequencerFunc func() EventSequence

func (e EventSequencerFunc) Get() EventSequence {
	return e()
}

/*
von item_group
- slicing mit []
- dynamik change mit + und -
- diatonische transposition mit ^
- chromatische transposition mit # und b
- octavierungen mit " und '
- microtimingverschiebung mit > und <
- zufallswahrscheinlichkeit mit xx%
- zufällige auswahlmit %()
*/

/*
- überschreiben mit /
- startsynchronisierung mit !

- slicing mit []
- dynamik change mit + und -
- diatonische transposition mit ^
- chromatische transposition mit # und b
- octavierungen mit " und '
- microtimingverschiebung mit > und <
- zufallswahrscheinlichkeit mit xx%
- zufällige auswahlmit %()
*/

/*
items.Pattern     -> EventSequence
items.PartRepeat  -> EventSequence
items.BarRepeater -> EventSequence
*/

type eventSequenceModifier struct {
	SyncFirst bool
	Repeat    uint
	firstPos  uint
	itemGroupModifier
}

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
	Original string
}

func (e *EventSequenceGetter) Get() *EventSequenceItem {
	es := e.mw.Get()
	return &EventSequenceItem{
		EventSequence: es,
		Original:      e.Original,
	}
}

func (e *EventSequenceGetter) String() string {
	return e.Original
}

func (e *EventSequenceGetter) Parse(data string, posIn32th uint) error {
	panic("TODO: implement")
}

func (e EventSequenceGetter) Dup() Item {
	return &e
}

type EventSequenceItem struct {
	EventSequence EventSequence
	Original      string
}

func (e *EventSequenceItem) String() string {
	return e.Original
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
