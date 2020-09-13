package patterncommands

import (
	"fmt"
	"strconv"

	"gitlab.com/gomidi/muskel/items"
)

// $from(=patt1,from)
// returns the tail of a pattern, starting at from
func From(params []string, helper Helper) (res []*items.Event, err error) {
	var from int

	if len(params) != 2 {
		return nil, fmt.Errorf("from needs a pattern and a from position")
	}

	evts, err := patternEvents(params[0], helper)

	if err != nil {
		return nil, fmt.Errorf("first parameter of from must be a pattern")
	}

	from, err = strconv.Atoi(params[1])

	if err != nil {
		return nil, fmt.Errorf("second parameter of from must be an int (from)")
	}

	from -= 1

	sl := [2]int{from, -1}

	res = items.SliceEvents(sl, evts)
	if from > 0 {
		res = items.MoveBySyncFirst(res)
	}
	return
}
