package sketch

import (
	"fmt"
	"strings"

	"gitlab.com/gomidi/muskel/items"
)

type lyrics struct {
	column *column
	call   *items.LyricsTable //interface{} // items.Template, items.Token or items.LyricsTable
}

/*
//func (pc *call) unrollPattern(start uint, until uint) (evt []*Event, diff uint, absoluteEnd uint, err error) {
func (pc *lyrics) unrollPattern(start uint, until uint) (evt []*items.Event, absoluteEnd uint, err error) {
	p := pc.column
	tc := pc.call
	var evts []*items.Event
	evts, absoluteEnd, err = p.call(until, tc.SyncFirst, tc.Params...)
	if err != nil {
		return
	}

	for _, ev := range evts {
		ev.PosShift += tc.PosShift
	}

	printEvents(fmt.Sprintf("after call.call(until: %v, syncfirst: %v, params: %v)\n", until, tc.SyncFirst, tc.Params), evts)

	//_evt, diff, absoluteEnd = pc.modifyEvents(start, until, evts)
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
*/

/*
func (c *lyrics) modifyItem(it items.Item) (items.Item, error) {
	//cc := c.call.(dynamicAdder)
	cc := c.call.(*items.Template)
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

	case *items.Template:
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
*/

func (c *lyrics) applyLyrics(evts []*items.Event) ([]*items.Event, error) {
	cc := c.call

	idx := strings.Index(cc.Name, ".")
	if idx == -1 {
		return nil, fmt.Errorf("invalid lyrics call name %q (missing dot)", cc.Name)
	}

	if idx >= len(cc.Name)-1 {
		return nil, fmt.Errorf("invalid lyrics call name %q", cc.Name)
	}

	l, err := c.column.sketch.Score.Lyric(cc.Name, cc.Slice[0], cc.Slice[1])
	if err != nil {
		return nil, err
	}
	return applyLyrics(evts, l), nil
}

/*
func (pc *lyrics) modifyEvents(start uint, until uint, evts []*items.Event) (evt []*items.Event, end uint, err error) {
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

	evt, end = sliceEvents(pc.call.Slice, evt, projectedBarEnd)

	if pc.call.Slice[0] > 0 {
		evt = moveBySyncFirst(evt)
		evt = forwardEvents(evt, start)
	}

	return
}
*/

/*
func (pc *lyrics) _getEventStream(start uint, endPos uint, isOverride bool) (*eventStream, error) {
	evts, diff, absoluteEnd, err := pc.unroll(start, endPos)
	if err != nil {
		return nil, err
	}
	_ = diff
	end := uint(endPos)

	//fmt.Printf("absoluteEnd: %v end: %v\n", absoluteEnd, end)

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
*/

/*
func (c *lyrics) getEventStream(start uint, end uint) (*eventStream, error) {
	cc := c.call
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

func (pc *lyrics) getOverrideEventStream(start uint, endPos uint) (*eventStream, error) {
	es, err := pc._getEventStream(start, endPos, true)
	es.end = es.start + es.end
	return es, err
}
*/
