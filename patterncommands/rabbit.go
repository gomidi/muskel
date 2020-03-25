package patterncommands

import "gitlab.com/gomidi/muskel/items"
import "fmt"

// $rabbit(=patt,startNote,factor,percentage,num) by the formula x(n+1) = x(0)*factor * (1-p) where p is the percentage as decimal e.g. $rabbit(=patt,c#',2.5,80%,10)
// rabbit generates num notes starting by startNote growing by factor, limited by percentage
// simulating a rabbit population growth. The generated notes are passed to patt as parameters
// maybe it should be changed to generate a chord of number items that would than be passed as parameter to a pattern, e.g.
// =patt($rabbit(c#',2.5,80%,10)...)
func Rabbit(params []string, helper Helper) (res []*items.Event, err error) {

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
