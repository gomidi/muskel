package muskel

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
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
	firstPos uint
	Events   []*positionedEvent // just a sausage of positions within one fictive infinite bar
	getter   func(name string) *PatternDefinition
}

func (p *PatternCall) String() string {
	var bf bytes.Buffer

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
	ip := &itemParser{}
	ip.GetDefinition = p.getter
	return ip.parseItem(data, posIn32th)
}

func (p *PatternCall) parseItems(data string, posIn32th uint) (item interface{}, err error) {
	if len(data) == 0 {
		return nil, nil
	}

	//	fmt.Printf("PatternCall#parseItems called with %q\n", data)

	switch data[0] {
	case '(':
		var items MultiItem
		data = strings.Trim(data, "()")
		its := strings.Split(data, ",")

		for _, it := range its {
			it = strings.TrimSpace(it)

			if it != "" {
				ite, err := p.parseItem(it, posIn32th)
				if err != nil {
					return nil, err
				}
				items = append(items, ite)
			}
		}

		return items, nil

	default:
		return p.parseItem(data, posIn32th)
	}

	return
}

func (p *PatternCall) mkEvent(position string, posIn32th uint, data string) (ev *positionedEvent, err error) {
	ev = &positionedEvent{}
	_, ev.positionIn32ths = positionTo32th("", position)
	ev.positionIn32ths += posIn32th
	ev.originalData = data
	item, err := p.parseItems(ev.originalData, posIn32th)
	ev.item = item
	if p.SyncFirst {
		//		fmt.Printf("removing %v from %v\n", p.firstPos, ev.positionIn32ths)
		ev.positionIn32ths -= p.firstPos
	}
	ev.position = pos32thToString(ev.positionIn32ths)
	return
}

func (p *PatternCall) parseEvent(idx int, data string, posIn32th uint) (ev *positionedEvent, err error) {

	var positionBf bytes.Buffer
	var bf bytes.Buffer
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
		case strings.IndexRune("123456789&;:.", rn) >= 0:
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
		_, p.firstPos = positionTo32th("", pos)
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

// replaceEvents replaces the events according to the pattern call replace definition
func (p *PatternCall) replaceEvents(posIn32th uint) error {
	lenRepl := len(p.Replacements)

	if lenRepl == 0 {
		return nil
	}
	//	panic("no replacements")

	i := 0

	for itidx, ev := range p.Events {

		if i < lenRepl && p.Replacements[i] != ":" {
			position, item := replaceItemWith(p.Replacements[i])

			if position == "" {
				position = ev.position
			}

			if item == "" {
				item = ev.originalData
			}

			if itidx == 0 {
				_, p.firstPos = positionTo32th("", position)
			}

			nu, err := p.mkEvent(position, posIn32th, item)

			if err != nil {
				return err
			}

			/*
				if itidx == 0 {
					firstPos = nu.positionIn32ths
				}
			*/
			p.Events[itidx] = nu

		} else {
			if itidx == 0 {
				p.firstPos = ev.positionIn32ths
			}
		}
		i++
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
		return fmt.Errorf("could not find definition for pattern %s", p.Name)
	}

	p.result, err = def.Call(p.Params...)
	if err != nil {
		return fmt.Errorf("could not call pattern %s with %q: %s", p.Name, data, err)
	}

	//	fmt.Printf("result: %q\n", p.result)

	return p.parseEvents(p.result, positionIn32th)
}

func (p *PatternCall) parseEvents(data string, posIn32th uint) error {
	p.Events = []*positionedEvent{}

	//fmt.Printf("parseEvents called with data: %v\n", data)

	evts := strings.Split(strings.TrimSpace(data), " ")

	for idx, ev := range evts {
		ev = strings.TrimSpace(ev)

		if ev != "" {
			e, err := p.parseEvent(idx, ev, posIn32th)
			if err != nil {
				return fmt.Errorf("could not parse event %q: %s", ev, err.Error())
			}

			//			fmt.Printf("e: %#v\n", e)

			//p.Events = append(p.Events, e)

			switch v := e.item.(type) {
			case *PatternCall:
				err := v.parseEvents(v.result, e.positionIn32ths)
				if err != nil {
					return err
				}
				p.Events = append(p.Events, v.Events...)
			default:
				p.Events = append(p.Events, e)
			}

		}
	}

	//fmt.Printf("len events: %v\n", len(p.Events))

	// slice the items
	p.sliceEvents()

	// do the replacement
	return p.replaceEvents(posIn32th)
}

func (p *PatternCall) Parse(call string) error {
	replacements := ""
	slice := ""
	params := ""

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

	p.Name = call

	if p.Name[0] == '\\' {
		p.SyncFirst = true
		p.Name = p.Name[1:]
	}

	if replacements != "" {
		repl := strings.Split(replacements, " ")

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
