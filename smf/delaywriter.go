package smf

import (
	"math"

	"gitlab.com/gomidi/midi"
	"gitlab.com/gomidi/midi/midimessage/channel"
	"gitlab.com/gomidi/midi/smf"
	"gitlab.com/gomidi/muskel/items"
)

type delayWriter struct {
	track                 string
	firstDeltaSet         bool
	delayTicks            int32
	deltaCached           int32
	noteDelay             int32
	noteDelayCompensation int32
	tempoBPM              float64
	resolution            smf.MetricTicks
	lastTick              uint64

	smf.Writer
}

func (i *delayWriter) SetNoteDelayCompensation(n int32) {
	i.noteDelayCompensation = n
}

func (i *delayWriter) ResolutionTicks() smf.MetricTicks {
	return i.resolution
}

func (i *delayWriter) setDelay(num, denom int) {
	// fmt.Printf("#0 setting delay to %v/%v\n", num, denom)
	if num == 0 {
		i.delayTicks = 0
		// fmt.Printf("#1 setting delay to %v\n", i.delayTicks)
		return
	}
	i.delayTicks = int32(math.Round(float64(i.resolution.Ticks4th()) * 4 * float64(num) / float64(denom)))
	// fmt.Printf("#2 setting delay to %v\n", i.delayTicks)
}

func (i *delayWriter) SetLaidBack() {
	i.setlaidBack()
}

func (i *delayWriter) setlaidBack() {
	// fmt.Println("setlaidBack")
	i.noteDelay = items.CalcNoteDelay(i.resolution)
	i.deltaCached += i.noteDelay
}

func (i *delayWriter) SetAhead() {
	i.setAhead()
}

func (i *delayWriter) setAhead() {
	// fmt.Println("setAhead")
	i.noteDelay = items.CalcNoteDelay(i.resolution) * (-1)
	i.deltaCached += i.noteDelay
}

func (i *delayWriter) SetStraight() {
	i.setStraight()
}

func (i *delayWriter) setStraight() {
	// fmt.Println("setStraight")
	i.noteDelay = 0
}

func (i *delayWriter) Write(msg midi.Message) error {
	// fmt.Printf("writing message: %s\n", msg.String())
	if i.track == "" {
		return i.Writer.Write(msg)
	}

	delta := i.deltaCached

	switch msg.(type) {
	case channel.NoteOn, channel.NoteOff, channel.NoteOffVelocity:
		// don't return but continue in this function
		// fmt.Printf("deltaCached is: %v noteDelayCompensation is: %v\n", i.deltaCached, i.noteDelayCompensation)
	default:
		if delta < 0 {
			delta = 0
		}
		i.setRealDelta(uint32(delta))
		i.deltaCached = 0
		return i.Writer.Write(msg)
	}

	if !i.firstDeltaSet {
		delta += i.delayTicks
		i.firstDeltaSet = true
	}

	if delta > 0 && i.noteDelayCompensation != 0 {
		delta += i.noteDelayCompensation
		i.noteDelayCompensation = 0
	}

	if delta < 0 {
		delta = 0
	}

	// fmt.Printf("setting real delta to %v before writing %s\n", delta, msg)
	i.setRealDelta(uint32(delta))
	i.deltaCached = 0
	if i.noteDelay != 0 {
		// remove noteDelayCompensation when we have a negative noteDelay at the start of the piece
		if delta == 0 {
			// fmt.Printf("reset noteDelayCompensation caused by delta == 0\n")
			i.noteDelayCompensation = 0
		} else {
			i.noteDelayCompensation -= i.noteDelay
			// fmt.Printf("setting noteDelayCompensation to %v\n", i.noteDelayCompensation)
		}
		i.noteDelay = 0
	}
	return i.Writer.Write(msg)
}

func (i *delayWriter) setRealDelta(ticks uint32) {
	i.lastTick += uint64(ticks)
	// fmt.Printf("setting real delta to %v\n", ticks)
	i.Writer.SetDelta(ticks)
}

func (i *delayWriter) SetDelta(ticks uint32) {
	if i.track == "" {
		i.setRealDelta(ticks)
		return
	}
	t := int32(ticks)

	i.deltaCached = t

	/*
		fmt.Printf("SetDelta(%v) noteDelay: %v, noteDelayComp: %v\n", ticks, i.noteDelay, i.noteDelayCompensation)
		//panic(ticks)

		if t > 0 && i.noteDelayCompensation != 0 {
			t += i.noteDelayCompensation
			i.noteDelayCompensation = 0
		}

		if i.noteDelay != 0 {
			t += i.noteDelay
			i.noteDelayCompensation -= i.noteDelay
			i.noteDelay = 0
		}

		if t < 0 {
			t = 0
		}

		i.Writer.SetDelta(uint32(t))
	*/
}
