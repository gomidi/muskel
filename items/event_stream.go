package items

import (
	"fmt"
	"sort"
)

func PipedEventStream(p Columner, start uint, endPos uint, evts []*Event) (*EventStream, error) {
	var cl Pattern
	pc := cl.newSketchPattern(p)
	evts, absoluteEnd, err := pc.modifyEvents(start, endPos, evts)
	if err != nil {
		return nil, err
	}
	_end := uint(endPos)

	if absoluteEnd < _end {
		_end = absoluteEnd
	}

	return NewEventStream(cl.SyncFirst, start, absoluteEnd, true, evts...), nil
}

// eventStream simulates a column of events
type EventStream struct {
	Start      uint     // start position of the sketch column
	Events     []*Event // the events (each character is an event)
	IsOverride bool     // simulates wether the sketch column call is an override
	IsTemplate bool
	End        uint
	Repeat     uint
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

func MergeEventStreams(mixed []*EventStream, endPos uint) (es *EventStream) {
	var (
		lastStreamPos uint
		keepLooping   bool = true
		lastStream    *EventStream
	)
	es = &EventStream{}

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
				es.Events = append(es.Events, lastStream.getEvents(lastStreamPos, toMod)...)
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
		es.Events = append(es.Events, stream.Events...)
	}

	if keepLooping && lastStream != nil {
		if lastStream.inRange(lastStreamPos, endPos) {
			es.Events = append(es.Events, lastStream.getEvents(lastStreamPos, endPos)...)
		}
	}

	es.SetEnd(endPos)
	return
}

func GetEventsInPosRange(from, to uint, evts []*Event) (res []*Event) {
	for _, ev := range evts {
		if ev.Position >= to {
			break
		}
		if ev.Position >= from {
			res = append(res, ev)
		}
	}
	return
}

func NewEventStream(syncFirst bool, start, end uint, isTemplate bool, events ...*Event) *EventStream {
	es := &EventStream{
		//Start:      start,
		//End:        end,
		IsOverride: false,
		IsTemplate: isTemplate,
		Events:     events,
	}
	es.SetStart(syncFirst, start)
	es.SetEnd(end)
	return es
}

func NewEventStreamOverride(syncFirst bool, start, end uint, isTemplate bool, events ...*Event) *EventStream {
	es := &EventStream{
		//Start:      start,
		//End:        end,
		IsOverride: true,
		IsTemplate: isTemplate,
		Events:     events,
	}
	es.SetStart(syncFirst, start)
	es.SetEnd(end)
	return es
}

func (p *EventStream) String() string {
	return fmt.Sprintf("[start: %v end: %v isOverride: %v isTemplate: %v events: %v]", p.Start, p.End, p.IsOverride, p.IsTemplate, p.Events)
}

func (p EventStream) inRange(from, to uint) bool {
	return p.Start <= to && p.End > from
}

func (l *EventStream) getEvents(from, to uint) (result []*Event) {
	return GetEventsInPosRange(from, to, l.Events)
}
