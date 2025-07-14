package lilypond

import "gitlab.com/gomidi/muskel/smf"

type bar struct {
	position            smf.TicksAbsPos
	no                  uint
	num, denom          uint8
	timeSignatureChange bool
	ticksPerQN          uint
	marker              string
}

func (b bar) dup() bar {
	return b
}

func (b bar) Length32() uint {
	return (32 * uint(b.num)) / uint(b.denom)
}

func (b bar) Length() uint {
	return (b.ticksPerQN * 4 * uint(b.num)) / uint(b.denom)
}

func (b bar) End() smf.TicksAbsPos {
	return b.position + smf.TicksAbsPos(b.Length())
}
