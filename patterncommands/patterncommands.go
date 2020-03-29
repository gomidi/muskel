package patterncommands

import (
	"gitlab.com/gomidi/muskel/items"
)

// not a patterncommand, but a more highlevel one to process one command after another
// if a parameter that the patterncommand expects to be a call, is just =, it is required to use the
// GetPipeEvents() method of the helper
// previous command and each
// pipe(patterncmd1,patterncmd2,...)

// Command is a command that receives different arguments (there might be pattern names within)
// and returns events
// we provide different functions to the PatternCMD
type Command func(params []string, helper Helper) ([]*items.Event, error)

var Commands = map[string]Command{
	"reverse":  Reverse,
	"euclid":   Euclid,
	"merge":    Merge,
	"rabbit":   Rabbit,
	"mirror":   Mirror,
	"negative": Negative,
}

type Helper interface {
	GetCallEvents(endPos uint, callDef string, params ...string) ([]*items.Event, error)
	GetPipeEvents() ([]*items.Event, error)
	GetSketchDefEvents(sketchDef [][2]string) ([]*items.Event, error)
}
