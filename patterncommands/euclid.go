package patterncommands

import "gitlab.com/gomidi/muskel/items"

func euclid(pos uint, params []string, helper Helper) (res []*items.Event, err error) {
	var eu items.EuclideanRhythm
	err = eu.Parse(pos, params...)
	if err != nil {
		return nil, err
	}

	return helper.GetPatternDefEvents(pos, eu.PatternDef)
}
