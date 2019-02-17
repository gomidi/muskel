package muskel

import (
	"math/rand"
)

type ScoreUnroller struct {
	src  *Score
	dest *Score

	// from unrollBarsAndJumps
	unrolledBars   []*Bar
	instrBarevents map[int][]BarEvents
	num            uint8
	denom          uint8
	newBarNo       int

	// from flattenInstrumentEvents
	currentlyRepeatingBars            []BarEvents
	indexWithinCurrentlyRepeatingBars int

	// from unfoldPatternCallWithFollowingEvent
	prevPos string
}

func (s *ScoreUnroller) createInstruments() {
	for _, instr := range s.src.Instruments {
		s.dest.Instruments = append(s.dest.Instruments, instr.Dup())
	}
}

// unrollJump unrolls a jump
func (s *ScoreUnroller) unrollJump(jump string) {
	for j, bar2ndloop := range s.src.Bars[s.src.Parts[jump][0] : s.src.Parts[jump][1]+1] {

		nub := bar2ndloop.Dup()
		if bar2ndloop.timeSigChange[0] > 0 {
			s.num, s.denom = bar2ndloop.timeSigChange[0], bar2ndloop.timeSigChange[1]
		}
		nub.barNo = s.newBarNo
		nub.timeSig[0] = s.num
		nub.timeSig[1] = s.denom
		s.unrolledBars = append(s.unrolledBars, nub)
		for iii, _instr := range s.src.Instruments {
			//			fmt.Printf("adding bar %v => %v\n", s.src.Parts[jump][0]+j, s.newBarNo)
			s.instrBarevents[iii] = append(s.instrBarevents[iii], _instr.events[s.src.Parts[jump][0]+j])
		}
		s.newBarNo++
	}
}

func (s *ScoreUnroller) unrollBar(bar *Bar) {
	bar.barNo = s.newBarNo
	if bar.timeSigChange[0] > 0 {
		s.num, s.denom = bar.timeSigChange[0], bar.timeSigChange[1]
	}

	bar.timeSig[0] = s.num
	bar.timeSig[1] = s.denom

	s.unrolledBars = append(s.unrolledBars, bar)
	s.newBarNo++

	for ii, instr := range s.src.Instruments {
		//		fmt.Printf("adding bar %v\n", bar.originalBarNo)
		s.instrBarevents[ii] = append(s.instrBarevents[ii], instr.events[bar.originalBarNo])
	}
}

// unrollBarsAndJumps unrolls the bars and jumps
func (s *ScoreUnroller) unrollBarsAndJumps() {

	if s.src.isUnrolled {
		panic("source must not be the unrolled score")
	}

	if !s.dest.isUnrolled {
		panic("destination must be the unrolled score")
	}

	s.unrolledBars = []*Bar{}
	s.instrBarevents = map[int][]BarEvents{}
	s.num = 4
	s.denom = 4
	s.newBarNo = 0

	s.createInstruments()

	//	fmt.Printf("len bars: %v\n", len(s.src.Bars))

	for _, bar := range s.src.Bars {
		//		fmt.Printf("bar.barNo: %v, bar.originalBarNo: %v\n", bar.barNo, bar.originalBarNo)

		if jump := bar.jumpTo; jump != "" {
			s.unrollJump(jump)
			continue
		}

		s.unrollBar(bar)
	}

	s.dest.Bars = s.unrolledBars

	for ii, instr := range s.dest.Instruments {
		//		fmt.Printf("[0] len bars in instr: %v\n", len(s.instrBarevents[ii]))
		instr.events = s.instrBarevents[ii]
	}
}

// dupEvents duplicates the given events and sets the bar number to the given ones
// empty items are skipped
func (s *ScoreUnroller) dupEvents(src []*Event, barNo, lenBar int) (dest []*Event) {
	for _, ev := range src {
		eNu := ev.Dup()
		eNu.BarNo = barNo
		// skip empty items
		if eNu.Item != nil && int(eNu.DistanceToStartOfBarIn32th) < lenBar {
			//			fmt.Printf("[0] setting bar no to : %v\n", barNo)
			dest = append(dest, eNu)
		}
	}
	return
}

// flattenInstrumentBarEvents flattens the instrument barevents for a given bar
func (s *ScoreUnroller) flattenInstrumentBarEvents(barNo int, bar BarEvents, instr *Instrument) (events []*Event) {
	numBar, denomBar := s.dest.Bars[barNo].timeSig[0], s.dest.Bars[barNo].timeSig[1]
	lenBar := length32ths(numBar, denomBar)

	numRep, untilNext := bar.RepeatingBars()

	if numRep > 0 {
		if !untilNext {
			return s.dupEvents(instr.events[barNo-1], barNo, lenBar)
		}

		s.indexWithinCurrentlyRepeatingBars = 0
		s.currentlyRepeatingBars = instr.events[barNo-numRep : barNo]
		return s.dupEvents(instr.events[barNo-numRep], barNo, lenBar)
	}

	if bar.isEmpty() {
		switch len(s.currentlyRepeatingBars) {
		case 0:
			return
		case 1:
			return s.dupEvents(s.currentlyRepeatingBars[0], barNo, lenBar)
		default:
			//			fmt.Printf("IDX: %v, len %v\n", s.indexWithinCurrentlyRepeatingBars, len(s.currentlyRepeatingBars))
			s.indexWithinCurrentlyRepeatingBars = (s.indexWithinCurrentlyRepeatingBars + 1) % len(s.currentlyRepeatingBars)
			//			fmt.Printf("new IDX: %v\n", s.indexWithinCurrentlyRepeatingBars)
			return s.dupEvents(s.currentlyRepeatingBars[s.indexWithinCurrentlyRepeatingBars], barNo, lenBar)
		}
	}

	s.currentlyRepeatingBars = []BarEvents{}
	s.indexWithinCurrentlyRepeatingBars = 0
	for barLine, ev := range bar {
		eNu := ev.Dup()
		eNu.BarNo = barNo
		eNu.DistanceToStartOfBarIn32th = s.dest.Bars[barNo].positions[barLine]

		// skip empty items
		if ev.Item != nil && int(eNu.DistanceToStartOfBarIn32th) < lenBar {
			//			fmt.Printf("[4] setting bar no to : %v\n", barNo)
			events = append(events, eNu)
		}
	}

	return
}

func (s *ScoreUnroller) flattenInstrumentEvents() {
	if !s.dest.isUnrolled {
		panic("must be unrolled")
	}
	s.currentlyRepeatingBars = []BarEvents{}
	s.indexWithinCurrentlyRepeatingBars = 0

	for _, instr := range s.dest.Instruments {
		var events []*Event
		for barNo, bar := range instr.events {
			events = append(events, s.flattenInstrumentBarEvents(barNo, bar, instr)...)
		}

		instr.unrolled = events

		/*
			for _, unev := range instr.unrolled {
				fmt.Printf("Events: bar: %v, dist from start: %v\n", unev.BarNo, unev.DistanceToStartOfBarIn32th)
			}
		*/
	}

}

func (s *ScoreUnroller) evalRandomItems() {
	if !s.dest.isUnrolled {
		panic("must be unrolled")
	}
	for _, instr := range s.dest.Instruments {
		var unrolled []*Event
		for _, ev := range instr.unrolled {
			switch v := ev.Item.(type) {
			case *RandomChooser:
				v.chosen = rand.Intn(len(v.alternatives))
				nuEv := ev.Dup()
				nuEv.Item = v.alternatives[v.chosen]
				unrolled = append(unrolled, nuEv)
			case *RandomProbability:
				if got := rand.Intn(int(v.prob)); got >= 50 {
					nuEv := ev.Dup()
					nuEv.Item = v.item
					unrolled = append(unrolled, nuEv)
				}

			default:
				unrolled = append(unrolled, ev)
			}
			// handle patterns
			_ = ev
		}

		instr.unrolled = unrolled
	}
}

func (s *ScoreUnroller) unfoldPatternCallNoFollowingEvent(ev *Event, v *PatternCall) (unrolled []*Event) {

	var positionOfNextBar int = 0

	timesig := s.dest.Bars[ev.BarNo].timeSig

	var evts []*positionedEvent

	//	fmt.Printf("offset: %v\n", ev.DistanceToStartOfBarIn32th)
	pvs := PatternEvents(v.Events)
	//	fmt.Printf("pattern events: %v\n", v.Events)

	evts, positionOfNextBar = pvs.Spread(positionOfNextBar, timesig[0], timesig[1])
	unrolled = append(unrolled, s.convertEvents(ev.BarNo, evts...)...)

	diffBars := (len(s.dest.Bars) - 1) - ev.BarNo

	// for each following empty bar
	for didx := 1; didx <= diffBars; didx++ {
		if positionOfNextBar == -1 {
			break
		}
		timesig = s.dest.Bars[ev.BarNo+didx].timeSig
		evts, positionOfNextBar = pvs.Spread(positionOfNextBar, timesig[0], timesig[1])
		unrolled = append(unrolled, s.convertEvents(ev.BarNo+didx, evts...)...)
		didx++
	}

	//	fmt.Printf("unrolled: %v\n", unrolled)

	return
}

func (s *ScoreUnroller) unfoldPatternCallWithFollowingEvent(idx int, instr *Instrument, ev *Event, v *PatternCall) (unrolled []*Event) {
	//	panic("hiho")
	/*
			TODO
			1. calc the distance of the next event in 32ths (respecting bar changes etc)
			2. for each pattern event:
			   a. calc the distance to the previous pattern event in 32ths
			   b. if it is smaller than the next normal event, write it to unrolled
			   c. otherwise skip the pattern event loop

			ad 1:
			  we need the distance between the bars and corresponding from the start of the bar to
			  the event. then for each bar the length in 32ths has to be calculated + distance from start
			  of event - distance from start of previous

		    ad 2:
			  a for the first pattern event it is either 0 (sync start) or based on the first position. for
			    the following pattern events it is the position related to current position minus the relative position
				of the previous pattern event

		Question: what to do, if there is no next event?
		I think then the pattern can be fully unrolled
	*/

	// there is a following event
	//	fmt.Printf("have next event")
	// 1. calc the distance of the next event in 32ths (respecting bar changes etc)

	var positionOfNextBar int = 0

	timesig := s.dest.Bars[ev.BarNo].timeSig

	var evts []*positionedEvent

	//	fmt.Printf("offset: %v\n", ev.DistanceToStartOfBarIn32th)
	pvs := PatternEvents(v.Events)

	// for each following empty bar
	nextEv := instr.unrolled[idx+1]
	//fmt.Printf("current event: %T %s at bar %v pos %v\n", ev.Item, ev.Item, ev.BarNo, ev.DistanceToStartOfBarIn32th)
	//fmt.Printf("nextEv: %T %s at bar %v pos %v\n", nextEv.Item, nextEv.Item, nextEv.BarNo, nextEv.DistanceToStartOfBarIn32th)
	//	diffBars := (nextEv.BarNo - ev.BarNo)

	//	fmt.Printf("diffBars: %v\n", diffBars)
	var didx int
barLoop:
	for {
		if positionOfNextBar == -1 {
			break
		}
		timesig = s.dest.Bars[ev.BarNo+didx].timeSig
		evts, positionOfNextBar = pvs.Spread(positionOfNextBar, timesig[0], timesig[1])
		ev2 := s.convertEvents(ev.BarNo+didx, evts...)

		for _, e2 := range ev2 {
			if ev.BarNo+didx > nextEv.BarNo ||
				(ev.BarNo+didx == nextEv.BarNo && e2.DistanceToStartOfBarIn32th >= nextEv.DistanceToStartOfBarIn32th) {
				break barLoop
			}
			unrolled = append(unrolled, e2)
		}
		didx++
		//		unrolled = append(unrolled, s.convertEvents(ev.BarNo+didx, evts...)...)
	}

	/*
		evts, positionOfNextBar = pvs.Spread(positionOfNextBar, timesig[0], timesig[1])
		unrolled = append(unrolled, s.convertEvents(ev.BarNo, evts...)...)

		// for each following empty bar
		diffBars := (instr.unrolled[idx+1].BarNo - ev.BarNo) - 1

		//	fmt.Printf("diffBars: %v\n", diffBars)
		for didx := 1; didx <= diffBars; didx++ {
			if positionOfNextBar == -1 {
				break
			}
			timesig = s.dest.Bars[ev.BarNo+didx].timeSig
			evts, positionOfNextBar = pvs.Spread(positionOfNextBar, timesig[0], timesig[1])
			unrolled = append(unrolled, s.convertEvents(ev.BarNo+didx, evts...)...)
		}
	*/
	//	fmt.Printf("unrolled: %v\n", unrolled)

	return

}

func (s *ScoreUnroller) convertEvents(barNo int, in ...*positionedEvent) (out []*Event) {
	out = make([]*Event, len(in))

	for i, pev := range in {
		ev := &Event{}
		ev.BarNo = barNo
		ev.DistanceToStartOfBarIn32th = pev.positionIn32ths
		ev.Item = pev.item
		ev.originalData = pev.originalData
		out[i] = ev
	}

	return
}

func (s *ScoreUnroller) unfoldPatternCall(idx int, instr *Instrument, ev *Event, v *PatternCall) (unrolled []*Event) {

	// no following event
	if idx+1 >= len(instr.unrolled) {
		return s.unfoldPatternCallNoFollowingEvent(ev, v)
	}

	return s.unfoldPatternCallWithFollowingEvent(idx, instr, ev, v)
}

func (s *ScoreUnroller) unfoldPattern(idx int, ev *Event, instr *Instrument) (unrolled []*Event) {
	switch v := ev.Item.(type) {
	case *PatternCall:
		unrolled = append(unrolled, s.unfoldPatternCall(idx, instr, ev, v)...)
	default:
		unrolled = append(unrolled, ev)
	}
	return
}

func (s *ScoreUnroller) unfoldPatterns() {
	if !s.dest.isUnrolled {
		panic("must be unrolled")
	}
	for _, instr := range s.dest.Instruments {
		var unrolled []*Event
		for idx, ev := range instr.unrolled {
			//			fmt.Printf("unrolled event: %#v\n", ev)
			// handle patterns
			unrolled = append(unrolled, s.unfoldPattern(idx, ev, instr)...)
		}
		instr.unrolled = unrolled
	}
}

func (s *ScoreUnroller) unrollInstruments() error {

	// attach the barnumbers and position in 32th relative to start to the event
	s.src.trackBarNumbers()

	// unroll the repetition within an instrument and the total jumps
	// thereby creating new bars and attaching those new bar numbers in addition
	// to the original bar numbers to the events
	s.unrollBarsAndJumps()

	// since now we have the barnumbers properly attached to events, we can flatten
	// the events so that we have a single line of events where we could easily look
	// for the next event
	s.flattenInstrumentEvents()

	// evaluate randomness here before the pattern unfolding, to be able to randomly choose
	// patterns
	s.evalRandomItems()

	// unfold the patterns in a way that they are interrupted by the next non empty event
	s.unfoldPatterns()

	// evalute randomness a second time, so that randomness produced by patterns can be evaluated
	s.evalRandomItems()

	return nil
}

func newScoreUnroller(src *Score) *ScoreUnroller {
	nu := NewScore()
	nu.isUnrolled = true
	nu.Meta = src.Meta
	nu.BodyComments = src.BodyComments
	nu.Parts = src.Parts
	nu.Temperament = src.Temperament
	nu.PatternDefinitions = src.PatternDefinitions
	nu.HeaderComments = src.HeaderComments
	nu.barNumbersTracked = true

	return &ScoreUnroller{
		src:  src,
		dest: nu,
	}
}
