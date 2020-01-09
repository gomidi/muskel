package items

import (
	"fmt"
	"strings"
	//"gitlab.com/gomidi/muskel/template"
)

type TemplateDefinition interface {
	TimeSignature() [2]uint8
	Definition() string
	Call(params ...string) (result string, err error)
}

type Parser struct {
	GetDefinition func(name string) TemplateDefinition
}

func (p *Parser) SetDefinitionGetter(getter func(string) TemplateDefinition) {
	p.GetDefinition = getter
}

func (p *Parser) ParseItem(data string, posIn32th uint) (Item, error) {
	/*
		if data == "=sk.col2" {
			panic("same col")
		}
		if data == "" {
			//panic("empty data")
			return nil, nil
		}
		fmt.Printf("Parser#ParseItem(%q,%v)\n", data, posIn32th)
	*/
	return parseItem(p, data, posIn32th)
}

func parseItem(p *Parser, data string, posIn32th uint) (it Item, err error) {
	data = strings.TrimSpace(data)

	// fmt.Printf("parseItem called with %q\n", data)
	switch len(data) {
	case 0:
		return nil, nil
	case 1:
		switch data[0] {
		case '%':
			return repeatLastEvent{}, nil
		case ':':
			return Hold, nil
		case '*':
			return Rest, nil
		case '~':
			gliss := GlideStart(false)
			gl := &gliss
			err = gl.Parse("", posIn32th)
			it = gl
			return
		default:
			var nt = &Note{}
			err = nt.Parse(data, posIn32th)
			if err != nil {
				err = fmt.Errorf("unknown item: %#v", data)
			}
			it = nt
			return
		}
	default:
		// fmt.Printf("data[0]: %#v\n", string(data[0]))
		switch data[0] {
		case '#':
			var p Placeholder
			var ph = &p
			err = ph.Parse(data[1:], posIn32th)
			if err != nil {
				err = fmt.Errorf("invalid placeholder: %#v", data)
			}
			it = ph
			return
		case '~':
			gliss := GlideStart(false)
			gl := &gliss
			err = gl.Parse(data[1:], posIn32th)
			it = gl
			return
		case '"':
			var ly = Lyric("")
			l := &ly
			err = l.Parse(data, posIn32th)
			it = l
			return
		case '{':
			var ntp = &NTuple{}
			ntp.Parser = p
			err = ntp.Parse(data[1:], posIn32th)
			it = ntp
			return
		case '$':
			var cmd = &CommandCall{}
			err = cmd.Parse(data[1:], posIn32th)
			if err != nil {
				return
			}
			switch cmd.Name {
			case "include":
				sketch := "=SCORE"
				if len(cmd.Params) > 1 {
					sketch = strings.Trim(cmd.Params[1], `"'`)
				}
				var inc Include
				inc.File = strings.Trim(cmd.Params[0], `"'`)
				inc.Sketch = sketch
				if len(cmd.Params) > 2 {
					inc.Params = cmd.Params[2:]
				}
				it = inc
			default:
				it = cmd
			}
			return
		case '(':
			var m = &MultiItem{}

			if idx := strings.Index(data, "..."); idx > 0 {
				m.Exploded = true
				data = data[:idx]
			}

			endIdx := strings.LastIndex(data, ")")
			/*
				if endIdx < 4 {
					return nil, fmt.Errorf("invalid MultiItem: %q must start and end with kleene star *", data)
				}
			*/

			if endIdx != len(data)-1 {
				return nil, fmt.Errorf("invalid MultiItem: %q must end with )", "("+data)
			}
			m.Parser = p
			err = m.Parse(data[1:endIdx], posIn32th)
			it = m
			return
			//return nil, fmt.Errorf("deprecated syntax: use  : instead to create a hold")
		//case 'S':
		//	return parseNote(data[1:])
		//case 'Z':
		//	return parseNote(data[1:])
		case '^':
			nt := &Note{}
			err = nt.parseScale(data[1:])
			it = nt
			return
		case '\\': // scale
			var sc Scale
			if idx := strings.Index(data, "..."); idx > 0 {
				sc.Exploded = true
				data = data[:idx]
			}
			//fmt.Printf("parsing scale: %q\n", data[1:])
			err = sc.Parse(data[1:], posIn32th)
			it = &sc
			return
			/*
				case 'Â':
			*/
			/*
				nt := &Note{}
				err = nt.parseScale(data[1:])
				it = nt
				return
			*/
		case '?':
			if data[1] == '(' {
				rd := &RandomChooser{}
				rd.Parser = p
				err = rd.Parse(data[1:], posIn32th)
				it = rd
				return
			}

			rd := &RandomProbability{}
			rd.Parser = p
			err = rd.Parse(data[1:], posIn32th)
			it = rd
			return
		//case 'O':
		//	return p.parseOSC(data[1:])
		case '/':
			var o Override
			o.Item, err = parseItem(p, data[1:], posIn32th)
			if err != nil {
				return
			}
			it = &o
			return
		/*
			case '!':
				pc := NewTemplateCall(p)
				pc.Parser = p
				err = pc.Parse(data, posIn32th)
				it = pc
				return
		*/
		case '.':
			var rp = &BarRepeater{}
			err = rp.Parse(data[1:], posIn32th)
			it = rp
			return

		default:

			// fmt.Printf("%v %v\n", data[0], int(data[0]))

			if len(data) > 3 && data[0:3] == "_MN" {
				mn := &MIDINote{}
				err = mn.Parse("_"+data[3:], posIn32th)
				it = mn
				return
			}

			switch data[0:2] {
			case "_^":
				nt := &Note{}
				err = nt.parseScale("_" + data[2:])
				it = nt
				return
			//case "_°": // hack for keyboads that don't allow to properly insert ^
			/*
				case "_\\":
					nt := &Note{}
					err = nt.parseScale("_" + data[2:])
					it = nt
					return
			*/
			case "MN":
				mn := &MIDINote{}
				err = mn.Parse(data[2:], posIn32th)
				it = mn
				return
			case "CC":
				cc := &MIDICC{}
				err = cc.Parse(data[2:], posIn32th)
				it = cc
				return
				//return parseMIDICC(data[2:])
			case "PB":
				pb := &MIDIPitchbend{}
				err = pb.Parse(data[2:], posIn32th)
				it = pb
				return
			case "AT":
				at := &MIDIAftertouch{}
				err = at.Parse(data[2:], posIn32th)
				it = at
				return
			case "PT":
				pt := &MIDIPolyAftertouch{}
				err = pt.Parse(data[2:], posIn32th)
				it = pt
				return
			}

			if regExTemplate0.MatchString(data) || regExTemplate1.MatchString(data) {
				//fmt.Printf("regExTemplate.MatchString(%q)\n", data)
				pc := NewCall(p)
				pc.Parser = p
				if idx := strings.Index(data, "..."); idx > 0 && idx+3 == len(data) {
					//fmt.Printf("len(data) = %v; idx = %v\n", len(data), idx)
					pc.Exploded = true
					data = data[:idx] //+ data[idx+3:]
				}
				//fmt.Printf("data cleaned: %q\n", data)
				err = pc.Parse(data, posIn32th)
				it = pc
				return
			}

			nt := &Note{}
			err := nt.Parse(data, posIn32th)
			if err != nil {
				return nil, fmt.Errorf("unknown item: %#v: %s", data, err)
			}
			return nt, nil
		}
	}

	return nil, nil

}
