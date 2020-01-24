package sketch

import (
	"fmt"

	"gitlab.com/gomidi/muskel/items"
)

type column struct {
	name   string
	sketch *Sketch
}

func (p *column) newCall(pc *items.Call) *call {
	return &call{
		column: p,
		call:   pc,
	}
}

func (p *column) euclidEventStream(start uint, endPos uint, evts []*items.Event) (*eventStream, error) {
	var cl items.Call
	pc := p.newCall(&cl)
	printEvents("euclidEventStream before modifyEvents", evts)
	evts, absoluteEnd, err := pc.modifyEvents(start, endPos, evts)
	printEvents("euclidEventStream after modifyEvents", evts)
	if err != nil {
		return nil, err
	}
	_end := uint(endPos)

	if absoluteEnd < _end {
		_end = absoluteEnd
	}

	return newEventStream(start, absoluteEnd, true, evts...), nil
}

func (c *column) replaceNtupleTokens(in *items.NTuple) (out *items.NTuple, err error) {
	out = in.Dup().(*items.NTuple)
	out.Events = nil

	for _, it := range in.Events {
		switch v := it.Item.(type) {
		case *items.Call:
			if v.IsToken() {
				pc := c.newCall(v)
				posEv, err := pc.getEventStream(0, 1)
				if err != nil {
					return nil, err
				}
				out.Events = append(out.Events, posEv.events[0])
			} else {
				return nil, fmt.Errorf("could not use pattern (%q) within ntuple %v", v.Name, in)
			}
		case *items.NTuple:
			nit, err := c.replaceNtupleTokens(v)
			if err != nil {
				return nil, fmt.Errorf("could not replace tokens in ntuple %q within ntuple %v: %s", v, in, err)
			}
			var ev items.Event
			ev.Item = nit
			out.Events = append(out.Events, &ev)
		default:
			out.Events = append(out.Events, it.Dup())
		}
	}

	return out, nil
}

func (p *column) unroll(inputEvents []*items.Event, endPos uint, params []string) (unrolled []*items.Event, err error) {
	s := p.sketch

	var evts []*items.Event
	for i, inEv := range inputEvents {
		if inEv.Item == nil || err != nil {
			continue
		}

		if inc, isInc := inEv.Item.(items.Include); isInc {
			_ = inc
			panic(fmt.Sprintf("unsupported: includes in columns"))
		}

		pos := s.Positions[i]
		//var ev Event
		inEv.Position = s.getAbsPos(pos[0], pos[1])

		//		fmt.Printf("item: %T %v\n", it, it)
		evts = append(evts, inEv)
	}

	if err != nil {
		return
	}

	//fmt.Printf("rolledEvts: %v\n", rolledEvts)
	//fmt.Printf(">> rolledEvts\n")

	/*
		for _, ev := range rolledEvts {
			fmt.Printf("at %v: %v\n", ev.Position, ev)
		}
	*/
	//fmt.Printf("<< rolledEvts\n")

	var forward uint

	var mixed []*eventStream

	for i, ev := range evts {
		//fmt.Printf("################## %v event: %v ##################\n", i, ev)
		if ev.Item == nil {
			continue
		}

		// fmt.Printf("[%v] ev.Item %T\n", i, ev.Item)
		var posEv *eventStream

		switch v := ev.Item.(type) {
		//case *items.BarRepeater:
		//	fmt.Printf("bar repeater: %v mixed: %v\n", v, mixed)
		case *items.CommandCall:
			switch v.Name {
			case "_euclid":
				var eu items.EuclideanRhythm
				err = eu.Parse(v.Position, v.Params...)
				if err != nil {
					continue
				}

				_evts, err := eventsFromPatternDef(v.Name, eu.PatternDef, p.sketch.Score, 0, v.Position, params)

				//DEBUG = true
				printEvents("from _euclid", _evts)
				//DEBUG = false

				if err != nil {
					continue
				}

				if len(_evts) == 0 {
					continue
				}
				posEv, err = p.euclidEventStream(forward+ev.Position, endPos, _evts)
				if err != nil {
					continue
				}
			default:
				continue
			}
		case *items.NTuple:
			// TODO look inside each item and replace random things and templatecalls if there
			var until = findNextPos(i, int(forward), evts)
			if until < 0 {
				until = int(endPos)
			}

			// fmt.Printf("inside ntuple: %v\n", v)
			newItm, err := p.replaceNtupleTokens(v)
			if err != nil {
				return nil, err
			}

			//v.Items = newItsm
			var nev = ev.Dup()
			nev.Position = ev.Position + forward
			nev.PosShift = 0
			nev.Item = newItm
			posEv = newEventStream(nev.Position, uint(until), false, nev)
		case *items.MultiItem:
			// TODO look inside each item and replace random things and templatecalls if there
			var until = findNextPos(i, int(forward), evts)
			if until < 0 {
				until = int(endPos)
			}

			// fmt.Printf("inside multiitem: %v\n", v)
			var newItsm = make([]*items.Event, len(v.Events))
			for itidx, itm := range v.Events {
				switch vv := itm.Item.(type) {
				case *items.Call:
					pc := p.newCall(vv)
					posEv, err = pc.getEventStream(forward+ev.Position, endPos)
					if err != nil {
						return nil, err
					}
					// fmt.Printf("got item: %v\n", posEv.events[0].Item)
					newItsm[itidx] = posEv.events[0]
				default:
					newItsm[itidx] = itm
				}
			}
			v.Events = newItsm
			var nev = ev.Dup()
			nev.Position = ev.Position + forward
			nev.Item = v.Dup()
			posEv = newEventStream(nev.Position, uint(until), false, nev)

		case *items.Call:
			// TODO prevent endless loops from templates calling each other like col1 -> col2 -> col1 by keeping a stack of template calls
			// and checking them for duplicates (the stack may as well be a map[string]bool; makes is easier; we need the complete names in there
			pc := p.newCall(v)
			posEv, err = pc.getEventStream(forward+ev.Position, endPos)
			if err != nil {
				return nil, err
			}

		case *items.Override:
			switch vv := v.Item.(type) {
			case *items.Call:
				pc := p.newCall(v.Item.(*items.Call))
				if vv.IsPattern() {
					posEv, err = pc.getOverrideEventStream(forward+ev.Position, endPos)
				} else {
					posEv, err = pc.getOverrideEventStream(forward+ev.Position, forward+ev.Position+1)
				}
				if err != nil {
					return nil, err
				}
			case *items.Scale:
				return nil, fmt.Errorf("item type not allowed in override: %#v", vv)
			default:
				nev := ev.Dup()
				nev.Item = vv
				nev.Position = ev.Position + forward
				posEv = newEventStream(nev.Position, nev.Position+1, true, nev)
				posEv.isOverride = true
				//fmt.Printf("override of %q: from: %v to: %v\n", v.Item.(*items.Call).Name, forward+ev.Position, endPos)
			}

		default:
			var until = findNextPos(i, int(forward), evts)
			if until < 0 {
				until = int(endPos)
			}
			var nev = ev.Dup()
			nev.Position = ev.Position + forward
			posEv = newEventStream(nev.Position, uint(until), false, nev)
		}
		mixed = append(mixed, posEv)
	}

	if DEBUG {
		fmt.Printf(">> mixed\n")

		for _, ev := range mixed {
			fmt.Printf("at %v: %v\n", ev.start, ev)
		}
		fmt.Printf("<< mixed\n")
	}

	//fmt.Printf("mixed premerged: %v\n", mixed)
	unrolled = mergeEventStreams(mixed, s.projectedBarEnd)
	//fmt.Printf("eevts: %v\n", eevts)

	if DEBUG {
		fmt.Printf(">> after merge\n")

		for _, ev := range unrolled {
			fmt.Printf("at %v: %v\n", ev.Position, ev)
		}
		fmt.Printf("<< after merge\n")
	}

	// fmt.Printf("«««««««««««««« UNROLL %v endPos: %v\n", p.sketch.Name+"."+p.name, endPos)

	return

	// return
}

// Call parameterizes the sketch in the given column, parses and unrolls it and returns the resulting events
// as absolut positioned items. If the column does not exist, nil is returned
// endPos is the absolute end position of the piece. If 0, the last bar of the SketchTable is used
func (c *column) call(endPos uint, syncFirst bool, params ...string) ([]*items.Event, error) {
	s := c.sketch
	if c.name == "" {
		for i, col := range s.colOrder {
			if i == 0 || col[0] == '!' {
				c.name = col
			}
		}
	}

	col, has := s.Columns[c.name]
	if !has {
		//fmt.Printf("has no column: %q\n", col)
		return nil, fmt.Errorf("has no column: %q\n", col)
	}

	//fmt.Printf("raw data of col %q: %#v\n", colName, col)

	if endPos == 0 {
		endPos = s.projectedBarEnd
	}
	//fmt.Printf("call %q endPos: %v\n", colName, endPos)

	//fmt.Printf("endPos: %v\n", endPos)

	data := c.sketch.injectParams(col, params)

	//fmt.Printf("data: %v\n", data)
	events, loop, err := s.parseEvents(data)

	if err != nil {
		return nil, err
	}

	//fmt.Printf("items: %v loop: %v\n", items, loop)
	_ = loop
	events, err = c.unroll(events, endPos, params) // replaces template calls and imports
	if err != nil {
		return nil, err
	}

	if syncFirst {
		events = moveBySyncFirst(events)
	}

	printEvents("after unroll of events", events)

	//fmt.Println("returning from unrolled")
	events = c.unrollBarRepetitions(events, endPos)
	printEvents("after unrollBarRepetitions of events", events)

	events = c.unrollIncludedBars(events)

	printEvents("after unrollIncludedBars of events", events)
	//fmt.Printf("unrolled: %v\n", events)
	events, err = c.unrollPartRepetitions(events, endPos)
	if err != nil {
		return nil, err
	}
	//fmt.Printf(unrolledPartRepetitions: %v\n", events)
	printBars("after unrollPartRepetitions of events", c.sketch.Bars...)
	printEvents("after call of col "+c.name+" in sketch "+c.sketch.Name+" in file "+c.sketch.File, events)
	return events, nil
}

func (p *column) unrollIncludedBars(evts []*items.Event) []*items.Event {

	s := p.sketch

	var res []*items.Event
	var lastBarEnd uint

	for _, bar := range s.Bars {
		//fmt.Printf("sketch: %q bar %v no: %v pos: %v jumpto %q\n", s.Name, i, bar.No, bar.Position, bar.JumpTo)

		if bar.Include != nil {
			lastBarEnd = bar.Include.Length32ths
			ets, err := s.includeCol(bar.Position, p.name, *bar.Include)
			if err != nil {
				continue
			}
			printEvents("added included events", ets)

			if DEBUG {
				fmt.Printf("end is: %v\n", lastBarEnd)
			}
			res = append(res, ets...)
			continue
		}

		if DEBUG {
			fmt.Printf("bar has start position: %v\n", bar.Position)
		}

		endPos := bar.Position + uint(bar.Length32th())
		lastBarEnd = endPos
		if endPos > s.projectedBarEnd {
			endPos = s.projectedBarEnd
		}

		ets := getEventsInPosRange(bar.Position, endPos, evts)
		res = append(res, ets...)

	}

	ets := getEventsInPosRange(lastBarEnd, s.projectedBarEnd, evts)
	res = append(res, ets...)

	return res
}

func (p *column) unrollPartRepetitions(evts []*items.Event, stopPos uint) ([]*items.Event, error) {

	s := p.sketch

	var res []*items.Event
	var lastBarEnd uint

	for i, bar := range s.Bars {
		_ = i
		if DEBUG {
			fmt.Printf("sketch: %q bar %v no: %v pos: %v jumpto %q include: %#v length: %v\n", s.Name, i, bar.No, bar.Position, bar.JumpTo, bar.Include, bar.Length32th())
		}

		lastBarEnd = bar.Position + uint(bar.Length32th())

		if bar.JumpTo != "" {
			part, has := s.Parts[bar.JumpTo]
			if !has {
				return nil, fmt.Errorf("can't jump to part %q from bar %v: part is not defined", bar.JumpTo, bar.No+1)
			}
			//fmt.Printf("part %q: %v\n", bar.JumpTo, part)
			startPos := part[0]
			endPos := part[1]
			diff := (endPos - startPos) - uint(bar.Length32th())
			_ = diff

			/*
				if i < len(s.Bars)-1 {
					forwardBars(s.Bars[i+1:], diff)
				}
			*/

			partEvents := getEventsInPosRange(startPos, endPos, evts)

			var nevts []*items.Event

			for _, ev := range partEvents {
				nue := ev.Dup()
				nue.Position += bar.Position
				nue.Position -= startPos
				nevts = append(nevts, nue)
			}

			printEvents("nevts", nevts)

			/*
				for _, nv := range nevts {
					fmt.Printf("partEvent at %v: %v\n", nv.Position, nv)
				}
			*/
			res = append(res, nevts...)

		} else {
			endPos := bar.Position + uint(bar.Length32th())
			if endPos > s.projectedBarEnd {
				endPos = s.projectedBarEnd
			}
			if endPos > stopPos {
				endPos = stopPos
			}
			ets := getEventsInPosRange(bar.Position, endPos, evts)
			printEvents("ets", ets)
			res = append(res, ets...)
		}

	}

	ets := getEventsInPosRange(lastBarEnd, stopPos, evts)
	printEvents("rest", ets)
	res = append(res, ets...)

	// TODO check if it is correct
	return res, nil
}

func (p *column) unrollBarRepetitions(ins []*items.Event, endPos uint) (out []*items.Event) {
	s := p.sketch

	for i, ev := range ins {
		switch v := ev.Item.(type) {
		case *items.BarRepeater:
			//	fmt.Printf("bar repeater: %v endPos: %v\n", v, endPos)
			stopPos := endPos
			if i < len(ins)-1 {
				stBarIdx := s.getBarIdxOf(ins[i+1].Position)
				stBarPos := s.Bars[stBarIdx].Position
				if stBarPos < stopPos {
					stopPos = stBarPos
				}
			}
			//fmt.Printf("stopPos: %v\n", stopPos)
			bidx := s.getBarIdxOf(ev.Position)
			startBar := bidx - v.LastN
			repevts := s.getEventsInBars(startBar, bidx-1, out)
			if len(repevts) == 0 {
				//fmt.Printf("no repevts\n")
				continue
			}

			currentPos := s.Bars[bidx].Position

			diff := currentPos - s.Bars[startBar].Position
			var res []*items.Event
			res, currentPos = s.repeatBars(repevts, diff, stopPos)
			if len(res) == 0 {
				continue
			}
			out = append(out, res...)

			if !v.OnlyOnce {
				for {
					/*
						barIdx := s.getBarIdxOf(res[len(res)-1].Position)
						if barIdx+1 >= len(s.Bars)-1 {
							break
						}

						currentPos = s.Bars[barIdx+1].Position


					*/
					if currentPos >= stopPos {
						break
					}
					diff = currentPos - s.Bars[startBar].Position
					res, currentPos = s.repeatBars(repevts, diff, stopPos)
					out = append(out, res...)
				}
			}

			/*
				for _, rev := range repevts {
					nev := rev.Dup()
					nev.Position += diff
					out = append(out, nev)
				}
			*/
			// TODO handle v.OnlyOnce
		default:
			out = append(out, ev.Dup())
		}
	}

	/*
		fmt.Printf(">>>>>>>>>>>>>[%q] endpos: %v\n", s.Name, endPos)
		for i, e := range out {
			fmt.Printf("unrollBarRepetitions event %v pos: %v ev: %v\n", i, e.Position, e)
		}
		fmt.Printf("<<<<<<<<<<<<<[%q]\n", s.Name)
	*/
	return
}
