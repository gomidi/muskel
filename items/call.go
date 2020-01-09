package items

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"regexp"
	"strconv"
	"strings"
)

type callEvent struct {
	PositionString string
	OriginalData   string
	Position       uint
	Item           Item
}

func (p *callEvent) dup() (nu *callEvent) {
	return &callEvent{
		PositionString: p.PositionString,
		OriginalData:   p.OriginalData,
		Position:       p.Position,
		Item:           p.Item,
	}
}

func (p *callEvent) String() string {
	return Pos32thToString(p.Position) + p.OriginalData
}

type Call struct {
	Name                               string
	Params                             []string
	Slice                              [2]int
	SyncFirst                          bool
	result                             string
	firstPos                           uint
	Events                             []*callEvent // just a sausage of positions within one fictive infinite bar
	Getter                             func(name string) TemplateDefinition
	DynamicAdd                         string
	ScaleMove                          int8 // 0: force to scale (take next note in scale if not exact matching, no movement for in scale notes), n >0 || n < 0: move by n steps along the scale
	ScaleMoveMode                      int8 // 0: no scale movement, 1: move only scale notes, 2: move only scale notes or non scale notes depending on the first item
	FirstNoteIsScaleNote               int  // 0: not set, 1: true, 2: false
	FirstNoteAbsKey                    uint8
	syncFirstThroughTemplateDefinition bool
	ItemParser                         itemParser
	TimeSignature                      [2]uint8
	Parser                             *Parser
	PosShift                           int // 0 = no, 1 = laidback, -1 = ahead of time
	Repeat                             uint
	Lyrics                             *Call
	Exploded                           bool // e.g. when token is (a b c) then =patt(token...) becomes =patt(a,b,c)
}

func (t *Call) IsLyrics() bool {
	return strings.Split(t.Name, ".")[0][0] == '@'
}

func NewCall(ip itemParser) *Call {
	return &Call{
		ItemParser:    ip,
		TimeSignature: [2]uint8{4, 4},
	}
}

func (c *Call) Dup() Item {
	return &Call{
		Name:                               c.Name,
		Params:                             c.Params,
		Slice:                              c.Slice,
		SyncFirst:                          c.SyncFirst,
		result:                             c.result,
		firstPos:                           c.firstPos,
		Events:                             c.Events,
		Getter:                             c.Getter,
		DynamicAdd:                         c.DynamicAdd,
		ScaleMove:                          c.ScaleMove,
		ScaleMoveMode:                      c.ScaleMoveMode,
		FirstNoteIsScaleNote:               c.FirstNoteIsScaleNote,
		FirstNoteAbsKey:                    c.FirstNoteAbsKey,
		syncFirstThroughTemplateDefinition: c.syncFirstThroughTemplateDefinition,
		ItemParser:                         c.ItemParser,
		TimeSignature:                      c.TimeSignature,
		Parser:                             c.Parser,
		PosShift:                           c.PosShift,
		Repeat:                             c.Repeat,
		Lyrics:                             c.Lyrics,
		Exploded:                           c.Exploded,
	}
}

func (c *Call) FirstItem() Item {
	if len(c.Events) > 0 {
		return c.Events[0].Item
	}
	return nil
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

type CallEvents []*callEvent

// Spread is called unless the returned positionOfNextBar is < 0
// The first call to spread is with start number of the position within the bar where the
// template is embedded.
// Subsequent calls will set start to the last returned positionOfNextBar
// when the last event is return, positionOfNextBar is -1
// the returned events are relative to the bar given by start
func (p CallEvents) Spread(start int, num, denom uint8) (barEvents []*callEvent, positionOfNextBar int, nextEv *callEvent) {
	barLength := Length32ths(num, denom)
	positionOfNextBar = start + barLength

	var lastIdx int

	for idx, ev := range p {
		if int(ev.Position) >= start && int(ev.Position) < positionOfNextBar {
			nu := ev.dup()
			nu.Position -= uint(start)
			nu.PositionString = Pos32thToString(nu.Position)
			barEvents = append(barEvents, nu)
			lastIdx = idx
		}
	}

	if lastIdx == len(p)-1 {
		positionOfNextBar = -1
	} else {
		nextEv = p[lastIdx+1].dup()
		nextEv.Position -= uint(start)
		nextEv.PositionString = Pos32thToString(nextEv.Position)
	}

	return
}

type itemParser interface {
	//GetDefinition(name string) *TemplateDefinition
	SetDefinitionGetter(func(name string) TemplateDefinition)
	ParseItem(data string, posIn32th uint) (item Item, err error)
}

func (p *Call) parseItem(data string, posIn32th uint) (item Item, err error) {
	if len(data) == 0 {
		return nil, nil
	}

	if p.ItemParser == nil {
		panic("set ItemParser first")
	}

	p.ItemParser.SetDefinitionGetter(p.Getter)
	return p.ItemParser.ParseItem(data, posIn32th)
}

func (p *Call) mkEvent(position string, posIn32th uint, data string) (ev *callEvent, err error) {
	ev = &callEvent{}
	_, ev.Position, err = PositionTo32th("", position)
	if err != nil {
		return nil, err
	}
	ev.Position += posIn32th
	ev.OriginalData = data
	item, err := p.parseItem(ev.OriginalData, posIn32th)

	if err != nil {
		return nil, err
	}

	ev.Item = item
	if p.SyncFirst || p.syncFirstThroughTemplateDefinition {
		ev.Position -= p.firstPos
	}
	ev.PositionString = Pos32thToString(ev.Position)
	return
}

func (p *Call) parseEvent(idx int, data string, posIn32th uint) (ev *callEvent, err error) {

	var positionBf strings.Builder
	var bf strings.Builder
	rd := bufio.NewReader(strings.NewReader(strings.TrimSpace(data)))
	_ = rd
	positionFound := false
	var rn rune

	for err == nil {
		rn, _, err = rd.ReadRune()

		if err != nil {
			break
		}

		switch {
		case strings.IndexRune("1234567890&;:.", rn) >= 0:
			if !positionFound {
				positionBf.WriteRune(rn)
			} else {
				bf.WriteRune(rn)
			}
		default:
			positionFound = true
			bf.WriteRune(rn)
		}

	}

	if err != nil && err != io.EOF {
		return nil, err
	}

	pos := positionBf.String()
	if idx == 0 {
		if pos == "" {
			p.syncFirstThroughTemplateDefinition = true
			p.firstPos = 0
			pos = "1"
		} else {
			_, p.firstPos, err = PositionTo32th("", pos)
			if err != nil {
				return nil, err
			}
		}
	}

	//	fmt.Printf("making event for event no %v: %#v\n", idx, bf.String())
	ev, err = p.mkEvent(pos, posIn32th, bf.String())

	if err != nil {
		return nil, err
	}

	return ev, nil
}

// sliceEvents slices the events according to the template call slice definition
func (p *Call) sliceEvents(posIn32th uint) {
	//fmt.Printf("slice events for template %q in call %v\n", p.Name, p)
	if p.Slice[0] < 0 {
		return
	}
	// slice the items
	evs := []*callEvent{}
	first := p.Slice[0]
	last := p.Slice[1]

	i := 0
	collecting := false

	for _, ev := range p.Events {
		if i == first {
			collecting = true
		}
		if last > 0 && i >= last {
			collecting = false
		}
		if collecting {
			evs = append(evs, ev)
		}
		i++
	}

	if len(evs) > 0 && first > 0 {
		firstEv := evs[0]

		firstPos := firstEv.Position

		if firstPos > p.BarLength(1) {

			num := firstPos / p.BarLength(1)

			sub := p.BarLength(int(num))

			//			fmt.Printf("slicing %q\n", p.Name)

			nevs := []*callEvent{}

			for _, ev := range evs {
				//	fmt.Printf("ev: %s %q %v\n", ev.String(), ev.Position, ev.PositionIn32ths)
				nev := ev.dup()
				nev.Position = posIn32th + ev.Position - sub
				nev.PositionString = Pos32thToString(nev.Position)
				//fmt.Printf("nev: %s %q %v\n", nev.String(), nev.Position, nev.PositionIn32ths)
				nevs = append(nevs, nev)
			}

			p.Events = nevs
			return
		}
	}

	p.Events = evs
}

func (p *Call) findEventThatIsNotRest(start int) (found int) {
	if start >= len(p.Events) {
		return -1
	}

	if p.Events[start].Item != Rest {
		return start
	}

	return p.findEventThatIsNotRest(start + 1)
}

func (p *Call) ParseTemplate(data string, positionIn32th uint) error {
	return p.SimpleParse(data)
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

func (p *Call) __parseEvent(idx int, ev string, posIn32th uint, firstScaleNoteDiff int8) (newDiff int8, err error) {
	e, err := p.parseEvent(idx, ev, posIn32th)
	if err != nil {
		return firstScaleNoteDiff, fmt.Errorf("could not parse event %q: %s", ev, err.Error())
	}

	p.Events = append(p.Events, e)
	return firstScaleNoteDiff, nil

	switch v := e.Item.(type) {

	case *Call:
		if v.ScaleMoveMode == 2 {
			return firstScaleNoteDiff, fmt.Errorf("mounting of templates into a scale is not allowed inside template definitions, only in the score")
		}

		err := v.parseEvents(v.result, e.Position)
		if err != nil {
			return firstScaleNoteDiff, err
		}

		for _, evv := range v.Events {
			switch vv := evv.Item.(type) {
			case *Note:
				vv.Dynamic = p.AddDynamic(vv.Dynamic)
				evv.OriginalData = vv.String()
				evv.Item = vv
			case *MIDINote:
				vv.Dynamic = p.AddDynamic(vv.Dynamic)
				evv.OriginalData = vv.String()
				evv.Item = vv
			}
			p.Events = append(p.Events, evv)
		}
	case *Note:
		ee := e.dup()
		nt := v.Dup().(*Note)

		if p.FirstNoteIsScaleNote == 0 {
			if nt.ScaleNote != 0 {
				firstScaleNoteValue := nt.ScaleNote
				if firstScaleNoteValue > 0 {
					firstScaleNoteValue -= 1
				}

				// mount mode: p.scaleMove is the mount target for the first note
				target := p.ScaleMove
				if target > 0 {
					target -= 1
				}

				firstScaleNoteDiff = target - firstScaleNoteValue
				p.FirstNoteIsScaleNote = 1
			} else {
				p.FirstNoteAbsKey = nt.ToMIDI()
				p.FirstNoteIsScaleNote = -1
			}
		}

		if p.ScaleMoveMode == 1 && nt.ScaleNote != 0 {
			sn := nt.ScaleNote
			if sn > 0 {
				sn -= 1
			}

			// p.scaleMove simply shifts everything up or down (0 doesn't make any sense here)
			sn += p.ScaleMove

			if sn >= 0 {
				sn += 1
			}

			// fmt.Printf("shift scale note: %v -> + %v = %v\n", nt.scaleNote, p.scaleMove, sn)

			nt.ScaleNote = sn
		}

		if p.ScaleMoveMode == 2 {
			if p.FirstNoteIsScaleNote == -1 && nt.ScaleNote == 0 {
				// that has to be done when unrolling
			}
			if p.FirstNoteIsScaleNote == 1 && nt.ScaleNote != 0 {
				sn := nt.ScaleNote
				if sn > 0 {
					sn -= 1
				}

				// mount mode: p.scaleMove is the mount target for the first note

				sn += firstScaleNoteDiff

				if sn >= 0 {
					sn += 1
				}

				// fmt.Printf("mount scale note: %v + %v = %v\n", nt.scaleNote, firstScaleNoteDiff, sn)

				nt.ScaleNote = sn
			}
		}

		nt.Dynamic = p.AddDynamic(nt.Dynamic)
		ee.Item = nt
		ee.OriginalData = nt.String()
		p.Events = append(p.Events, ee)
	case *MIDINote:
		ee := e.dup()
		nt := v.Dup().(*MIDINote)
		nt.Dynamic = p.AddDynamic(nt.Dynamic)
		ee.Item = nt
		ee.OriginalData = nt.String()
		p.Events = append(p.Events, ee)
	default:
		p.Events = append(p.Events, e)
	}

	return firstScaleNoteDiff, nil
}

var TemplateReg = regexp.MustCompile(regexp.QuoteMeta("#") + "([0-9]+)")

func (p *Call) parseEvents(data string, posIn32th uint) (err error) {
	err = p.ParseEventsLine(data, p.PosShift, posIn32th)
	if err != nil {
		return err
	}

	// slice the items
	p.sliceEvents(posIn32th)

	return nil

	// do the replacement
	//return p.ReplaceEvents(posIn32th)
}

func (p *Call) BarLength(barIdx int) uint {
	return uint(math.Round((float64(barIdx) * float64(p.TimeSignature[0]) * 32) / float64(p.TimeSignature[1])))
}

func (p *Call) ParseEventsLine(data string, posshift int, posIn32th uint) (err error) {
	p.Events = []*callEvent{}
	p.syncFirstThroughTemplateDefinition = false

	//fmt.Printf("parseEvents called with data: %v\n", data)
	var firstScaleNoteDiff int8

	bars := strings.Split(data, "|")

	for barIdx, barEvents := range bars {

		var add32ths uint

		if p.TimeSignature[1] > 0 {
			add32ths = p.BarLength(barIdx)
		}

		sc := bufio.NewScanner(strings.NewReader(strings.TrimSpace(barEvents)))
		sc.Split(bufio.ScanWords)

		var idx = barIdx

		for sc.Scan() {
			ev := sc.Text()
			if ev != "" {
				firstScaleNoteDiff, err = p.__parseEvent(idx, ev, posIn32th+add32ths, firstScaleNoteDiff)
			}
			idx++
		}

		err := sc.Err()

		if err != nil {
			return err
		}
	}

	ev := []*callEvent{}

	lenevents := len(p.Events)

	for idx, e := range p.Events {
		if idx < lenevents-1 && p.Events[idx+1].Position <= e.Position {
			continue
		}

		switch it := e.Item.(type) {
		case *Note:
			if it.PosShift == 0 && posshift != 0 {
				it.PosShift = posshift
				e.Item = it
			}
		case *MIDINote:
			if it.PosShift == 0 && posshift != 0 {
				it.PosShift = posshift
				e.Item = it
			}
		}
		ev = append(ev, e)
	}

	p.Events = ev

	// fmt.Printf("posIn32th: %v events: %v\n", posIn32th, p.Events)

	return nil
}

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

func (p *Call) SimpleParse(call string) error {
	//fmt.Printf("SimpleParse(%q)\n", call)
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
		err := lp.SimpleParse(lyrics)
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

func (v Call) WriteMIDI(wr SMFWriter) (addedNotes []uint8) {
	panic("don't call me")
}

var regExTemplate0 = regexp.MustCompile("^([=]?[" + regexp.QuoteMeta("!") + "]?[a-zA-Z][a-zA-Z][" + regexp.QuoteMeta(".") + "_0-9a-zA-Z]*).*")
var regExTemplate1 = regexp.MustCompile("^(" + regexp.QuoteMeta("=") + "[" + regexp.QuoteMeta("!") + "]?" + regexp.QuoteMeta(".") + "[" + regexp.QuoteMeta(".") + "_0-9a-zA-Z]*).*")

func (pc *Call) Parse(data string, positionIn32th uint) (err error) {
	if pc.Parser == nil {
		panic("must not happen")
	}

	err = pc.ParseTemplate(data, positionIn32th)
	return
}
