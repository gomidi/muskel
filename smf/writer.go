package smf

import (
	"fmt"
	"io"
	"sort"

	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/smf"
)

type Writer struct {
	ticks        uint16
	delta        uint32
	numTracks    uint16
	currentTrack uint16
	channel      uint8
	SMF          *smf.SMF

	// bool value indicates, if it is a monitored ("stoppable"), otherwise, it is still a noteOn, but just an unstoppable
	noteOns map[uint8]bool
}

type logger struct {
}

func (l logger) Printf(s string, other ...any) {
	fmt.Printf(s, other...)
}

func NewWriter(wr io.Writer, numTracks uint16, ticks uint16) *Writer {
	s := smf.NewSMF1()
	s.TimeFormat = smf.MetricTicks(ticks)
	s.Logger = logger{}

	return &Writer{
		SMF:       s,
		noteOns:   map[uint8]bool{},
		numTracks: numTracks,
		ticks:     ticks,
	}
}

func (wr *Writer) writeTrack(ch uint8, endPos int, evts events) (err error) {
	wr.noteOns = map[uint8]bool{}
	wr.delta = 0
	wr.channel = ch
	var lastPosition uint

	if wr.currentTrack > wr.numTracks {
		return fmt.Errorf("can't write track: all %v tracks written", wr.numTracks)
	}

	var tr smf.Track

	for _, ev := range evts {

		if ev == nil || (!ev.stopNotes && ev.message == nil) {
			continue
		}

		var ch uint8
		if ev.message.GetChannel(&ch) {
			if ch != wr.channel {
				return fmt.Errorf("track channel is %v, but got message for channel %v: %v at tick %v", wr.channel, ch, ev.message, ev.position)
			}
		}

		wr.delta = uint32(ev.position - lastPosition)

		if ev.stopNotes {
			var didStop bool
			didStop, err = wr.stopNotes(tr, false)
			if err != nil {
				return fmt.Errorf("error at tick %v: %s", ev.position, err)
			}

			// is a rest
			if ev.message == nil {
				if didStop {
					lastPosition = ev.position
				}
				continue
			}
		}

		var writeLastPos bool

		var channel, key, velocity uint8

		switch {

		case ev.message.GetNoteStart(&channel, &key, &velocity):
			if _, has := wr.noteOns[key]; has {
				tr.Add(wr.delta, midi.NoteOff(wr.channel, key).Bytes())
				if err != nil {
					return fmt.Errorf("could not stop repeating note %v: %s at tick %v", key, err, ev.position)
				}
				wr.delta = 0
			} else {
				writeLastPos = true
			}
			wr.noteOns[key] = ev.monitor
			tr.Add(wr.delta, ev.message.Bytes())

		case ev.message.GetNoteEnd(&channel, &key):
			if _, has := wr.noteOns[key]; has {
				delete(wr.noteOns, key)
				tr.Add(wr.delta, ev.message.Bytes())
				writeLastPos = true
			}

		default:
			tr.Add(wr.delta, ev.message.Bytes())
			writeLastPos = true
		}

		if writeLastPos {
			lastPosition = ev.position
		}
	}

	if uint(endPos) < lastPosition {
		if lastPosition-uint(endPos) > 3000000 {
			panic("must not happend endPos less than lastPosition")
		}
		endPos = int(lastPosition)
	}

	if endPos > 0 {
		wr.delta = uint32(uint(endPos) - lastPosition)
	} else {
		wr.delta = 0
	}

	_, err = wr.stopNotes(tr, true)
	if err != nil {
		return err
	}

	tr.Close(wr.delta)

	err = wr.SMF.Add(tr)
	if err == smf.ErrFinished {
		return nil
	}
	wr.currentTrack++

	return err
}

func (wr *Writer) stopNotes(tr smf.Track, all bool) (didStop bool, err error) {
	var stoppable []int

	for nt, on := range wr.noteOns {
		if on || all {
			didStop = true
			stoppable = append(stoppable, int(nt))
			delete(wr.noteOns, nt)
		}
	}

	sort.Ints(stoppable)

	for _, nt := range stoppable {
		tr.Add(wr.delta, midi.NoteOff(wr.channel, uint8(nt)))
		wr.delta = 0
	}
	return
}
