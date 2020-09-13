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

	if items.DEBUG {
		items.PrintEvents(" after EventsFromSketchDef "+name, evts)
	}

	return
}

func findPattern(fromSketch *Sketch, fromCol string, patternFile string, patternName string, args []string) (sk *Sketch, column string, err error) {
	var rs = &patternFinder{fromSketch}
	return rs.FindPattern(patternFile, fromCol, patternName, args)
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
	if idx := strings.Index(callDef, "..."); idx > 0 && idx+3 == len(callDef) {
		pc.Exploded = true
		callDef = callDef[:idx] //+ data[idx+3:]
	}
	err := pc.Parse(callDef, 0)
	if err != nil {
		return nil, err
	}

	pc.Params = params

	if endPos == 0 {
		endPos = p.column.sketch.projectedBarEnd
	}

	evts, _, _, err := p.column.UnrollPattern(0, endPos, pc)
	return evts, err
}

var _ patterncommands.Helper = &patterncmdHelper{}
