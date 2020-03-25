package patterncommands

import (
	"gitlab.com/gomidi/muskel/items"
)

func patternEvents(param string, helper Helper) (evts []*items.Event, err error) {
	if param == "=" {
		evts, err = helper.GetPipeEvents()
	} else {
		evts, err = helper.GetCallEvents(0, param)
	}

	if err != nil {
		evts = nil
	}
	return
}
