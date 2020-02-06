package patterncommands

import (
	"gitlab.com/gomidi/muskel/items"
)

/*
further ideas for patterncommands

generate(everydistance, item, times)
stretch(patt,factor) // time streching
chance(patt,percentage) // set the chance that something happens for randomness in the pattern. e.g. for ?30%a' a chance of 20% means nope and 31% means yes, or for ?(a,b,c) a chance of 20% means a, 45% means b and 80% means c
rabbit(x0,factor,percentage,number) generates a chord of number items by the formula x(n+1) = x(0)*factor * (1-p) where p is the percentage as decimal e.g. $rabbit(c#',2.5,80%,10)
merge(patt1,patt2) // on same time patt2 overrides patt1
append(patt1, patt2)
take(patt1,n) // take every nth item of pattern
transpose(patt, num)
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
	"reverse":  reverse,
	"euclid":   euclid,
	"merge":    merge,
	"rabbit":   rabbit,
	"mirror":   mirror,
	"negative": negative,
}

type Helper interface {
	GetCallEvents(endPos uint, callDef string, params ...string) ([]*items.Event, error)
	GetPipeEvents() ([]*items.Event, error)
	GetPatternDefEvents(def string) ([]*items.Event, error)
}
