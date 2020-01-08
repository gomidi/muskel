package items

import (
	"sort"

	"gitlab.com/gomidi/midi/smf"
)

type WriterWrapper interface {
	SetStraight()
	SetLaidBack()
	SetAhead()
	SetNoteDelayCompensation(n int32)
	ResolutionTicks() smf.MetricTicks
}

type miditrack struct {
	aftertouchGlide struct {
		active        bool
		startPosition uint64
		startValue    uint8
		glideFunc     func(wr SMFWriter, distance, diff int64, fn func(vl float64))
	}

	polyaftertouchGlide struct {
		active        bool
		key           uint8
		value         uint8
		startPosition uint64
		startValue    uint8
		glideFunc     func(wr SMFWriter, distance, diff int64, fn func(vl float64))
	}

	pitchbendGlide struct {
		active        bool
		startPosition uint64
		startValue    int16
		glideFunc     func(wr SMFWriter, distance, diff int64, fn func(vl float64))
	}

	ccGlide struct {
		active        bool
		controller    uint8
		startPosition uint64
		startValue    uint8
		glideFunc     func(wr SMFWriter, distance, diff int64, fn func(vl float64))
	}

	noteGlide struct {
		active        bool
		startNote     uint8
		startPosition uint64
		glideFunc     func(wr SMFWriter, distance, diff int64, fn func(vl float64))
	}

	//startGlissandoNote uint8
	PrevKey uint8
	// startGlissando uint64
	//glissandoFunc func(wr *mid.SMFWriter, distance, diff int64, fn func(vl float64))

	Track struct {
		MIDITranspose  int8
		PitchbendRange uint8
	}

	NoteOns       map[uint8]bool
	PrevVel       int8
	PrevPitchbend int16

	CurrentDistanceToStartOfBarIn32th uint

	LastNum32th uint

	CurrentNTuple *NTuple

	VelocityScale [5]uint8 // min max randomness steps, center

	WriterWrapper
}

func (mw *miditrack) Reset() {
	mw.aftertouchGlide.active = false
	mw.polyaftertouchGlide.active = false
	mw.pitchbendGlide.active = false
	mw.ccGlide.active = false
	mw.noteGlide.active = false
	mw.PrevKey = 0
	mw.Track.MIDITranspose = 0
	mw.Track.PitchbendRange = 2
	mw.NoteOns = map[uint8]bool{}
	mw.PrevVel = 63
	mw.PrevPitchbend = 0
	mw.CurrentDistanceToStartOfBarIn32th = 0
	mw.LastNum32th = 0
	mw.CurrentNTuple = nil
}

const glideResolution = 64 // 64th

type sortedNoteOns []uint8

func (s sortedNoteOns) Swap(a, b int) {
	s[a], s[b] = s[b], s[a]
}

func (s sortedNoteOns) Less(a, b int) bool {
	return s[a] < s[b]
}

func (s sortedNoteOns) Len() int {
	return len(s)
}

func (mw *miditrack) GlideResolution(wr SMFWriter) uint32 {
	return wr.Ticks256th() * (256 / glideResolution)
}

func (iw *miditrack) WritecurrentNTuple(wr SMFWriter, endPos uint) (addedNotes []uint8) {
	if iw.CurrentNTuple == nil {
		return nil
	}

	addedNotes = iw.CurrentNTuple.writeMIDI(wr, endPos)
	iw.CurrentNTuple = nil
	return addedNotes
}

// has to be changed in conjunction with GlissandoResolution
func (mw *miditrack) GlideForward(wr SMFWriter) {
	wr.Forward(0, 1, glideResolution)
}

// TODO check if filter is active and note is not written
// and plan note offs etc.
func (iw *miditrack) writeNoteOn(wr SMFWriter, key, velocity uint8) {
	/*
		if iw.track.Filter != nil {
			var end uint64
			if iw.track.Filter.SetVelocity() {
				// TODO here we would need the currentPos which we only get after having written (via iw.LastTick())
				velocity, end = iw.track.Filter.Filter(currentPos, key, velocity)

				// don't write
				if velocity == 0 {
					return
				}
			}
		}
	*/
	wr.NoteOn(key, velocity)
	//var nt filter.Note
	//nt.Velocity = velocity
	//nt.Key = key
	//nt.AbsTick = iw.LastTick()
	//iw.filterNotes = append(iw.filterNotes, nt)
}

// TODO check if noteoffs for missing noteons are critical - they should be handled by the midid/mid.Writer library
func (iw *miditrack) writeNoteOff(wr SMFWriter, key uint8) {
	wr.NoteOff(key)
	//var nt filter.Note
	//nt.Velocity = 0
	//nt.Key = key
	//nt.AbsTick = iw.LastTick()
	//iw.filterNotes = append(iw.filterNotes, nt)
}

func (mw *miditrack) StopNotes(wr SMFWriter) {
	var nons sortedNoteOns

	for nt, isOn := range mw.NoteOns {
		if isOn {
			nons = append(nons, nt)
			//fmt.Printf("NoteOff %v\n", nt)
			delete(mw.NoteOns, nt)
		}
	}

	sort.Sort(nons)

	for _, nt := range nons {
		mw.writeNoteOff(wr, nt)
		//iw.wr.NoteOff(nt)
	}
}

func (mw *miditrack) WriteItem(wr SMFWriter, item Item) (addedNotes []uint8) {
	if item == nil {
		return nil
	}

	return item.WriteMIDI(wr)
}

var MIDITrack = &miditrack{}
