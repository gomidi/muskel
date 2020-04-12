package items

import (
	"fmt"
)

type UnrollGetter interface {
	GetES(c Columner, ev *Event, startPos, endPos uint) ([]*EventStream, error)
}

type Columner interface {
	Call(until uint, syncFirst bool, params ...string) (evts []*Event, absoluteEnd uint, err error)
	EndPosition() uint
	UnrollPattern(start uint, until uint, cc *Pattern) (evt []*Event, diff uint, end uint, err error)
	ModifyToken(tk *Token) (Item, error)
	ApplyLyricsTable(lt *LyricsTable, evts []*Event) ([]*Event, error)
	ParseEvents(data []string) (evts []*Event, absEnd uint, err error)
	GetToken(origName string, params []string) (val string, err error)

	GetBarIdxOf(pos uint) int
	GetPart(partname string) (part [2]uint, has bool)
	GetBar(idx int) *Bar
	GetEventsInBars(start, end int, evts []*Event) []*Event
	UnUnroll(evts []*Event, originalEndPos uint, params []string) (unrolled []*Event, endPos uint, err error)
	RepeatBars(repevts []*Event, diff uint) (evts []*Event, absend uint)
}

type BarRepeaterES struct {
	Item   *BarRepeater
	Evts   []*Event
	NextEv *Event
}

var _ UnrollGetter = &BarRepeaterES{}

func (p *BarRepeaterES) GetES(c Columner, ev *Event, start, endPos uint) (x []*EventStream, err error) {
	stopPos := endPos
	if p.NextEv != nil {
		stBarIdx := c.GetBarIdxOf(p.NextEv.Position)
		stBarPos := c.GetBar(stBarIdx).Position
		if stBarPos < stopPos {
			stopPos = stBarPos
		}
	}
	bidx := c.GetBarIdxOf(ev.Position)
	startBar := bidx - p.Item.LastN
	repevts := c.GetEventsInBars(startBar, bidx-1, p.Evts)
	if len(repevts) == 0 {
		return nil, nil
	}

	currentPos := c.GetBar(bidx).Position

	diff := currentPos - c.GetBar(startBar).Position
	var ignoreEndPos uint
	repevts, ignoreEndPos, err = c.UnUnroll(repevts, stopPos, nil)
	_ = ignoreEndPos
	if err != nil {
		return nil, err
	}

	var res []*Event

	origDiff := diff
	endOfStream := origDiff

	var all []*Event

	if len(repevts) == 0 {
		return nil, nil
	}

	res, currentPos = c.RepeatBars(repevts, diff)
	if len(res) == 0 {
		return nil, nil
	}

	all = append(all, res...)

	var repetitions uint = 1

	if !p.Item.OnlyOnce {
		for {
			if currentPos >= stopPos {
				break
			}
			repetitions++
			endOfStream += origDiff
			res, currentPos = c.RepeatBars(repevts, diff*repetitions)
			all = append(all, res...)
		}
	}

	return []*EventStream{NewEventStream(c.GetBar(bidx).Position, currentPos, true, all...)}, nil
}

type PartRepeatES struct {
	Item *PartRepeat
	Evts []*Event
}

var _ UnrollGetter = &PartRepeatES{}

func (c *PartRepeatES) GetES(p Columner, ev *Event, start, endPos uint) (mixed []*EventStream, err error) {
	partname := c.Item.Part
	part, has := p.GetPart(partname)
	if !has {
		return nil, fmt.Errorf("can't repeat part %q at %v: part is not defined", partname, ev.Position)
	}

	startPos := part[0]
	_endPos := part[1]
	diff := (_endPos - startPos) //- uint(bar.Length32th())
	if _endPos > p.EndPosition() {
		_endPos = p.EndPosition()
	}

	partEvents := GetEventsInPosRange(startPos, _endPos, c.Evts)

	partEvents, _endPos, err = p.UnUnroll(partEvents, _endPos, nil)
	if err != nil {
		return nil, err
	}

	for r := 0; r < int(c.Item.Repeat); r++ {
		var nevts []*Event

		for _, _ev := range partEvents {
			nue := _ev.Dup()
			nue.Position += ev.Position + uint(r)*diff
			if c.Item.SyncFirst {
				nue.Position -= startPos
			}
			nevts = append(nevts, nue)
		}

		totalStart := startPos + ev.Position + (uint(r) * diff)

		mixed = append(mixed, NewEventStream(totalStart, totalStart+diff, true, nevts...))
	}

	return mixed, nil
}

type DefaultES struct{}

var _ UnrollGetter = DefaultES{}

func (c DefaultES) GetES(p Columner, ev *Event, start, endPos uint) (mixed []*EventStream, err error) {
	var nev = ev.Dup()
	nev.Position = start
	return []*EventStream{NewEventStream(nev.Position, endPos, false, nev)}, nil
}
