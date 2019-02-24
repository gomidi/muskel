package muskel

import (
	"strings"
)

type CommandCall struct {
	Name   string
	Params []string
}

func (p *CommandCall) Parse(call string) error {
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
