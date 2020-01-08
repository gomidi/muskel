package sketch

import (
	"fmt"

	"gitlab.com/gomidi/muskel/items"
)

type Bar struct {
	TempoChange   float64
	TimeSigChange [2]uint8
	Position      uint // number of 32th from start of Piece

	No int
	// real time sig (inherited or changed)
	TimeSig [2]uint8
	JumpTo  string
	//Include string
	Scale *items.Scale
	//IsEmpty bool
	//IsNaked bool
	Tilde   string
	Include *items.Include
	Comment string
	Part    string
}

func (b *Bar) Dup() (nuB *Bar) {
	return &Bar{
		TempoChange:   b.TempoChange,
		Position:      b.Position,
		TimeSigChange: b.TimeSigChange,
		No:            b.No,
		TimeSig:       b.TimeSig,
		JumpTo:        b.JumpTo,
		Include:       b.Include,
		Comment:       b.Comment,
		Part:          b.Part,
		Scale:         b.Scale,
		// IsEmpty: b.IsEmpty,
		// IsNaked: b.IsNaked,
		Tilde: b.Tilde,
	}
}

func (b *Bar) String() string {
	s := fmt.Sprintf("#%s [%v] %v/%v (<-%v/%v) @%v", b.Part, b.Position, b.TimeSig[0], b.TimeSig[1], b.TimeSigChange[0], b.TimeSigChange[1], b.TempoChange)

	if b.Scale != nil {
		var nt items.Note
		nt.Letter, nt.Augmenter, nt.Octave = items.KeyToNote(b.Scale.BaseNote)
		s += b.Scale.String()
	}

	s += fmt.Sprintf(" // #%v", b.No)
	return s
}

func NewBar() *Bar {
	return &Bar{
		TimeSig: [2]uint8{4, 4},
	}
}

func (b Bar) Mount(pos uint) *Bar {
	b.Position += pos
	return &b
}

func (b *Bar) Length32th() int {
	return items.Length32ths(b.TimeSig[0], b.TimeSig[1])
}

var eventDebug bool

func (e *Event) String() string {
	if e.Item == nil {
		if eventDebug {
			return items.Pos32thToString(e.Position) + "[nil]"
		}
		return ""
	}

	if eventDebug {
		return items.Pos32thToString(e.Position) + e.Item.String()
	}
	return e.Item.String()
}

func (e *Event) Dup() *Event {
	return &Event{
		Item:     e.Item.Dup(),
		Position: e.Position,
	}
}
