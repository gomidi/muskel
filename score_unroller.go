package muskel

import (
	"math/rand"
	"strings"
	"time"
	//"crypto/rand"
)

type ScoreUnroller struct {
	//	src  *Score
	dest *Score

	// from unrollBarsAndJumps
	unrolledBars   []*Bar
	instrBarevents map[string][]BarEvents
	num            uint8
	denom          uint8
	newBarNo       int

	// from flattenInstrumentEvents
	currentlyRepeatingBars            []BarEvents
	indexWithinCurrentlyRepeatingBars int

	// from unfoldPatternCallWithFollowingEvent
	prevPos string
}

func (s *ScoreUnroller) createInstruments(instruments []*Instrument) {
	for _, instr := range instruments {
		s.dest.AddInstrument(instr.Dup())
	}
}

// unrollJump unrolls a jump
func (s *ScoreUnroller) unrollJump(jump string) {
	for j, bar2ndloop := range s.dest.Bars[s.dest.Parts[jump][0]:s.dest.Parts[jump][1]] {

		nub := bar2ndloop.Dup()
		if bar2ndloop.timeSigChange[0] > 0 {
			s.num, s.denom = bar2ndloop.timeSigChange[0], bar2ndloop.timeSigChange[1]
		}
		nub.barNo = s.newBarNo
		//		fmt.Printf("adding new unrolled bar: %v\n", nub.barNo)
		nub.timeSig[0] = s.num
		nub.timeSig[1] = s.denom
		// nub.include = ""
		s.unrolledBars = append(s.unrolledBars, nub)
		for _, _instr := range s.dest.Instruments {
			//			fmt.Printf("adding bar %v => %v\n    %v\n", s.dest.Parts[jump][0]+j, s.newBarNo, _instr.events[s.dest.Parts[jump][0]+j])
			if _, has := s.instrBarevents[_instr.Name]; has {
				s.instrBarevents[_instr.Name] = append(s.instrBarevents[_instr.Name], _instr.events[s.dest.Parts[jump][0]+j])
			}
		}
		s.newBarNo++
	}
}

// unrollJump unrolls a jump
func (s *ScoreUnroller) unrollInclude(sc *Score) {
	for j, bar2ndloop := range sc.Bars {

		nub := bar2ndloop.Dup()
		if bar2ndloop.timeSigChange[0] > 0 {
			s.num, s.denom = bar2ndloop.timeSigChange[0], bar2ndloop.timeSigChange[1]
		}
		nub.barNo = s.newBarNo
		nub.originalBarNo = nub.barNo
		for prt, pvls := range sc.Parts {
			if pvls[0] == bar2ndloop.originalBarNo {
				if _, has := s.dest.Parts[prt]; !has {
					// fmt.Printf("found part %q for bar %v, setting to %v,%v\n", prt, bar2ndloop.barNo, nub.barNo, nub.barNo+pvls[1]-pvls[0])
					s.dest.Parts[prt] = [2]int{nub.barNo, nub.barNo + pvls[1] - pvls[0]}
				} else {
					nub.originalBarNo = s.dest.Parts[prt][0]
				}
			}
		}
		//		fmt.Printf("adding new unrolled bar: %v\n", nub.barNo)
		nub.timeSig[0] = s.num
		nub.timeSig[1] = s.denom
		s.unrolledBars = append(s.unrolledBars, nub)
		for _, _instr := range sc.Instruments {
			//			fmt.Printf("adding bar %v => %v\n    %v\n", s.dest.Parts[jump][0]+j, s.newBarNo, _instr.events[s.dest.Parts[jump][0]+j])
			s.instrBarevents[_instr.Name] = append(s.instrBarevents[_instr.Name], _instr.events[j].Dup())
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

	for _, instr := range s.dest.Instruments {
		//		fmt.Printf("adding bar %v\n", bar.originalBarNo)
		if len(instr.events)-1 >= bar.originalBarNo {
			s.instrBarevents[instr.Name] = append(s.instrBarevents[instr.Name], instr.events[bar.originalBarNo])
		}
	}
}

// unrollBarsAndJumps unrolls the bars and jumps
func (s *ScoreUnroller) unrollBarsAndJumps() {

	if !s.dest.isUnrolled {
		panic("destination must be the unrolled score")
	}

	s.unrolledBars = []*Bar{}
	s.instrBarevents = map[string][]BarEvents{}
	s.num = 4
	s.denom = 4
	s.newBarNo = 0

	//	s.createInstruments()

	//	fmt.Printf("len bars: %v\n", len(s.src.Bars))

	//for _, bar := range s.src.Bars {
	for _, bar := range s.dest.Bars {
		//		fmt.Printf("bar.barNo: %v, bar.originalBarNo: %v\n", bar.barNo, bar.originalBarNo)
		if include := bar.include; include != "" {
			s.unrollInclude(bar.includedScore)
			continue
		}

		if jump := bar.jumpTo; jump != "" {
			//fmt.Printf("doing the jump for bar: %v\n", bar.barNo)
			s.unrollJump(jump)
			continue
		}

		s.unrollBar(bar)
	}

	s.dest.Bars = s.unrolledBars

	for _, instr := range s.dest.Instruments {
		//		fmt.Printf("[0] len bars in instr: %v\n", len(s.instrBarevents[ii]))
		instr.events = s.instrBarevents[instr.Name]
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

	for _, instr := range s.dest.Instruments {
		s.currentlyRepeatingBars = []BarEvents{}
		s.indexWithinCurrentlyRepeatingBars = 0
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
				rand.Seed(time.Now().UnixNano())
				v.chosen = rand.Intn(len(v.alternatives))
				//				fmt.Printf("chosen: %v\n", v.chosen)
				nuEv := ev.Dup()
				nuEv.Item = v.alternatives[v.chosen]
				nuEv.originalData = v.alternativesOriginalData[v.chosen]
				unrolled = append(unrolled, nuEv)
			case *RandomProbability:
				switch int(v.prob) {
				case 0:
				// zero percent chance
				case 100:
					nuEv := ev.Dup()
					nuEv.originalData = v.itemOriginalData
					nuEv.Item = v.item
					unrolled = append(unrolled, nuEv)
				default:
					rand.Seed(time.Now().UnixNano())
					if got := rand.Intn(int(v.prob)); got >= 50 {
						nuEv := ev.Dup()
						nuEv.Item = v.item
						nuEv.originalData = v.itemOriginalData
						unrolled = append(unrolled, nuEv)
					}
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
	unrolled = append(unrolled, s.convertEvents(ev.BarNo, v, evts...)...)

	diffBars := (len(s.dest.Bars) - 1) - ev.BarNo

	// for each following empty bar
	for didx := 1; didx <= diffBars; didx++ {
		/*
			if positionOfNextBar == -1 || s.dest.Bars[ev.BarNo+didx].isEmpty {
				break
			}
		*/
		if positionOfNextBar == -1 {
			break
		}
		timesig = s.dest.Bars[ev.BarNo+didx].timeSig
		evts, positionOfNextBar = pvs.Spread(positionOfNextBar, timesig[0], timesig[1])
		unrolled = append(unrolled, s.convertEvents(ev.BarNo+didx, v, evts...)...)
		didx++
	}

	//	fmt.Printf("unrolled: %v\n", unrolled)

	return
}

func (s *ScoreUnroller) moveNoteAccordingToScale(sc *Scale, p *PatternCall, v Note) (abs Note) {
	mv := p.scaleMove
	if mv > 0 {
		mv -= 1
	}
	diff := int8(sc.StepToNote(mv)) - int8(p.firstNoteAbsKey)
	abs = v.Dup()

	if v.scaleNote == 0 {
		n := int8(v.toMIDI())
		n += diff

		nn := sc.StepToNote(sc.NoteToStep(uint8(n)))
		abs.letter, abs.augmenter, abs.octave = keyToNote(nn)
	}
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
		/*
			if positionOfNextBar == -1 || s.dest.Bars[ev.BarNo+didx].isEmpty {
				break
			}
		*/
		if positionOfNextBar == -1 {
			break
		}
		timesig = s.dest.Bars[ev.BarNo+didx].timeSig
		evts, positionOfNextBar = pvs.Spread(positionOfNextBar, timesig[0], timesig[1])
		ev2 := s.convertEvents(ev.BarNo+didx, v, evts...)

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

func (s *ScoreUnroller) convertEvents(barNo int, p *PatternCall, in ...*positionedEvent) (out []*Event) {
	out = make([]*Event, len(in))
	sc := s.scaleAt(barNo)

	for i, pev := range in {
		ev := &Event{}
		ev.BarNo = barNo
		ev.DistanceToStartOfBarIn32th = pev.positionIn32ths
		switch v := pev.item.(type) {
		case Note:
			if p.scaleMoveMode == 2 && p.firstNoteIsScaleNote == -1 && v.scaleNote == 0 {
				nnt := s.moveNoteAccordingToScale(sc, p, v)
				ev.Item = nnt
				ev.originalData = nnt.String()
			} else {
				ev.Item = pev.item
				ev.originalData = pev.originalData
			}
		case MultiItem:
			var nuMI MultiItem
			//inner := strings.Trim(pev.originalData, "()")
			ss := strings.Split(pev.originalData[1:], "=")
			for idx, mi := range v {
				switch vv := mi.(type) {
				case Note:
					if p.scaleMoveMode == 2 && p.firstNoteIsScaleNote == -1 && vv.scaleNote == 0 {
						nnnt := s.moveNoteAccordingToScale(sc, p, vv)
						nuMI = append(nuMI, nnnt)
						ss[idx] = nnnt.String()
					} else {
						nuMI = append(nuMI, mi)
					}
				}
			}

			ev.originalData = "=" + strings.Join(ss, "=")
			ev.Item = nuMI
		case NTuple:
			var nuNT NTuple
			nuNT.endPos = v.endPos
			inner := strings.Trim(pev.originalData, "{")
			is := strings.Split(inner, "}")
			inner = is[0]
			ss := strings.Split(inner, ",")

			for idx, mi := range v.items {
				switch vv := mi.(type) {
				case Note:
					if p.scaleMoveMode == 2 && p.firstNoteIsScaleNote == -1 && vv.scaleNote == 0 {
						nnnt := s.moveNoteAccordingToScale(sc, p, vv)
						nuNT.items = append(nuNT.items, nnnt)
						ss[idx] = nnnt.String()
					} else {
						nuNT.items = append(nuNT.items, mi)
					}
				}
			}

			ev.originalData = "{" + strings.Join(ss, ",") + "}" + pos32thToString(v.endPos)
			ev.Item = nuNT
		default:
			ev.Item = pev.item
			ev.originalData = pev.originalData
		}
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

func (s *ScoreUnroller) unfoldRepeated() {
	for _, instr := range s.dest.Instruments {
		var lastEvent *Event
		for i, ev := range instr.unrolled {
			switch ev.Item.(type) {
			case RepeatLastEvent:
				//				fmt.Printf("got repeated: %v\n", lastItem)
				ev.Item = lastEvent.Item
				ev.originalData = lastEvent.originalData
			case hold:
				// do nothing
			default:
				if ev.Item != nil {
					lastEvent = ev
					//					fmt.Printf("set lastitem to: %v\n", lastItem)
				}
			}
			instr.unrolled[i] = ev
		}
	}
}

func (s *ScoreUnroller) copyBars(bars []*Bar) {
	for _, b := range bars {
		s.dest.Bars = append(s.dest.Bars, b.Dup())
	}
}

func (s *ScoreUnroller) scaleAt(barNo int) (scale *Scale) {
	scale = &Scale{Mode: Ionian, BaseNote: 60}

	for _, bar := range s.dest.Bars {
		if bar.barNo > barNo {
			return
		}
		if bar.scale != nil {
			scale = bar.scale
		}
	}

	return
}

func (s *ScoreUnroller) convertScaleNoteToAbsNote(barNo int, nt Note) (abs Note) {
	if nt.scaleNote == 0 {
		return nt
	}
	abs = nt.Dup()
	sc := s.scaleAt(barNo)
	sn := nt.scaleNote
	if sn > 0 {
		sn -= 1
	}
	abs.letter, abs.augmenter, abs.octave = keyToNote(sc.StepToNote(sn))
	abs.scaleNote = 0
	return abs
}

func (s *ScoreUnroller) replaceScaleNotes() {
	if !s.dest.isUnrolled {
		panic("must be unrolled")
	}

	for _, instr := range s.dest.Instruments {
		var unrolled []*Event
		for _, ev := range instr.unrolled {
			switch v := ev.Item.(type) {
			case Note:
				switch {
				case v.scaleNote == 0:
					unrolled = append(unrolled, ev)
				default:
					nuEv := ev.Dup()
					nt := s.convertScaleNoteToAbsNote(ev.BarNo, v)
					nuEv.Item = nt
					nuEv.originalData = nt.String()
					unrolled = append(unrolled, nuEv)
				}
			case MultiItem:
				var nuMI MultiItem
				nuEv := ev.Dup()
				itms := strings.Split(ev.originalData[1:], "=")

				for it_idx, it := range v {
					switch vv := it.(type) {
					case Note:
						switch {
						case vv.scaleNote == 0:
							nuMI = append(nuMI, vv)
						default:
							nnt := s.convertScaleNoteToAbsNote(ev.BarNo, vv)
							nuMI = append(nuMI, nnt)
							itms[it_idx] = nnt.String()
						}
					default:
						nuMI = append(nuMI, vv)
					}
				}

				nuEv.Item = nuMI
				nuEv.originalData = "=" + strings.Join(itms, "=")
				unrolled = append(unrolled, nuEv)
			case NTuple:
				var nuNt NTuple
				nuEv := ev.Dup()
				nuNt.endPos = v.endPos
				inner := strings.Trim(ev.originalData, "{")
				is := strings.Split(inner, "}")
				inner = is[0]
				orig := strings.Split(inner, ",")

				for it_idx, it := range v.items {
					switch vv := it.(type) {
					case Note:
						switch {
						case vv.scaleNote == 0:
							nuNt.items = append(nuNt.items, vv)
						default:
							nnt := s.convertScaleNoteToAbsNote(ev.BarNo, vv)
							nuNt.items = append(nuNt.items, nnt)
							orig[it_idx] = nnt.String()
						}
					default:
						nuNt.items = append(nuNt.items, vv)
					}
				}

				nuEv.Item = nuNt
				nuEv.originalData = "{" + strings.Join(orig, ",") + "}" + pos32thToString(v.endPos)
				unrolled = append(unrolled, nuEv)
			default:
				unrolled = append(unrolled, ev)
			}
		}

		instr.unrolled = unrolled
	}
}

func (s *ScoreUnroller) unrollInstruments(bars []*Bar, instr []*Instrument) error {

	s.copyBars(bars)

	// attach the barnumbers and position in 32th relative to start to the event
	//s.src.trackBarNumbers()
	s.createInstruments(instr)
	s.trackBarNumbers()

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

	// replace scale notes with absolute Notes
	s.replaceScaleNotes()

	// unfold the repeated items
	// we don't want to repeat patterns or randomness via repeated items, because it is confusing
	// and does not bring much to the table
	// it is more interesting to be sure that repeating always means doubling the real item
	// before
	s.unfoldRepeated()

	return nil
}

func newScoreUnroller(src *Score) *ScoreUnroller {
	nu := NewScore()
	nu.isUnrolled = true
	nu.HeaderIncludes = src.HeaderIncludes
	nu.FileName = src.FileName
	nu.Meta = src.Meta
	nu.BodyComments = src.BodyComments
	nu.Parts = src.Parts
	nu.Temperament = src.Temperament
	nu.PatternDefinitions = src.PatternDefinitions
	nu.IncludedPatternDefinitions = src.IncludedPatternDefinitions
	nu.HeaderComments = src.HeaderComments
	nu.barNumbersTracked = true

	return &ScoreUnroller{
		//		src:  src,
		dest: nu,
	}
}
