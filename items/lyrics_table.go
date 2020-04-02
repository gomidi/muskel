package items

import (
	"fmt"
	"strconv"
	"strings"
)

type LyricsTable struct {
	Name   string
	Params []string
	Slice  [2]int
	Repeat uint
}

func (c *LyricsTable) Dup() Item {
	return &LyricsTable{
		Name:   c.Name,
		Params: c.Params,
		Slice:  c.Slice,
		Repeat: c.Repeat,
	}
}

func (p *LyricsTable) String() string {
	var bf strings.Builder

	bf.WriteString(p.Name)

	if len(p.Params) > 0 {
		bf.WriteString("(" + strings.Join(p.Params, ",") + ")")
	}

	if p.Slice[0] >= 0 || p.Slice[1] >= 0 {
		bf.WriteString("[")

		if p.Slice[0] >= 0 {
			bf.WriteString(fmt.Sprintf("%v", p.Slice[0]))
		}

		bf.WriteString(":")

		if p.Slice[1] >= 0 {
			bf.WriteString(fmt.Sprintf("%v", p.Slice[1]))
		}

		bf.WriteString("]")
	}

	if p.Repeat > 0 {
		fmt.Fprintf(&bf, "%%%v", p.Repeat)
	}

	return bf.String()
}

func (p *LyricsTable) Parse(call string, positionIn32th uint) error {
	slice := ""
	params := ""

	if idx := strings.Index(call, "["); idx > 0 {
		slice = strings.TrimSpace(strings.Trim(call[idx:], "[]"))
		call = call[:idx]
	}

	if idx := strings.Index(call, "("); idx > 0 {
		lidx := strings.LastIndex(call, ")")
		params = strings.TrimSpace(call[idx+1 : lidx])
		call = call[:idx]
	}

	mt := regPatternCallNameDyn.FindStringSubmatch(call)
	p.Name = mt[1]
	if p.Name[0] != '@' {
		return fmt.Errorf("invalid lyrics table name: %q", p.Name)
	}

	if mt[2] != "" && mt[2][0] == '%' {
		repeat, repErr := strconv.Atoi(mt[2][1:])
		if repErr != nil {
			return fmt.Errorf("invalid number of repetitions: %s", mt[2][1:])
		}
		p.Repeat = uint(repeat)
	}

	if params != "" {
		p.Params = splitParams(params)
	}

	p.Slice[0] = -1
	p.Slice[1] = -1

	if slice != "" {
		sl := strings.Split(slice, ":")
		if len(sl) != 2 {
			return fmt.Errorf("ERROR in call of lyrics %q: invalid slice %q", p.Name, "["+slice+"]")
		}

		from := strings.TrimSpace(sl[0])
		to := strings.TrimSpace(sl[1])

		if from == "" {
			p.Slice[0] = 0
		} else {
			fromI, err := strconv.Atoi(from)
			if err != nil {
				return fmt.Errorf("ERROR in call of lyrics %q: invalid slice %q", p.Name, "["+slice+"]")
			}
			p.Slice[0] = fromI
		}

		if to != "" {
			toI, err := strconv.Atoi(to)
			if err != nil || toI == 0 {
				return fmt.Errorf("ERROR in call of lyrics %q: invalid slice %q", p.Name, "["+slice+"]")
			}
			p.Slice[1] = toI
		}

	}
	return nil
}
