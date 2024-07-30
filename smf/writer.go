package smf

import (
	"fmt"
	"io"
	"sort"

	"gitlab.com/gomidi/midi/v2"

	//"gitlab.com/gomidi/midi/midimessage/channel"
	//"gitlab.com/gomidi/midi/midimessage/meta"
	"gitlab.com/gomidi/midi/v2/smf"
	//"gitlab.com/gomidi/midi/smf/smfwriter"
)

type Writer struct {
	ticks        uint16
	delta        uint32
	numTracks    uint16
	currentTrack uint16
	//currTrack    smfpkg.Track
	//tracks  []smfpkg.Track
	channel uint8
	SMF     *smf.SMF

	// bool value indicates, if it is a monitored ("stoppable"), otherwise, it is still a noteOn, but just an unstoppable
	noteOns map[uint8]bool
}

/*
s := smf.New()
	s.TimeFormat = clock
*/

type logger struct {
}

func (l logger) Printf(s string, other ...any) {
	fmt.Printf(s, other...)
}

// func newWriter(wr io.Writer, numTracks uint16, ticks uint16, options ...smfwriter.Option) *writer {
func NewWriter(wr io.Writer, numTracks uint16, ticks uint16) *Writer {
	/*
		options = append(
			[]smfwriter.Option{
				smfwriter.TimeFormat(smfpkg.MetricTicks(ticks)),
				smfwriter.NumTracks(numTracks),
			}, options...)
	*/
	s := smf.NewSMF1()
	s.TimeFormat = smf.MetricTicks(ticks)
	s.Logger = logger{}

	return &Writer{
		//smf:       smfwriter.New(wr, options...),
		//	tracks:    make([]smfpkg.Track, numTracks),
		SMF:       s,
		noteOns:   map[uint8]bool{},
		numTracks: numTracks,
		ticks:     ticks,
	}
}

/*
func (wr *writer) Add(delta uint32, bt []byte) {
	if int(wr.currentTrack) >= len(wr.tracks) {
		panic("not enough tracks")
	}
	wr.tracks[wr.currentTrack].Add(delta, bt)
}
*/

func (wr *Writer) writeTrack(ch uint8, endPos int, evts events) (err error) {

	//fmt.Printf("writeTrack called for channel %v endPos: %v\n", ch, endPos)

	wr.noteOns = map[uint8]bool{}
	wr.delta = 0
	wr.channel = ch
	var lastPosition uint

	if wr.currentTrack > wr.numTracks {
		return fmt.Errorf("can't write track: all %v tracks written", wr.numTracks)
	}

	var tr smf.Track

	for _, ev := range evts {

		//fmt.Printf("[writer] pos: %v msg: %v stopnotes: %v\n", ev.position, ev.message, ev.stopNotes)

		if ev == nil || (!ev.stopNotes && ev.message == nil) {
			continue
		}

		//if cm, isChannelMsg := ev.message.(channel.Message); isChannelMsg {
		var ch uint8
		if ev.message.GetChannel(&ch) {
			if ch != wr.channel {
				return fmt.Errorf("track channel is %v, but got message for channel %v: %v at tick %v", wr.channel, ch, ev.message, ev.position)
			}
		}

		//wr.currTrack.Add()

		wr.delta = uint32(ev.position - lastPosition)
		//fmt.Printf("ev.position: %v wr.delta: %v\n message: %v\n", ev.position, wr.delta, ev.message)

		//		fmt.Printf("ev: %T, %#v\n", ev.message, ev)

		if ev.stopNotes {
			//fmt.Printf("callig stopNotes I\n")
			var didStop bool
			didStop, err = wr.stopNotes(tr, false)
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

		var writeLastPos bool

		//wr.smf.SetDelta(wr.delta)

		//switch v := ev.message.(type) {
		var channel, key, velocity uint8

		switch {
		case ev.message.GetNoteStart(&channel, &key, &velocity):
			if _, has := wr.noteOns[key]; has {
				//fmt.Printf("remove note for repetiting: %v\n", v.Key())
				//err = wr.smf.Write(channel.Channel(wr.channel).NoteOff(v.Key()))
				tr.Add(wr.delta, midi.NoteOff(wr.channel, key).Bytes())
				if err != nil {
					//fmt.Printf("could not stop repeating note %v: %s at tick %v\n", v.Key(), err, ev.position)
					return fmt.Errorf("could not stop repeating note %v: %s at tick %v", key, err, ev.position)
				}
				//wr.smf.SetDelta(0)
				wr.delta = 0
			} else {
				writeLastPos = true
			}
			//fmt.Printf("setting noteOn tracking for note %v\n", v.Key())
			wr.noteOns[key] = ev.monitor
			tr.Add(wr.delta, ev.message.Bytes())
			//err = wr.smf.Write(ev.message)
			//if err != nil {
			//	return fmt.Errorf("could not write message %v at tick %v: %v", ev.message, ev.position, err)
			//}
		case ev.message.GetNoteEnd(&channel, &key):
			//case channel.NoteOff:
			if _, has := wr.noteOns[key]; has {
				//fmt.Printf("remove note via noteoff %v\n", v.Key())
				delete(wr.noteOns, key)
				tr.Add(wr.delta, ev.message.Bytes())
				//err = wr.smf.Write(ev.message)
				writeLastPos = true
			}
			/*
				if err != nil {
					return fmt.Errorf("could not write message %v at tick %v: %v", ev.message, ev.position, err)
				}
			*/
		default:
			tr.Add(wr.delta, ev.message.Bytes())
			/*
				err = wr.smf.Write(ev.message)
				if err != nil {
					return fmt.Errorf("could not write message %v at tick %v: %v", ev.message, ev.position, err)
				}
			*/
			writeLastPos = true
		}

		/*
			if err != nil {
				return fmt.Errorf("could not write message %v at tick %v: %v", ev.message, ev.position, err)
			}
		*/

		if writeLastPos {
			lastPosition = ev.position
			//fmt.Printf("setting lastPosition to: %v for %v\n", ev.position, ev.message)
		}
	}

	if uint(endPos) < lastPosition {
		fmt.Printf("endPos: %v lastPosition: %v\n", endPos, lastPosition)
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

	//fmt.Printf("callig stopNotes II last delta: %v (endPos: %v lastPosition: %v\n", wr.delta, endPos, lastPosition)
	_, err = wr.stopNotes(tr, true)
	if err != nil {
		return err
	}

	tr.Close(wr.delta)
	/*
		wr.smf.SetDelta(wr.delta)
		err = wr.smf.Write(meta.EndOfTrack)
	*/
	err = wr.SMF.Add(tr)
	if err == smf.ErrFinished {
		return nil
	}
	wr.currentTrack++

	return err
}

func (wr *Writer) stopNotes(tr smf.Track, all bool) (didStop bool, err error) {
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
		//fmt.Printf("stopping note: %v at delta: %v\n", nt, wr.delta)
		//wr.smf.SetDelta(wr.delta)
		//err = wr.smf.Write(channel.Channel(wr.channel).NoteOff(uint8(nt)))
		tr.Add(wr.delta, midi.NoteOff(wr.channel, uint8(nt)))
		/*
			if err != nil {
				err = fmt.Errorf("could not stop note %v: %s", nt, err)
				return
			}
		*/
		wr.delta = 0
		//fmt.Printf("stopping note: setting delta to 0\n")
	}
	return
}
