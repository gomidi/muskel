package patterncommands

import (
	"fmt"

	"gitlab.com/gomidi/muskel/items"
)

func reverse(pos uint, params []string, helper Helper) (res []*items.Event, err error) {
	if len(params) != 1 {
		return nil, fmt.Errorf("reverse just needs one parameter: a pattern")
	}

	if params[0][0] != '=' {
		return nil, fmt.Errorf("reverse just needs one parameter: a pattern")
	}

	var evts []*items.Event

	if params[0] == "=" {
		evts, err = helper.GetPipeEvents()
	} else {
		evts, err = helper.GetCallEvents(pos, params[0])
	}

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
