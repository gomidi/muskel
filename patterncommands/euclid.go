package patterncommands

import "gitlab.com/gomidi/muskel/items"

// $euclid(numTotal,numExposed,duration)
// generates a pattern of numTotal placeholders, where numExposed placeholders are #2 and the others #1.
// The distance between each placeholder is duration (in fraction of 4 beats, e.g. 0.25 would be a beat
func Euclid(params []string, helper Helper) (res []*items.Event, err error) {
	var eu items.EuclideanRhythm
	err = eu.Parse(0, params...)
	if err != nil {
		return nil, err
	}

	return helper.GetSketchDefEvents(eu.Sketch)
}
