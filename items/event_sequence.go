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
- überschreiben mit /
- startsynchronisierung mit !

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
items.Pattern     -> EventSequence
items.PartRepeat  -> EventSequence
items.BarRepeater -> EventSequence
items.Includes    -> EventSequence
items.NTuple      -> EventSequence
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

func (s EventSequence) Call(mod *eventSequenceModifier) EventSequence {
	panic("TODO implement")
}

/*
func (s EventSequence) Slice(mod *eventSequenceModifier) (evs EventSequence) {

	if mod.SingleSlice >= 0 {
		if len(s) <= mod.SingleSlice {
			return nil
		}
		return EventSequence{s[mod.SingleSlice]}
	}

	first, last := mod.Slice[0], mod.Slice[1]
	if first < 0 {
		return s
	}

	var collecting bool

	for i, ev := range s {
		if last > 0 && i >= last {
			break
		}

		if i == first {
			collecting = true
		}

		if collecting {
			evs = append(evs, ev.Dup())
		}
	}
	return
}
*/

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

type dynamicSetter interface {
	SetDynamic(string)
}

func (s EventSequence) ChangeDynamic(mod *eventSequenceModifier) (evs EventSequence) {
	for _, ev := range s {
		if _, is := ev.Item.(dynamicSetter); is {
			nu := ev.Dup()
			d := nu.Item.(dynamicSetter)
			d.SetDynamic(mod.DynamicAdd)
			evs = append(evs, nu)
		}
	}
	return

	/*
		switch v := it.(type) {
		case *items.Note:
			it := v.Dup().(*items.Note)
			if c != nil {
				it.Dynamic = cc.AddDynamic(v.Dynamic)
				if it.PosShift == 0 && cc.PosShift != 0 {
					it.PosShift = cc.PosShift
				}
			}
			return it, nil
		case *items.MIDINote:
			it := v.Dup().(*items.MIDINote)
			if c != nil {
				it.Dynamic = cc.AddDynamic(v.Dynamic)
				if it.PosShift == 0 && cc.PosShift != 0 {
					it.PosShift = cc.PosShift
				}
			}
			return it, nil

		case *items.Token:
			if c != nil {
				pc := c.column.newToken(v)
				//posEv, err = pc.getEventStream(forward+ev.Position, endPos)
				posEv, err := pc.getEventStream(0, 1)
				if err != nil {
					return nil, err
				}
				it := posEv.events[0].Item
				return it, nil
			} else {
				return it.Dup(), nil
			}

		case *items.LyricsTable:
			return it.Dup(), nil

		case *items.Pattern:
			return it.Dup(), nil

		case *items.NTuple:
			if c != nil {
				it := v.Dup().(*items.NTuple)
				it, err := c.column.replaceNtupleTokens(v)

				if err != nil {
					return nil, err
				}

				return it, nil
			} else {
				return v.Dup(), nil
			}
		default:
			return it.Dup(), nil
		}
	*/
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
