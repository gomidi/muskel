package sketch

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"gitlab.com/gomidi/muskel/items"
	"gitlab.com/gomidi/muskel/patterncommands"
)

func SketchFromSketchDef(name string, sketchDef [][2]string, sc Score) (*Sketch, error) {
	var sk = NewSketch(name, sc)
	sk.AddColumn("pattern")

	for i, df := range sketchDef {
		pos, val := df[0], df[1]

		err := sk.ParseLine([]string{pos, val, ""})
		if err != nil {
			return nil, fmt.Errorf("can't parse line %v: %s", i, err.Error())
		}
	}

	return sk, nil
}

func EventsFromSketchDef(name string, sketchDef [][2]string, sc Score, params []string) (evts []*items.Event, err error) {
	sk, err := SketchFromSketchDef(name, sketchDef, sc)

	if err != nil {
		return nil, err
	}

	_evts, err := sk.Unroll(nil, "pattern", params...)

	if err != nil {
		return nil, err
	}

	for _, pev := range _evts {
		ev := &items.Event{}
		ev.Position = pev.Position
		ev.Item = pev.Item.Dup()
		evts = append(evts, ev)
	}

	if DEBUG {
		printEvents(" after EventsFromSketchDef "+name, evts)
	}

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



func forwardBars(bars []*Bar, diff uint) {
	if diff == 0 {
		return
	}
	for _, b := range bars {
		b.Position += diff
	}
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

func printEvents(message string, evts []*items.Event) {
	if DEBUG {
		fmt.Printf("##> Events %s\n", message)
		for _, ev := range evts {
			fmt.Printf("[%v] %v ", ev.Position, ev.String())
		}
		fmt.Printf("\n##< Events %s\n", message)
	}
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

func (p *patterncmdHelper) GetSketchDefEvents(sketchDef [][2]string) ([]*items.Event, error) {
	return EventsFromSketchDef(p.cmdName, sketchDef, p.column.sketch.Score, p.params)

}

func (p *patterncmdHelper) GetCallEvents(endPos uint, callDef string, params ...string) ([]*items.Event, error) {
	pc := &items.Pattern{}
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

	pc.Params = params

	pcc := p.column.newPattern(pc)
	if endPos == 0 {
		endPos = p.column.sketch.projectedBarEnd
		//fmt.Printf("endPos: %v\n", endPos)
	}
	evts, _, _, err := pcc.unroll(0, endPos)
	return evts, err
}

var _ patterncommands.Helper = &patterncmdHelper{}
