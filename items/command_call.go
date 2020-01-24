package items

import (
	"strings"
)

type CommandCall struct {
	Name     string
	Params   []string
	original string
	Position uint
}

func (c CommandCall) Dup() Item {
	return &c
}

func (c CommandCall) String() string {
	return c.original
}

func (p *CommandCall) Parse(call string, posIn32th uint) error {
	call = strings.TrimSpace(call)
	p.original = call
	p.Position = posIn32th
	params := ""

	if idx := strings.Index(call, "("); idx > 0 {
		params = strings.TrimSpace(strings.Trim(call[idx:], "()"))
		call = call[:idx]
	}

	p.Name = strings.ToLower(call)

	if params != "" {
		ps := strings.Split(params, ",")

		for _, param := range ps {
			param = strings.TrimSpace(param)

			if param != "" {
				p.Params = append(p.Params, param)
			}
		}
	}

	return nil
}
