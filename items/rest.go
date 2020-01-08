package items

type rest struct{}

var Rest = rest{}

func (h rest) Parse(data string, posIn32th uint) (err error) {
	return nil
}

func (r rest) String() string {
	return "*"
}

func (r rest) Dup() Item {
	return Rest
}

var _ Item = Rest

func (v rest) WriteMIDI(wr SMFWriter) (addedNotes []uint8) {
	MIDITrack.noteGlide.active = false
	MIDITrack.StopNotes(wr)
	MIDITrack.SetStraight()
	return nil
}
