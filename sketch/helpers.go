package sketch

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"gitlab.com/gomidi/muskel/items"
)

func parseEventsLineItem(posItem string) (pos, value string, err error) {
	//fmt.Printf("parse eventslineItem: %v\n", posItem)
	var positionBf strings.Builder
	var bf strings.Builder
	rd := bufio.NewReader(strings.NewReader(strings.TrimSpace(posItem)))
	positionFound := false
	var rn rune

	for err == nil {
		rn, _, err = rd.ReadRune()

		if err != nil {
			break
		}

		switch {
		case strings.IndexRune("1234567890&;:.", rn) >= 0:
			if !positionFound {
				positionBf.WriteRune(rn)
			} else {
				bf.WriteRune(rn)
			}
		default:
			positionFound = true
			bf.WriteRune(rn)
		}

	}

	if err != nil && err != io.EOF {
		return
	}

	pos = positionBf.String()
	value = bf.String()
	err = nil
	return
}

func sketchFromEventsLine(name string, patternDef string, sc Score) (*Sketch, error) {
	posItems := strings.Split(patternDef, "|")
	var sk = NewSketch(name, sc)
	sk.AddColumn("pattern")
	err := sk.ParseLine([]string{"#"})
	//fmt.Printf("ParseLine: #\n")
	if err != nil {
		return nil, err
	}
	var barNo uint

	for _, posItem := range posItems {
		pos, val, err := parseEventsLineItem(posItem)

		if err != nil {
			return nil, err
		}

		var num32 uint
		_, num32, err = items.PositionTo32th(0, pos)

		if err != nil {
			return nil, err
		}

		b := num32 / 32
		rest := num32 % 32

		for b > barNo {
			//fmt.Printf("ParseLine: #\n")
			err = sk.ParseLine([]string{"#"})
			if err != nil {
				return nil, err
			}
			barNo++
		}

		//fmt.Printf("ParseLine: %s | %s\n", items.Pos32thToString(rest), val)

		err = sk.ParseLine([]string{items.Pos32thToString(rest), val, ""})
		if err != nil {
			return nil, err
		}
	}

	return sk, nil
}

func eventsFromPatternDef(name, patternDef string, sc Score, posshift int, posIn32th uint, params []string) (evts []*Event, err error) {
	replaced := replaceParams(patternDef, params)
	//fmt.Printf("eventsFromPatternDef replaced: %q\n", replaced)
	sk, err := sketchFromEventsLine(name, replaced, sc)

	if err != nil {
		return nil, err
	}

	_, _evts, err := sk.Unroll("pattern")

	if err != nil {
		return nil, err
	}

	//var p items.Call
	//var ps items.Parser
	//p.ItemParser = &ps
	// p.Params = params

	//p.ParseEventsLine(replaceParams(patternDef, params), posshift, posIn32th)

	for _, pev := range _evts {
		ev := &Event{}
		ev.Position = pev.Position
		ev.Item = pev.Item.Dup()
		evts = append(evts, ev)
	}

	if DEBUG {
		printEvents(" after eventsFromPatternDef "+patternDef, evts)
	}

	return
}

func mergeEventStreams(mixed []*eventStream, endPos uint) (evts []*Event) {
	var (
		lastStreamPos uint
		keepLooping   bool = true
		lastStream    *eventStream
	)

	lastStream = nil

	for idx, stream := range mixed {

		if keepLooping && lastStream != nil {
			if lastStream.inRange(lastStreamPos, stream.start) {
				var toMod = stream.start
				if idx+1 < len(mixed) {
					if mixed[idx+1].start <= stream.start {
						toMod = mixed[idx+1].start
					}
				}
				evts = append(evts, lastStream.getEvents(lastStreamPos, toMod)...)
			}
		}

		if stream.isTemplate && !stream.isOverride {
			keepLooping = true
			lastStream = stream
			continue
		}

		if !stream.isTemplate {
			keepLooping = false
			lastStream = nil
		}

		lastStreamPos = stream.end
		evts = append(evts, stream.events...)
	}

	if keepLooping && lastStream != nil {
		if lastStream.inRange(lastStreamPos, endPos) {
			evts = append(evts, lastStream.getEvents(lastStreamPos, endPos)...)
		}
	}
	return
}

func getEventsInPosRange(from, to uint, evts []*Event) (res []*Event) {
	if DEBUG {
		fmt.Printf("getEventsInPosRange(from: %v, to: %v)\n", from, to)
	}
	for _, ev := range evts {
		//fmt.Printf("event %v at %v: %v\n", i, ev.Position, ev)
		if ev.Position >= to {
			break
		}
		if ev.Position >= from {
			res = append(res, ev)
		}
	}
	//fmt.Printf("getEventsInPosRange from: %v to: %v => %v\n", from, to, res)

	return
}

func getBarsInPosRange(from, to uint, bars []*Bar) (res []*Bar) {
	for _, b := range bars {
		//fmt.Printf("event %v at %v: %v\n", i, ev.Position, ev)
		if b.Position >= to {
			break
		}
		if b.Position >= from {
			res = append(res, b)
		}
	}
	//fmt.Printf("getEventsInPosRange from: %v to: %v => %v\n", from, to, res)

	return
}

//func replaceParams(s string, params []interface{}) string {
func replaceParams(s string, params []string) string {
	if DEBUG {
		fmt.Printf("should replace params %#v in %q\n", params, s)
	}
	if len(params) == 0 {
		return s
	}
	paramslen := len(params)
	return items.TemplateReg.ReplaceAllStringFunc(s, func(in string) (out string) {
		i, err := strconv.Atoi(in[1:])
		if err != nil {
			// should be unreachable, if regexp is correct
			panic(in)
		}
		return params[(i-1)%paramslen]
	})
}

func isLoop(it items.Item) int {
	if it == nil {
		return -1
	}

	switch v := it.(type) {
	case *items.Override:
		return isLoop(v.Item)
	case *items.BarRepeater:
		if v.OnlyOnce {
			return 0
		}
		return v.LastN
	default:
		return 0
	}
}

func findEventThatIsNotRest(start int, evts []*Event) (found int) {
	if start >= len(evts) {
		return -1
	}

	if evts[start].Item != items.Rest {
		return start
	}

	return findEventThatIsNotRest(start+1, evts)
}

// sliceEvents slices the events according to the template call slice definition
func sliceEvents(slice [2]int, all []*Event, projectedBarEnd uint) (evs []*Event, absoluteEnd uint) {
	first, last := slice[0], slice[1]
	absoluteEnd = projectedBarEnd
	if first < 0 {
		return all, projectedBarEnd
	}

	var collecting bool

	for i, ev := range all {
		if last > 0 && i >= last {
			absoluteEnd = ev.Position
			break
		}

		if i == first {
			collecting = true
		}

		if collecting {
			evs = append(evs, ev.Dup())
		}
	}
	return
}

func printBars(text string, bars ...*Bar) {
	if !DEBUG {
		return
	}
	fmt.Printf(">>>>>>> bars %s\n", text)

	for _, b := range bars {
		fmt.Printf("%s\n", b.String())
	}

	fmt.Printf("<<<<<<< bars %s\n", text)
}

func multiHasNote(me *items.MultiItem) bool {
	for _, it := range me.Items {
		switch it.(type) {
		case *items.Note, *items.MIDINote:
			return true
		default:
		}
	}
	return false
}

func ntupleHasNote(me *items.NTuple) bool {
	for _, it := range me.Items {
		switch it.(type) {
		case *items.Note, *items.MIDINote:
			return true
		default:
		}
	}
	return false
}

func _applyLyric(ev *Event, lyric string) (applied *Event, wasApplied bool) {
	applied = ev.Dup()
	var l items.Lyric
	l.Text = lyric

	switch v := applied.Item.(type) {
	case *items.MultiItem:
		if !multiHasNote(v) {
			return
		}
		v.Items = append(v.Items, &l)
		applied.Item = v
		wasApplied = true
	case *items.Note:
		var me = &items.MultiItem{}
		l.PosShift = v.PosShift
		me.Items = append(me.Items, ev.Item.Dup(), &l)
		applied.Item = me
		wasApplied = true
	case *items.MIDINote:
		var me = &items.MultiItem{}
		l.PosShift = v.PosShift
		me.Items = append(me.Items, ev.Item.Dup(), &l)
		applied.Item = me
		wasApplied = true
	case *items.NTuple:
		if !ntupleHasNote(v) {
			return
		}
		var me = &items.MultiItem{}
		l.PosShift = v.PosShift
		me.Items = append(me.Items, ev.Item.Dup(), &l)
		applied.Item = me
		wasApplied = true
	default:
		if ev.Item != items.Hold {
			return
		}

		var me = &items.MultiItem{}
		me.Items = append(me.Items, ev.Item.Dup(), &l)
		applied.Item = me
		wasApplied = true
	}
	return
}

// applyLyrics applies the lyrics to the items of the pattern
func applyLyrics(events []*Event, lyrics []string) (applied []*Event) {
	if len(lyrics) == 0 {
		return events
	}
	var i int
	for _, ev := range events {
		if i >= len(lyrics) {
			i = i % len(lyrics)
		}
		nev, wasApplied := _applyLyric(ev, lyrics[i])
		if wasApplied {
			i++
		}
		applied = append(applied, nev)
	}
	return
}

func forwardBars(bars []*Bar, diff uint) {
	if diff == 0 {
		return
	}
	for _, b := range bars {
		b.Position += diff
	}
}

func forwardEvents(ins []*Event, diff uint) (outs []*Event) {
	if diff == 0 {
		return ins
	}
	for _, ev := range ins {
		nev := ev.Dup()
		nev.Position += diff
		outs = append(outs, nev)
	}
	return
}

func moveBySyncFirst(ins []*Event) (out []*Event) {
	var posDiff uint
	for i, ev := range ins {
		if ev == nil {
			continue
		}

		if i == 0 {
			posDiff = ev.Position
		}

		nev := ev.Dup()
		nev.Position = ev.Position - posDiff

		out = append(out, nev)
	}
	return
}

// findNextNotEmptyPos finds the next non empty position in evts and adds forward to it
func findNextPos(i int, forward int, evts []*Event) int {
	// TODO test, if it works
	for j := i + 1; j < len(evts); j++ {
		if evts[j].Item != nil {
			//_, isOver := evts[j].Item.(*items.Override)

			//if !isOver {
			return int(evts[j].Position) + forward
			//}
		}
	}

	return -1
}

func rollTheDiceForAnItem(it items.Item) items.Item {
	if it == nil {
		return nil
	}
	switch v := it.(type) {
	case *items.RandomChooser:
		rand.Seed(time.Now().UnixNano())
		return v.Alternatives[rand.Intn(len(v.Alternatives))].Dup()
	case *items.RandomProbability:
		if v.Prob == 0 {
			return nil
		}
		if v.Prob >= 100 {
			return v.Item.Dup()
		}
		rand.Seed(time.Now().UnixNano())
		if int(v.Prob) >= rand.Intn(100) {
			return v.Item.Dup()
		}
		return nil
	default:
		return it.Dup()
	}
}

func rollTheDice(events []*Event) []*Event {
	// resolv remaining randomness
	var evs = make([]*Event, len(events))
	for i, ev := range events {
		var nuEv = &Event{
			Position: ev.Position,
			Item:     rollTheDiceForAnItem(ev.Item),
		}
		evs[i] = nuEv
	}

	return evs
}

func findPattern(fromSketch *Sketch, fromCol string, patternName string) (sk *Sketch, column string, err error) {
	var rs = &patternFinder{fromSketch}
	return rs.FindPattern(fromCol, patternName)
}
