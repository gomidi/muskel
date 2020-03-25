package patterncommands

import (
	"fmt"

	"gitlab.com/gomidi/muskel/items"
)

// negative harmony
// This concept is derived, from the Theory of Harmony by Swiss composer Ernst Levy (1895 – 1981).
// $negative(=patt,mirrorNote)
func Negative(params []string, helper Helper) (res []*items.Event, err error) {

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

	ntVal := nt.ToMIDI() + 3

	noevts := len(evts)
	res = make([]*items.Event, noevts)

	for i, ev := range evts {
		res[i] = _mirrorEvent(ntVal, 1, ev)
	}

	return res, nil
}
