package items

import "strings"

// CalledEvents is a column of events, the events have the position relative to the start
/*
type CalledEvents struct {
	Events   []*Event
	Start uint
	End   uint
	Repeat   uint
}
*/

func (c *EventStream) Length32ths() uint {
	return c.End - c.Start
}

func (c *EventStream) SetStart(syncFirst bool, start uint) {
	if syncFirst {
		c.setStartToFirstEvent()
		return
	}
	c.Start = start
	return
}

func (c *EventStream) setStartToFirstEvent() {
	for _, ev := range c.Events {
		if ev == nil {
			continue
		}

		if ev.Item == nil {
			continue
		}

		if ev.Item == End {
			return
		}

		if ev.Item == Rest {
			continue
		}

		/*
			switch v := ev.Item.(type) {
			case BarChange:
				c.PosStart += v.Length32ths()
			default:
		*/
		c.Start += ev.Position
		return
		/*
			}
		*/
	}
}

// SetEndToEndEvent sets the PosEnd to the first End item
// and returns the position. If there is no End item, 0 is returned
func (c *EventStream) SetEnd(end uint) {
	for _, ev := range c.Events {
		if ev.Item == End {
			c.End = ev.Position
			return
		}
	}
	c.End = end
	return
}

func (c *EventStream) GetEvents(placementPos uint) (res []*Event) {
	diff := placementPos
	for i := c.Repeat; i >= 0; i-- {
		diff += i * c.Length32ths()
		res = append(res, c.forward(c.Events, diff)...)
	}
	return
}

func (c *EventStream) Normalize() {
	length := c.Length32ths()
	res := c.backward(c.Events, c.Start)
	c.Start = 0
	c.End = length
	c.Events = res
	return
}

func (c *EventStream) backward(src []*Event, diff uint) (res []*Event) {
	for _, ev := range src {
		res = append(res, backwardEvent(ev, diff))
	}
	return
}

func (c *EventStream) forward(src []*Event, diff uint) (res []*Event) {
	for _, ev := range src {
		/*
			switch v := ev.Item.(type) {
			case BarChange:
				diff += v.Length32ths()
			default:
		*/
		res = append(res, forwardEvent(ev, diff))
		//}
	}
	return
}

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
	//var parser Parser

	//fmt.Printf("parsing: %q\n", data)
	data = e.parsePosShift(data)

	it, err := Parse(data, 0)

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

func ForwardEvents(ins []*Event, diff uint) (outs []*Event) {
	if diff == 0 {
		return ins
	}
	for _, ev := range ins {
		outs = append(outs, forwardEvent(ev, diff))
	}
	return
}

func forwardEvent(in *Event, diff uint) (out *Event) {
	if diff == 0 {
		return in.Dup()
	}
	out = in.Dup()
	out.Position += diff
	return
}

func backwardEvent(in *Event, diff uint) (out *Event) {
	if diff == 0 {
		return in.Dup()
	}
	out = in.Dup()
	out.Position -= diff
	return
}

func MoveBySyncFirst(ins []*Event) (out []*Event) {
	var posDiff uint
	for i, ev := range ins {
		if ev == nil {
			continue
		}

		if i == 0 {
			posDiff = ev.Position
		}

		nev := ev.Dup()
		nev.Position = ev.Position - posDiff

		out = append(out, nev)
	}
	return
}
