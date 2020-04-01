package items

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

type ItemGrouperFunc func() ItemGroup

func (e ItemGrouperFunc) Get() ItemGroup {
	return e()
}

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
