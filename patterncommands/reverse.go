package patterncommands

import (
	"fmt"

	"gitlab.com/gomidi/muskel/items"
)

// $revers(=patt) reverses the events in the given pattern
func Reverse(params []string, helper Helper) (res []*items.Event, err error) {
	if len(params) != 1 {
		return nil, fmt.Errorf("reverse just needs one parameter: a pattern")
	}

	if params[0][0] != '=' {
		return nil, fmt.Errorf("reverse just needs one parameter: a pattern")
	}

	var evts []*items.Event

	evts, err = patternEvents(params[0], helper)

	if err != nil {
		return nil, err
	}

	noevts := len(evts)
	res = make([]*items.Event, noevts)

	for i, ev := range evts {
		nev := ev.Dup()
		nev.Position = evts[noevts-1-i].Position
		nev.PosShift = evts[noevts-1-i].PosShift
		res[noevts-1-i] = nev
	}

	return res, nil
}
