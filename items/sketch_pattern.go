package items

import "fmt"

type SketchPattern struct {
	column  Columner
	pattern *Pattern
}

func NewSketchPattern(column Columner, pattern *Pattern) *SketchPattern {
	return &SketchPattern{
		column:  column,
		pattern: pattern,
	}
}

type Columner interface {
	Call(until uint, syncFirst bool, params ...string) (evts []*Event, absoluteEnd uint, err error)
	//ReplaceNtupleTokens(*NTuple) (*NTuple, error)
	EndPosition() uint
	UnrollPattern(start uint, until uint, cc *Pattern) (evt []*Event, diff uint, end uint, err error)
	ModifyToken(tk *Token) (Item, error)
	ApplyLyricsTable(lt *LyricsTable, evts []*Event) ([]*Event, error)
	ParseEvents(data []string) (evts []*Event, absEnd uint, err error)
	GetToken(origName string, params []string) (val string, err error)
}

func ReplaceNtupleTokens(c Columner, in *NTuple) (out *NTuple, err error) {
	out = in.Dup().(*NTuple)
	out.Events = nil

	for _, it := range in.Events {
		switch v := it.Item.(type) {
		case *Token:
			pc := NewSketchToken(c, v)
			posEv, err := pc.GetEventStream(0, 1)
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

func (pc *SketchPattern) UnrollPattern(start uint, until uint) (evt []*Event, absoluteEnd uint, err error) {
	tc := pc.pattern
	var evts []*Event
	evts, absoluteEnd, err = pc.column.Call(until, tc.SyncFirst, tc.Params...)
	if err != nil {
		return
	}

	for _, ev := range evts {
		ev.PosShift += tc.PosShift
	}

	//printEvents(fmt.Sprintf("after call.call(until: %v, syncfirst: %v, params: %v)", until, tc.SyncFirst, tc.Params), evts)

	var _evt []*Event
	_evt, absoluteEnd, err = pc.ModifyEvents(start, start+absoluteEnd, evts)
	if err != nil {
		return
	}

	//fmt.Printf("until: %v absoluteEnd: %v\n", until, absoluteEnd)
	// TODO check, if until and absoluteEnd is correct

	// TODO check if length is correct
	length := absoluteEnd //- start

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

func (c *SketchPattern) modifyItem(it Item) (Item, error) {
	//cc := c.call.(dynamicAdder)
	cc := c.pattern
	/*
		if DEBUG {
			fmt.Printf("modify item %T %v\n", it, it)
		}
	*/
	switch v := it.(type) {
	case *Note:
		it := v.Dup().(*Note)
		if c != nil {
			it.Dynamic = cc.AddDynamic(v.Dynamic)
			if it.PosShift == 0 && cc.PosShift != 0 {
				it.PosShift = cc.PosShift
			}
		}
		return it, nil
	case *MIDINote:
		it := v.Dup().(*MIDINote)
		if c != nil {
			it.Dynamic = cc.AddDynamic(v.Dynamic)
			if it.PosShift == 0 && cc.PosShift != 0 {
				it.PosShift = cc.PosShift
			}
		}
		return it, nil

	case *Token:
		if c != nil {
			return c.column.ModifyToken(v)
		} else {
			return it.Dup(), nil
		}

	case *LyricsTable:
		return it.Dup(), nil

	case *Pattern:
		return it.Dup(), nil

	case *NTuple:
		if c != nil {
			it := v.Dup().(*NTuple)
			it, err := ReplaceNtupleTokens(c.column, v)

			if err != nil {
				return nil, err
			}

			return it, nil
		} else {
			return v.Dup(), nil
		}
	default:
		return it.Dup(), nil
	}
}

func (pc *SketchPattern) ModifyEvents(start uint, until uint, evts []*Event) (evt []*Event, end uint, err error) {
	/*
		if DEBUG {
			fmt.Printf("modifyEvents(start: %v until: %v)\n", start, until)
		}
	*/
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
		nuEv.Item, err = pc.modifyItem(ev.Item)
		if err != nil {
			return
		}
		evt = append(evt, nuEv)
	}

	evt, end = SliceEvents(pc.pattern.Slice, evt, projectedBarEnd)

	if pc.pattern.Slice[0] > 0 {
		evt = MoveBySyncFirst(evt)
		evt = ForwardEvents(evt, start)
	}

	return
}

func (pc *SketchPattern) _getEventStream(start uint, endPos uint, isOverride bool) (*EventStream, error) {
	evts, diff, absoluteEnd, err := pc.column.UnrollPattern(start, endPos, pc.pattern)
	if err != nil {
		return nil, err
	}
	_ = diff
	end := uint(endPos)

	if absoluteEnd < end && isOverride {
		end = absoluteEnd
	}

	_ = absoluteEnd

	//printEvents("after unrolling colum "+pc.column.name+" of sketch "+pc.column.sketch.Name, evts)

	var es *EventStream

	es = NewEventStream(start, end, true, evts...)

	es.IsOverride = isOverride
	return es, nil
}

func (c *SketchPattern) GetEventStream(start uint, end uint) (*EventStream, error) {
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

func (pc *SketchPattern) GetOverrideEventStream(start uint, endPos uint) (*EventStream, error) {
	es, err := pc._getEventStream(start, endPos, true)
	es.End = es.Start + es.End
	return es, err
}
