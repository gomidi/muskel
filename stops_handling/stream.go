package stops_handling

import (
	"fmt"
)

/*
	A stream is a collection of Events, returned from an PositionedItem via the Stream method
*/

type Stream struct {
	Item   Item
	Offset int // startposition relativ to Anchor
	Anchor uint
	Events []Event
	Length uint // length in ticks
}

func (it *Stream) FirstPos() uint {
	diff := int(it.Anchor) + it.Offset
	if diff < 0 {
		panic(fmt.Sprintf("Anchor %v is too late, need offset of %v", it.Anchor, it.Offset))
	}
	return uint(diff)
}

type Streams []*Stream

func (s Streams) Less(a, b int) bool {
	return s[a].FirstPos() < s[b].FirstPos()
}

func (s Streams) Swap(a, b int) {
	s[a], s[b] = s[b], s[a]
}

func (s Streams) Len() int {
	return len(s)
}
