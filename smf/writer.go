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
	targetWR     io.Writer
	logger       smf.Logger

	// bool value indicates, if it is a monitored ("stoppable"), otherwise, it is still a noteOn, but just an unstoppable
	noteOns map[uint8]bool
}

type Option func(*Writer)

func WithLogger(smf.Logger) Option {
	return func(wr *Writer) {
		wr.logger = nil
	}
}

type logger struct {
}

func (l logger) Printf(s string, other ...any) {
	fmt.Printf(s, other...)
}

func NewWriter(wr io.Writer, numTracks uint16, ticks uint16, opts ...Option) *Writer {
	w := &Writer{}
	w.logger = logger{}

	for _, opt := range opts {
		opt(w)
	}

	s := smf.NewSMF1()
	s.TimeFormat = smf.MetricTicks(ticks)
	s.Logger = w.logger

	w.SMF = s
	w.noteOns = map[uint8]bool{}
	w.numTracks = numTracks
	w.ticks = ticks
	w.targetWR = wr
	return w
}

func (wr *Writer) writeTrack(ch uint8, endPos int, evts Events) (tr smf.Track, err error) {
	wr.noteOns = map[uint8]bool{}
	wr.delta = 0
	wr.channel = ch
	var lastPosition TicksAbsPos

	if wr.currentTrack > wr.numTracks {
		err = fmt.Errorf("can't write track: all %v tracks written", wr.numTracks)
		return
	}

	for _, ev := range evts {

		if ev == nil || (!ev.StopNotes && ev.Message == nil) {
			continue
		}

		var ch uint8
		if ev.Message.GetChannel(&ch) {
			if ch != wr.channel {
				err = fmt.Errorf("track channel is %v, but got message for channel %v: %v at tick %v", wr.channel, ch, ev.Message, ev.Position)
				return
			}
		}

		wr.delta = uint32(ev.Position - lastPosition)

		if ev.StopNotes {
			var didStop bool
			didStop, err = wr.stopNotes(&tr, false)
			if err != nil {
				err = fmt.Errorf("error at tick %v: %s", ev.Position, err)
				return
			}

			// is a rest
			if ev.Message == nil {
				if didStop {
					lastPosition = ev.Position
				}
				continue
			}
		}

		var writeLastPos bool

		var channel, key, velocity uint8

		switch {

		case ev.Message.GetNoteStart(&channel, &key, &velocity):
			if _, has := wr.noteOns[key]; has {
				tr.Add(wr.delta, midi.NoteOff(wr.channel, key).Bytes())
				if err != nil {
					err = fmt.Errorf("could not stop repeating note %v: %s at tick %v", key, err, ev.Position)
					return
				}
				wr.delta = 0
			} else {
				writeLastPos = true
			}
			wr.noteOns[key] = ev.Monitor
			tr.Add(wr.delta, ev.Message.Bytes())

		case ev.Message.GetNoteEnd(&channel, &key):
			if _, has := wr.noteOns[key]; has {
				delete(wr.noteOns, key)
				tr.Add(wr.delta, ev.Message.Bytes())
				writeLastPos = true
			}

		default:
			tr.Add(wr.delta, ev.Message.Bytes())
			writeLastPos = true
		}

		if writeLastPos {
			lastPosition = ev.Position
		}
	}

	if TicksAbsPos(endPos) < lastPosition {
		if lastPosition-TicksAbsPos(endPos) > 3000000 {
			panic("must not happend endPos less than lastPosition")
		}
		endPos = int(lastPosition)
	}

	if endPos > 0 {
		wr.delta = uint32(TicksAbsPos(endPos) - lastPosition)
	} else {
		wr.delta = 0
	}

	_, err = wr.stopNotes(&tr, true)
	if err != nil {
		return
	}

	tr.Close(wr.delta)
	return
}

func (wr *Writer) stopNotes(tr *smf.Track, all bool) (didStop bool, err error) {
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
