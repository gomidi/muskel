package items

import (
	"fmt"
	"sort"
)

func PipedEventStream(p Columner, start uint, endPos uint, evts []*Event) (*EventStream, error) {
	var cl Pattern
	//pc := p.newPattern(&cl)
	pc := NewSketchPattern(p, &cl)
	/*
		printEvents("pipedEventStream before modifyEvents", evts)
		if DEBUG {
			fmt.Printf("pipedEventStream: start: %v endPos: %v\n", start, endPos)
		}
	*/
	evts, absoluteEnd, err := pc.ModifyEvents(start, endPos, evts)
	//printEvents("pipedEventStream after modifyEvents", evts)
	if err != nil {
		return nil, err
	}
	_end := uint(endPos)

	if absoluteEnd < _end {
		_end = absoluteEnd
	}

	return NewEventStream(start, absoluteEnd, true, evts...), nil
}

// eventStream simulates a column of events
type EventStream struct {
	Start      uint     // start position of the sketch column
	Events     []*Event // the events (each character is an event)
	IsOverride bool     // simulates wether the sketch column call is an override
	IsTemplate bool
	End        uint
}

func getFirstType(es *EventStream) string {
	if es == nil || len(es.Events) == 0 || es.Events[0] == nil {
		return ""
	}

	return fmt.Sprintf("%T", es.Events[0].Item)
}

type sortedEventStream []*EventStream

func (s sortedEventStream) Less(a, b int) bool {
	if s[a].Start == s[b].Start {
		if s[a].IsOverride {
			return false
		}

		ta := getFirstType(s[a])
		tb := getFirstType(s[b])

		//fmt.Printf("%q vs %q\n", ta, tb)

		if ta == "*items.Scale" {
			return true
		}

		if tb == "*items.Scale" {
			return false
		}

		return false

	}

	return s[a].Start < s[b].Start
}

func (s sortedEventStream) Swap(a, b int) {
	s[a], s[b] = s[b], s[a]
}

func (s sortedEventStream) Len() int {
	return len(s)
}

func MergeEventStreams(mixed []*EventStream, endPos uint) (evts []*Event) {
	var (
		lastStreamPos uint
		keepLooping   bool = true
		lastStream    *EventStream
	)

	lastStream = nil
	sorted := sortedEventStream(mixed)
	sort.Sort(sorted)

	for idx, stream := range sorted {

		if keepLooping && lastStream != nil {
			if lastStream.inRange(lastStreamPos, stream.Start) {
				var toMod = stream.Start
				if idx+1 < len(mixed) {
					if mixed[idx+1].Start <= stream.Start {
						toMod = mixed[idx+1].Start
					}
				}
				evts = append(evts, lastStream.getEvents(lastStreamPos, toMod)...)
			}
		}

		if stream.IsTemplate && !stream.IsOverride {
			keepLooping = true
			lastStream = stream
			continue
		}

		if !stream.IsTemplate {
			keepLooping = false
			lastStream = nil
		}

		lastStreamPos = stream.End
		evts = append(evts, stream.Events...)
	}

	if keepLooping && lastStream != nil {
		if lastStream.inRange(lastStreamPos, endPos) {
			evts = append(evts, lastStream.getEvents(lastStreamPos, endPos)...)
		}
	}
	return
}

func GetEventsInPosRange(from, to uint, evts []*Event) (res []*Event) {
	/*
		if DEBUG {
			fmt.Printf("getEventsInPosRange(from: %v, to: %v)\n", from, to)
		}
	*/
	/*
		if to < from {
			fmt.Printf("getEventsInPosRange(from: %v, to: %v)\n", from, to)
			panic("invalid pos range")
		}
	*/
	for _, ev := range evts {
		//fmt.Printf("event %v at %v: %v\n", i, ev.Position, ev)
		if ev.Position >= to {
			break
		}
		if ev.Position >= from {
			res = append(res, ev)
		}
	}
	//fmt.Printf("getEventsInPosRange from: %v to: %v => %v\n", from, to, res)

	return
}

func NewEventStream(start, end uint, isTemplate bool, events ...*Event) *EventStream {
	/*
		if DEBUG {
			fmt.Printf("newEventStream(start: %v, end: %v, isTemplate: %v, events: %v)\n", start, end, isTemplate, events)
		}
	*/
	return &EventStream{
		Start:      start,
		End:        end,
		IsOverride: false,
		IsTemplate: isTemplate,
		Events:     events,
	}
}

func NewEventStreamOverride(start, end uint, isTemplate bool, events ...*Event) *EventStream {
	return &EventStream{
		Start:      start,
		End:        end,
		IsOverride: true,
		IsTemplate: isTemplate,
		Events:     events,
	}
}

func (p *EventStream) String() string {
	return fmt.Sprintf("[start: %v end: %v isOverride: %v isTemplate: %v events: %v]", p.Start, p.End, p.IsOverride, p.IsTemplate, p.Events)
}

func (p EventStream) inRange(from, to uint) bool {
	res := p.Start <= to && p.End > from
	//fmt.Printf("%v:%v inRange from: %v to: %v: %v\n", p.start, p.end, from, to, res)
	return res
}

func (l *EventStream) getEvents(from, to uint) (result []*Event) {
	for _, ev := range l.Events {
		if ev.Position >= uint(from) && ev.Position < uint(to) {
			result = append(result, ev)
		}
	}

	return
}