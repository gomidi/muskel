package items

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type Pattern struct {
	Name                               string
	Params                             []string
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
	Lyrics                             *LyricsTable
	Exploded                           bool // e.g. when token is (a b c) then =patt(token...) becomes =patt(a,b,c)
	IncludeFile                        string
	Part                               string
}

func (c *Pattern) Dup() Item {
	return &Pattern{
		Name:                               c.Name,
		Params:                             c.Params,
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
		IncludeFile:                        c.IncludeFile,
		Part:                               c.Part,
	}
}

func (p *Pattern) newSketchPattern(column Columner) *sketchPattern {
	return &sketchPattern{column: column, pattern: p}
}

func (p *Pattern) GetEventStream(column Columner, start, end uint) (*EventStream, error) {
	return p.newSketchPattern(column).getEventStream(start, end)
}

func (p *Pattern) GetOverrideEventStream(column Columner, start, end uint) (*EventStream, error) {
	return p.newSketchPattern(column).getOverrideEventStream(start, end)
}

func (p *Pattern) Unroll(column Columner, start, until uint) (evt []*Event, end uint, err error) {
	return p.newSketchPattern(column).unrollPattern(start, until)
}

var _ UnrollGetter = &Pattern{}

func (c *Pattern) GetES(p Columner, ev *Event, start, endPos uint) (mixed []*EventStream, err error) {
	// TODO prevent endless loops from templates calling each other like col1 -> col2 -> col1 by keeping a stack of template calls
	// and checking them for duplicates (the stack may as well be a map[string]bool; makes is easier; we need the complete names in there
	posEv, err := c.GetEventStream(p, start, endPos)
	if err != nil {
		return nil, err
	}

	//PrintEvents("inside pattern GetES", posEv.Events)
	return []*EventStream{posEv}, nil
}

func (p *Pattern) String() string {
	var bf strings.Builder

	if p.SyncFirst {
		bf.WriteString("!")
	}

	bf.WriteString(p.Name)

	if len(p.Params) > 0 {
		bf.WriteString("(" + strings.Join(p.Params, ",") + ")")
	}

	return bf.String()
}

func (p *Pattern) parseItem(data string, posIn32th uint) (item Item, err error) {
	if len(data) == 0 {
		return nil, nil
	}

	return Parse(data, posIn32th)
}

func (p *Pattern) Parse(call string, positionIn32th uint) (err error) {
	params := ""
	lyrics := ""

	if idx := strings.Index(call, "..."); idx > 0 && idx+3 == len(call) {
		p.Exploded = true
		call = call[:idx] //+ data[idx+3:]
	}

	if idx := strings.Index(call, "@"); idx > 0 {
		lyrics = call[idx:]
		call = call[:idx]
	}

	_ = lyrics

	if lyrics != "" {
		var lp LyricsTable
		err := lp.Parse(lyrics, 0)
		if err != nil {
			return fmt.Errorf("can't parse lyrics call %q", lyrics)
		}
		p.Lyrics = &lp
	}

	if idx := strings.Index(call, "%"); idx > 0 {
		repeat, repErr := strconv.Atoi(call[idx+1:])
		if repErr != nil {
			return fmt.Errorf("invalid number of repetitions: %s", call[idx+1:])
		}
		p.Repeat = uint(repeat)
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

	mt := regPatternCallNameDyn.FindStringSubmatch(call)
	p.Name = mt[1]
	if p.Name[0] != '=' {
		return fmt.Errorf("invalid template name: %q", p.Name)
	}

	if mt[2] != "" {
		dt := mt[2][1:]
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

	p.DynamicAdd = mt[3]

	switch mt[4] {
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

	return nil
}

func (p *Pattern) AddDynamic(orig string) (nu string) {
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

var regPattStr = "^([=+" + regexp.QuoteMeta(".") + "@a-zA-Z][._~a-zA-Z0-9]+)(" +
	regexp.QuoteMeta("^") +
	"{1,2}[-0-9]+){0,1}([" +
	regexp.QuoteMeta("-+=") +
	"]*)([<>]{0,1})"
	//+
	//"?(" + regexp.QuoteMeta("@") + "[a-zA-Z0-9])$"

var regPatternCallNameDyn = regexp.MustCompile(regPattStr)

var regTokenStr = "^([+" + regexp.QuoteMeta(".") + "a-zA-Z][._~a-zA-Z0-9]+)(.*)"

//+
//"?(" + regexp.QuoteMeta("@") + "[a-zA-Z0-9])$"

var regTokenCallNameDyn = regexp.MustCompile(regTokenStr)

var regExTemplate0 = regexp.MustCompile("^(=[" + regexp.QuoteMeta("!") + "]?[a-zA-Z][" + regexp.QuoteMeta(".") + "_0-9a-zA-Z]*).*")
var regExTemplate1 = regexp.MustCompile("^(" + regexp.QuoteMeta("=") + "[" + regexp.QuoteMeta("!") + "]?" + regexp.QuoteMeta(".") + "[" + regexp.QuoteMeta(".") + "_0-9a-zA-Z]*).*")

var regExToken0 = regexp.MustCompile("^(" + regexp.QuoteMeta(".") + "[" + regexp.QuoteMeta("!") + "]?[a-zA-Z][" + regexp.QuoteMeta(".") + "_0-9a-zA-Z]*).*")
var regExToken1 = regexp.MustCompile("^(" + regexp.QuoteMeta(".") + "[" + regexp.QuoteMeta("!") + "]?" + regexp.QuoteMeta(".") + "[" + regexp.QuoteMeta(".") + "_0-9a-zA-Z]*).*")

var regExEndScaleNote = regexp.MustCompile("^_(-)?[0-9]+")
