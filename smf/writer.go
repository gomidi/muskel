package smf

import (
	"fmt"
	"io"
	"sort"

	"gitlab.com/gomidi/midi/midimessage/channel"
	"gitlab.com/gomidi/midi/midimessage/meta"
	smfpkg "gitlab.com/gomidi/midi/smf"
	"gitlab.com/gomidi/midi/smf/smfwriter"
)

type writer struct {
	ticks        uint32
	delta        uint32
	numTracks    uint16
	currentTrack uint16
	channel      uint8
	smf          smfpkg.Writer

	// bool value indicates, if it is a monitored ("stoppable"), otherwise, it is still a noteOn, but just an unstoppable
	noteOns map[uint8]bool
}

func newWriter(wr io.Writer, numTracks uint16, ticks uint32, options ...smfwriter.Option) *writer {
	options = append(
		[]smfwriter.Option{
			smfwriter.TimeFormat(smfpkg.MetricTicks(ticks)),
			smfwriter.NumTracks(numTracks),
		}, options...)

	return &writer{
		smf:       smfwriter.New(wr, options...),
		noteOns:   map[uint8]bool{},
		numTracks: numTracks,
		ticks:     ticks,
	}
}

func (wr *writer) stopNotes(all bool) (didStop bool, err error) {
	//fmt.Printf("stopNotes called (len: %v all: %v)\n", len(wr.noteOns), all)
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
		//	fmt.Printf("stopping note: %v\n", nt)
		wr.smf.SetDelta(wr.delta)
		err = wr.smf.Write(channel.Channel(wr.channel).NoteOff(uint8(nt)))
		if err != nil {
			err = fmt.Errorf("could not stop note %v: %s", nt, err)
			return
		}
		wr.delta = 0
	}
	return
}

func (wr *writer) writeTrack(ch uint8, endPos int, evts events) (err error) {

	//fmt.Printf("writeTrack called for channel %v endPos: %v\n", ch, endPos)

	wr.noteOns = map[uint8]bool{}
	wr.delta = 0
	wr.channel = ch
	var lastPosition uint

	if wr.currentTrack > wr.numTracks {
		return fmt.Errorf("can't write track: all %v tracks written", wr.numTracks)
	}

	for _, ev := range evts {

		if ev == nil || (!ev.stopNotes && ev.message == nil) {
			continue
		}

		if cm, isChannelMsg := ev.message.(channel.Message); isChannelMsg {
			if cm.Channel() != wr.channel {
				return fmt.Errorf("track channel is %v, but got message for channel %v: %v at tick %v", wr.channel, cm.Channel(), ev.message, ev.position)
			}
		}

		wr.delta = uint32(ev.position - lastPosition)

		//		fmt.Printf("ev: %T, %#v\n", ev.message, ev)

		if ev.stopNotes {
			//fmt.Printf("callig stopNotes I\n")
			var didStop bool
			didStop, err = wr.stopNotes(false)
			if err != nil {
				//	fmt.Printf("error at tick %v: %s", ev.position, err)
				return fmt.Errorf("error at tick %v: %s", ev.position, err)
			}

			// is a rest
			if ev.message == nil {
				//fmt.Printf("stopped from a rest: %v at %v\n", didStop, ev.position)
				if didStop {
					//fmt.Printf("setting lastPosition to rest: %v\n", ev.position)
					lastPosition = ev.position
				}
				continue
			}
		}

		lastPosition = ev.position
		//fmt.Printf("setting lastPosition to: %v for %v\n", ev.position, ev.message)
		wr.smf.SetDelta(wr.delta)

		switch v := ev.message.(type) {
		case channel.NoteOn:
			if _, has := wr.noteOns[v.Key()]; has {
				//fmt.Printf("remove note for repetiting: %v\n", v.Key())
				err = wr.smf.Write(channel.Channel(wr.channel).NoteOff(v.Key()))
				if err != nil {
					//fmt.Printf("could not stop repeating note %v: %s at tick %v\n", v.Key(), err, ev.position)
					return fmt.Errorf("could not stop repeating note %v: %s at tick %v", v.Key(), err, ev.position)
				}
				wr.smf.SetDelta(0)
			}
			//fmt.Printf("setting noteOn tracking for note %v\n", v.Key())
			wr.noteOns[v.Key()] = ev.monitor
			err = wr.smf.Write(ev.message)
			if err != nil {
				return fmt.Errorf("could not write message %v at tick %v: %v", ev.message, ev.position, err)
			}
		case channel.NoteOff:
			if _, has := wr.noteOns[v.Key()]; has {
				//fmt.Printf("remove note via noteoff %v\n", v.Key())
				delete(wr.noteOns, v.Key())
				err = wr.smf.Write(ev.message)
			}
			if err != nil {
				return fmt.Errorf("could not write message %v at tick %v: %v", ev.message, ev.position, err)
			}
		default:
			err = wr.smf.Write(ev.message)
			if err != nil {
				return fmt.Errorf("could not write message %v at tick %v: %v", ev.message, ev.position, err)
			}
		}

		if err != nil {
			return fmt.Errorf("could not write message %v at tick %v: %v", ev.message, ev.position, err)
		}
	}

	if uint(endPos) < lastPosition {
		fmt.Printf("endPos: %v lastPosition: %v\n", endPos, lastPosition)
		panic("must not happend endPos less than lastPosition")
	}

	if endPos > 0 {
		wr.delta = uint32(uint(endPos) - lastPosition)
	} else {
		wr.delta = 0
	}

	//fmt.Printf("callig stopNotes II last delta: %v (endPos: %v lastPosition: %v\n", wr.delta, endPos, lastPosition)
	_, err = wr.stopNotes(true)
	if err != nil {
		return err
	}
	wr.currentTrack++
	wr.smf.SetDelta(wr.delta)
	return wr.smf.Write(meta.EndOfTrack)
}
