package items

type SketchToken struct {
	column Columner
	token  *Token
}

func NewSketchToken(column Columner, token *Token) *SketchToken {
	return &SketchToken{
		column: column,
		token:  token,
	}
}

func (c *SketchToken) modifyItem(it Item) (Item, error) {
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

func (pc *SketchToken) modifyEvents(start uint, until uint, evts []*Event) (evt []*Event, end uint, err error) {
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

	evt, end = SliceEvents(pc.token.Slice, evt, projectedBarEnd)

	if pc.token.Slice[0] > 0 {
		evt = MoveBySyncFirst(evt)
		evt = ForwardEvents(evt, start)
	}

	return
}

func (pc *SketchToken) _getEventStream(start uint, endPos uint, isOverride bool) (*EventStream, error) {
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
		es = NewEventStream(start, 1, true, evts...)
	} else {
		es = NewEventStream(start, 1, false, evts...)
	}

	es.IsOverride = isOverride
	return es, nil
}

func (c *SketchToken) unroll(start uint, until uint) (evt []*Event, diff uint, end uint, err error) {
	// TODO allow parameters?
	var sc string
	sc, err = c.getToken()
	if err != nil {
		return
	}

	var pEvents []*Event

	pEvents, _, err = c.column.ParseEvents([]string{sc})
	if err != nil {
		return
	}
	for _, pEvent := range pEvents {
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

func (pc *SketchToken) getToken() (val string, err error) {
	return pc.column.GetToken(pc.token.Name, pc.token.Params)
}

func (c *SketchToken) GetEventStream(start uint, end uint) (*EventStream, error) {
	//cc := c.call
	return c._getEventStream(start, end, false)
}

func (pc *SketchToken) GetOverrideEventStream(start uint, endPos uint) (*EventStream, error) {
	es, err := pc._getEventStream(start, endPos, true)
	es.End = es.Start + es.End
	return es, err
}