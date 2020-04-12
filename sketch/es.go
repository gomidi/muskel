package sketch

import (
	"fmt"

	"gitlab.com/gomidi/muskel/items"
	"gitlab.com/gomidi/muskel/patterncommands"
)

type pipedPatternCommandES struct {
	item   *items.PipedPatternCommands
	params []string
}

var _ items.UnrollGetter = &pipedPatternCommandES{}

func (p *pipedPatternCommandES) GetES(c items.Columner, ev *items.Event, start, endPos uint) ([]*items.EventStream, error) {
	helper := &patterncmdHelper{
		column: c.(*column),
		params: p.params,
		inPipe: true,
	}

	for _, cmd := range p.item.Cmds {
		if fn, hasCmd := patterncommands.Commands[cmd.Name]; hasCmd {
			helper.cmdName = cmd.Name
			_evts, err := fn(cmd.Params, helper)

			if err != nil {
				return nil, err
			}

			helper.pipeEvents = _evts
		} else {
			return nil, fmt.Errorf("unknown command %q", cmd.Name)
		}
	}

	posEv, err := items.PipedEventStream(c, start, endPos, helper.pipeEvents)
	if err != nil {
		return nil, err
	}
	return []*items.EventStream{posEv}, nil
}

type commandES struct {
	item   *items.Command
	params []string
}

var _ items.UnrollGetter = &commandES{}

func (c *commandES) GetES(p items.Columner, ev *items.Event, start, endPos uint) (x []*items.EventStream, err error) {
	if fn, hasCmd := patterncommands.Commands[c.item.Name]; hasCmd {
		helper := &patterncmdHelper{
			cmdName: c.item.Name,
			column:  p.(*column),
			params:  c.params,
		}

		_evts, err := fn(c.item.Params, helper)

		if err != nil {
			return nil, err
		}

		if len(_evts) == 0 {
			return nil, nil
		}
		var posEv *items.EventStream
		posEv, err = items.PipedEventStream(p, start, endPos, _evts)
		if err != nil {
			return nil, err
		}
		return []*items.EventStream{posEv}, nil
	}
	return nil, fmt.Errorf("unknown command %q", c.item.Name)
}
