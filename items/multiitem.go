package items

import (
	"strings"
)

type ItemGroup []*Item

/*
- slicing mit []
- dynamik change mit + und -
- diatonische transposition mit ^
- chromatische transposition mit # und b
- octavierungen mit " und '
- zufallswahrscheinlichkeit mit xx%
- zufällige auswahlmit %()
- microtimingverschiebung mit > und <
*/

type ItemGrouper interface {
	Get() ItemGroup
}

type ItemGroupMiddleWare []func(in ItemGroup) ItemGroup

// first function must retrieve the ItemGroup (receives nil)
func (mw ItemGroupMiddleWare) Get() (es ItemGroup) {
	if len(mw) == 0 {
		return nil
	}

	for _, fn := range mw {
		es = fn(es)
	}

	return
}

func (s ItemGroup) ToParams() Params {
	panic("TODO implement")
}

func (s ItemGroup) Call(params Params) ItemGroup {
	panic("TODO implement")
}

func (s ItemGroup) Slice(from, to int) ItemGroup {
	panic("TODO implement")
}

func (s ItemGroup) DiatonicTranspose(steps int8) ItemGroup {
	panic("TODO implement")
}

func (s ItemGroup) Transpose(halftones int8) ItemGroup {
	panic("TODO implement")
}

func (s ItemGroup) TransposeOctave(octave int8) ItemGroup {
	panic("TODO implement")
}

func (s ItemGroup) MircoTimingShift(shift int8) ItemGroup {
	panic("TODO implement")
}

func (s ItemGroup) ChangeDynamic(diff int8) ItemGroup {
	panic("TODO implement")
}

type MultiItem struct {
	Events   []*Event
	Parser   *Parser
	Exploded bool // exploded means: =patt((c d b)...) will become  =patt(c,d,b)
}

func (m *MultiItem) Parse(data string, posIn32th uint) (err error) {
	d := strings.Split(data, " ")
	for _, dd := range d {
		var ev Event

		err := ev.Parse(strings.TrimSpace(dd))

		//it, err := parseItem(m.Parser, strings.TrimSpace(dd), posIn32th)

		if err != nil {
			return err
		}

		m.Events = append(m.Events, &ev)
		//m.Items = append(m.Items, it)
	}

	return nil
}

var _ Item = &MultiItem{}

func (m MultiItem) Dup() Item {
	return &m
}

func (m MultiItem) String() string {
	var bd strings.Builder
	bd.WriteString("(")
	for i, it := range m.Events {
		if it != nil {
			if i > 0 {
				bd.WriteString(" ")
			}
			bd.WriteString(it.String())
		}
	}

	bd.WriteString(")")

	return bd.String()
}

func (m MultiItem) IsHold() bool {
	var hasStoppingNote bool
	var hasHold bool

	for _, val := range m.Events {
		switch v := val.Item.(type) {
		case *Note:
			if !v.NoteOn && !v.NoteOff {
				hasStoppingNote = true
			}
		case *MIDINote:
			if !v.NoteOn && !v.NoteOff {
				hasStoppingNote = true
			}
		case hold:
			hasHold = true
		}
	}

	return hasHold && !hasStoppingNote
}

type sortMultiItem []*Event

func (s sortMultiItem) Len() int {
	return len(s)
}

func (s sortMultiItem) Swap(a, b int) {
	s[a], s[b] = s[b], s[a]
}

/*
func getPosShift(i Item) int {
	switch v := i.(type) {
	case *Note:
		return v.PosShift
	case *MIDINote:
		return v.PosShift
	case *NTuple:
		return v.PosShift
	case *Call:
		return v.PosShift
	case *Lyric:
		return v.PosShift
	default:
		return 0
	}
}
*/

func (s sortMultiItem) Less(a, b int) bool {
	if s[a].Position == s[b].Position {
		return s[a].PosShift < s[b].PosShift
	}
	return s[a].Position < s[b].Position
}
