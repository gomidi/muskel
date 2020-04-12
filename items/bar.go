package items

import (
	"fmt"
)

type Bar struct {
	TempoChange   float64
	TimeSigChange [2]uint8
	Position      uint // number of 32th from start of Piece

	No int
	// real time sig (inherited or changed)
	TimeSig           [2]uint8
	JumpTo            string
	Scale             *Scale
	Tilde             string
	Include           *Include
	Comment           string
	Part              string
	InnerTempoChanges map[uint]float64
	InnerScales       map[uint]*Scale
}

func (b *Bar) Dup() (nuB *Bar) {
	return &Bar{
		TempoChange:       b.TempoChange,
		Position:          b.Position,
		TimeSigChange:     b.TimeSigChange,
		No:                b.No,
		TimeSig:           b.TimeSig,
		JumpTo:            b.JumpTo,
		Include:           b.Include,
		Comment:           b.Comment,
		Part:              b.Part,
		Scale:             b.Scale,
		Tilde:             b.Tilde,
		InnerTempoChanges: b.InnerTempoChanges,
		InnerScales:       b.InnerScales,
	}
}

func (b *Bar) String() string {
	s := fmt.Sprintf("#%s [%v] %v/%v (<-%v/%v) @%v", b.Part, b.Position, b.TimeSig[0], b.TimeSig[1], b.TimeSigChange[0], b.TimeSigChange[1], b.TempoChange)

	if b.Scale != nil {
		var nt Note
		nt.Letter, nt.Augmenter, nt.Octave = KeyToNote(b.Scale.BaseNote)
		s += b.Scale.String()
	}

	s += fmt.Sprintf(" // #%v", b.No)
	return s
}

func NewBar() *Bar {
	return &Bar{
		TimeSig:           [2]uint8{4, 4},
		InnerTempoChanges: map[uint]float64{},
		InnerScales:       map[uint]*Scale{},
	}
}

func (b Bar) Mount(pos uint) *Bar {
	b.Position += pos
	return &b
}

func (b *Bar) Length32th() uint {
	if b.Include != nil {
		return b.Include.Length32ths
	}
	return Length32ths(b.TimeSig[0], b.TimeSig[1])
}
