package unroller

import (
	"math/rand"
	"strings"
	"time"

	"gitlab.com/gomidi/muskel/items"
	"gitlab.com/gomidi/muskel/score"
	"gitlab.com/gomidi/muskel/template"
	//"crypto/rand"
)

// Unroll unrolls repetitions, templates and randomness and returns a score
// without repetition, randomness and templates
func Unroll(s *score.Score) (dest *score.Score, err error) {

	for _, b := range s.Bars {
		b.Positions = make([]uint, len(b.OriginalPositions))
		var oldPos string

		for pi, pos := range b.OriginalPositions {
			oldPos, b.Positions[pi], err = items.PositionTo32th(oldPos, pos)
			if err != nil {
				return nil, err
			}
		}
	}

	ur := newScoreUnroller(s)
	ur.dest.SmallColumns = s.SmallColumns
	err = ur.unrollInstruments(s.Bars, s.Instruments)
	return ur.dest, err
}

type unroller struct {
	//	src  *Score
	dest *score.Score

	// from unrollBarsAndJumps
	unrolledBars   []*score.Bar
	instrBarevents map[string][]score.BarEvents
	num            uint8
	denom          uint8
	newBarNo       int

	// from flattenInstrumentEvents
	currentlyRepeatingBars            []score.BarEvents
	indexWithinCurrentlyRepeatingBars int

	// from unfoldPatternCallWithFollowingEvent
	prevPos string
}

// trackBarNumbers tracks the bar numbers and positions for each event
func (p *unroller) trackBarNumbers() {

	for _, instr := range p.dest.Instruments {
		var events []score.BarEvents

		for barNo, bar := range instr.Events {
			es := make(score.BarEvents, len(bar))
			events = append(events, es)

			//			if bar.

			for barLine, ev := range bar {
				nuev := ev.Dup()
				//
				/*
					set the bar number and position
				*/
				nuev.BarNo = barNo
				nuev.OriginalBarNo = barNo
				//fmt.Printf("barNo: %v, BarLine: %v p.Bar: %#v\n", barNo, barLine, p.Bars[barNo])
				nuev.DistanceToStartOfBarIn32th = p.dest.Bars[barNo].Positions[barLine]

				events[barNo][barLine] = nuev
			}
		}

		//		nu := instr.Dup()

		instr.Events = events

		//		p.dest.Instruments = append(p.dest.Instruments, nu)
		//instr.unrolled = events
	}
	//	p.dest.barNumbersTracked = true
}

func (s *unroller) createInstruments(instruments []*score.Instrument) {
	for _, instr := range instruments {
		s.dest.AddInstrument(instr.Dup())
	}
}

// unrollJump unrolls a jump
func (s *unroller) unrollJump(jump string) {
	for j, bar2ndloop := range s.dest.Bars[s.dest.Parts[jump][0]:s.dest.Parts[jump][1]] {

		nub := bar2ndloop.Dup()
		if bar2ndloop.TimeSigChange[0] > 0 {
			s.num, s.denom = bar2ndloop.TimeSigChange[0], bar2ndloop.TimeSigChange[1]
		}
		nub.BarNo = s.newBarNo
		//		fmt.Printf("adding new unrolled bar: %v\n", nub.barNo)
		nub.TimeSig[0] = s.num
		nub.TimeSig[1] = s.denom
		// nub.include = ""
		s.unrolledBars = append(s.unrolledBars, nub)
		for _, _instr := range s.dest.Instruments {
			//			fmt.Printf("adding bar %v => %v\n    %v\n", s.dest.Parts[jump][0]+j, s.newBarNo, _instr.events[s.dest.Parts[jump][0]+j])
			if _, has := s.instrBarevents[_instr.Name]; has {
				s.instrBarevents[_instr.Name] = append(s.instrBarevents[_instr.Name], _instr.Events[s.dest.Parts[jump][0]+j])
			}
		}
		s.newBarNo++
	}
}

// unrollJump unrolls a jump
func (s *unroller) unrollInclude(sc *score.Score) {
	for j, bar2ndloop := range sc.Bars {

		nub := bar2ndloop.Dup()
		if bar2ndloop.TimeSigChange[0] > 0 {
			s.num, s.denom = bar2ndloop.TimeSigChange[0], bar2ndloop.TimeSigChange[1]
		}
		nub.BarNo = s.newBarNo
		nub.OriginalBarNo = nub.BarNo
		for prt, pvls := range sc.Parts {
			if pvls[0] == bar2ndloop.OriginalBarNo {
				if _, has := s.dest.Parts[prt]; !has {
					// fmt.Printf("found part %q for bar %v, setting to %v,%v\n", prt, bar2ndloop.barNo, nub.barNo, nub.barNo+pvls[1]-pvls[0])
					s.dest.Parts[prt] = [2]int{nub.BarNo, nub.BarNo + pvls[1] - pvls[0]}
				} else {
					nub.OriginalBarNo = s.dest.Parts[prt][0]
				}
			}
		}
		//		fmt.Printf("adding new unrolled bar: %v\n", nub.barNo)
		nub.TimeSig[0] = s.num
		nub.TimeSig[1] = s.denom
		s.unrolledBars = append(s.unrolledBars, nub)
		for _, _instr := range sc.Instruments {
			//			fmt.Printf("adding bar %v => %v\n    %v\n", s.dest.Parts[jump][0]+j, s.newBarNo, _instr.events[s.dest.Parts[jump][0]+j])
			s.instrBarevents[_instr.Name] = append(s.instrBarevents[_instr.Name], _instr.Events[j].Dup())
		}
		s.newBarNo++
	}
}

func (s *unroller) unrollBar(bar *score.Bar) {
	bar.BarNo = s.newBarNo
	if bar.TimeSigChange[0] > 0 {
		s.num, s.denom = bar.TimeSigChange[0], bar.TimeSigChange[1]
	}

	bar.TimeSig[0] = s.num
	bar.TimeSig[1] = s.denom

	s.unrolledBars = append(s.unrolledBars, bar)
	s.newBarNo++

	for _, instr := range s.dest.Instruments {
		//		fmt.Printf("adding bar %v\n", bar.originalBarNo)
		if len(instr.Events)-1 >= bar.OriginalBarNo {
			s.instrBarevents[instr.Name] = append(s.instrBarevents[instr.Name], instr.Events[bar.OriginalBarNo])
		}
	}
}

// unrollBarsAndJumps unrolls the bars and jumps
func (s *unroller) unrollBarsAndJumps() {

	if !s.dest.IsUnrolled {
		panic("destination must be the unrolled score")
	}

	s.unrolledBars = []*score.Bar{}
	s.instrBarevents = map[string][]score.BarEvents{}
	s.num = 4
	s.denom = 4
	s.newBarNo = 0

	//	s.createInstruments()

	//	fmt.Printf("len bars: %v\n", len(s.src.Bars))

	//for _, bar := range s.src.Bars {
	for _, bar := range s.dest.Bars {
		//		fmt.Printf("bar.barNo: %v, bar.originalBarNo: %v\n", bar.barNo, bar.originalBarNo)
		if include := bar.Include; include != "" {
			s.unrollInclude(bar.IncludedScore)
			continue
		}

		if jump := bar.JumpTo; jump != "" {
			//fmt.Printf("doing the jump for bar: %v\n", bar.barNo)
			s.unrollJump(jump)
			continue
		}

		s.unrollBar(bar)
	}

	s.dest.Bars = s.unrolledBars

	for _, instr := range s.dest.Instruments {
		//		fmt.Printf("[0] len bars in instr: %v\n", len(s.instrBarevents[ii]))
		instr.Events = s.instrBarevents[instr.Name]
	}

}

// dupEvents duplicates the given events and sets the bar number to the given ones
// empty items are skipped
func (s *unroller) dupEvents(src []*score.Event, barNo, lenBar int) (dest []*score.Event) {
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
func (s *unroller) flattenInstrumentBarEvents(barNo int, bar score.BarEvents, instr *score.Instrument) (events []*score.Event) {
	numBar, denomBar := s.dest.Bars[barNo].TimeSig[0], s.dest.Bars[barNo].TimeSig[1]
	lenBar := items.Length32ths(numBar, denomBar)

	numRep, untilNext := bar.RepeatingBars()

	if numRep > 0 {
		if !untilNext {
			return s.dupEvents(instr.Events[barNo-1], barNo, lenBar)
		}

		s.indexWithinCurrentlyRepeatingBars = 0
		s.currentlyRepeatingBars = instr.Events[barNo-numRep : barNo]
		return s.dupEvents(instr.Events[barNo-numRep], barNo, lenBar)
	}

	if bar.IsEmpty() {
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

	s.currentlyRepeatingBars = []score.BarEvents{}
	s.indexWithinCurrentlyRepeatingBars = 0
	for barLine, ev := range bar {
		eNu := ev.Dup()
		eNu.BarNo = barNo
		eNu.DistanceToStartOfBarIn32th = s.dest.Bars[barNo].Positions[barLine]

		// skip empty items
		if ev.Item != nil && int(eNu.DistanceToStartOfBarIn32th) < lenBar {
			//			fmt.Printf("[4] setting bar no to : %v\n", barNo)
			events = append(events, eNu)
		}
	}

	return
}

func (s *unroller) flattenInstrumentEvents() {
	if !s.dest.IsUnrolled {
		panic("must be unrolled")
	}

	for _, instr := range s.dest.Instruments {
		s.currentlyRepeatingBars = []score.BarEvents{}
		s.indexWithinCurrentlyRepeatingBars = 0
		var events []*score.Event
		for barNo, bar := range instr.Events {
			events = append(events, s.flattenInstrumentBarEvents(barNo, bar, instr)...)
		}

		instr.Unrolled = events

		/*
			for _, unev := range instr.unrolled {
				fmt.Printf("Events: bar: %v, dist from start: %v\n", unev.BarNo, unev.DistanceToStartOfBarIn32th)
			}
		*/
	}

}

func (s *unroller) evalRandomItems() {
	if !s.dest.IsUnrolled {
		panic("must be unrolled")
	}
	for _, instr := range s.dest.Instruments {
		var unrolled []*score.Event
		for _, ev := range instr.Unrolled {
			switch v := ev.Item.(type) {
			case *items.RandomChooser:
				rand.Seed(time.Now().UnixNano())
				v.Chosen = rand.Intn(len(v.Alternatives))
				//				fmt.Printf("chosen: %v\n", v.chosen)
				nuEv := ev.Dup()
				nuEv.Item = v.Alternatives[v.Chosen]
				nuEv.OriginalData = v.AlternativesOriginalData[v.Chosen]
				unrolled = append(unrolled, nuEv)
			case *items.RandomProbability:
				switch int(v.Prob) {
				case 0:
				// zero percent chance
				case 100:
					nuEv := ev.Dup()
					nuEv.OriginalData = v.ItemOriginalData
					nuEv.Item = v.Item
					unrolled = append(unrolled, nuEv)
				default:
					rand.Seed(time.Now().UnixNano())
					if got := rand.Intn(int(v.Prob)); got >= 50 {
						nuEv := ev.Dup()
						nuEv.Item = v.Item
						nuEv.OriginalData = v.ItemOriginalData
						unrolled = append(unrolled, nuEv)
					}
				}

			default:
				unrolled = append(unrolled, ev)
			}
		}

		instr.Unrolled = unrolled
	}
}

func (s *unroller) unfoldTemplateCallNoFollowingEvent(ev *score.Event, v *template.Call) (unrolled []*score.Event) {

	var positionOfNextBar int = 0

	timesig := s.dest.Bars[ev.BarNo].TimeSig

	var evts []*template.PositionedEvent

	//	fmt.Printf("offset: %v\n", ev.DistanceToStartOfBarIn32th)
	pvs := template.TemplateEvents(v.Events)
	//	fmt.Printf("template events: %v\n", v.Events)

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
		timesig = s.dest.Bars[ev.BarNo+didx].TimeSig
		evts, positionOfNextBar = pvs.Spread(positionOfNextBar, timesig[0], timesig[1])
		unrolled = append(unrolled, s.convertEvents(ev.BarNo+didx, v, evts...)...)
		didx++
	}

	//	fmt.Printf("unrolled: %v\n", unrolled)

	return
}

func (s *unroller) moveNoteAccordingToScale(sc *items.Scale, p *template.Call, v items.Note) (abs items.Note) {
	mv := p.ScaleMove
	if mv > 0 {
		mv -= 1
	}
	diff := int8(sc.StepToNote(mv)) - int8(p.FirstNoteAbsKey)
	abs = v.Dup()

	if v.ScaleNote == 0 {
		n := int8(v.ToMIDI())
		n += diff

		nn := sc.StepToNote(sc.NoteToStep(uint8(n)))
		abs.Letter, abs.Augmenter, abs.Octave = items.KeyToNote(nn)
	}
	return
}

func (s *unroller) unfoldTemplateCallWithFollowingEvent(idx int, instr *score.Instrument, ev *score.Event, v *template.Call) (unrolled []*score.Event) {
	/*
			TODO
			1. calc the distance of the next event in 32ths (respecting bar changes etc)
			2. for each template event:
			   a. calc the distance to the previous template event in 32ths
			   b. if it is smaller than the next normal event, write it to unrolled
			   c. otherwise skip the template event loop

			ad 1:
			  we need the distance between the bars and corresponding from the start of the bar to
			  the event. then for each bar the length in 32ths has to be calculated + distance from start
			  of event - distance from start of previous

		    ad 2:
			  a for the first template event it is either 0 (sync start) or based on the first position. for
			    the following template events it is the position related to current position minus the relative position
				of the previous template event

		Question: what to do, if there is no next event?
		I think then the template can be fully unrolled
	*/

	// there is a following event
	//	fmt.Printf("have next event")
	// 1. calc the distance of the next event in 32ths (respecting bar changes etc)

	var positionOfNextBar int = 0

	timesig := s.dest.Bars[ev.BarNo].TimeSig

	var evts []*template.PositionedEvent

	//	fmt.Printf("offset: %v\n", ev.DistanceToStartOfBarIn32th)
	pvs := template.TemplateEvents(v.Events)

	// for each following empty bar
	nextEv := instr.Unrolled[idx+1]
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
		timesig = s.dest.Bars[ev.BarNo+didx].TimeSig
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

func (s *unroller) convertEvents(barNo int, p *template.Call, in ...*template.PositionedEvent) (out []*score.Event) {
	out = make([]*score.Event, len(in))
	sc := s.scaleAt(barNo)

	for i, pev := range in {
		ev := &score.Event{}
		ev.BarNo = barNo
		ev.DistanceToStartOfBarIn32th = pev.PositionIn32ths
		switch v := pev.Item.(type) {
		case items.Note:
			if p.ScaleMoveMode == 2 && p.FirstNoteIsScaleNote == -1 && v.ScaleNote == 0 {
				nnt := s.moveNoteAccordingToScale(sc, p, v)
				ev.Item = nnt
				ev.OriginalData = nnt.String()
			} else {
				ev.Item = pev.Item
				ev.OriginalData = pev.OriginalData
			}
		case items.MultiItem:
			var nuMI items.MultiItem
			//inner := strings.Trim(pev.originalData, "()")
			ss := strings.Split(pev.OriginalData[1:], "_")
			ss = ss[:len(ss)-1]
			for idx, mi := range v {
				switch vv := mi.(type) {
				case items.Note:
					if p.ScaleMoveMode == 2 && p.FirstNoteIsScaleNote == -1 && vv.ScaleNote == 0 {
						nnnt := s.moveNoteAccordingToScale(sc, p, vv)
						nuMI = append(nuMI, nnnt)
						ss[idx] = nnnt.String()
					} else {
						nuMI = append(nuMI, mi)
					}
				}
			}

			ev.OriginalData = "_" + strings.Join(ss, "_") + "_"
			ev.Item = nuMI
		case items.NTuple:
			var nuNT items.NTuple
			nuNT.EndPos = v.EndPos
			nuNT.PosShift = v.PosShift
			inner := strings.Trim(pev.OriginalData, "{")
			is := strings.Split(inner, "}")
			inner = is[0]
			ss := strings.Split(inner, ",")

			for idx, mi := range v.Items {
				switch vv := mi.(type) {
				case items.Note:
					if p.ScaleMoveMode == 2 && p.FirstNoteIsScaleNote == -1 && vv.ScaleNote == 0 {
						nnnt := s.moveNoteAccordingToScale(sc, p, vv)
						nuNT.Items = append(nuNT.Items, nnnt)
						ss[idx] = nnnt.String()
					} else {
						nuNT.Items = append(nuNT.Items, mi)
					}
				}
			}

			posShift := ""

			if v.PosShift == 1 {
				posShift = ">"
			}

			if v.PosShift == -1 {
				posShift = "<"
			}

			ev.OriginalData = "{" + strings.Join(ss, ",") + "}" + items.Pos32thToString(v.EndPos) + posShift
			ev.Item = nuNT
		default:
			ev.Item = pev.Item
			ev.OriginalData = pev.OriginalData
		}
		out[i] = ev
	}

	return
}

func (s *unroller) unfoldTemplateCall(idx int, instr *score.Instrument, ev *score.Event, v *template.Call) (unrolled []*score.Event) {

	// no following event
	if idx+1 >= len(instr.Unrolled) {
		return s.unfoldTemplateCallNoFollowingEvent(ev, v)
	}

	return s.unfoldTemplateCallWithFollowingEvent(idx, instr, ev, v)
}

func (s *unroller) unfoldTemplate(idx int, ev *score.Event, instr *score.Instrument) (unrolled []*score.Event) {
	switch v := ev.Item.(type) {
	case *template.Call:
		unrolled = append(unrolled, s.unfoldTemplateCall(idx, instr, ev, v)...)
	default:
		unrolled = append(unrolled, ev)
	}
	return
}

func (s *unroller) unfoldTemplates() {
	if !s.dest.IsUnrolled {
		panic("must be unrolled")
	}
	for _, instr := range s.dest.Instruments {
		var unrolled []*score.Event
		for idx, ev := range instr.Unrolled {
			//			fmt.Printf("unrolled event: %#v\n", ev)
			unrolled = append(unrolled, s.unfoldTemplate(idx, ev, instr)...)
		}
		instr.Unrolled = unrolled
	}
}

func (s *unroller) unfoldRepeated() {
	for _, instr := range s.dest.Instruments {
		var lastEvent *score.Event
		for i, ev := range instr.Unrolled {
			switch ev.Item.(type) {
			case items.RepeatLastEvent:
				//				fmt.Printf("got repeated: %v\n", lastItem)
				ev.Item = lastEvent.Item
				ev.OriginalData = lastEvent.OriginalData
			default:
				if ev.Item == items.Hold {
					// do nothing
					continue
				}
				if ev.Item != nil {
					lastEvent = ev
					//					fmt.Printf("set lastitem to: %v\n", lastItem)
				}
			}
			instr.Unrolled[i] = ev
		}
	}
}

func (s *unroller) copyBars(bars []*score.Bar) {
	for _, b := range bars {
		s.dest.Bars = append(s.dest.Bars, b.Dup())
	}
}

func (s *unroller) scaleAt(barNo int) (scale *items.Scale) {
	scale = &items.Scale{Mode: items.Ionian, BaseNote: 60}

	for _, bar := range s.dest.Bars {
		if bar.BarNo > barNo {
			return
		}
		if bar.Scale != nil {
			scale = bar.Scale
		}
	}

	return
}

func (s *unroller) convertScaleNoteToAbsNote(barNo int, nt items.Note) (abs items.Note) {
	if nt.ScaleNote == 0 {
		return nt
	}
	abs = nt.Dup()
	sc := s.scaleAt(barNo)
	sn := nt.ScaleNote
	if sn > 0 {
		sn -= 1
	}
	key := sc.StepToNote(sn)

	for _, aug := range nt.Augmenter {
		if aug == '#' {
			key += 1
		}
	}
	abs.Letter, abs.Augmenter, abs.Octave = items.KeyToNote(key)
	abs.ScaleNote = 0
	return abs
}

func (s *unroller) replaceScaleNotes() {
	if !s.dest.IsUnrolled {
		panic("must be unrolled")
	}

	for _, instr := range s.dest.Instruments {
		var unrolled []*score.Event
		for _, ev := range instr.Unrolled {
			switch v := ev.Item.(type) {
			case items.Note:
				switch {
				case v.ScaleNote == 0:
					unrolled = append(unrolled, ev)
				default:
					nuEv := ev.Dup()
					nt := s.convertScaleNoteToAbsNote(ev.BarNo, v)
					nuEv.Item = nt
					nuEv.OriginalData = nt.String()
					unrolled = append(unrolled, nuEv)
				}
			case items.MultiItem:
				var nuMI items.MultiItem
				nuEv := ev.Dup()
				itms := strings.Split(ev.OriginalData[1:], "_")
				itms = itms[:len(itms)-1]

				for it_idx, it := range v {
					switch vv := it.(type) {
					case items.Note:
						switch {
						case vv.ScaleNote == 0:
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
				nuEv.OriginalData = "_" + strings.Join(itms, "_") + "_"
				unrolled = append(unrolled, nuEv)
			case items.NTuple:
				var nuNt items.NTuple
				nuEv := ev.Dup()
				nuNt.EndPos = v.EndPos
				nuNt.PosShift = v.PosShift
				inner := strings.Trim(ev.OriginalData, "{")
				is := strings.Split(inner, "}")
				inner = is[0]
				orig := strings.Split(inner, ",")

				for it_idx, it := range v.Items {
					switch vv := it.(type) {
					case items.Note:
						switch {
						case vv.ScaleNote == 0:
							nuNt.Items = append(nuNt.Items, vv)
						default:
							nnt := s.convertScaleNoteToAbsNote(ev.BarNo, vv)
							nuNt.Items = append(nuNt.Items, nnt)
							orig[it_idx] = nnt.String()
						}
					default:
						nuNt.Items = append(nuNt.Items, vv)
					}
				}

				posShift := ""

				if v.PosShift == 1 {
					posShift = ">"
				}

				if v.PosShift == -1 {
					posShift = "<"
				}

				nuEv.Item = nuNt
				nuEv.OriginalData = "{" + strings.Join(orig, ",") + "}" + items.Pos32thToString(v.EndPos) + posShift
				unrolled = append(unrolled, nuEv)
			default:
				unrolled = append(unrolled, ev)
			}
		}

		instr.Unrolled = unrolled
	}
}

func (s *unroller) unrollInstruments(bars []*score.Bar, instr []*score.Instrument) error {

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

	// evaluate randomness here before the template unfolding, to be able to randomly choose
	// templates
	s.evalRandomItems()

	// unfold the templates in a way that they are interrupted by the next non empty event
	s.unfoldTemplates()

	// evalute randomness a second time, so that randomness produced by templates can be evaluated
	s.evalRandomItems()

	// replace scale notes with absolute Notes
	s.replaceScaleNotes()

	// unfold the repeated items
	// we don't want to repeat templates or randomness via repeated items, because it is confusing
	// and does not bring much to the table
	// it is more interesting to be sure that repeating always means doubling the real item
	// before
	s.unfoldRepeated()

	return nil
}

func newScoreUnroller(src *score.Score) *unroller {
	nu := score.New()
	nu.IsUnrolled = true
	nu.HeaderIncludes = src.HeaderIncludes
	nu.FileName = src.FileName
	nu.Meta = src.Meta
	nu.BodyComments = src.BodyComments
	nu.Parts = src.Parts
	nu.Temperament = src.Temperament
	nu.TemplateDefinitions = src.TemplateDefinitions
	nu.IncludedTemplateDefinitions = src.IncludedTemplateDefinitions
	nu.HeaderComments = src.HeaderComments
	nu.BarNumbersTracked = true

	return &unroller{
		//		src:  src,
		dest: nu,
	}
}
