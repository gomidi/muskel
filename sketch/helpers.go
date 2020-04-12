package sketch

import (
	"fmt"
	"strings"

	"gitlab.com/gomidi/muskel/items"
	"gitlab.com/gomidi/muskel/patterncommands"
)

var eventDebug bool

func SketchFromSketchDef(name string, sketchDef [][2]string, sc Score) (*Sketch, error) {
	var sk = New(name, sc)
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

func getBarsInPosRange(from, to uint, bars []*items.Bar) (res []*items.Bar) {
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

func printBars(text string, bars ...*items.Bar) {
	if !DEBUG {
		return
	}
	fmt.Printf(">>>>>>> bars %s\n", text)

	for _, b := range bars {
		fmt.Printf("%s\n", b.String())
	}

	fmt.Printf("<<<<<<< bars %s\n", text)
}

func forwardBars(bars []*items.Bar, diff uint) {
	if diff == 0 {
		return
	}
	for _, b := range bars {
		b.Position += diff
	}
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

	//pcc := p.column.newPattern(pc)
	if endPos == 0 {
		endPos = p.column.sketch.projectedBarEnd
		//fmt.Printf("endPos: %v\n", endPos)
	}

	evts, _, _, err := p.column.UnrollPattern(0, endPos, pc)
	return evts, err
}

var _ patterncommands.Helper = &patterncmdHelper{}
