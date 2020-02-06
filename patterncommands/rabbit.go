package patterncommands

import "gitlab.com/gomidi/muskel/items"
import "fmt"

func rabbit(params []string, helper Helper) (res []*items.Event, err error) {

	if len(params) != 5 {
		return nil, fmt.Errorf("rabbit needs 5 parameters: a pattern, a starting note, a factor, a percentage and the number of notes")
	}

	if params[0][0] != '=' {
		return nil, fmt.Errorf("rabbit must have a pattern as first parameter")
	}

	if params[0] == "=" {
		//fmt.Printf("reverse getting pipe events\n")
		//evts, err = helper.GetPipeEvents()
		return nil, fmt.Errorf("pipe events not supported within rabbit")
	} else {
		var r items.Rabbit
		err = r.Parse(0, params[1:]...)
		if err != nil {
			return nil, err
		}

		cparams := r.Strings()

		return helper.GetCallEvents(0, params[0], cparams...)
	}
}
