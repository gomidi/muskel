package items

type sketchToken struct {
	column Columner
	token  *Token
}

func (c *sketchToken) modifyItem(it Item) (Item, error) {
	cc := c.token
	switch v := it.(type) {
	case *Note:
		it := v.Dup().(*Note)
		if c != nil {
			it.Dynamic = AddDynamic(v.Dynamic, cc.DynamicAdd)
			if it.PosShift == 0 && cc.PosShift != 0 {
				it.PosShift = cc.PosShift
			}
		}
		return it, nil
	case *MIDINote:
		it := v.Dup().(*MIDINote)
		if c != nil {
			it.Dynamic = AddDynamic(v.Dynamic, cc.DynamicAdd)
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

func (pc *sketchToken) modifyEvents(start uint, until uint, evts []*Event) (evt []*Event, end uint, err error) {
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

	end = projectedBarEnd

	return
}

func (pc *sketchToken) _getEventStream(start uint, endPos uint, isOverride bool) (*EventStream, error) {
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

	//printEvents("after unrolling colum "+pc.column.name+" of sketch "+pc.column.sketch.Name, evts)

	var es *EventStream

	if isOverride {
		es = NewEventStream(false, start, 1, true, evts...)
	} else {
		es = NewEventStream(false, start, 1, false, evts...)
	}

	es.IsOverride = isOverride
	return es, nil
}

func (c *sketchToken) unroll(start uint, until uint) (evt []*Event, diff uint, end uint, err error) {
	// TODO allow parameters?
	var sc string
	sc, err = c.getToken()
	if err != nil {
		return
	}

	var pEvents *EventStream

	pEvents, err = c.column.ParseEvents(false, []string{sc})
	if err != nil {
		return
	}
	for _, pEvent := range pEvents.Events {
		pEvent.Position = start
		pEvent.Item, err = c.modifyItem(pEvent.Item)
		if err != nil {
			return
		}
		evt = append(evt, pEvent)
	}
	end = c.column.EndPosition()
	return
}

func (pc *sketchToken) getToken() (val string, err error) {
	return pc.column.GetToken(pc.token.Name, pc.token.Params)
}

func (c *sketchToken) getEventStream(start uint, end uint) (*EventStream, error) {
	//cc := c.call
	return c._getEventStream(start, end, false)
}

func (pc *sketchToken) getOverrideEventStream(start uint, endPos uint) (*EventStream, error) {
	es, err := pc._getEventStream(start, endPos, true)
	es.End = es.Start + es.End
	return es, err
}
