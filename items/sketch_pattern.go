package items

import "fmt"

type sketchPattern struct {
	column  Columner
	pattern *Pattern
}

func ReplaceNtupleTokens(c Columner, in *NTuple) (out *NTuple, err error) {
	out = in.Dup().(*NTuple)
	out.Events = nil

	for _, it := range in.Events {
		switch v := it.Item.(type) {
		case *Token:
			pc := v.newSketchToken(c)
			posEv, err := pc.getEventStream(0, 1)
			if err != nil {
				return nil, err
			}
			out.Events = append(out.Events, posEv.Events[0])
		case *Pattern:
			return nil, fmt.Errorf("could not use pattern (%q) within ntuple %v", v.Name, in)
		case *NTuple:
			nit, err := ReplaceNtupleTokens(c, v)
			if err != nil {
				return nil, fmt.Errorf("could not replace tokens in ntuple %q within ntuple %v: %s", v, in, err)
			}
			var ev Event
			ev.Item = nit
			out.Events = append(out.Events, &ev)
		default:
			out.Events = append(out.Events, it.Dup())
		}
	}

	return out, nil
}

func (pc *sketchPattern) unrollPattern(start uint, until uint) (evt []*Event, absoluteEnd uint, err error) {
	tc := pc.pattern
	var evts []*Event
	var length uint
	evts, length, err = pc.column.Call(until, tc.SyncFirst, tc.Params...)
	if err != nil {
		return
	}

	//length := absoluteEnd - start

	if DEBUG {
		fmt.Printf("length: %v start: %v\n", length, start)
	}

	PrintEvents("sketchPattern.unrollPattern", evts)

	for _, ev := range evts {
		ev.PosShift += tc.PosShift
	}

	//printEvents(fmt.Sprintf("after call.call(until: %v, syncfirst: %v, params: %v)", until, tc.SyncFirst, tc.Params), evts)

	var _evt []*Event
	_evt, absoluteEnd, err = pc.modifyEvents(start, start+length, evts)
	if err != nil {
		return
	}

	//fmt.Printf("until: %v absoluteEnd: %v\n", until, absoluteEnd)
	// TODO check, if until and absoluteEnd is correct

	// TODO check if length is correct

	if DEBUG {
		fmt.Printf("absoluteEnd: %v\n", absoluteEnd)
	}

	/*
		if DEBUG {
			fmt.Printf("absoluteEnd: %v start: %v until: %v length: %v\n", absoluteEnd, start, until, length)
		}
	*/

	//printEvents("before repeat", _evt)

	repeat := int(tc.Repeat)
	if repeat == 0 {
		repeat = 1
	}

	var __evts []*Event

	for i := 0; i < repeat; i++ {
		/*
			if DEBUG {
				fmt.Printf("forwarding by %v\n", length*uint(i))
			}
		*/
		__evts = append(__evts, ForwardEvents(_evt, length*uint(i))...)
		absoluteEnd += length * uint(i)
	}

	/*
		printEvents("after repeat", __evts)

		if DEBUG {
			fmt.Printf("absoluteEnd: %v start: %v until: %v length: %v\n", absoluteEnd, start, until, length)
		}
	*/

	for _, ev := range __evts {
		if ev.Position < until {
			evt = append(evt, ev)
		}
	}

	if absoluteEnd > until {
		absoluteEnd = until
	}

	/*
		printEvents("after unrollPattern", evt)

		if DEBUG {
			fmt.Printf("absoluteEnd: %v start: %v until: %v length: %v\n", absoluteEnd, start, until, length)
		}
	*/
	return
}

type dynamicAdder interface {
	AddDynamic(orig string) (nu string)
}

func (p *sketchPattern) modifyItem(it Item) (out Item, err error) {
	patt := p.pattern
	out = it

	switch v := out.(type) {
	case *Note:
		if p != nil {
			v.Dynamic = patt.AddDynamic(v.Dynamic)
			if v.PosShift == 0 && patt.PosShift != 0 {
				v.PosShift = patt.PosShift
			}
		}
	case *MIDINote:
		if p != nil {
			v.Dynamic = patt.AddDynamic(v.Dynamic)
			if v.PosShift == 0 && patt.PosShift != 0 {
				v.PosShift = patt.PosShift
			}
		}
	case *Token:
		if p != nil {
			return p.column.ModifyToken(v)
		}
	case *NTuple:
		if p != nil {
			out, err = ReplaceNtupleTokens(p.column, v)

			if err != nil {
				return nil, err
			}
		}
	}
	return
}

func (pc *sketchPattern) modifyEvents(start uint, until uint, evts []*Event) (evt []*Event, end uint, err error) {
	projectedBarEnd := pc.column.EndPosition()

	if until < projectedBarEnd {
		projectedBarEnd = until
	}

	for _, ev := range evts {
		pos := start + ev.Position
		if until > 0 && pos >= until {
			break
		}
		nuEv := ev.Dup()
		nuEv.Position = pos
		nuEv.Item, err = pc.modifyItem(ev.Item.Dup())
		if err != nil {
			return
		}
		evt = append(evt, nuEv)
	}

	end = projectedBarEnd
	return
}

func (pc *sketchPattern) _getEventStream(start uint, endPos uint, isOverride bool) (*EventStream, error) {
	evts, diff, absoluteEnd, err := pc.column.UnrollPattern(start, endPos, pc.pattern)
	if err != nil {
		return nil, err
	}
	PrintEvents("sketchPattern._getEventStream", evts)
	_ = diff
	end := uint(endPos)

	if absoluteEnd < end && isOverride {
		end = absoluteEnd
	}

	es := NewEventStream(pc.pattern.SyncFirst, start, end, true, evts...)
	es.IsOverride = isOverride
	return es, nil
}

func (c *sketchPattern) getEventStream(start uint, end uint) (*EventStream, error) {
	cc := c.pattern
	if cc.Lyrics != nil {
		es, err := c._getEventStream(start, end, false)
		if err != nil {
			return nil, err
		}
		es.Events, err = c.column.ApplyLyricsTable(cc.Lyrics, es.Events)
		if err != nil {
			return nil, err
		}
		return es, nil
	}
	return c._getEventStream(start, end, false)
}

func (pc *sketchPattern) getOverrideEventStream(start uint, endPos uint) (*EventStream, error) {
	es, err := pc._getEventStream(start, endPos, true)
	es.End = es.Start + es.End
	return es, err
}
