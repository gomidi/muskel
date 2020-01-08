package items

type repeatLastEvent struct{}

func (r repeatLastEvent) String() string {
	return "%"
}

func (r repeatLastEvent) Parse(data string, posIn32th uint) error {
	return nil
}

var _ Item = RepeatLastEvent

func (v repeatLastEvent) Dup() Item {
	return v
}

func (v repeatLastEvent) WriteMIDI(wr SMFWriter) (addedNotes []uint8) {
	panic("don't call me")
}

var RepeatLastEvent = repeatLastEvent{}
