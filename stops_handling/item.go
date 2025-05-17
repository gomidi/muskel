package stops_handling

type Item int

const (
	RestItem                 Item = 0
	NoteItem                 Item = 1
	NoteOnItem               Item = 2
	NoteOffItem              Item = 3
	PatternItem              Item = 4
	NotePatternOverwriteItem Item = 5
	RestPatternOverwriteItem Item = 6
	MidiSampleItem           Item = 7
)

// implies the generation of midi messages
func (i Item) Event() []Event {
	return nil
}

type PositionedItem struct {
	Offset  int
	Item    Item
	evts    Events
	pos32th uint
}

type PositionedItems []PositionedItem

func (it *PositionedItem) Stream(ticksPerQN uint, nextPatternStopItem *PositionedItem, endPos uint) *Stream {

	/*
		todo:
				inside patterns calculate the repetitions based on the nextPatternStopItem, if it is not nil
				just as a hint for not producing too many events/bar.

				the real stoping / ignoring comes from the collectEvents mechanics, but since we can have endless patterns
				and they must be able to calculate ahead their stream, they should know when to stop.
				therefor we also need the endPos

		wir müssen die temporären overrides in den pattern handhaben. daher muss schon beim parsen das letzte hintergrund pattern
		referenziert bleiben und die dann entdeckten overrides werden dann "an das pattern" struct "geschickt", so dass es sich drum kümmert (nachtürlich mit positionen)

		auf diese weise kann man auch gut eine fehlermeldung werfen, wenn ein override gesetzt ist, ohne dass es ein hintergrundpattern gibt.
	*/

	var s = &Stream{}
	s.Item = it.Item
	s.Anchor = it.pos32th * ticksPerQN / 8
	s.Offset = it.Offset

	var lastpos uint

	var firstNote = true

	for _, ev := range it.evts {
		abs := ev.pos32th * ticksPerQN / 8
		ev.absPosition = uint(int(abs) + int(s.Anchor) + int(s.Offset))
		lastpos = ev.absPosition
		var ch, key, vel uint8
		if it.Item != NoteOnItem && it.Item != MidiSampleItem && ev.msg.GetNoteStart(&ch, &key, &vel) {
			ev.trackNoteOn = true

			if !firstNote {
				stopev := makeStopEvent(ev.absPosition)
				s.Events = append(s.Events, stopev)
			}
			firstNote = false
			//	ev.stopNote = true
		}

		s.Events = append(s.Events, ev)
	}

	s.Length = lastpos - s.FirstPos()
	return s
}
