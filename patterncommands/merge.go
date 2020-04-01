package patterncommands

import (
	"sort"

	"gitlab.com/gomidi/muskel/items"
)

// $merge(=patt1,=patt2,...)
// merges the given patterns
func Merge(params []string, helper Helper) (res []*items.Event, err error) {
	var sorted items.EventSequence
	var evts []*items.Event

	for _, param := range params {

		evts, err = patternEvents(param, helper)

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
