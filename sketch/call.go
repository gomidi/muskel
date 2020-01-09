package sketch

import (
	"fmt"
	"strings"

	"gitlab.com/gomidi/muskel/items"
)

type call struct {
	column *column
	call   *items.Call
}

//func (pc *call) unrollPattern(start uint, until uint) (evt []*Event, diff uint, absoluteEnd uint, err error) {
func (pc *call) unrollPattern(start uint, until uint) (evt []*Event, absoluteEnd uint, err error) {
	p := pc.column
	tc := pc.call
	evts, err := p.call(until, tc.SyncFirst, tc.Params...)
	if err != nil {
		return
	}

	printEvents(fmt.Sprintf("after call.call(until: %v, syncfirst: %v, params: %v)\n", until, tc.SyncFirst, tc.Params), evts)

	//_evt, diff, absoluteEnd = pc.modifyEvents(start, until, evts)
	var _evt []*Event
	_evt, absoluteEnd, err = pc.modifyEvents(start, until, evts)
	if err != nil {
		return
	}

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

	var __evts []*Event

	for i := 0; i < repeat; i++ {
		if DEBUG {
			fmt.Printf("forwarding by %v\n", length*uint(i))
		}
		__evts = append(__evts, forwardEvents(_evt, length*uint(i))...)
		absoluteEnd += length * uint(i)
	}

	printEvents("after repeat", __evts)

	for _, ev := range __evts {
		if ev.Position < until {
			evt = append(evt, ev)
		}
	}

	if absoluteEnd > until {
		absoluteEnd = until
	}

	printEvents("after unrollPattern", evt)

	return
}

func (c *call) modifyItem(it items.Item) (items.Item, error) {
	cc := c.call
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

	case *items.Call:
		if v.IsToken() && c != nil {
			pc := c.column.newCall(v)
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

	case *items.NTuple:
		if c != nil {
			it := v.Dup().(*items.NTuple)
			it, err := c.column.replaceNtupleTokens(v)
			if it.PosShift == 0 && cc.PosShift != 0 {
				it.PosShift = cc.PosShift
			}

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

func (c *call) applyLyrics(evts []*Event) ([]*Event, error) {
	cc := c.call

	if !cc.IsLyrics() {
		return nil, fmt.Errorf("%q is not a lyrics call", cc.Name)
	}

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

//func (pc *call) modifyEvents(start uint, until uint, evts []*Event) (evt []*Event, diff uint, end uint) {
func (pc *call) modifyEvents(start uint, until uint, evts []*Event) (evt []*Event, end uint, err error) {
	//p := pc.column
	projectedBarEnd := pc.column.sketch.projectedBarEnd
	for _, ev := range evts {
		pos := start + ev.Position
		if until > 0 && pos >= until {
			//diff = 0
			break
		}
		//diff += ev.Position
		nuEv := ev.Dup()
		nuEv.Position = pos
		nuEv.Item, err = pc.modifyItem(ev.Item)
		if err != nil {
			return
		}
		//nuEv := modEventByTemplateCall(ev, pos, tc)
		evt = append(evt, nuEv)
	}

	//fmt.Printf("slice: %v\n", pc.call.Slice)
	evt, end = sliceEvents(pc.call.Slice, evt, projectedBarEnd)

	if pc.call.Slice[0] > 0 {
		evt = moveBySyncFirst(evt)
		evt = forwardEvents(evt, start)
	}

	//fmt.Printf("sliced events: %v\n", evt)

	/*
		Each Sketch Column may have the repetition signs including the "endless repetitions" .n. and ... somewhere in the middle or
		at the end. an "endless repetition" is tracked in the column. When Call is called, the next breaking event or if there is
		none, the last position from the outermost table is passed and a callback function is called that may overwrite the events.
	*/

	// TODO rewrite replaceEvents (or remove it) since we convert back and forth between Events and TemplatePositionedEvents
	// evts = replaceEvents(absPos, tc, evts)
	/*
		Houston, we have a problem here: Replacements imply that we do reparse the replaced strings after doing
		the replacement. therefor we need to reparse the items after that replacement / need the strings inside the
		TemplatePositionedEvents. It would be the best to have all this slice and replace and repeat junk here in the sketch
		and let TemplateCall just be simply the parsed call definition

		MAYBE instead of replacements we allow a special character to be placed before an item to mean "override at the current position
		without interrupting the underlying template or (bar)repetitions". This could be an alternative way to have fills etc. esp. when used with templates
		"pausing" underlying templates for the duration of the overriding template.

		It also makes the syntax more simple.

		The special precharacter could be / since it is free if we remove replaceEvents.
	*/
	return
}

func printEvents(message string, evts []*Event) {
	if DEBUG {
		fmt.Printf("##> Events %s\n", message)
		for _, ev := range evts {
			fmt.Printf("[%v] %v ", ev.Position, ev.String())
		}
		fmt.Printf("\n##< Events %s\n", message)
	}
}

func (pc *call) _getEventStream(start uint, endPos uint, isOverride bool) (*eventStream, error) {
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

	if pc.call.IsToken() {
		if isOverride {
			es = newEventStream(start, 1, true, evts...)
		} else {
			es = newEventStream(start, 1, false, evts...)
		}
	} else {
		es = newEventStream(start, end, true, evts...)

	}

	es.isOverride = isOverride
	return es, nil
}

func (c *call) unroll(start uint, until uint) (evt []*Event, diff uint, end uint, err error) {
	cc := c.call

	switch cc.Name[0] {
	case '@':
		err = fmt.Errorf("can't unroll lyrics %q", cc.Name)
		return
	case '=':
		var sk *Sketch
		var colname string
		sk, colname, err = findPattern(c.column.sketch, c.column.name, cc.Name)
		if err != nil {
			return
		}
		var patt = sk.newCol(colname)
		var pcc = patt.newCall(cc)
		//evt, diff, end, err = pcc.unrollPattern(start, until)
		evt, end, err = pcc.unrollPattern(start, until)
		return
	default:
		// TODO allow parameters?
		var sc string
		sc, err = c.getToken()
		if err != nil {
			return
		}

		items, loop, err2 := c.column.sketch.parseItems([]string{sc})
		if err2 != nil {
			err = err2
			return
		}
		_ = loop
		for _, it := range items {
			ev := &Event{}
			ev.Position = start
			ev.Item, err = c.modifyItem(it)
			if err != nil {
				return
			}
			evt = append(evt, ev)
		}
		end = c.column.sketch.projectedBarEnd
		return
	}
}

func (pc *call) getToken() (val string, err error) {
	var table, colname string
	idx := strings.Index(pc.call.Name, ".")

	switch {
	case idx < 0:
		table = pc.call.Name
	case idx == 0 || idx == 1:
		err = fmt.Errorf("invalid token %q", pc.call.Name)
		return
	default:
		table, colname = pc.call.Name[:idx], pc.call.Name[idx+1:]
	}

	var token = table
	if colname != "" {
		token += "." + colname
	}

	val, err = pc.column.sketch.Score.GetToken(token)
	val = replaceParams(val, pc.call.Params)
	return
}

func (c *call) getEventStream(start uint, end uint) (*eventStream, error) {
	cc := c.call
	if cc.Lyrics != nil {
		es, err := c._getEventStream(start, end, false)
		if err != nil {
			return nil, err
		}
		var lc = c.column.newCall(cc.Lyrics)
		es.events, err = lc.applyLyrics(es.events)
		if err != nil {
			return nil, err
		}
		return es, nil
	}
	return c._getEventStream(start, end, false)
}

func (pc *call) getOverrideEventStream(start uint, endPos uint) (*eventStream, error) {
	es, err := pc._getEventStream(start, endPos, true)
	es.end = es.start + es.end
	return es, err
}
