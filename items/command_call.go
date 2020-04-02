package items

import (
	"strings"
)

type Command struct {
	Name     string
	Params   []string
	original string
}

func (c Command) Dup() Item {
	return &c
}

func (c Command) String() string {
	return c.original
}

func (p *Command) Parse(call string, posIn32th uint) error {
	//call = call[1:]
	call = strings.TrimSpace(strings.TrimSpace(call)[1:])
	p.original = call
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
