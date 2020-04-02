package items

import (
	"fmt"
	"strings"
)

type PipedPatternCommands struct {
	Cmds []*Command
}

func (p *PipedPatternCommands) Dup() Item {
	var n PipedPatternCommands
	for _, cc := range p.Cmds {
		n.Cmds = append(n.Cmds, cc.Dup().(*Command))
	}
	return &n
}

func (p *PipedPatternCommands) Parse(data string, pos uint) error {
	cmds := strings.Split(data, "/")
	for _, c := range cmds {
		var cc Command
		err := cc.Parse(c, 0)
		if err != nil {
			return err
		}
		if cc.Name[1] == '$' {
			return fmt.Errorf("only pattern commands can be piped")
		}
		p.Cmds = append(p.Cmds, &cc)
	}
	return nil
}

func (p *PipedPatternCommands) String() string {
	var bd strings.Builder
	for i, pp := range p.Cmds {
		if i > 0 {
			bd.WriteString("/")
		}
		bd.WriteString(pp.String())
	}
	return bd.String()
}

var _ Item = &PipedPatternCommands{}
