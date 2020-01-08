package items

import (
	"fmt"
	//	"gitlab.com/gomidi/midi/mid"
)

type GlideStart bool

func (g GlideStart) Dup() Item {
	return &g
}

func (g GlideStart) String() string {
	if g {
		return "~~"
	}
	return "~"
}

func (g *GlideStart) Parse(data string, posIn32th uint) error {
	if data == "" {
		*g = GlideLinear
		return nil
	}

	if data == "~" {
		*g = GlideExponential
		return nil
	}
	return fmt.Errorf("invalid glissando: \"~%s\"", data)
}

func (v GlideStart) WriteMIDI(wr SMFWriter) (addedNotes []uint8) {
	MIDITrack.noteGlide.startNote = MIDITrack.PrevKey
	delete(MIDITrack.NoteOns, MIDITrack.noteGlide.startNote)
	MIDITrack.StopNotes(wr)
	wr.Pitchbend(0)
	MIDITrack.noteGlide.startPosition = wr.Position()
	MIDITrack.noteGlide.active = true
	wr.BackupTimeline()
	MIDITrack.noteGlide.glideFunc = linearGlide
	if v == GlideExponential {
		MIDITrack.noteGlide.glideFunc = exponentialGlide
	}

	addedNotes = append(addedNotes, MIDITrack.noteGlide.startNote)
	return addedNotes
}

var GlideLinear = GlideStart(false)
var GlideExponential = GlideStart(true)
