package patterncommands

import (
	"gitlab.com/gomidi/muskel/items"
)

/*
further ideas for patterncommands

generate(everydistance, item, times)
merge(patt1,patt2)
append(patt1, patt2)
transpose(patt, num)
stretch(patt, factor)
replace(patt,no1,repl1,no2,repl2,...)
replacePos(patt,pos1,repl1,pos2,repl2,...)
slice(patt, from, to)
repeat(patt, times, distanceBetweenRepetitions)
moveAll(patt, distance)
moveNotes(patt, distance)
mirror(patt,zeroNote)
rotate(patt, every, times)
addVelocity(patt,every,velDiff)
addVelocityPos(patt,posDiff,velDiff)
setVelocity(patt,every,velDiff)
setVelocityPos(patt,posDiff,velDiff)
linVelocity(patt,startVel,endVel)
expVelocityPos(patt,startVel,endVel)
linCCVal(patt,cc,startVal,endVal)
expCCVal(patt,cc,startVal,endVal)
setCCVal(patt,cc,every,valDiff)
setCCValPos(patt,cc,posDiff,valDiff)
add(patt,every,items...)
addPos(patt,posDiff,items...)
delete(patt, every)
deletePos(patt, posDiff)
cut(patt,fromPos,toPos)
*/

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
	"reverse": reverse,
	"euclid":  euclid,
	"merge":   merge,
}

type Helper interface {
	GetCallEvents(endPos uint, callDef string) ([]*items.Event, error)
	GetPipeEvents() ([]*items.Event, error)
	GetPatternDefEvents(def string) ([]*items.Event, error)
}
