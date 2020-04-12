package sketch

import (
	"fmt"
	"strings"

	"gitlab.com/gomidi/muskel/items"
	"gitlab.com/gomidi/muskel/patterncommands"
	"gitlab.com/gomidi/muskel/track"
)

type column struct {
	name   string
	sketch *Sketch
	track  *track.Track
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

func (c *column) ParseEvents(data []string) (evts []*items.Event, absEnd uint, err error) {
	return c.sketch.parseEvents(data, c.sketch.projectedBarEnd)
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
	return
}

func (p *column) _unroll(evts []*items.Event, originalEndPos uint, params []string) (unrolled []*items.Event, endPos uint, err error) {

	endPos = originalEndPos

	s := p.sketch

	var forward uint

	var mixed []*items.EventStream

	for i, ev := range evts {
		if ev.Item == nil {
			continue
		}

		if ev.Item == items.End {
			endPos = ev.Position
			break
		}

		var posEv *items.EventStream

		switch v := ev.Item.(type) {
		case *items.PipedPatternCommands:
			helper := &patterncmdHelper{
				column: p,
				params: params,
				inPipe: true,
			}

			for _, cmd := range v.Cmds {
				if fn, hasCmd := patterncommands.Commands[cmd.Name]; hasCmd {
					helper.cmdName = cmd.Name
					_evts, err := fn(cmd.Params, helper)

					if err != nil {
						return nil, endPos, err
					}

					helper.pipeEvents = _evts
				} else {
					return nil, endPos, fmt.Errorf("unknown command %q", cmd.Name)
				}
			}

			posEv, err = items.PipedEventStream(p, forward+ev.Position, endPos, helper.pipeEvents)
			if err != nil {
				return nil, endPos, err
			}

		case *items.Command:
			if fn, hasCmd := patterncommands.Commands[v.Name]; hasCmd {
				helper := &patterncmdHelper{
					cmdName: v.Name,
					column:  p,
					params:  params,
				}

				_evts, err := fn(v.Params, helper)

				if err != nil {
					return nil, endPos, err
				}

				if len(_evts) == 0 {
					continue
				}
				posEv, err = items.PipedEventStream(p, forward+ev.Position, endPos, _evts)
				if err != nil {
					return nil, endPos, err
				}
			} else {
				return nil, endPos, fmt.Errorf("unknown command %q", v.Name)
			}

		case *items.NTuple:
			// TODO look inside each item and replace random things and templatecalls if there
			var until = items.FindNextPos(i, int(forward), evts)
			if until < 0 {
				until = int(endPos)
			}

			newItm, err := items.ReplaceNtupleTokens(p, v)
			if err != nil {
				return nil, endPos, err
			}

			var nev = ev.Dup()
			nev.Position = ev.Position + forward
			nev.PosShift = 0
			nev.Item = newItm
			posEv = items.NewEventStream(nev.Position, uint(until), false, nev)
		case *items.BarRepeater:
			stopPos := endPos
			if i < len(evts)-1 {
				stBarIdx := s.getBarIdxOf(evts[i+1].Position)
				stBarPos := s.Bars[stBarIdx].Position
				if stBarPos < stopPos {
					stopPos = stBarPos
				}
			}
			bidx := s.getBarIdxOf(ev.Position)
			startBar := bidx - v.LastN
			repevts := s.getEventsInBars(startBar, bidx-1, evts)
			if len(repevts) == 0 {
				continue
			}

			currentPos := s.Bars[bidx].Position

			diff := currentPos - s.Bars[startBar].Position
			var ignoreEndPos uint
			repevts, ignoreEndPos, err = p._unroll(repevts, stopPos, nil)
			_ = ignoreEndPos
			if err != nil {
				return nil, endPos, err
			}

			var res []*items.Event

			origDiff := diff
			endOfStream := origDiff

			var all []*items.Event

			if len(repevts) == 0 {
				continue
			}

			res, currentPos = s.repeatBars(repevts, diff)
			if len(res) == 0 {
				continue
			}

			all = append(all, res...)

			var repetitions uint = 1

			if !v.OnlyOnce {
			barRepeatLoop:
				for {
					if currentPos >= stopPos {
						break barRepeatLoop
					}
					repetitions++
					endOfStream += origDiff
					res, currentPos = s.repeatBars(repevts, diff*repetitions)
					all = append(all, res...)
				}
			}

			mixed = append(mixed, items.NewEventStream(s.Bars[bidx].Position, currentPos, true, all...))
			posEv = nil

		case *items.PartRepeat:
			partname := v.Part
			part, has := s.Parts[partname]
			if !has {
				return nil, endPos, fmt.Errorf("can't repeat part %q at %v: part is not defined", partname, ev.Position)
			}

			startPos := part[0]
			_endPos := part[1]
			diff := (_endPos - startPos) //- uint(bar.Length32th())
			if _endPos > s.projectedBarEnd {
				_endPos = s.projectedBarEnd
			}

			partEvents := items.GetEventsInPosRange(startPos, _endPos, evts)

			partEvents, _endPos, err = p._unroll(partEvents, _endPos, nil)
			if err != nil {
				return nil, endPos, err
			}

			for r := 0; r < int(v.Repeat); r++ {
				var nevts []*items.Event

				for _, _ev := range partEvents {
					nue := _ev.Dup()
					nue.Position += ev.Position + uint(r)*diff
					if v.SyncFirst {
						nue.Position -= startPos
					}
					nevts = append(nevts, nue)
				}

				totalStart := startPos + ev.Position + (uint(r) * diff)

				mixed = append(mixed, items.NewEventStream(totalStart, totalStart+diff, true, nevts...))
			}

			posEv = nil

		case *items.MultiItem:
			// TODO look inside each item and replace random things and templatecalls if there
			var until = items.FindNextPos(i, int(forward), evts)
			if until < 0 {
				until = int(endPos)
			}

			var newItsm = make([]*items.Event, len(v.Events))
			for itidx, itm := range v.Events {
				switch vv := itm.Item.(type) {
				case *items.Pattern:
					posEv, err = vv.GetEventStream(p, forward+ev.Position, endPos)
					if err != nil {
						return nil, endPos, err
					}
					newItsm[itidx] = posEv.Events[0]
				case *items.Token:
					posEv, err = vv.GetEventStream(p, forward+ev.Position, endPos)
					if err != nil {
						return nil, endPos, err
					}
					newItsm[itidx] = posEv.Events[0]
				default:
					newItsm[itidx] = itm
				}
			}
			v.Events = newItsm
			var nev = ev.Dup()
			nev.Position = ev.Position + forward
			nev.Item = v.Dup()
			posEv = items.NewEventStream(nev.Position, uint(until), false, nev)

		case *items.Pattern:
			// TODO prevent endless loops from templates calling each other like col1 -> col2 -> col1 by keeping a stack of template calls
			// and checking them for duplicates (the stack may as well be a map[string]bool; makes is easier; we need the complete names in there
			posEv, err = v.GetEventStream(p, forward+ev.Position, endPos)
			if err != nil {
				return nil, endPos, err
			}

		case *items.Token:
			// TODO prevent endless loops from templates calling each other like col1 -> col2 -> col1 by keeping a stack of template calls
			// and checking them for duplicates (the stack may as well be a map[string]bool; makes is easier; we need the complete names in there
			posEv, err = v.GetEventStream(p, forward+ev.Position, endPos)
			if err != nil {
				return nil, endPos, err
			}

		case *items.Override:
			switch vv := v.Item.(type) {
			case *items.Pattern:
				posEv, err = vv.GetOverrideEventStream(p, forward+ev.Position, endPos)

				if err != nil {
					return nil, endPos, err
				}
			case *items.Token:
				tk := v.Item.(*items.Token)
				posEv, err = tk.GetOverrideEventStream(p, forward+ev.Position, forward+ev.Position+1)
				if err != nil {
					return nil, endPos, err
				}
			case *items.Scale:
				return nil, endPos, fmt.Errorf("item type not allowed in override: %#v", vv)
			default:
				nev := ev.Dup()
				nev.Item = vv
				nev.Position = ev.Position + forward
				posEv = items.NewEventStream(nev.Position, nev.Position+1, true, nev)
				posEv.IsOverride = true
			}

		default:
			var until = items.FindNextPos(i, int(forward), evts)
			if until < 0 {
				until = int(endPos)
			}
			var nev = ev.Dup()
			nev.Position = ev.Position + forward
			posEv = items.NewEventStream(nev.Position, uint(until), false, nev)
		}
		if posEv != nil {
			mixed = append(mixed, posEv)
		}
	}

	if DEBUG {
		fmt.Printf(">> mixed\n")

		for _, ev := range mixed {
			fmt.Printf("at %v: %v\n", ev.Start, ev)
		}
		fmt.Printf("<< mixed\n")
	}

	unrolled = items.MergeEventStreams(mixed, s.projectedBarEnd)

	if DEBUG {
		fmt.Printf(">> after merge\n")

		for _, ev := range unrolled {
			fmt.Printf("at %v: %v\n", ev.Position, ev)
		}
		fmt.Printf("<< after merge\n")
	}

	return
}

func (p *column) unroll(inputEvents []*items.Event, originalEndPos uint, params []string) (unrolled []*items.Event, endPos uint, err error) {
	s := p.sketch
	//endPos = originalEndPos

	var evts []*items.Event
	for i, inEv := range inputEvents {
		if inEv.Item == nil || err != nil {
			continue
		}

		if inc, isInc := inEv.Item.(items.Include); isInc {
			_ = inc
			panic(fmt.Sprintf("unsupported: includes in columns"))
		}

		pos := s.Positions[i]
		inEv.Position = s.getAbsPos(pos[0], pos[1])

		evts = append(evts, inEv)
	}

	if err != nil {
		return
	}

	return p._unroll(evts, originalEndPos, params)
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

	var loop uint

	events, loop, err = s.parseEvents(data, endPos)

	if err != nil {
		return
	}

	_ = loop

	events, endPos, err = c.unroll(events, endPos, params) // replaces template calls and imports
	if err != nil {
		return
	}

	if syncFirst {
		events = items.MoveBySyncFirst(events)
	}

	events = c.unrollIncludedBars(events)
	events, err = c.unrollPartRepetitionsOfBars(events, endPos)
	return
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

				mixed = append(mixed, items.NewEventStream(startPos, endPos, true, nevts...))

				// overrides
				otherEvents := items.GetEventsInPosRange(endPos, endPos+diff, evts)

				for _, otherE := range otherEvents {
					switch otherE.Item.(type) {
					case *items.PartRepeat:
						// ignore
					case *items.Override:
						fmt.Printf("got override: %v\n", otherE)
						oes := items.NewEventStream(otherE.Position, otherE.Position+1, false, otherE)
						oes.IsOverride = true
						mixed = append(mixed, oes)
					default:
						nes := items.NewEventStream(otherE.Position, otherE.Position+1, false, otherE)
						mixed = append(mixed, nes)
					}
				}

				skipPos = endPos + diff
			} else {
				if evt.Position >= skipPos {
					nes := items.NewEventStream(evt.Position, evt.Position+1, false, evt)
					mixed = append(mixed, nes)
				}
			}
		}
	}

	unrolled := items.MergeEventStreams(mixed, s.projectedBarEnd)

	return unrolled, nil
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
