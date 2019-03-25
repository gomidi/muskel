package parser

import (
	"fmt"
	"strconv"
	"strings"

	"gitlab.com/gomidi/muskel/muskellib"
	"gitlab.com/gomidi/muskel/template"
)

type Items struct {
	GetDefinition func(name string) *template.Definition
}

// TODO implement parseOSC
func (p *Items) parseOSC(data string) (om muskellib.OSCMessage, err error) {
	return om, fmt.Errorf("parseOSC is not implemented yet")
}

func (p *Items) SetDefinitionGetter(getter func(string) *template.Definition) {
	p.GetDefinition = getter
}

func (p *Items) parseTemplate(data string, positionIn32th uint) (pc *template.Call, err error) {
	if p.GetDefinition == nil {
		return nil, fmt.Errorf("no definition resolver defined")
	}
	pc = template.NewTemplateCall(p)
	//  &template.TemplateCall{}
	pc.Getter = p.GetDefinition
	err = pc.ParseTemplate(data, positionIn32th)
	return
}

// ntuple has the form {c,e,d}3&
// where 3& is the ending position that defines the total length
func (p *Items) parseNTuple(data string, posIn32th uint) (nt muskellib.NTuple, err error) {

	var ntp muskellib.NTuple
	orig := data

	dd := strings.Split(data, "}")
	if len(dd) != 2 {
		return nt, fmt.Errorf("invalid n-tuple: %#v", "{"+orig)
	}

	pos := strings.TrimSpace(dd[1])

	if strings.HasSuffix(pos, ">") {
		ntp.PosShift = 1
		pos = strings.TrimRight(pos, ">")
	}

	if strings.HasSuffix(pos, "<") {
		ntp.PosShift = -1
		pos = strings.TrimRight(pos, "<")
	}

	// fmt.Printf("pos: %q postShift: %v\n", pos, ntp.posShift)

	if len(pos) == 0 {
		err = fmt.Errorf("invalid n-tuple %#v: position missing", "{"+orig)
		return
	}

	_, endPos, err := muskellib.PositionTo32th("", pos)

	if err != nil {
		return nt, err
	}

	data = strings.TrimSpace(dd[0])

	d := strings.Split(data, ",")

	if len(d) < 2 {
		err = fmt.Errorf("empty or single item n-tuples are not allowed: %#v", "{"+orig)
		return
	}

	ntp.EndPos = endPos

	for ppos, it := range d {
		itt, err := p.ParseItem(it, posIn32th)
		if err != nil {
			return nt, fmt.Errorf("ERROR invalid n-tuple at position %v: %#v: %s", ppos, it, err)
		}
		ntp.Items = append(ntp.Items, itt)
	}
	return ntp, nil
}

func (p *Items) parseRandom(data string, posIn32th uint) (item interface{}, err error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("invalid random value: ?%s", data)
	}

	// strings.IndexAny()

	if data[0] == '(' {
		alternatives := strings.Trim(data, "()")
		alt := strings.Split(alternatives, ",")
		var r muskellib.RandomChooser
		for _, a := range alt {
			a = strings.TrimSpace(a)
			if a != "" {
				item, err = p.ParseItem(a, posIn32th)

				if err != nil {
					return nil, fmt.Errorf("invalid random value item: %s", a)
				}
				r.Alternatives = append(r.Alternatives, item)
				r.AlternativesOriginalData = append(r.AlternativesOriginalData, a)
			}
		}
		return &r, nil
	}

	idx := strings.Index(data, "%")
	if idx < 0 {
		return nil, fmt.Errorf("invalid random value: ?%s", data)
	}

	num := data[:idx]

	var rp muskellib.RandomProbability
	var n int
	n, err = strconv.Atoi(num)

	if err != nil {
		return nil, fmt.Errorf("invalid random value: ?%s, syntax must be ?n%% where n is a number between 0 and 100", data)
	}

	rp.Prob = uint8(n)

	if rp.Prob < 0 || rp.Prob > 100 {
		return nil, fmt.Errorf("invalid random value: ?%s, syntax must be ?n%% where n is a number between 0 and 100, but number is %v", data, rp.Prob)
	}

	rp.Item, err = p.ParseItem(data[idx+1:], posIn32th)

	if err != nil {
		return nil, fmt.Errorf("invalid random value item: ?%s", data)
	}
	rp.ItemOriginalData = data[idx+1:]
	return &rp, nil

}

func (p *Items) parseCommand(data string, posIn32th uint) (interface{}, error) {
	panic("TODO implement")
}

func (p *Items) parseGlissando(data string, posIn32th uint) (interface{}, error) {
	if data == "" {
		return muskellib.GlissandoLinear, nil
	}

	if data == "~" {
		return muskellib.GlissandoExponential, nil
	}
	return nil, fmt.Errorf("invalid glissando: \"~%s\"", data)
}

func (p *Items) ParseItem(data string, posIn32th uint) (interface{}, error) {
	data = strings.TrimSpace(data)
	// fmt.Printf("parseItem called with %q\n", data)
	switch len(data) {
	case 0:
		return nil, nil
	case 1:
		switch data[0] {
		case '%':
			return muskellib.RepeatLastEvent{}, nil
		case ':':
			return muskellib.Hold, nil
		case '*':
			return muskellib.Rest, nil
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
			return muskellib.Lyric(strings.Trim(data, `"`)), nil
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
			var m muskellib.MultiItem
			d := strings.Split(data[1:endIdx], "_")
			for _, dd := range d {
				it, err := p.ParseItem(dd, posIn32th)

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
			return p.parseTemplate(data, posIn32th)
		case '.':
			if data == "..." {
				return muskellib.RepeatLastBarUntilChange{}, nil
			}
			if data == "./." {
				return muskellib.RepeatLastBar{}, nil
			}
			if strings.Index("123456789", data[1:2]) != -1 && data[2] == '.' {
				n, err := strconv.Atoi(data[1:2])
				return muskellib.RepeatLastNBarsUntilChange(n), err
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

			if regExTemplate.MatchString(data) {
				return p.parseTemplate(data, posIn32th)
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
