package patterncommands

import (
	"sort"

	"gitlab.com/gomidi/muskel/items"
)

func merge(params []string, helper Helper) (res []*items.Event, err error) {
	//merge(patt1,patt2)
	var sorted items.SortEvents

	for _, param := range params {
		var evts []*items.Event

		if param == "=" {
			evts, err = helper.GetPipeEvents()
		} else {
			evts, err = helper.GetCallEvents(0, param)
		}

		if err != nil {
			return nil, err
		}

		for _, ev := range evts {
			nev := ev.Dup()
			//nev.Position += pos
			sorted = append(sorted, nev)
		}

	}

	sort.Sort(sorted)

	return sorted, nil
}
