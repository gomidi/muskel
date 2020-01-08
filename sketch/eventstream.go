package sketch

import (
	"fmt"
)

// eventStream simulates a column of events
type eventStream struct {
	start      uint     // start position of the sketch column
	events     []*Event // the events (each character is an event)
	isOverride bool     // simulates wether the sketch column call is an override
	isTemplate bool
	end        uint
}

func newEventStream(start, end uint, isTemplate bool, events ...*Event) *eventStream {
	if DEBUG {
		fmt.Printf("newEventStream(start: %v, end: %v, isTemplate: %v, events: %v)", start, end, isTemplate, events)
	}
	return &eventStream{
		start:      start,
		end:        end,
		isOverride: false,
		isTemplate: isTemplate,
		events:     events,
	}
}

func newEventStreamOverride(start, end uint, isTemplate bool, events ...*Event) *eventStream {
	return &eventStream{
		start:      start,
		end:        end,
		isOverride: true,
		isTemplate: isTemplate,
		events:     events,
	}
}

func (p *eventStream) String() string {
	return fmt.Sprintf("[start: %v end: %v isOverride: %v isTemplate: %v events: %v]", p.start, p.end, p.isOverride, p.isTemplate, p.events)
}

func (p eventStream) inRange(from, to uint) bool {
	res := p.start <= to && p.end > from
	//fmt.Printf("%v:%v inRange from: %v to: %v: %v\n", p.start, p.end, from, to, res)
	return res
}

func (l *eventStream) getEvents(from, to uint) (result []*Event) {
	for _, ev := range l.events {
		if ev.Position >= uint(from) && ev.Position < uint(to) {
			result = append(result, ev)
		}
	}

	return
}
