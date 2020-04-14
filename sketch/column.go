package sketch

import (
	"fmt"
	"strings"

	"gitlab.com/gomidi/muskel/items"
	"gitlab.com/gomidi/muskel/track"
)

type column struct {
	name   string
	sketch *Sketch
	track  *track.Track
}

func (c *column) GetEventsInBars(start, end int, evts []*items.Event) []*items.Event {
	return c.sketch.getEventsInBars(start, end, evts)
}

func (c *column) GetBarIdxOf(pos uint) int {
	return c.sketch.getBarIdxOf(pos)
}

func (c *column) GetBar(idx int) *items.Bar {
	return c.sketch.Bars[idx]
}

func (c *column) Lyric(name string, from, to int) ([]string, error) {
	return c.sketch.Score.Lyric(name, from, to)
}

func (c *column) findPattern(pattern string) (sk *Sketch, colname string, err error) {
	return findPattern(c.sketch, c.name, pattern)
}

func (c *column) EndPosition() uint {
	return c.sketch.projectedBarEnd
}

func (c *column) GetPart(partname string) (part [2]uint, has bool) {
	part, has = c.sketch.Parts[partname]
	return
}

func (cl *column) ModifyToken(tk *items.Token) (items.Item, error) {
	posEv, err := tk.GetEventStream(cl, 0, 1)
	if err != nil {
		return nil, err
	}
	it := posEv.Events[0].Item
	return it, nil
}

func (cl *column) ApplyLyricsTable(lt *items.LyricsTable, evts []*items.Event) ([]*items.Event, error) {
	return lt.ApplyLyrics(cl, evts)
}

func (c *column) ParseEvents(syncfirst bool, data []string) (es *items.EventStream, err error) {
	return c.sketch.parseEvents(syncfirst, data, c.sketch.projectedBarEnd)
}

func (c *column) GetToken(origName string, params []string) (val string, err error) {
	var table, colname string
	idx := strings.Index(origName[1:], ".")

	switch {
	case idx < 0:
		table = origName
	case idx == 0 || idx == 1:
		err = fmt.Errorf("invalid token %q", origName)
		return
	default:
		table, colname = origName[:idx+1], origName[idx+2:]
	}

	var token = table
	if colname != "" {
		if colname[len(colname)-1] == '.' {
			if c.track == nil {
				colname += c.name
			} else {
				colname += c.track.Name
			}
		}
		token += "." + colname
	}

	val, err = c.sketch.Score.GetToken(token)
	val = items.ReplaceParams(val, params)
	return

}

func (c *column) UnrollPattern(start uint, until uint, patt *items.Pattern) (evt []*items.Event, diff uint, end uint, err error) {
	var sk *Sketch
	var colname string
	sk, colname, err = c.findPattern(patt.Name)
	if err != nil {
		return
	}
	tr, _ := c.sketch.Score.GetTrack(colname)
	evt, end, err = patt.Unroll(sk.newCol(tr, colname), start, until)
	items.PrintEvents(fmt.Sprintf("column.UnrollPattern start: %v until: %v sk.Name: %q.%q", start, until, sk.Name, colname), evt)
	return
}

func (p *column) unrollItem(evts []*items.Event, ev, nextEv *items.Event, until, endPos uint, params []string) (mixed []*items.EventStream, err error) {
	var (
		es    items.UnrollGetter
		start = ev.Position
	)

	/*
		if items.DEBUG {
			fmt.Printf("unrollItem(%s)\n", ev)
		}
	*/

	switch it := ev.Item.(type) {
	case *items.PipedPatternCommands:
		es = &pipedPatternCommandES{item: it, params: params}
	case *items.Command:
		es = &commandES{item: it, params: params}
	case *items.NTuple, *items.MultiItem:
		es = it.(items.UnrollGetter)
		endPos = until
	case *items.BarRepeater:
		es = it.NewUnrollGetter(evts, nextEv)
		start = 0
	case *items.PartRepeat:
		es = it.NewUnrollGetter(evts)
		start, endPos = 0, 0
	default:
		if ug, isUG := it.(items.UnrollGetter); isUG {
			es = ug
		} else {
			es = items.DefaultES{}
			endPos = until
		}
	}

	mixed, err = es.GetES(p, ev, start, endPos)
	//items.PrintEvents(fmt.Sprintf("GetES(%s)", ev.Item), mixed[0].Events)

	if err != nil {
		return nil, err
	}

	return
}

func (p *column) RepeatBars(repevts []*items.Event, diff uint) (evts []*items.Event, absend uint) {
	return p.sketch.repeatBars(repevts, diff)
}

func (p *column) UnrollAndMerge(es *items.EventStream, params []string) (mix *items.EventStream, err error) {
	var mixed []*items.EventStream

	end := p.sketch.projectedBarEnd

	for i, ev := range es.Events {
		if ev.Item == nil {
			continue
		}

		if ev.Item == items.End {
			end = ev.Position
			break
		}

		var until = items.FindNextPos(i, 0, es.Events)
		if until < 0 {
			until = int(es.End)
		}

		var nextEv *items.Event

		if i < len(es.Events)-1 {
			nextEv = es.Events[i+1]
		}

		_mixed, err := p.unrollItem(es.Events, ev, nextEv, uint(until), es.End, params)

		//items.PrintEvents(fmt.Sprintf("after unrollItem %s", ev.Item), _mixed[0].Events)

		if err != nil {
			return nil, err
		}

		mixed = append(mixed, _mixed...)
	}

	items.PrintEventStream("mixed", mixed...)
	return items.MergeEventStreams(mixed, end), nil
}

func (p *column) setAbsolutePositions(es *items.EventStream, params []string) (mix *items.EventStream, err error) {
	var evts []*items.Event
	for i, inEv := range es.Events {
		if inEv.Item == nil || err != nil {
			continue
		}

		if _, isInc := inEv.Item.(items.Include); isInc {
			panic(fmt.Sprintf("unsupported: includes in columns"))
		}

		pos := p.sketch.Positions[i]
		inEv.Position = p.sketch.getAbsPos(pos[0], pos[1])

		evts = append(evts, inEv)
	}

	if err != nil {
		return
	}

	es.Events = evts

	return es, err
}

func (p *column) unroll(es *items.EventStream, params []string) (mix *items.EventStream, err error) {
	es, err = p.setAbsolutePositions(es, params)
	if err != nil {
		return
	}

	items.PrintEventStream("after setAbsolutePositions", es)

	return p.UnrollAndMerge(es, params)
}

// Call parameterizes the sketch in the given column, parses and unrolls it and returns the resulting events
// as absolut positioned items. If the column does not exist, nil is returned
// endPos is the absolute end position of the piece. If 0, the last bar of the SketchTable is used
func (c *column) Call(originalEndPos uint, syncFirst bool, params ...string) (events []*items.Event, endPos uint, err error) {
	s := c.sketch
	endPos = originalEndPos
	if c.name == "" {
		for i, col := range s.colOrder {
			if i == 0 || col[0] == '!' {
				c.name = col
			}
		}
	}

	col, has := s.Columns[c.name]
	if !has {
		err = fmt.Errorf("has no column: %q\n", col)
		return
	}

	if endPos == 0 {
		endPos = s.projectedBarEnd
	}

	data := c.sketch.injectParams(col, params)

	//var loop uint
	var es *items.EventStream

	es, err = s.parseEvents(syncFirst, data, endPos)

	if err != nil {
		return
	}

	//	items.PrintEvents("after parse", es.Events)

	// _ = loop

	es, err = c.unroll(es, params) // replaces template calls and imports
	if err != nil {
		return
	}

	items.PrintEventStream("after unroll", es)

	if syncFirst {
		es.SetStart(true, 0)
		es.Normalize()
		//items.
		//es.Events =
		//items.MoveBySyncFirst(es)
		items.PrintEventStream("after syncFirst", es)
	}

	es.Events = c.unrollIncludedBars(es.Events)

	items.PrintEventStream("after unrollIncludedBars", es)

	es.Events, err = c.unrollPartRepetitionsOfBars(es.Events, endPos)

	items.PrintEventStream("after unrollPartRepetitionsOfBars", es)
	return es.Events, es.Length32ths(), err
}

func (p *column) unrollIncludedBars(evts []*items.Event) []*items.Event {
	s := p.sketch

	var res []*items.Event
	var lastBarEnd uint

	for _, bar := range s.Bars {

		if bar.Include != nil {
			lastBarEnd = bar.Include.Length32ths
			ets, err := s.includeCol(bar.Position, p.name, *bar.Include)
			if err != nil {
				continue
			}
			res = append(res, ets...)
			continue
		}

		endPos := bar.Position + uint(bar.Length32th())
		lastBarEnd = endPos
		if endPos > s.projectedBarEnd {
			endPos = s.projectedBarEnd
		}

		ets := items.GetEventsInPosRange(bar.Position, endPos, evts)
		res = append(res, ets...)
	}

	ets := items.GetEventsInPosRange(lastBarEnd, s.projectedBarEnd, evts)
	res = append(res, ets...)
	return res
}

func (p *column) unrollPartRepetitions(evts []*items.Event) (res []*items.Event, err error) {
	s := p.sketch

	var skipPos uint

	var mixed []*items.EventStream

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

				startPos := part[0]
				endPos := part[1]
				diff := (endPos - startPos) //- uint(bar.Length32th())
				if endPos > s.projectedBarEnd {
					endPos = s.projectedBarEnd
				}

				partEvents := items.GetEventsInPosRange(startPos, endPos, evts)

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

				mixed = append(mixed, items.NewEventStream(false, startPos, endPos, true, nevts...))

				// overrides
				otherEvents := items.GetEventsInPosRange(endPos, endPos+diff, evts)

				for _, otherE := range otherEvents {
					switch otherE.Item.(type) {
					case *items.PartRepeat:
						// ignore
					case *items.Override:
						fmt.Printf("got override: %v\n", otherE)
						oes := items.NewEventStream(false, otherE.Position, otherE.Position+1, false, otherE)
						oes.IsOverride = true
						mixed = append(mixed, oes)
					default:
						nes := items.NewEventStream(false, otherE.Position, otherE.Position+1, false, otherE)
						mixed = append(mixed, nes)
					}
				}

				skipPos = endPos + diff
			} else {
				if evt.Position >= skipPos {
					nes := items.NewEventStream(false, evt.Position, evt.Position+1, false, evt)
					mixed = append(mixed, nes)
				}
			}
		}
	}

	unrolled := items.MergeEventStreams(mixed, s.projectedBarEnd)

	return unrolled.Events, nil
}

func (p *column) unrollPartRepetitionsOfBars(evts []*items.Event, stopPos uint) ([]*items.Event, error) {

	s := p.sketch

	var res []*items.Event
	var lastBarEnd uint

	for i, bar := range s.Bars {
		_ = i

		lastBarEnd = bar.Position + uint(bar.Length32th())

		if bar.JumpTo != "" {
			part, has := s.Parts[bar.JumpTo]
			if !has {
				return nil, fmt.Errorf("can't jump to part %q from bar %v: part is not defined", bar.JumpTo, bar.No+1)
			}
			startPos := part[0]
			endPos := part[1]
			diff := (endPos - startPos) - uint(bar.Length32th())
			_ = diff

			partEvents := items.GetEventsInPosRange(startPos, endPos, evts)

			var nevts []*items.Event

			for _, ev := range partEvents {
				nue := ev.Dup()
				nue.Position += bar.Position
				nue.Position -= startPos
				nevts = append(nevts, nue)
			}

			_lastBarEnd := 2*endPos - startPos
			if lastBarEnd < _lastBarEnd {
				lastBarEnd = _lastBarEnd
			}
			res = append(res, nevts...)

		} else {
			endPos := bar.Position + uint(bar.Length32th())
			if endPos > s.projectedBarEnd {
				endPos = s.projectedBarEnd
			}
			if endPos > stopPos {
				endPos = stopPos
			}
			ets := items.GetEventsInPosRange(bar.Position, endPos, evts)
			res = append(res, ets...)
		}

	}

	ets := items.GetEventsInPosRange(lastBarEnd, stopPos, evts)
	res = append(res, ets...)

	// TODO check if it is correct
	return res, nil
}
