package stops_handling

import (
	"encoding/hex"
	"slices"
	"sort"

	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/smf"
)

// not tested, just to get an idea, how to do it
func MakeTrack(ticksperQN uint, endPos uint, its PositionedItems) (track smf.Track) {

	evts := GetTrackEvents(ticksperQN, endPos, its)

	var noteOns = [128]bool{}
	var prev uint

	for _, evt := range evts {

		delta := uint32(evt.absPosition - prev)

		var ch, key, vel uint8

		switch {
		case evt.msg.GetNoteStart(&ch, &key, &vel) && evt.trackNoteOn:
			noteOns[key] = true
		case evt.msg.GetNoteEnd(&ch, &key):
			noteOns[key] = false
		case isStopEvent(evt.msg):
			first := true
			didChange := false
			for k, v := range noteOns {
				if v {
					didChange = true
					d := delta
					if !first {
						d = 0
					}
					track.Add(d, midi.NoteOff(ch, uint8(k)))
				}
			}

			if didChange {
				track.Add(0, midi.Pitchbend(ch, 0))
			}

			prev = evt.absPosition
			continue
		}
		track.Add(delta, evt.msg)
		prev = evt.absPosition
	}

	return
}

// GetTrackEvents receives all PositionedItems in a track and returns the Events of a track
func GetTrackEvents(ticksperQN uint, endPos uint, its PositionedItems) Events {
	var s Streams
	for i, it := range its {

		var str *Stream
		if it.Item == PatternItem && i < len(its)-1 {
			str = it.Stream(ticksperQN, nextPatternStopItem(its[i:]), endPos)
		} else {
			str = it.Stream(ticksperQN, nil, endPos)
		}

		s = append(s, str)
	}

	return filterEvents(s)
}

func nextPatternStopItem(pits PositionedItems) *PositionedItem {
	for _, it := range pits {
		switch it.Item {
		case PatternItem:
			return &it
		case NoteItem:
			return &it
		case RestItem:
			return &it
		case MidiSampleItem:
			return &it
		}
	}

	return nil
}

/*

wir müssen die streams rückwärts durchlaufen (also letzter anchor zuerst), um zu sehen, welche streams andere abschneiden.

also:

1. alle streams durchlaufen und deren absolute Anchor-Position setzen.
2. alle streams rückwärts durchlaufen und dabei nur die events bis zum startpunkt des vorigen (also weil rückwärts nächsten streams) berücksichtigen

*/

type lastOverwritePos struct {
	pattern               uint
	midisampler           uint
	note                  uint
	noteon                uint
	break_                uint
	notepatternoverride   uint
	break_patternoverride uint
}

func (l lastOverwritePos) LeastPosOf(poses ...uint) (stopPosition uint) {
	var positions = []uint{}

	for _, pos := range poses {
		if pos > 0 {
			positions = append(positions, pos)
		}
	}

	slices.Sort(positions)

	if len(positions) > 0 {
		return positions[0]
	}

	return 0
}

/*
the pattern handling should before already mix the pattern with override patterns and override notes
and return them as part of just the pattern, but where you have overridenoteitems for the notes within
the override notes or the override notes themselves, this way they can still stop each other (if we want it this way)
*/

func isStopEvent(msg smf.Message) bool {
	var bt []byte
	if !msg.GetMetaSeqData(&bt) {
		return false
	}

	return hex.EncodeToString(bt) == hex.EncodeToString([]byte{0x7f, 0x7f})
}

func makeStopEvent(lastStop uint) Event {
	var stopEvent Event
	stopEvent.absPosition = lastStop
	stopEvent.msg = smf.MetaSequencerData([]byte{0x7f, 0x7f})
	return stopEvent
}

func (l lastOverwritePos) mystopPosition(it Item) uint {
	switch it {
	case RestItem:
		return l.LeastPosOf(l.pattern, l.midisampler, l.note, l.break_)
	case NoteItem:
		return l.LeastPosOf(l.pattern, l.midisampler, l.note, l.break_)
	case NoteOnItem:
		// NoteOff we don't take here, because it is stopped directly for the note and not for all running notes
		return l.LeastPosOf(l.pattern, l.midisampler)
	case PatternItem:
		return l.LeastPosOf(l.pattern, l.midisampler, l.note, l.break_, l.noteon)
	case NotePatternOverwriteItem: // this is difficult and should be solved in another way (as part of the pattern)
		return l.LeastPosOf(l.break_patternoverride, l.notepatternoverride)
	case RestPatternOverwriteItem: // this is difficult and should be solved in another way (as part of the pattern)
		return l.LeastPosOf(l.break_patternoverride, l.notepatternoverride)
	case MidiSampleItem:
		return l.LeastPosOf(l.pattern, l.midisampler, l.note, l.break_, l.noteon)
	}
	return 0
}

// filterEvents filters out and makes stop events, based on our rules
func filterEvents(streams Streams) (evts Events) {
	// sort by first position
	sort.Sort(streams)

	var lastOver lastOverwritePos

	for i := len(streams) - 1; i >= 0; i-- {
		str := streams[i]
		myevents := str.Events

		lastStop := lastOver.mystopPosition(str.Item)

		if lastStop == 0 {
			evts = append(evts, myevents...)
		} else {
			stopEvent := makeStopEvent(lastStop)
			// TODO make a special type of it
			// on every stopevent all runninig notes should be stopped and pitchbend should be send back ti neutral
			evts = append(evts, stopEvent)
			for _, ev := range myevents {
				if ev.absPosition < lastStop {
					evts = append(evts, ev)
				}
			}
		}

		pos := str.FirstPos()

		switch str.Item {
		case RestItem:
			lastOver.break_ = pos
		case NoteItem:
			lastOver.note = pos
		case NoteOnItem:
			lastOver.noteon = pos
		case NoteOffItem:
		//	lastOver.no = pos
		case PatternItem:
			lastOver.pattern = pos
		case NotePatternOverwriteItem:
			lastOver.notepatternoverride = pos
		case MidiSampleItem:
			lastOver.midisampler = pos
		}
	}

	sort.Sort(evts)

	return
}
