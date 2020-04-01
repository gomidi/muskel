package items

import (
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

	if cl, is := it.(*Pattern); is {
		cl.PosShift = e.PosShift
		e.PosShift = 0
	}

	if cl, is := it.(*Token); is {
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
