package sketch

import (
	"fmt"

	"gitlab.com/gomidi/muskel/items"
	"gitlab.com/gomidi/muskel/patterncommands"
	"gitlab.com/gomidi/muskel/track"
)

type column struct {
	name   string
	sketch *Sketch
	track  *track.Track
}

func (p *column) newCall(pc *items.Call) *call {
	return &call{
		column: p,
		call:   pc,
	}
}

func (p *column) pipedEventStream(start uint, endPos uint, evts []*items.Event) (*eventStream, error) {
	var cl items.Call
	pc := p.newCall(&cl)
	printEvents("pipedEventStream before modifyEvents", evts)
	evts, absoluteEnd, err := pc.modifyEvents(start, endPos, evts)
	printEvents("pipedEventStream after modifyEvents", evts)
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

func (p *column) _unroll(evts []*items.Event, originalEndPos uint, params []string) (unrolled []*items.Event, endPos uint, err error) {

	endPos = originalEndPos

	s := p.sketch

	var forward uint

	var mixed []*eventStream

	for i, ev := range evts {
		//fmt.Printf("################## %v event: %v ##################\n", i, ev)
		if ev.Item == nil {
			continue
		}

		if ev.Item == items.End {
			//fmt.Printf("###### END ITEM in col %q of sketch %q ##########\n", p.name, p.sketch.Name)
			endPos = ev.Position
			break
		}

		// fmt.Printf("[%v] ev.Item %T\n", i, ev.Item)
		var posEv *eventStream

		switch v := ev.Item.(type) {
		//case *items.BarRepeater:
		//	fmt.Printf("bar repeater: %v mixed: %v\n", v, mixed)
		case *items.PipedPatternCommands:
			//fmt.Printf("got: *items.PipedPatternCommands\n")
			helper := &patterncmdHelper{
				column: p,
				params: params,
				inPipe: true,
			}

			for _, cmd := range v.Cmds {
				//fmt.Printf("command: %q\n", cmd.Name)
				if fn, hasCmd := patterncommands.Commands[cmd.Name[1:]]; hasCmd {
					//fmt.Printf("found command: %q\n", cmd.Name)
					helper.cmdName = cmd.Name[1:]
					_evts, err := fn(cmd.Params, helper)

					if err != nil {
						return nil, endPos, err
					}

					printEvents(cmd.Name, _evts)

					helper.pipeEvents = _evts
				} else {
					return nil, endPos, fmt.Errorf("unknown command %q", cmd.Name)
				}
			}

			posEv, err = p.pipedEventStream(forward+ev.Position, endPos, helper.pipeEvents)
			if err != nil {
				return nil, endPos, err
			}

		case *items.CommandCall:
			if fn, hasCmd := patterncommands.Commands[v.Name]; hasCmd {
				helper := &patterncmdHelper{
					cmdName: v.Name,
					column:  p,
					params:  params,
				}

				_evts, err := fn(v.Params, helper)

				if err != nil {
					return nil, endPos, err
				}

				printEvents(v.Name, _evts)

				if len(_evts) == 0 {
					continue
				}
				posEv, err = p.pipedEventStream(forward+ev.Position, endPos, _evts)
				if err != nil {
					return nil, endPos, err
				}
			} else {
				return nil, endPos, fmt.Errorf("unknown command %q", v.Name)
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
				return nil, endPos, err
			}

			//v.Items = newItsm
			var nev = ev.Dup()
			nev.Position = ev.Position + forward
			nev.PosShift = 0
			nev.Item = newItm
			posEv = newEventStream(nev.Position, uint(until), false, nev)
		case *items.BarRepeater:
			//fmt.Printf("bar repeater: %v endPos: %v in col %q of sketch %q\n", v, endPos, p.name, p.sketch.Name)
			stopPos := endPos
			if i < len(evts)-1 {
				stBarIdx := s.getBarIdxOf(evts[i+1].Position)
				stBarPos := s.Bars[stBarIdx].Position
				if stBarPos < stopPos {
					stopPos = stBarPos
				}
				//fmt.Printf("stopPos: %v next bars position: %v in col %q of sketch %q\n", stopPos, stBarPos, p.name, p.sketch.Name)
			}
			bidx := s.getBarIdxOf(ev.Position)
			startBar := bidx - v.LastN
			repevts := s.getEventsInBars(startBar, bidx-1, evts)
			if len(repevts) == 0 {
				//fmt.Printf("no repevts\n")
				continue
			}

			currentPos := s.Bars[bidx].Position

			diff := currentPos - s.Bars[startBar].Position
			var ignoreEndPos uint
			repevts, ignoreEndPos, err = p._unroll(repevts, stopPos, nil)
			_ = ignoreEndPos
			//fmt.Printf("setting endPos to %v from BarRepeater in col %q of sketch %q via stopPos %v\n", endPos, p.name, p.sketch.Name, stopPos)
			if err != nil {
				//fmt.Printf("error while unrolling: %v\n", err)
				return nil, endPos, err
			}

			//DEBUG = true
			printEvents("bar repeater I, after unroll", repevts)
			//DEBUG = false

			var res []*items.Event

			//fmt.Printf("stopPos: %v\n", stopPos)
			origDiff := diff
			endOfStream := origDiff

			//res, currentPos = s.repeatBars(repevts, diff, stopPos)
			res, currentPos = s.repeatBars(repevts, diff)
			if len(res) == 0 {
				//fmt.Printf("no repeatBars\n")
				continue
			}

			//DEBUG = true
			printEvents("bar repeater I, after repeatBars", res)
			//DEBUG = false

			var all []*items.Event

			all = append(all, res...)

			//mixed = append(mixed, newEventStream(s.Bars[bidx].Position, currentPos, true, res...))
			//out = append(out, res...)

			var repetitions uint = 1

			if !v.OnlyOnce {
			barRepeatLoop:
				for {
					/*
						barIdx := s.getBarIdxOf(res[len(res)-1].Position)
						if barIdx+1 >= len(s.Bars)-1 {
							break
						}

						currentPos = s.Bars[barIdx+1].Position


					*/
					if currentPos >= stopPos {
						break barRepeatLoop
					}
					repetitions++
					endOfStream += origDiff
					//diff = currentPos - s.Bars[startBar].Position
					//oldStart := currentPos
					//res, currentPos = s.repeatBars(repevts, diff, stopPos)
					res, currentPos = s.repeatBars(repevts, diff*repetitions)

					//printEvents("bar repeater II, after repeatBars", res)
					all = append(all, res...)
					/*
						res, err = p._unroll(res, s.projectedBarEnd, nil)
						if err != nil {
							return nil, err
						}
					*/
					//printEvents("bar repeater II, after unroll", res)
					//mixed = append(mixed, newEventStream(oldStart, currentPos, true, res...))
					//out = append(out, res...)
				}
			}

			//sorted := items.SortEvents(all)
			//sort.Sort(sorted)

			//DEBUG = true
			//printEvents("bar repeater II, sorted", sorted)
			//DEBUG = false
			//fmt.Printf("bar repeat: create stream events from %v to %v\n", s.Bars[bidx].Position, currentPos)
			//mixed = append(mixed, newEventStream(s.Bars[bidx].Position, currentPos, true, sorted...))
			mixed = append(mixed, newEventStream(s.Bars[bidx].Position, currentPos, true, all...))
			posEv = nil

		case *items.PartRepeat:
			partname := v.Part
			part, has := s.Parts[partname]
			if !has {
				return nil, endPos, fmt.Errorf("can't repeat part %q at %v: part is not defined", partname, ev.Position)
			}

			//fmt.Printf("part %q: %v\n", bar.JumpTo, part)
			startPos := part[0]
			_endPos := part[1]
			diff := (_endPos - startPos) //- uint(bar.Length32th())
			//_ = diff
			if _endPos > s.projectedBarEnd {
				_endPos = s.projectedBarEnd
			}

			/*
				if endPos >= ev.Position {
					endPos = ev.Position - 1
				}
			*/
			partEvents := getEventsInPosRange(startPos, _endPos, evts)

			partEvents, _endPos, err = p._unroll(partEvents, _endPos, nil)
			if err != nil {
				return nil, endPos, err
			}

			printEvents(fmt.Sprintf("partEvents from %v to %v", startPos, _endPos), partEvents)

			for r := 0; r < int(v.Repeat); r++ {
				var nevts []*items.Event

				for _, _ev := range partEvents {
					nue := _ev.Dup()
					nue.Position += ev.Position + uint(r)*diff
					if v.SyncFirst {
						nue.Position -= startPos
					}
					nevts = append(nevts, nue)
				}

				totalStart := startPos + ev.Position + (uint(r) * diff)

				mixed = append(mixed, newEventStream(totalStart, totalStart+diff, true, nevts...))
			}

			posEv = nil

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
						return nil, endPos, err
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
				return nil, endPos, err
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
					return nil, endPos, err
				}
			case *items.Scale:
				return nil, endPos, fmt.Errorf("item type not allowed in override: %#v", vv)
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
		if posEv != nil {
			mixed = append(mixed, posEv)
		}
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

	return
}

func (p *column) unroll(inputEvents []*items.Event, originalEndPos uint, params []string) (unrolled []*items.Event, endPos uint, err error) {
	s := p.sketch
	//endPos = originalEndPos

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

	/*
		unrolled, err = p._unroll(evts, endPos, params)

		if err != nil {
			return
		}
	*/

	return p._unroll(evts, originalEndPos, params)

	//fmt.Printf("rolledEvts: %v\n", rolledEvts)
	//fmt.Printf(">> rolledEvts\n")

	/*
		for _, ev := range rolledEvts {
			fmt.Printf("at %v: %v\n", ev.Position, ev)
		}
	*/
	//fmt.Printf("<< rolledEvts\n")

	// fmt.Printf("«««««««««««««« UNROLL %v endPos: %v\n", p.sketch.Name+"."+p.name, endPos)

	return

	// return
}

// Call parameterizes the sketch in the given column, parses and unrolls it and returns the resulting events
// as absolut positioned items. If the column does not exist, nil is returned
// endPos is the absolute end position of the piece. If 0, the last bar of the SketchTable is used
func (c *column) call(originalEndPos uint, syncFirst bool, params ...string) (events []*items.Event, endPos uint, err error) {
	s := c.sketch
	endPos = originalEndPos
	if c.name == "" {
		for i, col := range s.colOrder {
			if i == 0 || col[0] == '!' {
				c.name = col
			}
		}
	}

	//if DEBUG {
	//fmt.Printf("at start of call in col %q of sketch %q originalEndPos: %v\n", c.name, c.sketch.Name, originalEndPos)
	//}

	col, has := s.Columns[c.name]
	if !has {
		//fmt.Printf("has no column: %q\n", col)
		return nil, endPos, fmt.Errorf("has no column: %q\n", col)
	}

	//fmt.Printf("raw data of col %q: %#v\n", colName, col)

	if endPos == 0 {
		endPos = s.projectedBarEnd
	}
	//fmt.Printf("call %q endPos: %v\n", colName, endPos)

	//fmt.Printf("endPos: %v\n", endPos)

	data := c.sketch.injectParams(col, params)

	var loop uint

	//fmt.Printf("data: %v\n", data)
	events, loop, err = s.parseEvents(data, endPos)

	if err != nil {
		return nil, endPos, err
	}

	//fmt.Printf("items: %v loop: %v\n", items, loop)
	_ = loop
	//if DEBUG {
	//fmt.Printf("before unroll in col %q of sketch %q endPos: %v\n", c.name, c.sketch.Name, endPos)
	//}

	events, endPos, err = c.unroll(events, endPos, params) // replaces template calls and imports
	if err != nil {
		return nil, endPos, err
	}

	//if DEBUG {
	//fmt.Printf("after unroll in col %q of sketch %q endPos: %v\n", c.name, c.sketch.Name, endPos)
	//}

	if syncFirst {
		events = moveBySyncFirst(events)
	}

	printEvents("after unroll of events", events)

	//fmt.Println("returning from unrolled")
	//events = c.unrollBarRepetitions(events, endPos)
	//printEvents("after unrollBarRepetitions of events", events)

	events = c.unrollIncludedBars(events)

	printEvents("after unrollIncludedBars of events", events)
	//fmt.Printf("unrolled: %v\n", events)
	//if DEBUG {
	//fmt.Printf("unrollPartRepetitionsOfBars in col %q of sketch %q\n", c.name, c.sketch.Name)
	//}
	events, err = c.unrollPartRepetitionsOfBars(events, endPos)
	if err != nil {
		return nil, endPos, err
	}

	/*
		events, err = c.unrollPartRepetitions(events)
		if err != nil {
			return nil, err
		}
	*/

	//fmt.Printf(unrolledPartRepetitions: %v\n", events)
	printBars("after unrollPartRepetitions of events", c.sketch.Bars...)
	printEvents("after call of col "+c.name+" in sketch "+c.sketch.Name+" in file "+c.sketch.File, events)
	return events, endPos, nil
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

func (p *column) unrollPartRepetitions(evts []*items.Event) (res []*items.Event, err error) {
	s := p.sketch

	var skipPos uint

	var mixed []*eventStream

	for _, evt := range evts {
		if evt.Position > s.projectedBarEnd {
			continue
		}

		if evt.Item != nil {
			if partItem, is := evt.Item.(*items.PartRepeat); is {
				partname := partItem.Part
				part, has := s.Parts[partname]
				if !has {
					return nil, fmt.Errorf("can't repeat part %q at %v: part is not defined", partname, evt.Position)
				}

				//fmt.Printf("part %q: %v\n", bar.JumpTo, part)
				startPos := part[0]
				endPos := part[1]
				diff := (endPos - startPos) //- uint(bar.Length32th())
				//_ = diff
				if endPos > s.projectedBarEnd {
					endPos = s.projectedBarEnd
				}

				partEvents := getEventsInPosRange(startPos, endPos, evts)

				// TODO:
				//
				// 1. handle overrides somehow
				// 2. handle repeats of parts (like Calls)
				var nevts []*items.Event

				for _, ev := range partEvents {
					nue := ev.Dup()
					nue.Position += evt.Position
					if partItem.SyncFirst {
						nue.Position -= startPos
					}
					nevts = append(nevts, nue)
				}

				mixed = append(mixed, newEventStream(startPos, endPos, true, nevts...))

				//printEvents("nevts", nevts)

				// overrides
				otherEvents := getEventsInPosRange(endPos, endPos+diff, evts)

				for _, otherE := range otherEvents {
					switch otherE.Item.(type) {
					case *items.PartRepeat:
						// ignore
					case *items.Override:
						fmt.Printf("got override: %v\n", otherE)
						oes := newEventStream(otherE.Position, otherE.Position+1, false, otherE)
						oes.isOverride = true
						mixed = append(mixed, oes)
					default:
						nes := newEventStream(otherE.Position, otherE.Position+1, false, otherE)
						mixed = append(mixed, nes)
					}
				}

				skipPos = endPos + diff

				//res = append(res, nevts...)

			} else {
				if evt.Position >= skipPos {
					nes := newEventStream(evt.Position, evt.Position+1, false, evt)
					mixed = append(mixed, nes)
				}
				//res = append(res, evt)
			}
		}
	}

	unrolled := mergeEventStreams(mixed, s.projectedBarEnd)

	return unrolled, nil
}

func (p *column) unrollPartRepetitionsOfBars(evts []*items.Event, stopPos uint) ([]*items.Event, error) {

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
			//fmt.Printf("endPos: %v s.projectedBarEnd: %v stopPos: %v\n", endPos, s.projectedBarEnd, stopPos)
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
