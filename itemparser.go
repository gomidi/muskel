package muskel

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var regexMIDINote = regexp.MustCompile("^([0-9]{1,3})([-+" + regexp.QuoteMeta("=") + "]*)(:{0,1})([<>]{0,1})$")

func parseMIDINote(data string) (nt MIDINote, err error) {
	mt := regexMIDINote.FindStringSubmatch(data)
	if len(mt) < 2 {
		return nt, fmt.Errorf("not a valid MIDINote: %q", data)
	}
	//	fmt.Printf("parse MIDInotes: %#v\n", mt)
	i, err := strconv.Atoi(mt[1])
	if err != nil {
		return nt, fmt.Errorf("not a valid MIDINote: %q: %s", data, err.Error())
	}
	if i < 1 && i > 128 {
		return nt, fmt.Errorf("not a valid MIDINote: %q: must be >= 1 and <= 128", data)
	}
	nt.note = int8(i)
	nt.velocity = dynamicToVelocity(mt[2])
	nt.dotted = mt[3] == ":"

	if mt[4] == ">" {
		nt.posShift = 1
	}

	if mt[4] == "<" {
		nt.posShift = -1
	}
	return nt, nil
}

func parseMIDICC(data string) (cc MIDICC, err error) {
	var ccval int = -1

	if len(data) > 2 {
		if strings.ToLower(data[0:2]) == "on" {
			ccval = 127
			data = data[2:]
		}

		if strings.ToLower(data[0:3]) == "off" {
			ccval = 0
			data = data[3:]
		}
	}

	if idx := strings.Index(data, "~"); idx > -1 {
		cc.tilde = data[idx:]
		data = data[:idx]
	}

	if idx := strings.Index(data, ":"); idx > -1 {
		cc.dotted = true
		data = data[:idx]
	}

	dt := strings.Trim(data, "()")
	d := strings.Split(dt, ",")
	if ccval == -1 && len(d) != 2 {
		err = fmt.Errorf("invalid format for MIDI CC: %#v, must be CC(nn,mm) where nn is the controller number and mm is the controler value", data)
		return
	}

	var ccnum int
	ccnum, err = strconv.Atoi(d[0])

	if err != nil {
		return
	}

	if ccval > -1 {
		cc.controller = uint8(ccnum)
		cc.value = uint8(ccval)
		return
	}

	ccval, err = strconv.Atoi(d[1])

	if err != nil {
		return
	}

	cc.controller = uint8(ccnum)
	cc.value = uint8(ccval)

	return
}

func parseMIDIPitchbend(data string) (pb MIDIPitchbend, err error) {
	if idx := strings.Index(data, "~"); idx > -1 {
		pb.tilde = data[idx:]
		data = data[:idx]
	}

	if idx := strings.Index(data, ":"); idx > -1 {
		pb.dotted = true
		data = data[:idx]
	}

	dt := strings.Trim(data, "()")
	var val int
	val, err = strconv.Atoi(dt)

	if err != nil {
		return
	}

	// fmt.Printf("parsing pitchbend %q: %v, normalized: %v\n", dt, val, int16(val))

	pb.value = int16(val)
	return
}

func parseMIDIAftertouch(data string) (at MIDIAftertouch, err error) {
	if idx := strings.Index(data, "~"); idx > -1 {
		at.tilde = data[idx:]
		data = data[:idx]
	}

	if idx := strings.Index(data, ":"); idx > -1 {
		at.dotted = true
		data = data[:idx]
	}

	dt := strings.Trim(data, "()")
	var val int
	val, err = strconv.Atoi(dt)

	if err != nil {
		return
	}

	at.value = uint8(val)
	return
}

func parseMIDIPolyAftertouch(data string) (pt MIDIPolyAftertouch, err error) {
	if idx := strings.Index(data, "~"); idx > -1 {
		pt.tilde = data[idx:]
		data = data[:idx]
	}

	if idx := strings.Index(data, ":"); idx > -1 {
		pt.dotted = true
		data = data[:idx]
	}

	dt := strings.Trim(data, "()")
	d := strings.Split(dt, ",")
	if len(d) != 2 {
		err = fmt.Errorf("invalid format for MIDI PT: %#v, must be PT(nt,mm) where nt is the note and mm is the value", data)
		return
	}

	var nt Note
	nt, err = parseNote(d[0])
	if err != nil {
		err = fmt.Errorf("invalid format for MIDI PT can't parse note: %s", err.Error())
		return
	}

	var val int
	val, err = strconv.Atoi(d[1])

	if err != nil {
		err = fmt.Errorf("invalid format for MIDI PT can't parse value: %s", err.Error())
		return
	}

	pt.key = nt.toMIDI()
	pt.value = uint8(val)

	return
}

var regScaleNote = regexp.MustCompile("^(-){0,1}([0-9]+)(.*)$")

func parseScaleNote(data string) (nt Note, err error) {
	mt := regScaleNote.FindStringSubmatch(data)
	// fmt.Printf("scale note %q match: %#v\n", data, mt)
	var i int
	i, err = strconv.Atoi(mt[1] + mt[2])

	if err != nil {
		err = fmt.Errorf("invalid scale note: %#v", data)
		return
	}

	nt.scaleNote = int8(i)

	var dynamic string

	for _, l := range mt[3] {
		switch l {
		case ':':
			nt.dotted = true
		case '+':
			dynamic += "+"
		case '-':
			dynamic += "-"
		case '=':
			dynamic = "="
		case '~':
			nt.glissandoStart = true
		case '#':
			nt.augmenter += "#"
		case '^':
			nt.augmenter += "^"
		case '°':
			nt.augmenter += "°"
		case '>':
			nt.posShift = 1
		case '<':
			nt.posShift = -1
		/*
			case '#':
				nt.augmenter = "#"
			case '^':
				nt.augmenter = "^"
			case '°':
				nt.augmenter = "°"
			case '"':
				if nt.octave > 0 {
					nt.octave += 2
				}

				if nt.octave < 0 {
					nt.octave -= 2
				}
			case '\'':
				if nt.octave > 0 {
					nt.octave += 1
				}

				if nt.octave < 0 {
					nt.octave -= 1
				}
		*/
		default:
			err = fmt.Errorf("invalid scale note: %#v", data)
			return
		}
	}

	nt.velocity = dynamicToVelocity(dynamic)

	return
}

func parseNote(data string) (nt Note, err error) {

	switch data[:1] {
	case "a", "b", "c", "d", "e", "f", "g":
		nt.letter = data[:1]
		nt.octave = 1
	case "A", "B", "C", "D", "E", "F", "G":
		nt.letter = strings.ToLower(data[:1])
		nt.octave = -1
	default:
		err = fmt.Errorf("invalid note: %#v", data)
		return
	}

	var dynamic string
	var gliss int

	for _, l := range data[1:] {
		switch l {
		case ':':
			nt.dotted = true
		case '+':
			dynamic += "+"
		case '-':
			dynamic += "-"
		case '=':
			dynamic = "="
		case '~':
			gliss++
		case '>':
			nt.posShift = 1
		case '<':
			nt.posShift = -1
		case '#':
			nt.augmenter += "#"
		case '^':
			nt.augmenter += "^"
		case '°':
			nt.augmenter += "°"
		case '"':
			if nt.octave > 0 {
				nt.octave += 2
			}

			if nt.octave < 0 {
				nt.octave -= 2
			}
		case '\'':
			if nt.octave > 0 {
				nt.octave += 1
			}

			if nt.octave < 0 {
				nt.octave -= 1
			}
		default:
			err = fmt.Errorf("invalid note: %#v", data)
			return
		}
	}

	if gliss > 0 {
		nt.glissandoStart = true
	}

	if gliss > 1 {
		nt.glissandoExp = true
	}

	nt.velocity = dynamicToVelocity(dynamic)

	return
}

// TODO implement parseOSC
func (p *itemParser) parseOSC(data string) (om OSCMessage, err error) {
	return om, fmt.Errorf("parseOSC is not implemented yet")
}

func (p *itemParser) parsePattern(data string, positionIn32th uint) (pc *PatternCall, err error) {
	if p.GetDefinition == nil {
		return nil, fmt.Errorf("no definition resolver defined")
	}
	pc = &PatternCall{}
	pc.getter = p.GetDefinition
	err = pc.parsePattern(data, positionIn32th)
	return
}

// ntuple has the form {c,e,d}3&
// where 3& is the ending position that defines the total length
func (p *itemParser) parseNTuple(data string, posIn32th uint) (nt NTuple, err error) {

	var ntp NTuple
	orig := data

	dd := strings.Split(data, "}")
	if len(dd) != 2 {
		return nt, fmt.Errorf("invalid n-tuple: %#v", "{"+orig)
	}

	pos := strings.TrimSpace(dd[1])

	if strings.HasSuffix(pos, ">") {
		ntp.posShift = 1
		pos = strings.TrimRight(pos, ">")
	}

	if strings.HasSuffix(pos, "<") {
		ntp.posShift = -1
		pos = strings.TrimRight(pos, "<")
	}

	// fmt.Printf("pos: %q postShift: %v\n", pos, ntp.posShift)

	if len(pos) == 0 {
		err = fmt.Errorf("invalid n-tuple %#v: position missing", "{"+orig)
		return
	}

	_, endPos, err := positionTo32th("", pos)

	if err != nil {
		return nt, err
	}

	data = strings.TrimSpace(dd[0])

	d := strings.Split(data, ",")

	if len(d) < 2 {
		err = fmt.Errorf("empty or single item n-tuples are not allowed: %#v", "{"+orig)
		return
	}

	ntp.endPos = endPos

	for ppos, it := range d {
		itt, err := p.parseItem(it, posIn32th)
		if err != nil {
			return nt, fmt.Errorf("ERROR invalid n-tuple at position %v: %#v: %s", ppos, it, err)
		}
		ntp.items = append(ntp.items, itt)
	}
	return ntp, nil
}

type itemParser struct {
	GetDefinition func(name string) *PatternDefinition
}

func (p *itemParser) parseRandom(data string, posIn32th uint) (item interface{}, err error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("invalid random value: ?%s", data)
	}

	// strings.IndexAny()

	if data[0] == '(' {
		alternatives := strings.Trim(data, "()")
		alt := strings.Split(alternatives, ",")
		var r RandomChooser
		for _, a := range alt {
			a = strings.TrimSpace(a)
			if a != "" {
				item, err = p.parseItem(a, posIn32th)

				if err != nil {
					return nil, fmt.Errorf("invalid random value item: %s", a)
				}
				r.alternatives = append(r.alternatives, item)
				r.alternativesOriginalData = append(r.alternativesOriginalData, a)
			}
		}
		return &r, nil
	}

	idx := strings.Index(data, "%")
	if idx < 0 {
		return nil, fmt.Errorf("invalid random value: ?%s", data)
	}

	num := data[:idx]

	var rp RandomProbability
	var n int
	n, err = strconv.Atoi(num)

	if err != nil {
		return nil, fmt.Errorf("invalid random value: ?%s, syntax must be ?n%% where n is a number between 0 and 100", data)
	}

	rp.prob = uint8(n)

	if rp.prob < 0 || rp.prob > 100 {
		return nil, fmt.Errorf("invalid random value: ?%s, syntax must be ?n%% where n is a number between 0 and 100, but number is %v", data, rp.prob)
	}

	rp.item, err = p.parseItem(data[idx+1:], posIn32th)

	if err != nil {
		return nil, fmt.Errorf("invalid random value item: ?%s", data)
	}
	rp.itemOriginalData = data[idx+1:]
	return &rp, nil

}

func (p *itemParser) parseCommand(data string, posIn32th uint) (interface{}, error) {
	panic("TODO implement")
}

func (p *itemParser) parseGlissando(data string, posIn32th uint) (interface{}, error) {
	if data == "" {
		return GlissandoLinear, nil
	}

	if data == "~" {
		return GlissandoExponential, nil
	}
	return nil, fmt.Errorf("invalid glissando: \"~%s\"", data)
}

func (p *itemParser) parseItem(data string, posIn32th uint) (interface{}, error) {
	data = strings.TrimSpace(data)
	// fmt.Printf("parseItem called with %q\n", data)
	switch len(data) {
	case 0:
		return nil, nil
	case 1:
		switch data[0] {
		case '%':
			return RepeatLastEvent{}, nil
		case ':':
			return Hold, nil
		case '*':
			return Rest, nil
		case '~':
			return p.parseGlissando("", posIn32th)
		default:
			res, err := parseNote(data)
			if err != nil {
				return nil, fmt.Errorf("unknown item: %#v", data)
			}
			return res, nil
		}
	default:
		//fmt.Printf("data[0]: %#v\n", string(data[0]))
		switch data[0] {
		case '~':
			return p.parseGlissando(data[1:], posIn32th)
		case '"':
			return Lyric(strings.Trim(data, `"`)), nil
		case '{':
			return p.parseNTuple(data[1:], posIn32th)
		case '$':
			return p.parseCommand(data[1:], posIn32th)
		case '_':
			endIdx := strings.LastIndex(data, "_")
			if endIdx < 4 {
				return nil, fmt.Errorf("invalid MultiItem: %q must start and end with underscore _", data)
			}

			if endIdx != len(data)-1 {
				return nil, fmt.Errorf("invalid MultiItem: %q must start and end with underscore _", data)
			}
			var m MultiItem
			d := strings.Split(data[1:endIdx], "_")
			for _, dd := range d {
				it, err := p.parseItem(dd, posIn32th)

				if err != nil {
					return nil, err
				}

				m = append(m, it)
			}
			return m, nil
			//return nil, fmt.Errorf("deprecated syntax: use  : instead to create a hold")
		//case 'S':
		//	return parseNote(data[1:])
		//case 'Z':
		//	return parseNote(data[1:])
		case '^':
			return parseScaleNote(data[1:])
		case '?':
			return p.parseRandom(data[1:], posIn32th)
		//case 'O':
		//	return p.parseOSC(data[1:])
		case '!':
			return p.parsePattern(data, posIn32th)
		case '.':
			if data == "..." {
				return RepeatLastBarUntilChange{}, nil
			}
			if data == "./." {
				return RepeatLastBar{}, nil
			}
			if strings.Index("123456789", data[1:2]) != -1 && data[2] == '.' {
				n, err := strconv.Atoi(data[1:2])
				return RepeatLastNBarsUntilChange(n), err
			}
		default:
			switch data[0:2] {
			case "MN":
				return parseMIDINote(data[2:])
			case "CC":
				return parseMIDICC(data[2:])
			case "PB":
				return parseMIDIPitchbend(data[2:])
			case "AT":
				return parseMIDIAftertouch(data[2:])
			case "PT":
				return parseMIDIPolyAftertouch(data[2:])
			}

			if regExPattern.MatchString(data) {
				return p.parsePattern(data, posIn32th)
			}
			res, err := parseNote(data)
			if err != nil {
				return nil, fmt.Errorf("unknown item: %#v", data)
			}
			return res, nil
		}
	}

	return nil, nil

}
