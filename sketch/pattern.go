package sketch

import (
	"fmt"

	"gitlab.com/gomidi/muskel/items"
)

type pattern struct {
	column  *column
	pattern *items.Pattern
}

func (pc *pattern) unrollPattern(start uint, until uint) (evt []*items.Event, absoluteEnd uint, err error) {
	p := pc.column
	tc := pc.pattern
	var evts []*items.Event
	evts, absoluteEnd, err = p.call(until, tc.SyncFirst, tc.Params...)
	if err != nil {
		return
	}

	for _, ev := range evts {
		ev.PosShift += tc.PosShift
	}

	printEvents(fmt.Sprintf("after call.call(until: %v, syncfirst: %v, params: %v)\n", until, tc.SyncFirst, tc.Params), evts)

	var _evt []*items.Event
	_evt, absoluteEnd, err = pc.modifyEvents(start, absoluteEnd, evts)
	if err != nil {
		return
	}

	//fmt.Printf("until: %v absoluteEnd: %v\n", until, absoluteEnd)
	// TODO check, if until and absoluteEnd is correct

	// TODO check if length is correct
	length := absoluteEnd //- start

	if DEBUG {
		fmt.Printf("absoluteEnd: %v start: %v until: %v length: %v\n", absoluteEnd, start, until, length)
	}

	printEvents("before repeat", _evt)

	repeat := int(tc.Repeat)
	if repeat == 0 {
		repeat = 1
	}

	var __evts []*items.Event

	for i := 0; i < repeat; i++ {
		if DEBUG {
			fmt.Printf("forwarding by %v\n", length*uint(i))
		}
		__evts = append(__evts, forwardEvents(_evt, length*uint(i))...)
		absoluteEnd += length * uint(i)
	}

	printEvents("after repeat", __evts)

	if DEBUG {
		fmt.Printf("absoluteEnd: %v start: %v until: %v length: %v\n", absoluteEnd, start, until, length)
	}

	for _, ev := range __evts {
		if ev.Position < until {
			evt = append(evt, ev)
		}
	}

	if absoluteEnd > until {
		absoluteEnd = until
	}

	printEvents("after unrollPattern", evt)

	if DEBUG {
		fmt.Printf("absoluteEnd: %v start: %v until: %v length: %v\n", absoluteEnd, start, until, length)
	}

	return
}

type dynamicAdder interface {
	AddDynamic(orig string) (nu string)
}

func (c *pattern) modifyItem(it items.Item) (items.Item, error) {
	//cc := c.call.(dynamicAdder)
	cc := c.pattern
	if DEBUG {
		fmt.Printf("modify item %T %v\n", it, it)
	}
	switch v := it.(type) {
	case *items.Note:
		it := v.Dup().(*items.Note)
		if c != nil {
			it.Dynamic = cc.AddDynamic(v.Dynamic)
			if it.PosShift == 0 && cc.PosShift != 0 {
				it.PosShift = cc.PosShift
			}
		}
		return it, nil
	case *items.MIDINote:
		it := v.Dup().(*items.MIDINote)
		if c != nil {
			it.Dynamic = cc.AddDynamic(v.Dynamic)
			if it.PosShift == 0 && cc.PosShift != 0 {
				it.PosShift = cc.PosShift
			}
		}
		return it, nil

	case *items.Token:
		if c != nil {
			pc := c.column.newToken(v)
			//posEv, err = pc.getEventStream(forward+ev.Position, endPos)
			posEv, err := pc.getEventStream(0, 1)
			if err != nil {
				return nil, err
			}
			it := posEv.events[0].Item
			return it, nil
		} else {
			return it.Dup(), nil
		}

	case *items.LyricsTable:
		return it.Dup(), nil

	case *items.Pattern:
		return it.Dup(), nil

	case *items.NTuple:
		if c != nil {
			it := v.Dup().(*items.NTuple)
			it, err := c.column.replaceNtupleTokens(v)

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

func (pc *pattern) modifyEvents(start uint, until uint, evts []*items.Event) (evt []*items.Event, end uint, err error) {
	projectedBarEnd := pc.column.sketch.projectedBarEnd

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

	evt, end = sliceEvents(pc.pattern.Slice, evt, projectedBarEnd)

	if pc.pattern.Slice[0] > 0 {
		evt = moveBySyncFirst(evt)
		evt = forwardEvents(evt, start)
	}

	return
}

func printEvents(message string, evts []*items.Event) {
	if DEBUG {
		fmt.Printf("##> Events %s\n", message)
		for _, ev := range evts {
			fmt.Printf("[%v] %v ", ev.Position, ev.String())
		}
		fmt.Printf("\n##< Events %s\n", message)
	}
}

func (pc *pattern) _getEventStream(start uint, endPos uint, isOverride bool) (*eventStream, error) {
	evts, diff, absoluteEnd, err := pc.unroll(start, endPos)
	if err != nil {
		return nil, err
	}
	_ = diff
	end := uint(endPos)

	if absoluteEnd < end && isOverride {
		end = absoluteEnd
	}

	_ = absoluteEnd

	printEvents("after unrolling colum "+pc.column.name+" of sketch "+pc.column.sketch.Name, evts)

	var es *eventStream

	es = newEventStream(start, end, true, evts...)

	es.isOverride = isOverride
	return es, nil
}

func (c *pattern) unroll(start uint, until uint) (evt []*items.Event, diff uint, end uint, err error) {
	cc := c.pattern

	var sk *Sketch
	var colname string
	sk, colname, err = findPattern(c.column.sketch, c.column.name, cc.Name)
	if err != nil {
		return
	}
	tr, _ := c.column.sketch.Score.GetTrack(colname)
	var patt = sk.newCol(tr, colname)
	var pcc = patt.newPattern(cc)
	//evt, diff, end, err = pcc.unrollPattern(start, until)
	evt, end, err = pcc.unrollPattern(start, until)
	return
	//default:

}

func (c *pattern) getEventStream(start uint, end uint) (*eventStream, error) {
	cc := c.pattern
	if cc.Lyrics != nil {
		es, err := c._getEventStream(start, end, false)
		if err != nil {
			return nil, err
		}
		var lc = c.column.newLyrics(cc.Lyrics)
		es.events, err = lc.applyLyrics(es.events)
		if err != nil {
			return nil, err
		}
		return es, nil
	}
	return c._getEventStream(start, end, false)
}

func (pc *pattern) getOverrideEventStream(start uint, endPos uint) (*eventStream, error) {
	es, err := pc._getEventStream(start, endPos, true)
	es.end = es.start + es.end
	return es, err
}
