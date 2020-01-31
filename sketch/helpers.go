package sketch

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"

	"gitlab.com/gomidi/muskel/items"
	"gitlab.com/gomidi/muskel/patterncommands"
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
		case strings.IndexRune("1234567890&;:,", rn) >= 0:
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
	//	fmt.Printf("ParseLine: %q\n", patternDef)
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

func eventsFromPatternDef(name, patternDef string, sc Score, params []string) (evts []*items.Event, err error) {
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
		ev := &items.Event{}
		ev.Position = pev.Position
		ev.Item = pev.Item.Dup()
		evts = append(evts, ev)
	}

	if DEBUG {
		printEvents(" after eventsFromPatternDef "+patternDef, evts)
	}

	return
}

func getFirstType(es *eventStream) string {
	if es == nil || len(es.events) == 0 || es.events[0] == nil {
		return ""
	}

	return fmt.Sprintf("%T", es.events[0].Item)
}

type sortedEventStream []*eventStream

func (s sortedEventStream) Less(a, b int) bool {
	if s[a].start == s[b].start {
		if s[a].isOverride {
			return false
		}

		ta := getFirstType(s[a])
		tb := getFirstType(s[b])

		//fmt.Printf("%q vs %q\n", ta, tb)

		if ta == "*items.Scale" {
			return true
		}

		if tb == "*items.Scale" {
			return false
		}

		return false

	}

	return s[a].start < s[b].start
}

func (s sortedEventStream) Swap(a, b int) {
	s[a], s[b] = s[b], s[a]
}

func (s sortedEventStream) Len() int {
	return len(s)
}

func mergeEventStreams(mixed []*eventStream, endPos uint) (evts []*items.Event) {
	var (
		lastStreamPos uint
		keepLooping   bool = true
		lastStream    *eventStream
	)

	lastStream = nil
	sorted := sortedEventStream(mixed)
	sort.Sort(sorted)

	for idx, stream := range sorted {

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

func getEventsInPosRange(from, to uint, evts []*items.Event) (res []*items.Event) {
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

func findEventThatIsNotRest(start int, evts []*items.Event) (found int) {
	if start >= len(evts) {
		return -1
	}

	if evts[start].Item != items.Rest {
		return start
	}

	return findEventThatIsNotRest(start+1, evts)
}

// sliceEvents slices the events according to the template call slice definition
func sliceEvents(slice [2]int, all []*items.Event, projectedBarEnd uint) (evs []*items.Event, absoluteEnd uint) {
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
	for _, it := range me.Events {
		switch it.Item.(type) {
		case *items.Note, *items.MIDINote:
			return true
		default:
		}
	}
	return false
}

func ntupleHasNote(me *items.NTuple) bool {
	for _, it := range me.Events {
		switch it.Item.(type) {
		case *items.Note, *items.MIDINote:
			return true
		default:
		}
	}
	return false
}

func _applyLyric(ev *items.Event, lyric string) (applied *items.Event, wasApplied bool) {
	applied = ev.Dup()
	var l items.Lyric
	l.Text = lyric

	switch v := applied.Item.(type) {
	case *items.MultiItem:
		if !multiHasNote(v) {
			return
		}
		var _ev = ev.Dup()
		_ev.PosShift = 0
		_ev.Item = &l
		v.Events = append(v.Events, _ev)
		applied.Item = v
		wasApplied = true
	case *items.Note:
		var me = &items.MultiItem{}
		//l.PosShift = v.PosShift
		_ev := ev.Dup()
		_ev.PosShift = 0
		_ev.Item = &l

		_ev2 := ev.Dup()
		_ev2.PosShift = 0
		me.Events = append(me.Events, _ev2, _ev)
		applied.Item = me
		wasApplied = true
	case *items.MIDINote:
		var me = &items.MultiItem{}
		_ev := ev.Dup()
		_ev.PosShift = 0
		_ev.Item = &l
		_ev2 := ev.Dup()
		_ev2.PosShift = 0
		me.Events = append(me.Events, _ev2, _ev)
		applied.Item = me
		wasApplied = true
	case *items.NTuple:
		if !ntupleHasNote(v) {
			return
		}
		var me = &items.MultiItem{}
		var _ev = ev.Dup()
		_ev.Item = &l
		_ev.PosShift = 0
		_ev2 := ev.Dup()
		_ev2.PosShift = 0
		//l.PosShift = v.PosShift
		me.Events = append(me.Events, _ev2, _ev)
		applied.Item = me
		wasApplied = true
	default:
		if ev.Item != items.Hold {
			return
		}

		var me = &items.MultiItem{}
		_ev := ev.Dup()
		_ev.Item = &l
		_ev.PosShift = 0
		_ev2 := ev.Dup()
		_ev2.PosShift = 0
		me.Events = append(me.Events, _ev2, _ev)
		applied.Item = me
		wasApplied = true
	}
	return
}

// applyLyrics applies the lyrics to the items of the pattern
func applyLyrics(events []*items.Event, lyrics []string) (applied []*items.Event) {
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

func forwardEvents(ins []*items.Event, diff uint) (outs []*items.Event) {
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

func moveBySyncFirst(ins []*items.Event) (out []*items.Event) {
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
func findNextPos(i int, forward int, evts []*items.Event) int {
	// TODO test, if it works
	for j := i + 1; j < len(evts); j++ {
		if evts[j].Item != nil {
			return int(evts[j].Position) + forward
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

func rollTheDice(events []*items.Event) []*items.Event {
	// resolv remaining randomness
	var evs = make([]*items.Event, len(events))
	for i, ev := range events {
		var nuEv = ev.Dup()
		nuEv.Item = rollTheDiceForAnItem(ev.Item)
		evs[i] = nuEv
	}

	return evs
}

func findPattern(fromSketch *Sketch, fromCol string, patternName string) (sk *Sketch, column string, err error) {
	var rs = &patternFinder{fromSketch}
	return rs.FindPattern(fromCol, patternName)
}

type patterncmdHelper struct {
	cmdName    string
	column     *column
	params     []string
	pipeEvents []*items.Event
	inPipe     bool
}

func (p *patterncmdHelper) GetPipeEvents() ([]*items.Event, error) {
	if !p.inPipe {
		return nil, fmt.Errorf("not in pattern command pipe")
	}
	return p.pipeEvents, nil
}

func (p *patterncmdHelper) GetPatternDefEvents(patternDef string) ([]*items.Event, error) {
	return eventsFromPatternDef(p.cmdName, patternDef, p.column.sketch.Score, p.params)

}

func (p *patterncmdHelper) GetCallEvents(endPos uint, callDef string) ([]*items.Event, error) {
	pc := &items.Call{}
	//pc.Parser = p
	if idx := strings.Index(callDef, "..."); idx > 0 && idx+3 == len(callDef) {
		//fmt.Printf("len(data) = %v; idx = %v\n", len(data), idx)
		pc.Exploded = true
		callDef = callDef[:idx] //+ data[idx+3:]
	}
	//fmt.Printf("data cleaned: %q\n", data)
	err := pc.Parse(callDef, 0)
	if err != nil {
		return nil, err
	}
	pcc := p.column.newCall(pc)
	if endPos == 0 {
		endPos = p.column.sketch.projectedBarEnd
		//fmt.Printf("endPos: %v\n", endPos)
	}
	evts, _, _, err := pcc.unroll(0, endPos)
	return evts, err
}

var _ patterncommands.Helper = &patterncmdHelper{}
