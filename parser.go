package muskel

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
	"strings"
)

type Parser struct {
	Score             *Score
	rd                *bufio.Reader
	inComment         bool
	inHeader          bool
	lineNo            int
	lineNosystemStart int
	currentBar        int
	numInstruments    int
}

func NewParser(r io.Reader) *Parser {
	return &Parser{
		Score:          NewScore(),
		rd:             bufio.NewReader(r),
		currentBar:     -1,
		numInstruments: -1,
	}
}

func ParseFile(filepath string) (sc *Score, err error) {
	b, err := ioutil.ReadFile(filepath)

	if err != nil {
		return nil, err
	}

	return Parse(bytes.NewReader(b))
}

func Parse(rd io.Reader) (sc *Score, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error: %v", r)
		}
	}()

	ps := NewParser(rd)
	err = ps.Parse()

	if err != nil {
		return nil, err
	}

	return ps.Score, nil
}

func (p *Parser) newBar(b *Bar) {
	p.currentBar++
	b.barNo = p.currentBar
	b.originalBarNo = p.currentBar
	p.Score.Bars = append(p.Score.Bars, b)
}

func (p *Parser) addHeaderComment(line []byte) {
	p.Score.HeaderComments = append(p.Score.HeaderComments, string(line))
}

func (p *Parser) GetPatternDefinition(name string) *PatternDefinition {
	return p.Score.PatternDefinitions[name]
}

func (p *Parser) parseData(data string) (interface{}, error) {
	if len(data) == 0 {
		return nil, nil
	}

	switch {
	case data[0] == '"':
		return Lyric(strings.Trim(data, `"`)), nil
	default:
		d := strings.Split(data, " ")
		switch len(d) {
		case 1:
			it, err := (&itemParser{p.GetPatternDefinition}).parseItem(d[0])
			if err != nil {
				return nil, err
			}
			return it, nil
		default:
			var m MultiItem

			for _, dd := range d {
				it, err := (&itemParser{p.GetPatternDefinition}).parseItem(strings.TrimSpace(dd))

				if err != nil {
					return nil, err
				}

				m = append(m, it)
			}
			return m, nil
		}

	}

	return nil, nil

}

func (p *Parser) parseSystemLine(line string) error {
	tabs := strings.Split(line, "|")

	//	parts := map[string]int{}

	switch len(tabs) {
	case 1:
		tab0 := strings.TrimSpace(tabs[0])
		// we have a bar-change, i.e. no pipes anywhere
		switch tab0 {
		case "":
			var b Bar
			p.newBar(&b)
			return nil
		// normal bar change without anything
		default:

			var b Bar

			// its a jump
			if tab0[0] == '[' {
				part := strings.Trim(tab0, "[]")
				if _, has := p.Score.Parts[part]; !has {
					return fmt.Errorf("could not jump to part %#v: not found", part)
				}
				p.Score.Jumps[p.currentBar] = part
				return nil
			}

			if idx := strings.Index(tab0, "@"); idx >= 0 {
				bpm, err := strconv.ParseFloat(tab0[idx+1:], 64)
				if err != nil {
					return fmt.Errorf("error in tempo change %#v. must be a number", tab0[idx+1:])
				}
				b.tempoChange = bpm

				tab0 = strings.TrimSpace(tab0[:idx])
			}

			if tab0 == "" {
				// add bar
				p.newBar(&b)
				return nil
			}

			timeSig := strings.Split(tab0, "/")
			if len(timeSig) != 2 {
				return fmt.Errorf("error in time signature %#v. must be in format n/m where n and m are numbers > 0", tab0)
			}

			num, err := strconv.Atoi(timeSig[0])
			if err != nil {
				return fmt.Errorf("error in time signature %#v. must be in format n/m where n and m are numbers > 0", tab0)
			}

			denom, err := strconv.Atoi(timeSig[1])
			if err != nil {
				return fmt.Errorf("error in time signature %#v. must be in format n/m where n and m are numbers > 0", tab0)
			}

			b.timeSigChange[0] = uint8(num)
			b.timeSigChange[1] = uint8(denom)

			p.newBar(&b)
			return nil
		}
	case 0:
		panic("must not happen")
	default:
		if p.numInstruments == -1 {
			p.numInstruments = len(tabs) - 2
		}

		position := strings.TrimSpace(tabs[0])

		instrData := tabs[1 : len(tabs)-1]

		//fmt.Printf("instruments: [%v]: %v (should %v)\n", lineNo, len(tabs)-2, numInstruments)
		if missingTabs := p.numInstruments - (len(tabs) - 2); missingTabs > 0 {
			for mt := 0; mt < missingTabs; mt++ {
				instrData = append(instrData, "")
			}
		}
		/*
			if len(tabs)-2 != p.numInstruments {
				return fmt.Errorf("line %v has %v instruments, but must have %v", p.lineNo, len(tabs)-2, p.numInstruments)
			}
		*/

		for i, data := range instrData {
			var instr *Instrument
			if i > len(p.Score.Instruments)-1 {
				instr = &Instrument{}
				instr.MIDIProgram = -1
				instr.MIDIChannel = -1
				instr.MIDIVolume = -1
				instr.MIDIBank = -1
				p.Score.Instruments = append(p.Score.Instruments, instr)
			} else {
				instr = p.Score.Instruments[i]
			}

			data = strings.TrimSpace(data)

			switch strings.ToLower(position) {
			case "":
				if instr.Name == "" {
					instr.Name = data
				}
			case "ch", "channel":
				if instr.MIDIChannel == -1 && data != "-" {
					cc, errC := strconv.Atoi(data)
					if errC != nil {
						return fmt.Errorf("%#v is not a valid MIDI channel", data)
					}
					if cc > 16 || cc < 1 {
						return fmt.Errorf("%#v is not a valid MIDI channel", cc)
					}

					instr.MIDIChannel = int8(cc - 1)
				}
			case "prog", "program":
				if instr.MIDIProgram == -1 && data != "-" {
					pr, errP := strconv.Atoi(data)
					if errP != nil {
						return fmt.Errorf("%#v is not a valid MIDI program value", data)
					}
					if pr > 128 || pr < 1 {
						return fmt.Errorf("%#v is not a valid MIDI program value", pr)
					}
					instr.MIDIProgram = int8(pr - 1)
				}
			case "vol", "volume":
				if instr.MIDIVolume == -1 && data != "-" {
					pr, errP := strconv.Atoi(data)
					if errP != nil {
						return fmt.Errorf("%#v is not a valid MIDI volume value", data)
					}
					if pr > 128 || pr < 1 {
						return fmt.Errorf("%#v is not a valid MIDI volume value", pr)
					}
					instr.MIDIVolume = int8(pr - 1)
				}
			case "bank":
				if instr.MIDIBank == -1 && data != "-" {
					pr, errP := strconv.Atoi(data)
					if errP != nil {
						return fmt.Errorf("%#v is not a valid MIDI bank value", data)
					}
					if pr > 128 || pr < 1 {
						return fmt.Errorf("%#v is not a valid MIDI bank value", pr)
					}
					instr.MIDIBank = int8(pr - 1)
				}
			default:
				var barEvents BarEvents
				//fmt.Printf("currentBar: %v, len(instr.events)-1: %v\n", p.currentBar, len(instr.events)-1)
				if diff := (p.currentBar - (len(instr.events) - 1)); diff > 0 {
					for _d := 0; _d < diff; _d++ {
						instr.events = append(instr.events, barEvents)
					}
				}
				barEvents = instr.events[p.currentBar]
				ev := &Event{}
				//ev.originalPosition = position
				//ev.Position = 0 // TODO parse the position from position
				ev.originalData = data

				item, err := p.parseData(ev.originalData)

				if err != nil {
					return fmt.Errorf("column %v: %s", i+2, err)
				}

				//				fmt.Printf("item: %v\n", item)

				if item != nil {
					ev.Item = item
				}
				barEvents = append(barEvents, ev)
				instr.events[p.currentBar] = barEvents

				if i == 0 {
					p.Score.Bars[p.currentBar].originalPositions = append(p.Score.Bars[p.currentBar].originalPositions, position)
					//p.Score.Bars[p.currentBar].positions = append(p.Score.Bars[p.currentBar].positions, positionTo32th(position))
				}
			}
		}

		part := strings.TrimSpace(tabs[len(tabs)-1])

		switch len(part) {
		case 1:
			// part definition
			if _, hasPart := p.Score.Parts[part]; hasPart {
				return fmt.Errorf("part %#v already exists", part)
			} else {
				p.Score.Parts[part] = p.currentBar
			}

		case 0:
		// do nothing
		default:
			// probably a jump; check
			p.Score.Jumps[p.currentBar] = part
			//fmt.Printf("line: %v, position: %v part: %v\n", lineNo, position, part)
		}

	}

	return nil
}

func (p *Parser) parseHeader(lines string) error {
	ll := strings.Split(lines, "\n")
	//var props = map[string]string{}
	//var temperament = map[string]string{}
	//var patterns = map[string]string{}
	for _, l := range ll {
		if len(l) == 0 {
			continue
		}
		switch l[0] {
		case '/':
			keyVal := strings.Split(l[1:], "/")
			if len(keyVal) != 2 {
				return fmt.Errorf("invalid header line %#v, format must be /note/frequency", l)
			}

			p.Score.Temperament[strings.TrimSpace(keyVal[0])] = strings.TrimSpace(keyVal[1])
		case '$':
			var pd PatternDefinition
			err := pd.Parse(l[1:])
			if err != nil {
				return fmt.Errorf("invalid pattern definition line %#v: %s", l, err)
			}
			p.Score.PatternDefinitions[pd.Name] = &pd
		/*
			keyVal := strings.Split(l[1:], ":")
			if len(keyVal) != 2 {
				return fmt.Errorf("invalid header line %#v, format must be $name: definition", l)
			}



			p.patterns[strings.TrimSpace(keyVal[0])] = strings.TrimSpace(keyVal[1])
		*/
		default:
			keyVal := strings.Split(l, ":")
			if len(keyVal) != 2 {
				return fmt.Errorf("invalid header line %#v, format must be key: value", l)
			}

			key := strings.ToLower(strings.TrimSpace(keyVal[0]))
			val := strings.TrimSpace(keyVal[1])

			p.Score.Meta[key] = val

		}

	}

	//fmt.Printf("properties: %#v\n", props)
	//fmt.Printf("temperament: %#v\n", temperament)
	//fmt.Printf("patterns: %#v\n", patterns)

	return nil
}

func (p *Parser) addComment(line []byte) {
	if p.inHeader {
		p.addHeaderComment(line)
		return
	}
	p.addSystemComment(p.lineNo-p.lineNosystemStart, line)
}

func (p *Parser) addSystemComment(relLine int, line []byte) {
	p.Score.Comments[relLine] = string(line)
}

func (p *Parser) Parse() (err error) {
	var line []byte
	var header bytes.Buffer

	p.inHeader = true
	p.inComment = false
	p.lineNo = 0
	p.lineNosystemStart = 0

	for {
		line, _, err = p.rd.ReadLine()

		if err != nil {
			break
		}

		p.lineNo++

		if len(line) > 0 && string(line[0]) == "=" {
			p.inHeader = false
			err := p.parseHeader(header.String())
			p.lineNosystemStart = p.lineNo + 1
			if err != nil {
				return err
			}
			continue
		}

		if len(line) > 1 {
			switch string(line[0:2]) {
			case "//":
				// comment: ignore
				p.addComment(line)
			case "/*":
				p.inComment = true
				p.addComment(line)
			case "*/":
				p.inComment = false
				p.addComment(line)
			default:
				if !p.inComment {
					//					fmt.Printf("line %v\n", p.lineNo)
					if p.inHeader {
						header.Write(line)
						header.WriteString("\n")
					} else {
						err = p.parseSystemLine(string(line))
						if err != nil {
							return fmt.Errorf("Error in line %v:\n%s\n", p.lineNo, err.Error())
						}
					}
				} else {
					p.addComment(line)
				}
			}
		} else {
			if p.inComment {
				p.addComment(line)
			} else {
				if !p.inHeader {
					err = p.parseSystemLine(string(line))
					if err != nil {
						return fmt.Errorf("Error in line %v:\n%s\n", p.lineNo, err.Error())
					}
				}
			}
		}

	}

	//	fmt.Printf("HEADER:\n%s\n", header.String())

	//fmt.Printf("instruments: %v\n", instrumentNames)
	//fmt.Printf("instruments midi channels: %v\n", instrumentMIDIChannels)
	//fmt.Printf("instruments midi programs: %v\n", instrumentMIDIPrograms)

	if err != io.EOF {
		return err
	}
	return nil
}
