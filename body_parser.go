package muskel

import (
	"fmt"
	"strconv"
	"strings"
)

type BodyParser struct {
	*Score
	currentBarNo      int
	jumpInLineBefore  bool
	inPart            string
	numInstruments    int
	lastPosition      string
	currentPosition   string
	currentPosIn32ths uint
}

func (p *BodyParser) newBar(b *Bar) {
	p.currentBarNo++
	//fmt.Printf("parser: adding bar %v\n", p.currentBar)
	b.barNo = p.currentBarNo
	b.originalBarNo = p.currentBarNo
	p.Score.Bars = append(p.Score.Bars, b)
}

func (p *BodyParser) GetPatternDefinition(name string) *PatternDefinition {
	if def := p.Score.PatternDefinitions[name]; def != nil {
		return def
	}
	return p.Score.IncludedPatternDefinitions[name]
}

// parseSingleItem parses a single item (no multiitem)
func (p *BodyParser) parseSingleItem(data string, posIn32th uint) (interface{}, error) {
	return (&itemParser{p.GetPatternDefinition}).parseItem(strings.TrimSpace(data), posIn32th)
}

// parseItem parses an item that may be a lyric, a multiitem or a single item
func (p *BodyParser) parseItem(data string, posIn32th uint) (interface{}, error) {
	if len(data) == 0 {
		return nil, nil
	}

	switch {
	case data[0] == '"':
		return Lyric(strings.Trim(data, `"`)), nil
	default:
		d := strings.Split(data, " ")
		switch len(d) {
		case 0:
			return nil, nil
		case 1:
			return p.parseSingleItem(d[0], posIn32th)
		default:
			var m MultiItem

			for _, dd := range d {
				it, err := p.parseSingleItem(dd, posIn32th)

				if err != nil {
					return nil, err
				}

				m = append(m, it)
			}
			return m, nil
		}

	}
}

// handleEmptyLine handles an empty line
func (p *BodyParser) handleEmptyLine() error {
	if !p.jumpInLineBefore {
		//		fmt.Println("new bar added in empty line")
		p.newBar(NewBar())
	}
	//	p.jumpInLineBefore = false
	return nil
}

// handleJump handles a jump
func (p *BodyParser) handleJump(data string) error {
	b := NewBar()
	part := strings.Trim(data, "[]")
	/*
		if _, has := p.Score.Parts[part]; !has {
			return fmt.Errorf("could not jump to part %#v: not found", part)
		}
	*/

	if p.inPart != "" {
		p.Score.Parts[part] = [2]int{p.Score.Parts[part][0], p.currentBarNo}
		p.inPart = ""
	}

	b.jumpTo = part
	p.newBar(b)
	p.jumpInLineBefore = true
	return nil
}

// handleTempoChange handles a tempo change
func (p *BodyParser) handleTempoChange(b *Bar, data string) error {
	bpm, err := strconv.ParseFloat(data, 64)
	if err != nil {
		return fmt.Errorf("error in tempo change %#v. must be a number", data)
	}
	b.tempoChange = bpm
	return nil
}

// handleTimeSigChange handles a time signature change
func (p *BodyParser) handleTimeSigChange(b *Bar, data string) error {
	timeSig := strings.Split(data, "/")
	if len(timeSig) != 2 {
		return fmt.Errorf("error in time signature %#v. must be in format n/m where n and m are numbers > 0", data)
	}

	num, err := strconv.Atoi(timeSig[0])
	if err != nil {
		return fmt.Errorf("error in time signature %#v. must be in format n/m where n and m are numbers > 0", data)
	}

	denom, err := strconv.Atoi(timeSig[1])
	if err != nil {
		return fmt.Errorf("error in time signature %#v. must be in format n/m where n and m are numbers > 0", data)
	}

	b.timeSigChange[0] = uint8(num)
	b.timeSigChange[1] = uint8(denom)

	//	fmt.Println("new bar added based on time sig change")
	p.newBar(b)
	p.jumpInLineBefore = false
	return nil
}

func (p *BodyParser) includeScore(sc *Score, file string) error {
	b := NewBar()
	b.include = file

	for _, instr := range sc.Instruments {
		if !p.Score.hasInstrument(instr.Name) {
			return fmt.Errorf("instrument %q not found in main score", instr.Name)
		}
	}
	//b.includedBars = make([]*Bar, len(sc.Bars))
	ur, err := sc.Unroll()
	if err != nil {
		return fmt.Errorf("could not unroll score: %s", err.Error())
	}

	b.includedScore = ur
	p.newBar(b)
	p.jumpInLineBefore = true
	return nil
}

func (p *BodyParser) include(file string) error {
	sc, err := p.Score.include(file)
	if err != nil {
		return fmt.Errorf("can't include %q in score: %s", file, err.Error())
	}

	if !sc.isPartial() {
		return fmt.Errorf("can't include %q: it is no partial", file)
	}

	if sc.Meta["partial"] != "score" {
		return fmt.Errorf("can't include %q in score: it must be partial \"score\" but is %q", file, sc.Meta["partial"])
	}

	err = p.includeScore(sc, file)
	if err != nil {
		return fmt.Errorf("can't include %q in score: %s", file, err.Error())
	}
	return nil
}

func (p *BodyParser) parseCommand(data string) error {
	var c CommandCall
	err := c.Parse(data)
	if err != nil {
		return err
	}
	switch c.Name {
	case "include":
		p.finishPart(p.currentBarNo + 1)
		if len(c.Params) != 1 {
			return fmt.Errorf("include command needs one parameter")
		}
		return p.include(strings.Trim(c.Params[0], "\""))
	default:
		return fmt.Errorf("unsupported command in body: %q", c.Name)
	}
}

func (p *BodyParser) finishPart(end int) {
	if p.inPart != "" {
		if _, has := p.Score.Parts[p.inPart]; has {
			p.Score.Parts[p.inPart] = [2]int{p.Score.Parts[p.inPart][0], end}
		}
		p.inPart = ""
	}
}

// parseBarLine parses a bar line (i.e. either simple bar change, or jump, or time signature change or tempo change and combinations of the last two
func (p *BodyParser) parseBarLine(data string) error {

	if strings.TrimSpace(data) == "" {
		return p.handleEmptyLine()
	}

	if data[0] == '$' {
		return p.parseCommand(data[1:])
	}

	// its a jump
	if data[0] == '[' {
		p.finishPart(p.currentBarNo + 1)
		return p.handleJump(data)

	}
	var b = NewBar()
	p.jumpInLineBefore = false

	if idx := strings.Index(data, "@"); idx >= 0 {
		err := p.handleTempoChange(b, data[idx+1:])
		if err != nil {
			return err
		}
		data = strings.TrimSpace(data[:idx])
		if data == "" {
			// add bar
			//			fmt.Println("new bar added based on tempo change")
			p.newBar(b)
			return nil
		}
	}

	return p.handleTimeSigChange(b, data)
}

func (p *BodyParser) getInstrData(tabs []string) (instrData []string) {
	instrData = tabs[1 : len(tabs)-1]

	if missingTabs := p.numInstruments - (len(tabs) - 2); missingTabs > 0 {
		for mt := 0; mt < missingTabs; mt++ {
			instrData = append(instrData, "")
		}
	}
	return instrData
}

func (p *BodyParser) getInstrument(i int) (instr *Instrument) {
	if i > len(p.Score.Instruments)-1 {
		instr = &Instrument{}
		instr.MIDIProgram = -1
		instr.MIDIChannel = -1
		instr.MIDIVolume = -1
		instr.MIDIBank = -1
		p.Score.AddInstrument(instr)
		return
	}
	return p.Score.Instruments[i]
}

// setInstrumentName sets the name of an instrument
func (p *BodyParser) setInstrumentName(i int, instr *Instrument, name string) error {
	if instr.Name == "" {
		instr.Name = name
		return nil
	}

	if name != instr.Name {
		return fmt.Errorf("can't set name for instrument in column %v to %q: name already set to: %q", i+1, name, instr.Name)
	}

	return nil
}

func (p *BodyParser) setInstrumentPitchbendrange(i int, instr *Instrument, data string) error {
	if len(data) == 0 {
		return nil
	}
	rg, errC := strconv.Atoi(data)
	if errC != nil {
		return fmt.Errorf("%#v is not a valid MIDI pitchbendrange", data)
	}
	if rg > 24 || rg < 1 {
		return fmt.Errorf("%#v is not a valid MIDI pitchbendrange", rg)
	}

	if instr.PitchbendRange == 0 {
		instr.PitchbendRange = uint8(rg)
		return nil
	}

	if uint8(rg) != instr.PitchbendRange {
		return fmt.Errorf("can't set MIDI pitchbendrange for instrument in column %v to %v: MIDI pitchbendrange already set to: %v", i+1, rg, instr.PitchbendRange)
	}

	return nil
}

// setInstrumentChannel sets the MIDI channel of an instrument
func (p *BodyParser) setInstrumentChannel(i int, instr *Instrument, data string) error {
	if data == "-" {
		instr.MIDIChannel = -1
		return nil
	}
	cc, errC := strconv.Atoi(data)
	if errC != nil {
		return fmt.Errorf("%#v is not a valid MIDI channel", data)
	}
	if cc > 16 || cc < 1 {
		return fmt.Errorf("%#v is not a valid MIDI channel", cc)
	}

	ch := int8(cc - 1)

	if instr.MIDIChannel == -1 && data != "-" {
		instr.MIDIChannel = ch
		return nil
	}

	if ch != instr.MIDIChannel {
		return fmt.Errorf("can't set MIDI channel for instrument in column %v to %v: MIDI channel already set to: %v", i+1, ch, instr.MIDIChannel)
	}

	return nil
}

// setInstrumentProgram sets the MIDI program value of an instrument
func (p *BodyParser) setInstrumentProgram(i int, instr *Instrument, data string) error {
	if data == "-" {
		instr.MIDIProgram = -1
		return nil
	}
	pr, err := strconv.Atoi(data)
	if err != nil {
		return fmt.Errorf("%#v is not a valid MIDI program value", data)
	}
	if pr > 128 || pr < 1 {
		return fmt.Errorf("%#v is not a valid MIDI program value", pr)
	}

	prog := int8(pr - 1)

	if instr.MIDIProgram == -1 && data != "-" {
		instr.MIDIProgram = prog
		return nil
	}

	if prog != instr.MIDIProgram {
		return fmt.Errorf("can't set MIDI program for instrument in column %v to %v: MIDI program already set to: %v", i+1, prog, instr.MIDIProgram)
	}

	return nil
}

// setInstrumentVolume sets the MIDI volume of an instrument
func (p *BodyParser) setInstrumentVolume(i int, instr *Instrument, data string) error {
	if data == "-" {
		instr.MIDIVolume = -1
		return nil
	}
	pr, err := strconv.Atoi(data)
	if err != nil {
		return fmt.Errorf("%#v is not a valid MIDI volume value", data)
	}
	if pr > 128 || pr < 1 {
		return fmt.Errorf("%#v is not a valid MIDI volume value", pr)
	}

	vol := int8(pr - 1)

	if instr.MIDIVolume == -1 && data != "-" {
		instr.MIDIVolume = vol
		return nil
	}

	if vol != instr.MIDIVolume {
		return fmt.Errorf("can't set MIDI volume for instrument in column %v to %v: MIDI volume already set to: %v", i+1, vol, instr.MIDIVolume)
	}

	return nil
}

// setInstrumentTranspose sets the MIDITranspose for an instrument
func (p *BodyParser) setInstrumentTranspose(i int, instr *Instrument, data string) error {
	if data == "" {
		instr.MIDITranspose = 0
		return nil
	}
	pr, err := strconv.Atoi(data)
	if err != nil {
		return fmt.Errorf("%#v is not a valid MIDI transpose value", data)
	}
	if pr >= 84 || pr <= -84 {
		return fmt.Errorf("%#v is not a valid MIDI transpose value (must be >=-84 and <=84", pr)
	}

	tr := int8(pr)

	instr.MIDITranspose = tr
	return nil
}

// setInstrumentBank sets the MIDI bank for an instrument
func (p *BodyParser) setInstrumentBank(i int, instr *Instrument, data string) error {
	if data == "-" {
		instr.MIDIBank = -1
		return nil
	}
	pr, err := strconv.Atoi(data)
	if err != nil {
		return fmt.Errorf("%#v is not a valid MIDI bank value", data)
	}
	if pr > 128 || pr < 1 {
		return fmt.Errorf("%#v is not a valid MIDI bank value", pr)
	}

	bank := int8(pr - 1)

	if instr.MIDIBank == -1 && data != "-" {
		instr.MIDIBank = bank
		return nil
	}

	if bank != instr.MIDIBank {
		return fmt.Errorf("can't set MIDI bank for instrument in column %v to %v: MIDI bank already set to: %v", i+1, bank, instr.MIDIBank)
	}

	return nil
}

// parseInstrEvents parses the events of a single instrument in a non-bar line
func (p *BodyParser) parseInstrEvents(i int, instr *Instrument, data string) (err error) {
	var barEvents BarEvents

	if diff := (p.currentBarNo - (len(instr.events) - 1)); diff > 0 {
		for _d := 0; _d < diff; _d++ {
			instr.events = append(instr.events, barEvents)
		}
	}
	barEvents = instr.events[p.currentBarNo]
	ev := &Event{}
	ev.originalData = data

	item, err := p.parseItem(ev.originalData, p.currentPosIn32ths)

	if err != nil {
		return fmt.Errorf("column %v: %s", i+1, err)
	}

	if item != nil {
		ev.Item = item
	}

	barEvents = append(barEvents, ev)
	instr.events[p.currentBarNo] = barEvents

	return nil
}

// handleLastColumn handles the data of the last column in a non-bar line / event line
func (p *BodyParser) handleLastColumn(data string) error {
	if data == "" {
		return nil
	}

	// part definition
	if _, hasPart := p.Score.Parts[data]; hasPart {
		return nil
	}

	p.finishPart(p.currentBarNo)

	p.Score.Parts[data] = [2]int{p.currentBarNo, -1}
	p.inPart = data
	return nil
}

// parseEventsLine parses a non-bar line / event line
func (p *BodyParser) parseEventsLine(tabs []string) (err error) {
	//	fmt.Printf("parseEventsLine in score %q\n", p.Score.FileName)

	if p.numInstruments == -1 {
		p.numInstruments = len(tabs) - 2
	}

	instrData := p.getInstrData(tabs)
	firstColumn := strings.TrimSpace(tabs[0])

	for i, data := range instrData {
		instr := p.getInstrument(i)
		data = strings.TrimSpace(data)

		switch strings.ToLower(firstColumn) {
		case "":
			err = p.setInstrumentName(i, instr, data)
		case "pbrange", "pitchbendrange":
			err = p.setInstrumentPitchbendrange(i, instr, data)
		case "ch", "channel":
			err = p.setInstrumentChannel(i, instr, data)
		case "prog", "program":
			err = p.setInstrumentProgram(i, instr, data)
		case "vol", "volume":
			err = p.setInstrumentVolume(i, instr, data)
		case "trans", "transpose":
			err = p.setInstrumentTranspose(i, instr, data)
		case "bank":
			err = p.setInstrumentBank(i, instr, data)
		default:
			if p.jumpInLineBefore { //&& p.firstBarSet {
				//				fmt.Println("new bar in parse events")
				p.newBar(NewBar())
			}
			p.jumpInLineBefore = false
			if i == 0 {
				p.lastPosition, p.currentPosIn32ths, err = positionTo32th(p.lastPosition, firstColumn)
				if err != nil {
					return err
				}

				p.Score.Bars[p.currentBarNo].originalPositions = append(p.Score.Bars[p.currentBarNo].originalPositions, firstColumn)
				p.Score.Bars[p.currentBarNo].positions = append(p.Score.Bars[p.currentBarNo].positions, p.currentPosIn32ths)
			}

			err = p.parseInstrEvents(i, instr, data)
		}

		if err != nil {
			return err
		}

	}

	return p.handleLastColumn(strings.TrimSpace(tabs[len(tabs)-1]))
}

// parseLine parses a line in the body (everything after the =)
func (p *BodyParser) parseLine(line string) error {
	//	fmt.Printf("parse body line %q\n", line)
	tabs := strings.Split(line, "|")

	switch len(tabs) {
	case 1:
		return p.parseBarLine(strings.TrimSpace(tabs[0]))
	case 0:
		panic("must not happen")
	default:
		return p.parseEventsLine(tabs)
	}
}

// addComment adds a comment to the body
func (p *BodyParser) addComment(relLine int, line []byte) {
	p.Score.BodyComments[relLine] = string(line)
}
