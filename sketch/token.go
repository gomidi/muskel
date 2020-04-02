package sketch

import (
	"fmt"
	"strings"

	"gitlab.com/gomidi/muskel/items"
)

type token struct {
	column *column
	token  *items.Token
}

func (c *token) modifyItem(it items.Item) (items.Item, error) {
	cc := c.token
	if DEBUG {
		fmt.Printf("modify item %T %v\n", it, it)
	}
	switch v := it.(type) {
	case *items.Note:
		it := v.Dup().(*items.Note)
		if c != nil {
			it.Dynamic = items.AddDynamic(v.Dynamic, cc.DynamicAdd)
			if it.PosShift == 0 && cc.PosShift != 0 {
				it.PosShift = cc.PosShift
			}
		}
		return it, nil
	case *items.MIDINote:
		it := v.Dup().(*items.MIDINote)
		if c != nil {
			it.Dynamic = items.AddDynamic(v.Dynamic, cc.DynamicAdd)
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
			it := posEv.Events[0].Item
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

func (pc *token) modifyEvents(start uint, until uint, evts []*items.Event) (evt []*items.Event, end uint, err error) {
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

	evt, end = items.SliceEvents(pc.token.Slice, evt, projectedBarEnd)

	if pc.token.Slice[0] > 0 {
		evt = items.MoveBySyncFirst(evt)
		evt = items.ForwardEvents(evt, start)
	}

	return
}

func (pc *token) _getEventStream(start uint, endPos uint, isOverride bool) (*items.EventStream, error) {
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

	var es *items.EventStream

	if isOverride {
		es = items.NewEventStream(start, 1, true, evts...)
	} else {
		es = items.NewEventStream(start, 1, false, evts...)
	}

	es.IsOverride = isOverride
	return es, nil
}

func (c *token) unroll(start uint, until uint) (evt []*items.Event, diff uint, end uint, err error) {
	// TODO allow parameters?
	var sc string
	sc, err = c.getToken()
	if err != nil {
		return
	}

	var pEvents []*items.Event
	var loop uint

	pEvents, loop, err = c.column.sketch.parseEvents([]string{sc}, c.column.sketch.projectedBarEnd)
	if err != nil {
		return
	}
	_ = loop
	for _, pEvent := range pEvents {
		pEvent.Position = start
		pEvent.Item, err = c.modifyItem(pEvent.Item)
		if err != nil {
			return
		}
		evt = append(evt, pEvent)
	}
	end = c.column.sketch.projectedBarEnd
	return
}

func (pc *token) getToken() (val string, err error) {
	var table, colname string
	idx := strings.Index(pc.token.Name[1:], ".")

	switch {
	case idx < 0:
		table = pc.token.Name
	case idx == 0 || idx == 1:
		err = fmt.Errorf("invalid token %q", pc.token.Name)
		return
	default:
		table, colname = pc.token.Name[:idx+1], pc.token.Name[idx+2:]
	}

	var token = table
	if colname != "" {
		if colname[len(colname)-1] == '.' {
			if pc.column.track == nil {
				colname += pc.column.name
			} else {
				colname += pc.column.track.Name
			}
		}
		token += "." + colname
	}

	val, err = pc.column.sketch.Score.GetToken(token)
	val = items.ReplaceParams(val, pc.token.Params)
	return
}

func (c *token) getEventStream(start uint, end uint) (*items.EventStream, error) {
	//cc := c.call
	return c._getEventStream(start, end, false)
}

func (pc *token) getOverrideEventStream(start uint, endPos uint) (*items.EventStream, error) {
	es, err := pc._getEventStream(start, endPos, true)
	es.End = es.Start + es.End
	return es, err
}
