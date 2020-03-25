
further ideas for patterncommands

$generate(everydistance, item, times)
$stretch(patt,factor) // time streching
$chance(patt,percentage) // set the chance that something happens for randomness in the pattern. e.g. for ?30%a' a chance of 20% means nope and 31% means yes, or for ?(a,b,c) a chance of 20% means a, 45% means b and 80% means c
$append(patt1, patt2)
$take(patt1,n) // take every nth item of pattern
$transpose(patt, num)
$replace(patt,no1,repl1,no2,repl2,...)
$replacePos(patt,pos1,repl1,pos2,repl2,...)
$slice(patt, from, to)
$repeat(patt, times, distanceBetweenRepetitions)
$moveAll(patt, distance)
$moveNotes(patt, distance)
$rotate(patt, every, times)
$addVelocity(patt,every,velDiff)
$addVelocityPos(patt,posDiff,velDiff)
$setVelocity(patt,every,velDiff)
$setVelocityPos(patt,posDiff,velDiff)
$linVelocity(patt,startVel,endVel)
$expVelocityPos(patt,startVel,endVel)
$linCCVal(patt,cc,startVal,endVal)
$expCCVal(patt,cc,startVal,endVal)
$setCCVal(patt,cc,every,valDiff)
$setCCValPos(patt,cc,posDiff,valDiff)
$add(patt,every,items...)
$addPos(patt,posDiff,items...)
$delete(patt, every)
$deletePos(patt, posDiff)
$cut(patt,fromPos,toPos)
$call(=patt,params,...) call the pattern with params, e.g. same as =patt(params,...) but could also be piped

$pick(=source,=picker) picks events from =source at the positions in =picker where there are placeholders. Any other events in  =picker will be taken as is. Should also work as pipe.

$impose(=source,=imposter) overrides events at =source at position of events within =imposter with that events. should be pipeable.

$silence(=source,=silencer) injects a break into =source at the positions of events within =silencer. Should be pipeable.

- generalbassfunktion: (siehe https://de.wikipedia.org/wiki/Generalbass) z.B.
  $gb(^2,6,4)     => (^2 ^8 ^6)
  $gb(^2)         => (^2 ^4 ^6) 
  $gb(^2,2)       => (^2 ^3 ^8 ^6) 
  bekommt eine Stufe (für den Bass) und die Generalbassziffern für die Tonleitereigenen Intervalle
  kann dann einem Pattern als Parameter übergeben werden, z.B.
  =patt($gb(^2,6,4)...)

$arp takes two Patterns: One has Multievents and one has Placeholders. The second pattern is invoked at each appearance of a multievent in the first pattern and the events of the corresponding multievent are passed as parameters for each invocation. So that the first pattern is kind of a meta-melody triggering the rhythmic figure of the second pattern on each instance. If there is more than two patterns given, the first pattern keeps being the multievent trigger and the others form a pool of rhythmic figures that is rotated through on each invocation. Should also work as pipe.

  also z.B.

  =rhythm |    |
  1       | #1 |
  1&      | #2 |
  2       | #3 |

  =chords |            |
  1       | (^1 ^3 ^5) |
  #
  1       | (^9 ^4 ^6) |
  #
  1       | (^3 ^5)    |

  =SCORE |                      |
  1      |$arp(=chords,=rhythm) |

  würde resultieren in 

  =SCORE |   |
  1      |^1 |
  1&     |^3 |
  2      |^5 |
  #
  1      |^9 |
  1&     |^4 |
  2      |^6 |
  #
  1      |^3 |
  1&     |^5 |
  2      |^3 |

- dann könnte man patterns für die oktavregel in dur und moll (jeweils auf und absteigend) anlegen (siehe https://de.wikipedia.org/wiki/Oktavregel)
