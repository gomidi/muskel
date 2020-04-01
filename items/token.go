package items

import (
	"fmt"
	"strconv"
	"strings"
)

type Token struct {
	Name                               string
	Params                             []string
	Slice                              [2]int
	SyncFirst                          bool
	firstPos                           uint
	DynamicAdd                         string
	ScaleMove                          int8 // 0: force to scale (take next note in scale if not exact matching, no movement for in scale notes), n >0 || n < 0: move by n steps along the scale
	ScaleMoveMode                      int8 // 0: no scale movement, 1: move only scale notes, 2: move only scale notes or non scale notes depending on the first item
	FirstNoteIsScaleNote               int  // 0: not set, 1: true, 2: false
	FirstNoteAbsKey                    uint8
	syncFirstThroughTemplateDefinition bool
	PosShift                           int // 0 = no, 1 = laidback, -1 = ahead of time
	Repeat                             uint
	//Lyrics                             *Call
	Exploded bool // e.g. when token is (a b c) then =patt(token...) becomes =patt(a,b,c)
}

/*
func (t *Call) IsLyrics() bool {
	return strings.Split(t.Name, ".")[0][0] == '@'
}
*/

func (pc *Token) Parse(data string, positionIn32th uint) (err error) {
	return pc.parseTemplate(data, positionIn32th)
}

func (c *Token) Dup() Item {
	return &Token{
		Name:                               c.Name,
		Params:                             c.Params,
		Slice:                              c.Slice,
		SyncFirst:                          c.SyncFirst,
		firstPos:                           c.firstPos,
		DynamicAdd:                         c.DynamicAdd,
		ScaleMove:                          c.ScaleMove,
		ScaleMoveMode:                      c.ScaleMoveMode,
		FirstNoteIsScaleNote:               c.FirstNoteIsScaleNote,
		FirstNoteAbsKey:                    c.FirstNoteAbsKey,
		syncFirstThroughTemplateDefinition: c.syncFirstThroughTemplateDefinition,
		PosShift:                           c.PosShift,
		Repeat:                             c.Repeat,
		//Lyrics:                             c.Lyrics,
		Exploded: c.Exploded,
	}
}

func (p *Token) String() string {
	var bf strings.Builder

	if p.SyncFirst {
		bf.WriteString("!")
	}

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

	return bf.String()
}

func (p *Token) parseItem(data string, posIn32th uint) (item Item, err error) {
	if len(data) == 0 {
		return nil, nil
	}

	var parser Parser
	return parser.ParseItem(data, posIn32th)
}

func (p *Token) parseTemplate(call string, positionIn32th uint) error {
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
	if p.Name[0] != '.' {
		return fmt.Errorf("invalid token name: %q", p.Name)
	}

	if mt[3] != "" {
		dt := mt[3][1:]
		p.ScaleMoveMode = 1 // move only scale notes
		if dt[0] == '^' {
			p.ScaleMoveMode = 2 // move depending on first note
			dt = dt[1:]
		}

		si, err := strconv.Atoi(dt)
		if err != nil {
			return fmt.Errorf("error in scale moving of token %q: %q is not a number", p.Name, dt)
		}

		p.ScaleMove = int8(si)
	}

	p.DynamicAdd = mt[4]

	switch mt[5] {
	case "<":
		p.PosShift = -1
	case ">":
		p.PosShift = 1
	default:
		p.PosShift = 0

	}

	if params != "" {
		p.Params = splitParams(params)
	}

	p.Slice[0] = -1
	p.Slice[1] = -1

	if slice != "" {
		sl := strings.Split(slice, ":")
		if len(sl) != 2 {
			return fmt.Errorf("ERROR in call of token %q: invalid slice %q", p.Name, "["+slice+"]")
		}

		from := strings.TrimSpace(sl[0])
		to := strings.TrimSpace(sl[1])

		if from == "" {
			p.Slice[0] = 0
		} else {
			fromI, err := strconv.Atoi(from)
			if err != nil {
				return fmt.Errorf("ERROR in call of token %q: invalid slice %q", p.Name, "["+slice+"]")
			}
			p.Slice[0] = fromI
		}

		if to != "" {
			toI, err := strconv.Atoi(to)
			if err != nil || toI == 0 {
				return fmt.Errorf("ERROR in call of token %q: invalid slice %q", p.Name, "["+slice+"]")
			}
			p.Slice[1] = toI
		}

	}
	return nil
}

func (p *Token) AddDynamic(orig string) (nu string) {
	//fmt.Printf("addDynamic %q to %q\n", p.DynamicAdd, orig)
	if orig == "=" {
		return orig
	}

	if p.DynamicAdd == "" {
		return orig
	}

	if p.DynamicAdd == "=" {
		return p.DynamicAdd
	}

	return reduceDynamic(orig + p.DynamicAdd)
}
