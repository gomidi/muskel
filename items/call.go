package items

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type Call struct {
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
	Lyrics                             *Call
	Exploded                           bool // e.g. when token is (a b c) then =patt(token...) becomes =patt(a,b,c)
}

func (t *Call) IsLyrics() bool {
	return strings.Split(t.Name, ".")[0][0] == '@'
}

func (c *Call) Dup() Item {
	return &Call{
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
		Lyrics:                             c.Lyrics,
		Exploded:                           c.Exploded,
	}
}

func (c *Call) IsPattern() bool {
	return c.Name[0] == '='
}

func (c *Call) IsToken() bool {
	return c.Name[0] != '@' && c.Name[0] != '='
}

func (p *Call) String() string {
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

func (p *Call) parseItem(data string, posIn32th uint) (item Item, err error) {
	if len(data) == 0 {
		return nil, nil
	}

	var parser Parser
	return parser.ParseItem(data, posIn32th)
}

func (p *Call) ParseTemplate(call string, positionIn32th uint) error {
	slice := ""
	params := ""
	lyrics := ""

	if idx := strings.Index(call, "@"); idx > 0 {
		lyrics = call[idx:]
		call = call[:idx]
	}

	_ = lyrics

	if lyrics != "" {
		var lp Call
		err := lp.ParseTemplate(lyrics, 0)
		if err != nil {
			return fmt.Errorf("can't parse lyrics call %q", lyrics)
		}
		p.Lyrics = &lp
	}

	if idx := strings.Index(call, "["); idx > 0 {
		slice = strings.TrimSpace(strings.Trim(call[idx:], "[]"))
		call = call[:idx]
	}

	if idx := strings.Index(call, "("); idx > 0 {
		lidx := strings.LastIndex(call, ")")
		params = strings.TrimSpace(call[idx+1 : lidx])
		call = call[:idx]
	}

	if call[0:2] == "=!" {
		p.SyncFirst = true
		call = "=" + call[2:]
	}

	mt := regTemplateCallNameDyn.FindStringSubmatch(call)
	p.Name = mt[1]

	if mt[2] != "" && mt[2][0] == '%' {
		repeat, repErr := strconv.Atoi(mt[2][1:])
		if repErr != nil {
			return fmt.Errorf("invalid number of repetitions: %s", mt[2][1:])
		}
		p.Repeat = uint(repeat)
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
			return fmt.Errorf("error in scale moving of template %q: %q is not a number", p.Name, dt)
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
			return fmt.Errorf("ERROR in call of template %q: invalid slice %q", p.Name, "["+slice+"]")
		}

		from := strings.TrimSpace(sl[0])
		to := strings.TrimSpace(sl[1])

		if from == "" {
			p.Slice[0] = 0
		} else {
			fromI, err := strconv.Atoi(from)
			if err != nil {
				return fmt.Errorf("ERROR in call of template %q: invalid slice %q", p.Name, "["+slice+"]")
			}
			p.Slice[0] = fromI
		}

		if to != "" {
			toI, err := strconv.Atoi(to)
			if err != nil || toI == 0 {
				return fmt.Errorf("ERROR in call of template %q: invalid slice %q", p.Name, "["+slice+"]")
			}
			p.Slice[1] = toI
		}

	}
	return nil
}

func reduceDynamic(dyn string) string {

	var counter int
	for _, sym := range dyn {
		switch sym {
		case '-':
			counter--
		case '+':
			counter++
		}
	}

	//fmt.Printf("reduceDynamic: %q: %v\n", dyn, counter)

	switch {
	case counter > 0:
		return strings.Repeat("+", counter)
	case counter < 0:
		return strings.Repeat("-", counter*(-1))
	default:
		return ""
	}
}

func (p *Call) AddDynamic(orig string) (nu string) {
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

var TemplateReg = regexp.MustCompile(regexp.QuoteMeta("#") + "([0-9]+)")

var regTempStr = "^([=@a-zA-Z][._~a-zA-Z0-9]+)(" +
	regexp.QuoteMeta("%") +
	"?[0-9]*)(" +
	regexp.QuoteMeta("^") +
	"{1,2}[-0-9]+){0,1}([" +
	regexp.QuoteMeta("-+=") +
	"]*)([<>]{0,1})"
	//+
	//"?(" + regexp.QuoteMeta("@") + "[a-zA-Z0-9])$"

var regTemplateCallNameDyn = regexp.MustCompile(regTempStr)

var regExTemplate0 = regexp.MustCompile("^([=]?[" + regexp.QuoteMeta("!") + "]?[a-zA-Z][a-zA-Z][" + regexp.QuoteMeta(".") + "_0-9a-zA-Z]*).*")
var regExTemplate1 = regexp.MustCompile("^(" + regexp.QuoteMeta("=") + "[" + regexp.QuoteMeta("!") + "]?" + regexp.QuoteMeta(".") + "[" + regexp.QuoteMeta(".") + "_0-9a-zA-Z]*).*")

func (pc *Call) Parse(data string, positionIn32th uint) (err error) {
	return pc.ParseTemplate(data, positionIn32th)
}
