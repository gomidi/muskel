package muskel

import (
	"fmt"
	"strconv"
	"strings"
)

// TODO implement parseMIDICC
func parseMIDICC(data string) (cc MIDICC, err error) {
	return cc, fmt.Errorf("parseMIDICC is not implemented yet")
}

// TODO implement parseMIDIPitchbend
func parseMIDIPitchbend(string) (pb MIDIPitchbend, err error) {
	return pb, fmt.Errorf("parseMIDIPitchbend is not implemented yet")
}

// TODO implement parseMIDIAftertouch
func parseMIDIAftertouch(string) (at MIDIAftertouch, err error) {
	return at, fmt.Errorf("parseMIDIAftertouch is not implemented yet")
}

// TODO implement parseMIDIPolyAftertouch
func parseMIDIPolyAftertouch(string) (pt MIDIPolyAftertouch, err error) {
	return pt, fmt.Errorf("parseMIDIPolyAftertouch is not implemented yet")
}

func parseNote(data string) (item interface{}, err error) {
	if len(data) > 1 {

		switch data[:2] {
		case "kd", "sn", "ho", "hc", "rd", "tb", "tl", "tm", "th", "sh":
			var dn DrumNote
			dn.name = data[:2]
			for _, l := range data[2:] {
				switch l {
				case '+':
					dn.dynamic += "+"
				case '-':
					dn.dynamic += "-"
				case '*':
					dn.dynamic += "*"
				default:
					return nil, fmt.Errorf("invalid drum note: %#v", data)
				}
			}
			return dn, nil
		default:
		}
	}
	var nt Note

	switch data[:1] {
	case "a", "b", "c", "d", "e", "f", "g":
		nt.letter = data[:1]
		nt.octave = 1
	case "A", "B", "C", "D", "E", "F", "G":
		nt.letter = strings.ToLower(data[:1])
		nt.octave = -1
	default:
		return nil, fmt.Errorf("invalid note: %#v", data)
	}

	for _, l := range data[1:] {
		switch l {
		case '+':
			nt.dynamic += "+"
		case '-':
			nt.dynamic += "-"
		case '*':
			nt.dynamic += "*"
		case '~':
			nt.glissandoStart = true
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
		default:
			return nil, fmt.Errorf("invalid note: %#v", data)
		}
	}

	return nt, nil
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

	orig := data

	dd := strings.Split(data, "}")
	if len(dd) != 2 {
		return nt, fmt.Errorf("invalid n-tuple: %#v", "{"+orig)
	}

	pos := strings.TrimSpace(dd[1])

	if len(pos) == 0 {
		err = fmt.Errorf("invalid n-tuple %#v: position missing", "{"+orig)
		return
	}

	completedEndPos, endPos, err := positionTo32th("", dd[1])

	if err != nil {
		return nt, err
	}

	_ = completedEndPos

	data = strings.TrimSpace(dd[0])

	d := strings.Split(data, ",")

	if len(d) < 2 {
		err = fmt.Errorf("empty or single item n-tuples are not allowed: %#v", "{"+orig)
		return
	}

	var ntp NTuple
	ntp.endPos = endPos

	for pos, it := range d {
		itt, err := p.parseItem(it, posIn32th)
		if err != nil {
			return nt, fmt.Errorf("ERROR invalid n-tuple at position %v: %#v: %s", pos, it, err)
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

	switch data[0] {
	case '[':
		idx := strings.Index(data, "]")
		if idx < 0 {
			return nil, fmt.Errorf("invalid random value: ?%s", data)
		}

		num := data[1:idx]

		var rp RandomProbability
		var n int
		n, err = strconv.Atoi(num)

		if err != nil {
			return nil, fmt.Errorf("invalid random value: ?%s, syntax must be [n] where n is a number between 0 and 100", data)
		}

		rp.prob = uint8(n)

		if rp.prob < 0 || rp.prob > 100 {
			return nil, fmt.Errorf("invalid random value: ?%s, syntax must be [n] where n is a number between 0 and 100", data)
		}

		rp.item, err = p.parseItem(data[idx+1:], posIn32th)

		if err != nil {
			return nil, fmt.Errorf("invalid random value item: ?%s", data)
		}
		rp.itemOriginalData = data[idx+1:]
		return &rp, nil
	case '(':
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
	default:
		return nil, fmt.Errorf("invalid random value: ?%s", data)
	}
}

func (p *itemParser) parseCommand(data string, posIn32th uint) (interface{}, error) {
	panic("TODO implement")
}

func (p *itemParser) parseItem(data string, posIn32th uint) (interface{}, error) {
	data = strings.TrimSpace(data)
	//	fmt.Printf("parseItem called with %q\n", data)
	switch len(data) {
	case 0:
		return nil, nil
	default:
		//fmt.Printf("data[0]: %#v\n", string(data[0]))
		switch data[0] {
		case '"':
			return Lyric(strings.Trim(data, `"`)), nil
		case '%':
			return RepeatLastEvent{}, nil
		case '=':
			return Hold, nil
		case '_':
			return Rest{}, nil
		case '{':
			return p.parseNTuple(data[1:], posIn32th)
		case '$':
			return p.parseCommand(data[1:], posIn32th)
		case 'S':
			return parseNote(data[1:])
		case 'Z':
			return parseNote(data[1:])
		case '?':
			return p.parseRandom(data[1:], posIn32th)
		case 'O':
			return p.parseOSC(data[1:])
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
			if len(data) > 1 {
				switch data[0:2] {
				case "CC":
					return parseMIDICC(data[2:])
				case "PB":
					return parseMIDIPitchbend(data[2:])
				case "AT":
					return parseMIDIAftertouch(data[2:])
				case "PT":
					return parseMIDIPolyAftertouch(data[2:])
				default:
					if regExPattern.MatchString(data) {
						return p.parsePattern(data, posIn32th)
					}
				}
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
