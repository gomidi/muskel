package patterncommands

import (
	"fmt"

	"gitlab.com/gomidi/muskel/items"
)

func _mirrorNote(ntVal uint8, itVal uint8) uint8 {
	if itVal == ntVal {
		return ntVal
	}

	if itVal > 2*ntVal {
		return 0
	}

	res := 2*ntVal - itVal
	if res > 127 {
		return 127
	}

	return res
}

func _setNote(nt *items.Note, key uint8) {
	nt.Letter, nt.Augmenter, nt.Octave = items.KeyToNote(key)
}

func _mirrorEvent(ntVal uint8, add uint8, in *items.Event) (out *items.Event) {
	switch v := in.Item.(type) {
	case *items.Note:
		out = in.Dup()
		it := v.Dup().(*items.Note)
		_setNote(it, _mirrorNote(ntVal, it.ToMIDI())+add)
		out.Item = it
	case *items.MultiItem:
		out = in.Dup()
		it := v.Dup().(*items.MultiItem)
		for i, ev := range v.Events {
			it.Events[i] = _mirrorEvent(ntVal, add, ev)
		}
		out.Item = it
	case *items.NTuple:
		out = in.Dup()
		it := v.Dup().(*items.NTuple)
		for i, ev := range v.Events {
			it.Events[i] = _mirrorEvent(ntVal, add, ev)
		}
		out.Item = it
	default:
		out = in.Dup()
	}
	return
}

// $mirror(=patt,zeroNote)
// mirrors the given pattern at zeroNote axis
func Mirror(params []string, helper Helper) (res []*items.Event, err error) {

	if len(params) != 2 {
		return nil, fmt.Errorf("mirror needs two parameters: a pattern and the mirror note")
	}

	if params[0][0] != '=' {
		return nil, fmt.Errorf("mirror first parameter needs to be a pattern")
	}

	var evts []*items.Event

	evts, err = patternEvents(params[0], helper)

	if err != nil {
		return nil, err
	}

	var nt items.Note
	err = nt.Parse(params[1], 0)
	if err != nil {
		return nil, fmt.Errorf("mirror second parameter needs to be a note")
	}

	ntVal := nt.ToMIDI()

	noevts := len(evts)
	res = make([]*items.Event, noevts)

	for i, ev := range evts {
		res[i] = _mirrorEvent(ntVal, 0, ev)
	}

	return res, nil
}
