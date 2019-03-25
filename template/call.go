package template

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"gitlab.com/gomidi/muskel/items"
)

type PositionedEvent struct {
	Position     string
	OriginalData string
	//barNo           int
	PositionIn32ths uint
	Item            interface{}
}

func (p *PositionedEvent) dup() (nu *PositionedEvent) {
	return &PositionedEvent{
		Position:        p.Position,
		OriginalData:    p.OriginalData,
		PositionIn32ths: p.PositionIn32ths,
		Item:            p.Item,
	}
}

func (p *PositionedEvent) String() string {
	//return p.position + p.originalData
	return items.Pos32thToString(p.PositionIn32ths) + p.OriginalData
}

type Call struct {
	Name         string
	Params       []string
	Replacements []string
	Slice        [2]int
	SyncFirst    bool
	result       string
	//	offset       uint
	firstPos                           uint
	Events                             []*PositionedEvent // just a sausage of positions within one fictive infinite bar
	Getter                             func(name string) *Definition
	VelocityAdd                        string
	ScaleMove                          int8 // 0: force to scale (take next note in scale if not exact matching, no movement for in scale notes), n >0 || n < 0: move by n steps along the scale
	ScaleMoveMode                      int8 // 0: no scale movement, 1: move only scale notes, 2: move only scale notes or non scale notes depending on the first item
	FirstNoteIsScaleNote               int  // 0: not set, 1: true, 2: false
	FirstNoteAbsKey                    uint8
	syncFirstThroughTemplateDefinition bool
	ItemParser                         ItemParser
}

func NewTemplateCall(ip ItemParser) *Call {
	return &Call{
		ItemParser: ip,
	}
}

func (p *Call) String() string {
	var bf strings.Builder

	//	fmt.Printf("len events: %v\n", len(p.Events))

	for _, ev := range p.Events {
		//		fmt.Printf("event: position: %#v item: %#v\n", ev.position, ev.item)
		bf.WriteString(ev.String() + " ")
	}

	return bf.String()
}

type TemplateEvents []*PositionedEvent

// Spread is called unless the returned positionOfNextBar is < 0
// The first call to spread is with start number of the position within the bar where the
// template is embedded.
// Subsequent calls will set start to the last returned positionOfNextBar
// when the last event is return, positionOfNextBar is -1
// the returned events are relative to the bar given by start
func (p TemplateEvents) Spread(start int, num, denom uint8) (barEvents []*PositionedEvent, positionOfNextBar int) {
	barLength := items.Length32ths(num, denom)
	positionOfNextBar = start + barLength

	var lastIdx int

	for idx, ev := range p {
		if int(ev.PositionIn32ths) >= start && int(ev.PositionIn32ths) < positionOfNextBar {
			nu := ev.dup()
			nu.PositionIn32ths -= uint(start)
			nu.Position = items.Pos32thToString(nu.PositionIn32ths)
			barEvents = append(barEvents, nu)
			lastIdx = idx
		}
	}

	if lastIdx == len(p)-1 {
		positionOfNextBar = -1
	}

	return
}

type ItemParser interface {
	//GetDefinition(name string) *TemplateDefinition
	SetDefinitionGetter(func(name string) *Definition)
	ParseItem(data string, posIn32th uint) (item interface{}, err error)
}

func (p *Call) parseItem(data string, posIn32th uint) (item interface{}, err error) {
	if len(data) == 0 {
		return nil, nil
	}

	if p.ItemParser == nil {
		panic("set ItemParser first")
	}

	p.ItemParser.SetDefinitionGetter(p.Getter)
	return p.ItemParser.ParseItem(data, posIn32th)

	// return

	/*
		ip := &itemParser{}
		ip.GetDefinition = p.getter
		return ip.parseItem(data, posIn32th)
	*/
}

func (p *Call) mkEvent(position string, posIn32th uint, data string) (ev *PositionedEvent, err error) {
	ev = &PositionedEvent{}
	_, ev.PositionIn32ths, err = items.PositionTo32th("", position)
	if err != nil {
		return nil, err
	}
	ev.PositionIn32ths += posIn32th
	ev.OriginalData = data
	item, err := p.parseItem(ev.OriginalData, posIn32th)
	ev.Item = item
	if p.SyncFirst || p.syncFirstThroughTemplateDefinition {
		//		fmt.Printf("removing %v from %v\n", p.firstPos, ev.positionIn32ths)
		ev.PositionIn32ths -= p.firstPos
	}
	ev.Position = items.Pos32thToString(ev.PositionIn32ths)
	return
}

func (p *Call) parseEvent(idx int, data string, posIn32th uint) (ev *PositionedEvent, err error) {

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
			_, p.firstPos, err = items.PositionTo32th("", pos)
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
func (p *Call) sliceEvents() {
	if p.Slice[0] < 0 {
		return
	}
	// slice the items
	evs := []*PositionedEvent{}
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
	p.Events = evs
}

func (p *Call) findEventThatIsNotRest(start int) (found int) {
	if start >= len(p.Events) {
		return -1
	}

	if p.Events[start].Item != items.Rest {
		return start
	}

	return p.findEventThatIsNotRest(start + 1)
}

// replaceEvents replaces the events according to the template call replace definition
func (p *Call) replaceEvents(posIn32th uint) (err error) {
	lenRepl := len(p.Replacements)

	if lenRepl == 0 {
		return nil
	}

	lenev := len(p.Events)

	offset := 0

	for i, rpl := range p.Replacements {
		itidx := i + offset
		if itidx >= lenev {
			break
		}

		ev := p.Events[itidx]
		if rpl == ":" {
			if itidx == 0 {
				p.firstPos = ev.PositionIn32ths
			}
			continue
		}

		if rpl[0] == '!' {
			//offset += 1
			found := p.findEventThatIsNotRest(itidx)
			if found == -1 {
				break
			}
			ev = p.Events[found]
			offset += found - itidx
			itidx = found
			rpl = rpl[1:]
		}

		position, item := items.ReplaceItemWith(rpl)

		if position == "" {
			position = ev.Position
		}

		if item == "" {
			item = ev.OriginalData
		}

		if itidx == 0 {
			_, p.firstPos, err = items.PositionTo32th("", position)
			if err != nil {
				return err
			}
		}

		var nu *PositionedEvent

		nu, err = p.mkEvent(position, posIn32th, item)

		if err != nil {
			return err
		}

		p.Events[itidx] = nu

	}

	return nil
}

func (p *Call) ParseTemplate(data string, positionIn32th uint) error {
	if p.Getter == nil {
		return fmt.Errorf("need template getter")
	}

	//	fmt.Printf("parse template called with: %q\n", data)

	err := p.Parse(data)

	if err != nil {
		return fmt.Errorf("could not parse template call: %s", err)
	}

	def := p.Getter(p.Name)

	if def == nil {
		return fmt.Errorf("could not find definition for template %q", p.Name)
	}

	p.result, err = def.Call(p.Params...)
	if err != nil {
		return fmt.Errorf("could not call template %s with %q: %s", p.Name, data, err)
	}

	//	fmt.Printf("result: %q\n", p.result)

	return p.parseEvents(p.result, positionIn32th)
}

func (p *Call) addVelocity(orig int8) (vel int8) {
	if orig == -1 {
		return orig
	}

	if p.VelocityAdd == "" {
		return orig
	}

	if p.VelocityAdd == "=" {
		return items.DynamicToVelocity("=")
	}

	middle := items.DynamicToVelocity("=")

	diff := items.DynamicToVelocity(p.VelocityAdd) - middle

	vel = orig + diff
	if vel > 127 {
		vel = 123
	}

	if vel < 5 {
		vel = 5
	}

	return
}

func (p *Call) __parseEvent(idx int, ev string, posIn32th uint, firstScaleNoteDiff int8) (newDiff int8, err error) {
	e, err := p.parseEvent(idx, ev, posIn32th)
	if err != nil {
		return firstScaleNoteDiff, fmt.Errorf("could not parse event %q: %s", ev, err.Error())
	}

	//			fmt.Printf("e: %#v\n", e)

	//p.Events = append(p.Events, e)

	switch v := e.Item.(type) {
	case *Call:
		if v.ScaleMoveMode == 2 {
			return firstScaleNoteDiff, fmt.Errorf("mounting of templates into a scale is not allowed inside template definitions, only in the score")
		}
		err := v.parseEvents(v.result, e.PositionIn32ths)
		if err != nil {
			return firstScaleNoteDiff, err
		}

		for _, evv := range v.Events {
			switch vv := evv.Item.(type) {
			case items.Note:
				vv.Velocity = p.addVelocity(vv.Velocity)
				evv.OriginalData = vv.String()
				evv.Item = vv
			case items.MIDINote:
				vv.Velocity = p.addVelocity(vv.Velocity)
				evv.OriginalData = vv.String()
				evv.Item = vv
			}
			p.Events = append(p.Events, evv)
		}
		//p.Events = append(p.Events, v.Events...)
	case items.Note:
		ee := e.dup()
		nt := v.Dup()

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

		nt.Velocity = p.addVelocity(nt.Velocity)
		ee.Item = nt
		ee.OriginalData = nt.String()
		p.Events = append(p.Events, ee)
	case items.MIDINote:
		ee := e.dup()
		nt := v.Dup()
		nt.Velocity = p.addVelocity(nt.Velocity)
		ee.Item = nt
		ee.OriginalData = nt.String()
		p.Events = append(p.Events, ee)
	default:
		p.Events = append(p.Events, e)
	}

	return firstScaleNoteDiff, nil
}

var templateReg = regexp.MustCompile(regexp.QuoteMeta("#") + "([1-9]+)")

func (p *Call) parseEvents(data string, posIn32th uint) (err error) {
	p.Events = []*PositionedEvent{}
	p.syncFirstThroughTemplateDefinition = false

	//fmt.Printf("parseEvents called with data: %v\n", data)
	var firstScaleNoteDiff int8

	sc := bufio.NewScanner(strings.NewReader(data))
	sc.Split(bufio.ScanWords)

	var idx int

	for sc.Scan() {
		ev := sc.Text()
		if ev != "" {
			firstScaleNoteDiff, err = p.__parseEvent(idx, ev, posIn32th, firstScaleNoteDiff)
		}
		idx++
	}

	ev := []*PositionedEvent{}

	lenevents := len(p.Events)

	for idx, e := range p.Events {
		if idx < lenevents-1 && p.Events[idx+1].PositionIn32ths <= e.PositionIn32ths {
			continue
		}
		ev = append(ev, e)
	}

	p.Events = ev

	//fmt.Printf("len events: %v\n", len(p.Events))

	// slice the items
	p.sliceEvents()

	// do the replacement
	return p.replaceEvents(posIn32th)
}

//var regPatternCallNameDyn =
var regTemplateCallNameDyn = regexp.MustCompile("^([a-zA-Z][_~a-zA-Z0-9]+)(" + regexp.QuoteMeta("^") + "{1,2}[-0-9]+){0,1}([" + regexp.QuoteMeta("-+=") + "]*)$")

func (p *Call) Parse(call string) error {
	replacements := ""
	slice := ""
	params := ""

	//fmt.Printf("call: %q\n", call)

	if idx := strings.Index(call, "/"); idx > 0 {
		replacements = strings.TrimSpace(strings.Trim(call[idx:], "/"))
		call = call[:idx]
	}

	if idx := strings.Index(call, "["); idx > 0 {
		slice = strings.TrimSpace(strings.Trim(call[idx:], "[]"))
		call = call[:idx]
	}

	if idx := strings.Index(call, "("); idx > 0 {
		params = strings.TrimSpace(strings.Trim(call[idx:], "()"))
		call = call[:idx]
	}

	if call[0] == '!' {
		p.SyncFirst = true
		call = call[1:]
	}

	//fmt.Printf("replacements: %q\n", replacements)

	mt := regTemplateCallNameDyn.FindStringSubmatch(call)

	p.Name = mt[1]

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

	p.VelocityAdd = mt[3]

	if replacements != "" {
		repl := strings.Split(replacements, ",")

		for _, rep := range repl {
			rep = strings.TrimSpace(rep)

			if rep != "" {
				p.Replacements = append(p.Replacements, rep)
			}
		}
	}

	if params != "" {
		ps := strings.Split(params, ",")

		for _, param := range ps {
			param = strings.TrimSpace(param)

			if param != "" {
				p.Params = append(p.Params, param)
			}
		}
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
