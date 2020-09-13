package patterncommands

import (
	"fmt"
	"strconv"

	"gitlab.com/gomidi/muskel/items"
)

// $until(=patt1,to)
// returns the head of a pattern until to position
func Until(params []string, helper Helper) (res []*items.Event, err error) {
	var to int

	if len(params) != 2 {
		return nil, fmt.Errorf("until needs a pattern and a to position")
	}

	evts, err := patternEvents(params[0], helper)

	if err != nil {
		return nil, fmt.Errorf("first parameter of until must be a pattern")
	}

	to, err = strconv.Atoi(params[1])

	if err != nil {
		return nil, fmt.Errorf("second parameter of until must be an int (to)")
	}

	to += 1

	sl := [2]int{0, to}
	res = items.SliceEvents(sl, evts)

	return
}
