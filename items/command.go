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

func (c Command) GetInclude() (inc Include) {
	inc.File = strings.Trim(c.Params[0], `"'`)
	inc.Sketch = "=SCORE"
	if len(c.Params) > 1 {
		inc.Sketch = strings.Trim(c.Params[1], `"'`)
	}

	if len(c.Params) > 2 {
		inc.Params = c.Params[2:]
	}
	return
}

func (c Command) IsInclude() bool {
	return c.Name == "$include"
}

func (c Command) String() string {
	return c.original
}

func (p *Command) Parse(data string, posIn32th uint) error {
	data = strings.TrimSpace(strings.TrimSpace(data)[1:])
	p.original = data
	params := ""

	if idx := strings.Index(data, "("); idx > 0 {
		params = strings.TrimSpace(strings.Trim(data[idx:], "()"))
		data = data[:idx]
	}

	p.Name = strings.ToLower(data)

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
