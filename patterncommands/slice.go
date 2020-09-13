package patterncommands

import (
	"fmt"
	"strconv"

	"gitlab.com/gomidi/muskel/items"
)

// $slice(=patt1,from,to)
// slices the given pattern
func Slice(params []string, helper Helper) (res []*items.Event, err error) {
	var from, to int

	if len(params) != 3 {
		return nil, fmt.Errorf("slice needs a pattern, a from position and a to position")
	}

	evts, err := patternEvents(params[0], helper)

	if err != nil {
		return nil, fmt.Errorf("first parameter of slice must be a pattern")
	}

	from, err = strconv.Atoi(params[1])

	if err != nil {
		return nil, fmt.Errorf("second parameter of slice must be an int (from)")
	}

	from -= 1

	to, err = strconv.Atoi(params[2])

	if err != nil {
		return nil, fmt.Errorf("third parameter of slice must be an int (to)")
	}

	to += 1

	sl := [2]int{from, to}
	res = items.SliceEvents(sl, evts)

	if from > 0 {
		res = items.MoveBySyncFirst(res)
	}

	return
}
