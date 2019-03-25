package muskel

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

type PatternCall struct {
	Name         string
	Params       []string
	Replacements []string
	Slice        [2]int
	SyncFirst    bool
	result       string
	//	offset       uint
	firstPos                          uint
	Events                            []*positionedEvent // just a sausage of positions within one fictive infinite bar
	getter                            func(name string) *PatternDefinition
	velocityAdd                       string
	scaleMove                         int8 // 0: force to scale (take next note in scale if not exact matching, no movement for in scale notes), n >0 || n < 0: move by n steps along the scale
	scaleMoveMode                     int8 // 0: no scale movement, 1: move only scale notes, 2: move only scale notes or non scale notes depending on the first item
	firstNoteIsScaleNote              int  // 0: not set, 1: true, 2: false
	firstNoteAbsKey                   uint8
	syncFirstThroughPatternDefinition bool
}

func (p *PatternCall) String() string {
	var bf strings.Builder

	//	fmt.Printf("len events: %v\n", len(p.Events))

	for _, ev := range p.Events {
		//		fmt.Printf("event: position: %#v item: %#v\n", ev.position, ev.item)
		bf.WriteString(ev.String() + " ")
	}

	return bf.String()
}

type PatternEvents []*positionedEvent

// Spread is called unless the returned positionOfNextBar is < 0
// The first call to spread is with start number of the position within the bar where the
// pattern is embedded.
// Subsequent calls will set start to the last returned positionOfNextBar
// when the last event is return, positionOfNextBar is -1
// the returned events are relative to the bar given by start
func (p PatternEvents) Spread(start int, num, denom uint8) (barEvents []*positionedEvent, positionOfNextBar int) {
	barLength := length32ths(num, denom)
	positionOfNextBar = start + barLength

	var lastIdx int

	for idx, ev := range p {
		if int(ev.positionIn32ths) >= start && int(ev.positionIn32ths) < positionOfNextBar {
			nu := ev.dup()
			nu.positionIn32ths -= uint(start)
			nu.position = pos32thToString(nu.positionIn32ths)
			barEvents = append(barEvents, nu)
			lastIdx = idx
		}
	}

	if lastIdx == len(p)-1 {
		positionOfNextBar = -1
	}

	return
}

func (p *PatternCall) parseItem(data string, posIn32th uint) (item interface{}, err error) {
	if len(data) == 0 {
		return nil, nil
	}
	ip := &itemParser{}
	ip.GetDefinition = p.getter
	return ip.parseItem(data, posIn32th)
}

func (p *PatternCall) mkEvent(position string, posIn32th uint, data string) (ev *positionedEvent, err error) {
	ev = &positionedEvent{}
	_, ev.positionIn32ths, err = positionTo32th("", position)
	if err != nil {
		return nil, err
	}
	ev.positionIn32ths += posIn32th
	ev.originalData = data
	item, err := p.parseItem(ev.originalData, posIn32th)
	ev.item = item
	if p.SyncFirst || p.syncFirstThroughPatternDefinition {
		//		fmt.Printf("removing %v from %v\n", p.firstPos, ev.positionIn32ths)
		ev.positionIn32ths -= p.firstPos
	}
	ev.position = pos32thToString(ev.positionIn32ths)
	return
}

func (p *PatternCall) parseEvent(idx int, data string, posIn32th uint) (ev *positionedEvent, err error) {

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
			p.syncFirstThroughPatternDefinition = true
			p.firstPos = 0
			pos = "1"
		} else {
			_, p.firstPos, err = positionTo32th("", pos)
			if err != nil {
				return nil, err
			}
		}
	}

	//	fmt.Printf("making event for event no %v: %#v\n", idx, bf.String())
	//	panic("heyho")
	ev, err = p.mkEvent(pos, posIn32th, bf.String())

	if err != nil {
		return nil, err
	}

	return ev, nil
}

// sliceEvents slices the events according to the pattern call slice definition
func (p *PatternCall) sliceEvents() {
	if p.Slice[0] < 0 {
		return
	}
	// slice the items
	evs := []*positionedEvent{}
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

func (p *PatternCall) findEventThatIsNotRest(start int) (found int) {
	if start >= len(p.Events) {
		return -1
	}

	if p.Events[start].item != Rest {
		return start
	}

	return p.findEventThatIsNotRest(start + 1)
}

// replaceEvents replaces the events according to the pattern call replace definition
func (p *PatternCall) replaceEvents(posIn32th uint) (err error) {
	lenRepl := len(p.Replacements)

	if lenRepl == 0 {
		return nil
	}

	lenev := len(p.Events)
	//	panic("no replacements")

	//i := 0

	offset := 0

	for i, rpl := range p.Replacements {
		itidx := i + offset
		if itidx >= lenev {
			break
		}

		ev := p.Events[itidx]
		if rpl == ":" {
			if itidx == 0 {
				p.firstPos = ev.positionIn32ths
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

		position, item := replaceItemWith(rpl)

		if position == "" {
			position = ev.position
		}

		if item == "" {
			item = ev.originalData
		}

		if itidx == 0 {
			_, p.firstPos, err = positionTo32th("", position)
			if err != nil {
				return err
			}
		}

		var nu *positionedEvent

		nu, err = p.mkEvent(position, posIn32th, item)

		if err != nil {
			return err
		}

		p.Events[itidx] = nu

	}

	return nil
}

func (p *PatternCall) parsePattern(data string, positionIn32th uint) error {
	if p.getter == nil {
		return fmt.Errorf("need pattern getter")
	}

	//	fmt.Printf("parse pattern called with: %q\n", data)

	err := p.Parse(data)

	if err != nil {
		return fmt.Errorf("could not parse pattern call: %s", err)
	}

	def := p.getter(p.Name)

	if def == nil {
		return fmt.Errorf("could not find definition for pattern %q", p.Name)
	}

	p.result, err = def.Call(p.Params...)
	if err != nil {
		return fmt.Errorf("could not call pattern %s with %q: %s", p.Name, data, err)
	}

	//	fmt.Printf("result: %q\n", p.result)

	return p.parseEvents(p.result, positionIn32th)
}

func (p *PatternCall) addVelocity(orig int8) (vel int8) {
	if orig == -1 {
		return orig
	}

	if p.velocityAdd == "" {
		return orig
	}

	if p.velocityAdd == "=" {
		return dynamicToVelocity("=")
	}

	middle := dynamicToVelocity("=")

	diff := dynamicToVelocity(p.velocityAdd) - middle

	vel = orig + diff
	if vel > 127 {
		vel = 123
	}

	if vel < 5 {
		vel = 5
	}

	return
}

func (p *PatternCall) __parseEvent(idx int, ev string, posIn32th uint, firstScaleNoteDiff int8) (newDiff int8, err error) {
	e, err := p.parseEvent(idx, ev, posIn32th)
	if err != nil {
		return firstScaleNoteDiff, fmt.Errorf("could not parse event %q: %s", ev, err.Error())
	}

	//			fmt.Printf("e: %#v\n", e)

	//p.Events = append(p.Events, e)

	switch v := e.item.(type) {
	case *PatternCall:
		if v.scaleMoveMode == 2 {
			return firstScaleNoteDiff, fmt.Errorf("mounting of pattern into a scale is not allowed inside pattern definitions, only in the score")
		}
		err := v.parseEvents(v.result, e.positionIn32ths)
		if err != nil {
			return firstScaleNoteDiff, err
		}

		for _, evv := range v.Events {
			switch vv := evv.item.(type) {
			case Note:
				vv.velocity = p.addVelocity(vv.velocity)
				evv.originalData = vv.String()
				evv.item = vv
			case MIDINote:
				vv.velocity = p.addVelocity(vv.velocity)
				evv.originalData = vv.String()
				evv.item = vv
			}
			p.Events = append(p.Events, evv)
		}
		//p.Events = append(p.Events, v.Events...)
	case Note:
		ee := e.dup()
		nt := v.Dup()

		if p.firstNoteIsScaleNote == 0 {
			if nt.scaleNote != 0 {
				firstScaleNoteValue := nt.scaleNote
				if firstScaleNoteValue > 0 {
					firstScaleNoteValue -= 1
				}

				// mount mode: p.scaleMove is the mount target for the first note
				target := p.scaleMove
				if target > 0 {
					target -= 1
				}

				firstScaleNoteDiff = target - firstScaleNoteValue
				p.firstNoteIsScaleNote = 1
			} else {
				p.firstNoteAbsKey = nt.toMIDI()
				p.firstNoteIsScaleNote = -1
			}
		}

		if p.scaleMoveMode == 1 && nt.scaleNote != 0 {
			sn := nt.scaleNote
			if sn > 0 {
				sn -= 1
			}

			// p.scaleMove simply shifts everything up or down (0 doesn't make any sense here)
			sn += p.scaleMove

			if sn >= 0 {
				sn += 1
			}

			// fmt.Printf("shift scale note: %v -> + %v = %v\n", nt.scaleNote, p.scaleMove, sn)

			nt.scaleNote = sn
		}

		if p.scaleMoveMode == 2 {
			if p.firstNoteIsScaleNote == -1 && nt.scaleNote == 0 {
				// TODO that has to be done when unrolling
			}
			if p.firstNoteIsScaleNote == 1 && nt.scaleNote != 0 {
				sn := nt.scaleNote
				if sn > 0 {
					sn -= 1
				}

				// mount mode: p.scaleMove is the mount target for the first note

				sn += firstScaleNoteDiff

				if sn >= 0 {
					sn += 1
				}

				// fmt.Printf("mount scale note: %v + %v = %v\n", nt.scaleNote, firstScaleNoteDiff, sn)

				nt.scaleNote = sn
			}
		}

		nt.velocity = p.addVelocity(nt.velocity)
		ee.item = nt
		ee.originalData = nt.String()
		p.Events = append(p.Events, ee)
	case MIDINote:
		ee := e.dup()
		nt := v.Dup()
		nt.velocity = p.addVelocity(nt.velocity)
		ee.item = nt
		ee.originalData = nt.String()
		p.Events = append(p.Events, ee)
	default:
		p.Events = append(p.Events, e)
	}

	return firstScaleNoteDiff, nil
}

func (p *PatternCall) parseEvents(data string, posIn32th uint) (err error) {
	p.Events = []*positionedEvent{}
	p.syncFirstThroughPatternDefinition = false

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

	ev := []*positionedEvent{}

	lenevents := len(p.Events)

	for idx, e := range p.Events {
		if idx < lenevents-1 && p.Events[idx+1].positionIn32ths <= e.positionIn32ths {
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
var regPatternCallNameDyn = regexp.MustCompile("^([a-zA-Z][_~a-zA-Z0-9]+)(" + regexp.QuoteMeta("^") + "{1,2}[-0-9]+){0,1}([" + regexp.QuoteMeta("-+=") + "]*)$")

func (p *PatternCall) Parse(call string) error {
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

	mt := regPatternCallNameDyn.FindStringSubmatch(call)

	p.Name = mt[1]

	if mt[2] != "" {
		dt := mt[2][1:]
		p.scaleMoveMode = 1 // move only scale notes
		if dt[0] == '^' {
			p.scaleMoveMode = 2 // move depending on first note
			dt = dt[1:]
		}

		si, err := strconv.Atoi(dt)
		if err != nil {
			return fmt.Errorf("error in scale moving of pattern %q: %q is not a number", p.Name, dt)
		}

		p.scaleMove = int8(si)
	}

	p.velocityAdd = mt[3]

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
			return fmt.Errorf("ERROR in call of pattern %q: invalid slice %q", p.Name, "["+slice+"]")
		}

		from := strings.TrimSpace(sl[0])
		to := strings.TrimSpace(sl[1])

		if from == "" {
			p.Slice[0] = 0
		} else {
			fromI, err := strconv.Atoi(from)
			if err != nil {
				return fmt.Errorf("ERROR in call of pattern %q: invalid slice %q", p.Name, "["+slice+"]")
			}
			p.Slice[0] = fromI
		}

		if to != "" {
			toI, err := strconv.Atoi(to)
			if err != nil || toI == 0 {
				return fmt.Errorf("ERROR in call of pattern %q: invalid slice %q", p.Name, "["+slice+"]")
			}
			p.Slice[1] = toI
		}

	}
	return nil
}
