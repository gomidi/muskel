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

func (p *column) euclidEventStream(start uint, endPos uint, evts []*Event) (*eventStream, error) {
	var cl items.Call
	pc := p.newCall(&cl)
	//evts, diff, absoluteEnd := pc.modifyEvents(start, endPos, evts)
	printEvents("euclidEventStream before modifyEvents", evts)
	//fmt.Printf("calling modifyEvents(%v,%v)\n", start, endPos)
	evts, absoluteEnd, err := pc.modifyEvents(start, endPos, evts)
	printEvents("euclidEventStream after modifyEvents", evts)
	if err != nil {
		return nil, err
	}
	//_ = diff
	_end := uint(endPos)

	if absoluteEnd < _end {
		_end = absoluteEnd
	}

	return newEventStream(start, absoluteEnd, true, evts...), nil
}

func (c *column) replaceNtupleTokens(in *items.NTuple) (out *items.NTuple, err error) {
	out = in.Dup().(*items.NTuple)
	out.Items = nil
	//var parser items.Parser

	for _, it := range in.Items {
		switch v := it.(type) {
		case *items.Call:
			/*
				var newItsm = make([]items.Item, len(v.Items))
					for itidx, itm := range v.Items {
						switch vv := itm.(type) {
						case *items.Call:
							//fmt.Printf("templateCall at position: %v\n", ev.Position)
							pc := p.newCall(vv)
							posEv, err = pc.getEventStream(forward+ev.Position, endPos)
							if err != nil {
								return nil, err
							}
							// fmt.Printf("got item: %v\n", posEv.events[0].Item)
							newItsm[itidx] = posEv.events[0].Item
						default:
							newItsm[itidx] = itm
						}
					}
			*/
			if v.IsToken() {
				pc := c.newCall(v)
				//posEv, err = pc.getEventStream(forward+ev.Position, endPos)
				posEv, err := pc.getEventStream(0, 1)
				if err != nil {
					return nil, err
				}
				/*
					tok, err0 := c.sketch.Score.GetToken(v.Name)
					if err != nil {
						return nil, fmt.Errorf("could not find token %q within ntuple %v: %s", v.Name, in, err0)
					}
					nit, err := parser.ParseItem(tok, 0)
					if err != nil {
						return nil, fmt.Errorf("could not parse token %q resolving to %q within ntuple %v: %s", v.Name, tok, in, err)
					}
				*/
				out.Items = append(out.Items, posEv.events[0].Item)
			} else {
				return nil, fmt.Errorf("could not use pattern (%q) within ntuple %v", v.Name, in)
			}
		case *items.NTuple:
			nit, err := c.replaceNtupleTokens(v)
			if err != nil {
				return nil, fmt.Errorf("could not replace tokens in ntuple %q within ntuple %v: %s", v, in, err)
			}
			out.Items = append(out.Items, nit)
		default:
			out.Items = append(out.Items, it.Dup())
		}
	}

	return out, nil
}

func (p *column) unroll(ims []items.Item, endPos uint, params []string) (unrolled []*Event, err error) {
	s := p.sketch
	// "Templates" und imports werden ersetzt (triggert beim Template Phase 1-4)
	// absolute zeiten (vom nullpunkt der Sketchspalte werden gesetzt

	//findNextNotEmptyPos()
	//fmt.Printf("posDiff: %v\n", posDiff)
	//fmt.Printf("»»»»»»»»»»»»»»»» UNROLL %v endPos: %v\n", p.sketch.Name+"."+p.name, endPos)

	//p.findNextNotEmptyPos()
	//findNextPos()
	//var posDiff uint

	//firstNonNil := true
	//var lastPos0 int = -1

	var evts []*Event
	for i, it := range ims {
		if it == nil || err != nil {
			continue
		}

		if inc, isInc := it.(items.Include); isInc {
			_ = inc
			panic(fmt.Sprintf("unsupported: includes in columns"))
			/*
				incEvts, err := p.includeCol(inc)
				if err != nil {
					return nil, err
				}
				diff := includesDiff
				if lastPos0 >= 0 {
					br := s.Bars[lastPos0]
					diff += br.Position + uint(br.Length32th())
				}
				if posDiff > 0 {
					diff -= posDiff
				}

				incEvts = forwardEvents(incEvts, diff)
				printEvents("after including in col "+p.name, incEvts)
				evts = append(evts, incEvts...)
				includesDiff += inc.Length32ths
				s.projectedBarEnd += inc.Length32ths
				continue
			*/
		}

		pos := s.Positions[i]
		var ev Event
		ev.Position = s.getAbsPos(pos[0], pos[1])

		ev.Item = it
		//		fmt.Printf("item: %T %v\n", it, it)
		evts = append(evts, &ev)
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
			//var posEv = newPositionedEvents(ev.Position+forward, 0, false, false)
			//mixed = append(mixed, posEv)
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
			var nev Event
			nev.Position = ev.Position + forward
			nev.Item = newItm
			posEv = newEventStream(nev.Position, uint(until), false, &nev)
		case *items.MultiItem:
			// TODO look inside each item and replace random things and templatecalls if there
			var until = findNextPos(i, int(forward), evts)
			if until < 0 {
				until = int(endPos)
			}

			// fmt.Printf("inside multiitem: %v\n", v)
			var newItsm = make([]items.Item, len(v.Items))
			for itidx, itm := range v.Items {
				switch vv := itm.(type) {
				case *items.Call:
					pc := p.newCall(vv)
					posEv, err = pc.getEventStream(forward+ev.Position, endPos)
					if err != nil {
						return nil, err
					}
					// fmt.Printf("got item: %v\n", posEv.events[0].Item)
					newItsm[itidx] = posEv.events[0].Item
				default:
					newItsm[itidx] = itm
				}
			}
			v.Items = newItsm
			var nev Event
			nev.Position = ev.Position + forward
			nev.Item = v.Dup()
			posEv = newEventStream(nev.Position, uint(until), false, &nev)

		case *items.Call:
			// TODO prevent endless loops from templates calling each other like col1 -> col2 -> col1 by keeping a stack of template calls
			// and checking them for duplicates (the stack may as well be a map[string]bool; makes is easier; we need the complete names in there
			pc := p.newCall(v)
			//fmt.Printf("->> call of %q: from: %v to: %v\n", v.Name, forward+ev.Position, endPos)
			posEv, err = pc.getEventStream(forward+ev.Position, endPos)
			if err != nil {
				return nil, err
			}
			/*
				fmt.Printf(">> response\n")

				for _, __ev := range posEv.events {
					fmt.Printf("at %v: %v\n", __ev.start, __ev)
				}

				fmt.Printf("<< response\n")
			*/
			//fmt.Printf("<<- response of %q: from: %v to: %v\n%v\n", v.Name, forward+ev.Position, endPos, posEv)
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

		/*
			case *items.CommandCall:
				if v.Name == "import" {
					var file, sketch string
					var params []string
					switch len(v.Params) {
					case 0:
						panic("import must have 1 parameter")
					case 1:
						file = v.Params[0]
					case 2:
						file = v.Params[0]
						sketch = v.Params[1]
					default:
						file = v.Params[0]
						sketch = v.Params[1]
						params = v.Params[2:]
					}
					evts, diff := s.unrollImportTemplateCall(ev.Position+forward, file, sketch, params)
					forward += diff
					evt = append(evt, evts...)
				} else {
					panic("unknown command: " + v.Name)
				}
		*/
		default:
			var until = findNextPos(i, int(forward), evts)
			if until < 0 {
				until = int(endPos)
			}
			var nev Event
			nev.Position = ev.Position + forward
			nev.Item = ev.Item.Dup()
			posEv = newEventStream(nev.Position, uint(until), false, &nev)
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

/*
func (c *column) includeCol(inc items.Include) (evts []*Event, err error) {
	sk, err := c.sketch.Score.GetIncludedSketch(inc.File, inc.Sketch)
	if err != nil {
		return nil, fmt.Errorf("can't find sketch %q in file %q", inc.Sketch, inc.File)
	}
	if !sk.Score.HasTrack(c.name) {
		return nil, fmt.Errorf("can't find track %q in sketch %q in file %q", c.name, inc.Sketch, inc.File)
	}
	var patt = sk.newCol(c.name)
	return patt.call(0, false, inc.Params...)
}
*/

// Call parameterizes the sketch in the given column, parses and unrolls it and returns the resulting events
// as absolut positioned items. If the column does not exist, nil is returned
// endPos is the absolute end position of the piece. If 0, the last bar of the SketchTable is used
func (c *column) call(endPos uint, syncFirst bool, params ...string) ([]*Event, error) {
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
	items, loop, err := s.parseItems(data)

	if err != nil {
		return nil, err
	}

	//fmt.Printf("items: %v loop: %v\n", items, loop)
	_ = loop
	events, err := c.unroll(items, endPos, params) // replaces template calls and imports
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

	events, err = c.unrollIncludedBars(events)
	if err != nil {
		return nil, err
	}

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

func (p *column) unrollIncludedBars(evts []*Event) ([]*Event, error) {

	s := p.sketch

	var res []*Event
	var lastBarEnd uint

	for _, bar := range s.Bars {
		//fmt.Printf("sketch: %q bar %v no: %v pos: %v jumpto %q\n", s.Name, i, bar.No, bar.Position, bar.JumpTo)

		if bar.Include != nil {
			ets, end, err := s.includeCol(bar.Position, p.name, *bar.Include)
			if err != nil {
				return nil, err
			}
			//ets = forwardEvents(ets, bar.Position)
			printEvents("added included events", ets)
			res = append(res, ets...)
			lastBarEnd = end
			continue
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

	return res, nil
}

func (p *column) unrollPartRepetitions(evts []*Event, stopPos uint) ([]*Event, error) {

	s := p.sketch

	var res []*Event
	var lastBarEnd uint

	for i, bar := range s.Bars {
		_ = i
		//fmt.Printf("sketch: %q bar %v no: %v pos: %v jumpto %q\n", s.Name, i, bar.No, bar.Position, bar.JumpTo)

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

			var nevts []*Event

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

func (p *column) unrollBarRepetitions(ins []*Event, endPos uint) (out []*Event) {
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
			var res []*Event
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
