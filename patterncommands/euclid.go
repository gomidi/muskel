package patterncommands

import "gitlab.com/gomidi/muskel/items"

func euclid(params []string, helper Helper) (res []*items.Event, err error) {
	var eu items.EuclideanRhythm
	err = eu.Parse(0, params...)
	if err != nil {
		return nil, err
	}

	return helper.GetPatternDefEvents(eu.PatternDef)
}
