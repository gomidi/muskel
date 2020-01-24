package smf

type MidiTrack struct {
	aftertouchGlide struct {
		active        bool
		startPosition uint64
		startValue    uint8
		glideFunc     func(distance, diff int64, fn func(pos uint, vl float64))
	}

	polyaftertouchGlide struct {
		active        bool
		key           uint8
		value         uint8
		startPosition uint64
		startValue    uint8
		glideFunc     func(distance, diff int64, fn func(pos uint, vl float64))
	}

	pitchbendGlide struct {
		active        bool
		startPosition uint64
		startValue    int16
		glideFunc     func(distance, diff int64, fn func(pos uint, vl float64))
	}

	ccGlide struct {
		active        bool
		controller    uint8
		startPosition uint64
		startValue    uint8
		glideFunc     func(distance, diff int64, fn func(pos uint, vl float64))
	}

	noteGlide struct {
		active        bool
		startNote     uint8
		startPosition uint64
		glideFunc     func(distance, diff int64, fn func(pos uint, vl float64))
		noteOffBefore uint8
	}

	Track struct {
		MIDITranspose  int8
		PitchbendRange uint8
	}

	PrevKey       uint8
	PrevVel       int8
	PrevPitchbend int16
	VelocityScale [5]uint8 // min max randomness steps, center
}

func (mw *MidiTrack) Reset() {
	mw.aftertouchGlide.active = false
	mw.polyaftertouchGlide.active = false
	mw.pitchbendGlide.active = false
	mw.ccGlide.active = false
	mw.noteGlide.active = false
	mw.Track.MIDITranspose = 0
	mw.Track.PitchbendRange = 2
	mw.PrevKey = 0
	mw.PrevVel = 63
	mw.PrevPitchbend = 0
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

func (mw *MidiTrack) GlideResolution(ticks uint32) uint32 {
	/*
		ticks = 1/4
		  /2 = 1/8
		  /4 = 1/16
		  /8 = 1/32
		  /16 = 1/64
		  /32 = 1/128
		  /64 = 1/256
	*/
	return ticks * (256 / glideResolution) / 64
}
