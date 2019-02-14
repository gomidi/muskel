package muskel

import (
	"fmt"
	"math/rand"

	"gitlab.com/gomidi/midi/mid"
)

type SMFWriter struct {
	score        *Score
	unrolledbars []*Bar
	wr           *mid.SMFWriter
}

func NewSMFWriter(s *Score) *SMFWriter {
	return &SMFWriter{score: s}
}

func (p *SMFWriter) Write(wr *mid.SMFWriter) error {

	/*
		err := p.unrollInstruments()
		if err != nil {
			return err
		}
	*/
	err := p.writeFirstTrack(wr)

	if err != nil {
		return err
	}

	for _, instr := range p.score.Instruments {
		// ignore instruments without a MIDIchannel
		if instr.MIDIChannel < 0 {
			continue
		}

		fmt.Println("EOT")
		wr.EndOfTrack()
		iw := newInstrumentSMFWriter(p, wr, instr)
		err = iw.writeUnrolled()
		//err = iw.writeTrack()
		//		err = p.writeInstrumentTrack(wr, instr)
		if err != nil {
			return err
		}
	}

	fmt.Println("EOT")
	wr.EndOfTrack()
	return nil
}

func (p *SMFWriter) writeFirstTrack(wr *mid.SMFWriter) error {
	num := uint8(4)
	denom := uint8(4)

	for i, b := range p.unrolledbars {
		/*
			b.positions = make([]uint8, len(b.originalPositions))
			var oldPos string

			for pi, pos := range b.originalPositions {
				oldPos, b.positions[pi] = positionTo32th(oldPos, pos)
			}
		*/

		fmt.Printf("bar %v: %#v\n", i, b)

		if i > 0 {
			fmt.Printf("[firstrack] Forward(1, 0, 0)\n")
			wr.Forward(1, 0, 0)
			/*
				fmt.Printf("Forward(0, %v, %v)\n", uint32(num), uint32(denom))
				wr.Forward(0, uint32(num), uint32(denom))
			*/
		}

		if b.timeSigChange[0] > 0 {
			num = b.timeSigChange[0]
			denom = b.timeSigChange[1]
		}
		fmt.Printf("write meter %v/%v\n", num, denom)

		if b.tempoChange > 0 {
			wr.TempoBPM(b.tempoChange)
			fmt.Printf("write tempo change %v\n", b.tempoChange)
		}
		wr.Meter(num, denom)
		//wr.Forward(1, 0, 0)
	}

	return nil
}

func (p *Score) trackBarNumbers() {
	if p.barNumbersTracked {
		return
	}
	for _, instr := range p.Instruments {
		var events []BarEvents

		for barNo, bar := range instr.events {
			es := make(BarEvents, len(bar))
			events = append(events, es)

			//			if bar.

			for barLine, ev := range bar {
				//
				/*
					set the bar number and position
				*/
				ev.BarNo = barNo
				ev.OriginalBarNo = barNo
				//fmt.Printf("barNo: %v, BarLine: %v p.Bar: %#v\n", barNo, barLine, p.Bars[barNo])
				ev.DistanceToStartOfBarIn32th = p.Bars[barNo].positions[barLine]

				events[barNo][barLine] = ev
			}
		}

		instr.events = events
		//instr.unrolled = events
	}
	p.barNumbersTracked = true
}

func (p *Score) unrollBarsAndJumps(nu *Score) {

	if p.isUnrolled {
		panic("p must not be the unrolled score")
	}

	if !nu.isUnrolled {
		panic("nu must be the unrolled score")
	}

	var unrolledBars []*Bar
	var instrBarevents = map[int][]BarEvents{}
	var num, denom = uint8(4), uint8(4)

	var newBarNo int
	//var barMoved int

	for _, instr := range p.Instruments {
		nu.Instruments = append(nu.Instruments, instr.Dup())
	}

	fmt.Printf("len bars: %v\n", len(p.Bars))

	for _, bar := range p.Bars {
		fmt.Printf("bar.barNo: %v, bar.originalBarNo: %v\n", bar.barNo, bar.originalBarNo)

		if jump := bar.jumpTo; jump != "" {
			//startJump := bar.originalBarNo

			/*
				var nub *Bar
			*/
			// unroll the repeating bars in p.Bars for the first track and for reference
			//		(and renumber them properly)

			// TODO track the end of the current part
			for j, bar2ndloop := range p.Bars[p.Parts[jump][0] : p.Parts[jump][1]+1] {

				nub := bar2ndloop.Dup()
				if bar2ndloop.timeSigChange[0] > 0 {
					num, denom = bar2ndloop.timeSigChange[0], bar2ndloop.timeSigChange[1]
				}
				nub.barNo = newBarNo
				nub.timeSig[0] = num
				nub.timeSig[1] = denom
				unrolledBars = append(unrolledBars, nub)
				for iii, _instr := range p.Instruments {
					fmt.Printf("adding bar %v => %v\n", p.Parts[jump][0]+j, newBarNo)
					instrBarevents[iii] = append(instrBarevents[iii], _instr.events[p.Parts[jump][0]+j])
				}
				newBarNo++
			}

			/*
				nub = nub.Dup()
				nub.barNo = newBarNo
				unrolledBars = append(unrolledBars, nub)
				for iii, _instr := range p.Instruments {
						fmt.Printf("adding bar %v => %v\n", p.Parts[jump]+j, newBarNo)
						instrBarevents[iii] = append(instrBarevents[iii], Event{Item: nil, BarNo: newBarNo})
					}
				newBarNo++
			*/
		} else {
			bar.barNo = newBarNo
			if bar.timeSigChange[0] > 0 {
				num, denom = bar.timeSigChange[0], bar.timeSigChange[1]
			}

			bar.timeSig[0] = num
			bar.timeSig[1] = denom

			unrolledBars = append(unrolledBars, bar)
			newBarNo++

			for ii, instr := range p.Instruments {
				fmt.Printf("adding bar %v\n", bar.originalBarNo)
				instrBarevents[ii] = append(instrBarevents[ii], instr.events[bar.originalBarNo])
			}

		}
	}

	nu.Bars = unrolledBars

	for ii, instr := range nu.Instruments {
		fmt.Printf("[0] len bars in instr: %v\n", len(instrBarevents[ii]))
		instr.events = instrBarevents[ii]
	}
}

func (p *Score) flattenInstrumentEvents() {
	if !p.isUnrolled {
		panic("must be unrolled")
	}
	var currentlyRepeatingBars []BarEvents
	var indexWithinCurrentlyRepeatingBars int

	for _, instr := range p.Instruments {
		var events []*Event
		for barNo, bar := range instr.events {

			numBar, denomBar := p.Bars[barNo].timeSig[0], p.Bars[barNo].timeSig[1]
			lenBar := length32ths(numBar, denomBar)

			numRep, untilNext := bar.RepeatingBars()

			if numRep > 0 {
				if !untilNext {
					for _, ev := range instr.events[barNo-1] {
						eNu := ev.Dup()
						eNu.BarNo = barNo
						// skip empty items
						if eNu.Item != nil && int(eNu.DistanceToStartOfBarIn32th) < lenBar {
							fmt.Printf("[0] setting bar no to : %v\n", barNo)
							events = append(events, eNu)
						}
					}
				} else {
					indexWithinCurrentlyRepeatingBars = 0
					currentlyRepeatingBars = instr.events[barNo-numRep : barNo]
					for _, ev := range instr.events[barNo-numRep] {
						eNu := ev.Dup()
						eNu.BarNo = barNo
						// skip empty items
						if eNu.Item != nil && int(eNu.DistanceToStartOfBarIn32th) < lenBar {
							fmt.Printf("[1] setting bar no to : %v\n", barNo)
							events = append(events, eNu)
						}
					}
				}
				continue
			}

			if bar.isEmpty() {
				switch len(currentlyRepeatingBars) {
				case 0:
					continue
				case 1:
					for _, ev := range currentlyRepeatingBars[0] {
						eNu := ev.Dup()
						eNu.BarNo = barNo
						// skip empty items
						if eNu.Item != nil && int(eNu.DistanceToStartOfBarIn32th) < lenBar {
							fmt.Printf("[2] setting bar no to : %v\n", barNo)
							events = append(events, eNu)
						}
					}
					continue
				default:
					fmt.Printf("IDX: %v, len %v\n", indexWithinCurrentlyRepeatingBars, len(currentlyRepeatingBars))
					indexWithinCurrentlyRepeatingBars = (indexWithinCurrentlyRepeatingBars + 1) % len(currentlyRepeatingBars)
					fmt.Printf("new IDX: %v\n", indexWithinCurrentlyRepeatingBars)
					for _, ev := range currentlyRepeatingBars[indexWithinCurrentlyRepeatingBars] {
						eNu := ev.Dup()
						fmt.Printf("[3] setting bar no to : %v\n", barNo)
						eNu.BarNo = barNo
						// skip empty items
						if eNu.Item != nil && int(eNu.DistanceToStartOfBarIn32th) < lenBar {
							events = append(events, eNu)
						}
					}
					continue
				}
			} else {
				currentlyRepeatingBars = []BarEvents{}
				indexWithinCurrentlyRepeatingBars = 0
				for barLine, ev := range bar {
					//
					/*
						set the bar number and position
					*/
					eNu := ev.Dup()
					eNu.BarNo = barNo
					eNu.DistanceToStartOfBarIn32th = p.Bars[barNo].positions[barLine]

					// skip empty items
					if ev.Item != nil && int(eNu.DistanceToStartOfBarIn32th) < lenBar {
						fmt.Printf("[4] setting bar no to : %v\n", barNo)
						events = append(events, eNu)
					}
				}
			}
		}

		instr.unrolled = events

		for _, unev := range instr.unrolled {
			fmt.Printf("Events: bar: %v, dist from start: %v\n", unev.BarNo, unev.DistanceToStartOfBarIn32th)
		}
	}

}

func (p *Score) evalRandomItems() {
	if !p.isUnrolled {
		panic("must be unrolled")
	}
	for _, instr := range p.Instruments {
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

func (p *Score) unfoldPatterns() {
	if !p.isUnrolled {
		panic("must be unrolled")
	}
	for _, instr := range p.Instruments {
		var unrolled []*Event
		for idx, ev := range instr.unrolled {
			// handle patterns

			switch v := ev.Item.(type) {
			case *PatternCall:
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
				if idx+1 < len(instr.unrolled) {
					// 1. calc the distance of the next event in 32ths (respecting bar changes etc)
					diffBars := instr.unrolled[idx+1].BarNo - ev.BarNo
					var diff32ths int
					for didx := 0; didx < diffBars; didx++ {
						diff32ths += p.Bars[ev.BarNo+didx].length32th()
					}

					diff32ths += int(instr.unrolled[idx+1].DistanceToStartOfBarIn32th)
					diff32ths -= int(ev.DistanceToStartOfBarIn32th)

					// 2a. calc the distance to the previous pattern event in 32ths
					preventRelPosition := uint8(0)
					_ = preventRelPosition
					prevPos := ""
					barLenInBetween := 0
				pattenEventLoop:
					for idxBarPev, barpev := range v.Events {
						currentBarNo := ev.BarNo + idxBarPev

						for idxPev, pev := range barpev {
							_ = pev
							nuev := ev.Dup()
							nuev.BarNo = currentBarNo
							nuev.Item = pev.item

							if idxBarPev == 0 && idxPev == 0 {
								//dist := pev.position
								if !v.SyncFirst {
									prevPos, nuev.DistanceToStartOfBarIn32th = positionTo32th(prevPos, pev.position)
								}
								preventRelPosition = nuev.DistanceToStartOfBarIn32th

								// 2c. otherwise skip the pattern event loop
								if int(nuev.DistanceToStartOfBarIn32th) >= diff32ths {
									break pattenEventLoop
								}

								// 2b. if it is smaller than the next normal event, write it to unrolled
								unrolled = append(unrolled, nuev)

							} else {
								prevPos, nuev.DistanceToStartOfBarIn32th = positionTo32th(prevPos, pev.position)

								// 2c. otherwise skip the pattern event loop
								if barLenInBetween+int(nuev.DistanceToStartOfBarIn32th) >= diff32ths {
									break pattenEventLoop
								}

								// 2b. if it is smaller than the next normal event, write it to unrolled
								unrolled = append(unrolled, nuev)
							}
						}

						prevPos = ""
						barLenInBetween += p.Bars[currentBarNo].length32th()

					}
				}

				//v.Events
			default:
				unrolled = append(unrolled, ev)
			}
		}

		instr.unrolled = unrolled
	}
}

func (p *Score) unrollInstruments(nu *Score) error {

	// attach the barnumbers and position in 32th relative to start to the event
	p.trackBarNumbers()

	// unroll the repetition within an instrument and the total jumps
	// thereby creating new bars and attaching those new bar numbers in addition
	// to the original bar numbers to the events
	p.unrollBarsAndJumps(nu)

	// since now we have the barnumbers properly attached to events, we can flatten
	// the events so that we have a single line of events where we could easily look
	// for the next event
	nu.flattenInstrumentEvents()

	// evaluate randomness here before the pattern unfolding, to be able to randomly choose
	// patterns
	nu.evalRandomItems()

	// unfold the patterns in a way that they are interrupted by the next non empty event
	nu.unfoldPatterns()

	// evalute randomness a second time, so that randomness produced by patterns can be evaluated
	nu.evalRandomItems()

	return nil
}
