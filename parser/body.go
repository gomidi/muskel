package parser

import (
	"fmt"
	"strconv"
	"strings"

	"gitlab.com/gomidi/muskel/items"
	"gitlab.com/gomidi/muskel/score"
	"gitlab.com/gomidi/muskel/template"
	"gitlab.com/gomidi/muskel/unroller"
)

type body struct {
	*score.Score
	currentBarNo      int
	jumpInLineBefore  bool
	inPart            string
	numTracks         int
	lastPosition      string
	currentPosition   string
	currentPosIn32ths uint
}

func (p *body) newBar(b *score.Bar) {
	p.currentBarNo++
	//fmt.Printf("parser: adding bar %v\n", p.currentBar)
	b.BarNo = p.currentBarNo
	b.IsNaked = true
	b.OriginalBarNo = p.currentBarNo
	p.Score.Bars = append(p.Score.Bars, b)
}

func (p *body) GetTemplateDefinition(name string) *template.Definition {
	if def := p.Score.TemplateDefinitions[name]; def != nil {
		return def
	}
	return p.Score.IncludedTemplateDefinitions[name]
}

// parseSingleItem parses a single item (no multiitem)
func (p *body) parseSingleItem(data string, posIn32th uint) (interface{}, error) {
	return (&Items{p.GetTemplateDefinition}).ParseItem(strings.TrimSpace(data), posIn32th)
}

// parseItem parses an item that may be a lyric, a multiitem or a single item
func (p *body) parseItem(data string, posIn32th uint) (interface{}, error) {
	if len(data) == 0 {
		return nil, nil
	}

	return p.parseSingleItem(data, posIn32th)
}

// handleEmptyLine handles an empty line
func (p *body) handleEmptyLine() {
	if !p.jumpInLineBefore {
		//		fmt.Println("new bar added in empty line")
		p.newBar(score.NewBar())
	}
	// p.jumpInLineBefore = false
}

// handleJump handles a jump
func (p *body) handleJump(data string) error {
	b := score.NewBar()
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

	b.JumpTo = part
	p.newBar(b)
	p.jumpInLineBefore = true
	return nil
}

// handleTempoChange handles a tempo change
func (p *body) handleTempoChange(b *score.Bar, data string) error {
	if idx := strings.Index(data, "~"); idx > 0 {
		b.Tilde = data[idx:]
		data = data[:idx]
	}

	bpm, err := strconv.ParseFloat(data, 64)
	if err != nil {
		return fmt.Errorf("error in tempo change %#v. must be a number", data)
	}
	b.TempoChange = bpm
	return nil
}

// handleTimeSigChange handles a time signature change
func (p *body) handleTimeSigChange(b *score.Bar, data string) error {
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

	b.TimeSigChange[0] = uint8(num)
	b.TimeSigChange[1] = uint8(denom)

	//	fmt.Println("new bar added based on time sig change")
	p.newBar(b)
	p.jumpInLineBefore = false
	return nil
}

func (p *body) includeScore(sc *score.Score, file string) error {
	b := score.NewBar()
	b.Include = file

	for _, instr := range sc.Tracks {
		if !p.Score.HasTrack(instr.Name) {
			return fmt.Errorf("instrument %q not found in main score", instr.Name)
		}
	}
	//b.includedBars = make([]*Bar, len(sc.Bars))
	ur, err := unroller.Unroll(sc)
	if err != nil {
		return fmt.Errorf("could not unroll score: %s", err.Error())
	}

	b.IncludedScore = ur
	p.newBar(b)
	p.jumpInLineBefore = true
	return nil
}

func (p *body) include(file string) error {
	sc, err := IncludeScore(p.Score, file)
	if err != nil {
		return fmt.Errorf("can't include %q in score: %s", file, err.Error())
	}

	err = p.includeScore(sc, file)
	if err != nil {
		return fmt.Errorf("can't include %q in score: %s", file, err.Error())
	}
	return nil
}

func (p *body) parseScale(data string, b *score.Bar) error {
	idx := strings.Index(data, "^")
	if idx < 1 {
		return fmt.Errorf("invalid scale syntax: \\%s", data)
	}

	mod := data[:idx]
	md, has := items.Modes[mod]
	if !has {
		return fmt.Errorf("unknown mode: %q", mod)
	}
	nt, err := parseNote(data[idx+1:])
	if err != nil {
		return fmt.Errorf("invalid base note for scale: %q", data[idx+1:])
	}
	b.Scale = &items.Scale{BaseNote: nt.ToMIDI(), Mode: md}
	return nil
}

func (p *body) parseCommand(data string) error {
	var c items.CommandCall
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

func (p *body) finishPart(end int) {
	if p.inPart != "" {
		if _, has := p.Score.Parts[p.inPart]; has {
			p.Score.Parts[p.inPart] = [2]int{p.Score.Parts[p.inPart][0], end}
		}
		p.inPart = ""
	}
}

// parseBarLine parses a bar line (i.e. either simple bar change, or jump, or time signature change or tempo change and combinations of the last two
func (p *body) parseBarLine(data string) error {

	if strings.TrimSpace(data) == "" {
		p.handleEmptyLine()
		return nil
	}

	if data[0] == '$' {
		return p.parseCommand(data[1:])
	}

	// its a jump
	if data[0] == '[' {
		if p.currentBarNo == -1 {
			return fmt.Errorf("can't start with a jump: we need bars and parts first")
		}
		p.finishPart(p.currentBarNo + 1)
		return p.handleJump(data)

	}

	if data[0] == '*' {
		num, err := strconv.Atoi(data[1:])
		if err != nil {
			return err
		}
		for n := 0; n < num-1; n++ {
			b := score.NewBar()
			b.IsEmpty = true
			p.newBar(b)
		}
		p.jumpInLineBefore = false

		p.newBar(score.NewBar())
		return nil
	}
	var b = score.NewBar()
	p.jumpInLineBefore = false

	if idx := strings.Index(data, "\\"); idx >= 0 {
		err := p.parseScale(data[idx+1:], b)
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

func (p *body) getInstrData(tabs []string) (instrData []string) {
	//instrData = tabs[1 : len(tabs)-1]
	to := p.numTracks + 1
	if len(tabs)-1 < to {
		to = len(tabs) - 1
	}
	instrData = tabs[1:to]

	if missingTabs := p.numTracks - (len(tabs) - 2); missingTabs > 0 {
		for mt := 0; mt < missingTabs; mt++ {
			instrData = append(instrData, "")
		}
	}
	return instrData
}

func (p *body) getTrack(i int) (instr *score.Track) {
	if i > len(p.Score.Tracks)-1 {
		instr = score.NewTrack("")
		p.Score.AddTrack(instr)
		return
	}
	return p.Score.Tracks[i]
}

// setTrackName sets the name of an instrument
func (p *body) setTrackName(i int, instr *score.Track, name string) error {
	name = strings.Trim(name, "<>")
	if instr.Name == "" {
		instr.Name = name
		return nil
	}

	if name != instr.Name {
		return fmt.Errorf("can't set name for instrument in column %v to %q: name already set to: %q", i+1, name, instr.Name)
	}

	return nil
}

func (p *body) setTrackPitchbendrange(i int, instr *score.Track, data string) error {
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

// setTrackChannel sets the MIDI channel of an instrument
func (p *body) setTrackChannel(i int, instr *score.Track, data string) error {
	if len(data) == 0 {
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

	if instr.MIDIChannel == -1 && data != "" {
		instr.MIDIChannel = ch
		return nil
	}

	if ch != instr.MIDIChannel {
		return fmt.Errorf("can't set MIDI channel for instrument in column %v to %v: MIDI channel already set to: %v", i+1, ch, instr.MIDIChannel)
	}

	return nil
}

func (p *body) setTrackFileGroup(i int, instr *score.Track, data string) error {
	instr.FileGroup = strings.TrimSpace(data)
	return nil
}

// setTrackProgram sets the MIDI program value of an instrument
func (p *body) setTrackProgram(i int, instr *score.Track, data string) error {
	if len(data) == 0 {
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

	if instr.MIDIProgram == -1 && data != "" {
		instr.MIDIProgram = prog
		return nil
	}

	if prog != instr.MIDIProgram {
		return fmt.Errorf("can't set MIDI program for instrument in column %v to %v: MIDI program already set to: %v", i+1, prog, instr.MIDIProgram)
	}

	return nil
}

// setTrackVolume sets the MIDI volume of an instrument
func (p *body) setTrackVolume(i int, instr *score.Track, data string) error {
	if len(data) == 0 {
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

	if instr.MIDIVolume == -1 && data != "" {
		instr.MIDIVolume = vol
		return nil
	}

	if vol != instr.MIDIVolume {
		return fmt.Errorf("can't set MIDI volume for instrument in column %v to %v: MIDI volume already set to: %v", i+1, vol, instr.MIDIVolume)
	}

	return nil
}

// setTrackTranspose sets the MIDITranspose for an instrument
func (p *body) setTrackTranspose(i int, instr *score.Track, data string) error {
	if len(data) == 0 || data == "" {
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

// setTrackBank sets the MIDI bank for an instrument
func (p *body) setTrackBank(i int, instr *score.Track, data string) error {
	if len(data) == 0 {
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

	if instr.MIDIBank == -1 && data != "" {
		instr.MIDIBank = bank
		return nil
	}

	if bank != instr.MIDIBank {
		return fmt.Errorf("can't set MIDI bank for instrument in column %v to %v: MIDI bank already set to: %v", i+1, bank, instr.MIDIBank)
	}

	return nil
}

// setDelay sets the delay for an instrument
func (p *body) setDelay(i int, instr *score.Track, data string) error {
	if len(data) == 0 {
		instr.Delay = [2]int{0, 0}
		return nil
	}

	d := strings.Split(data, "/")

	if len(d) != 2 {
		return fmt.Errorf("%#v is not a valid delay value", data)
	}

	num, err := strconv.Atoi(d[0])
	if err != nil {
		return fmt.Errorf("%#v is not a valid delay numerator", d[0])
	}

	denom, err := strconv.Atoi(d[1])
	if err != nil || denom == 0 {
		return fmt.Errorf("%#v is not a valid delay denominator", d[1])
	}

	instr.Delay = [2]int{num, denom}
	//fmt.Printf("setting delay for instrument %q to %v/%v\n", instr.Name, num, denom)

	return nil
}

// parseInstrEvents parses the events of a single instrument in a non-bar line
func (p *body) parseInstrEvents(i int, instr *score.Track, data string) (err error) {
	var barEvents score.BarEvents

	if diff := (p.currentBarNo - (len(instr.Events) - 1)); diff > 0 {
		for _d := 0; _d < diff; _d++ {
			instr.Events = append(instr.Events, barEvents)
		}
	}
	barEvents = instr.Events[p.currentBarNo]
	ev := &score.Event{}
	ev.OriginalData = data

	item, err := p.parseItem(ev.OriginalData, p.currentPosIn32ths)

	if err != nil {
		return fmt.Errorf("column %v: %s", i+1, err)
	}

	if item != nil {
		ev.Item = item
	}

	barEvents = append(barEvents, ev)
	instr.Events[p.currentBarNo] = barEvents

	return nil
}

// handleLastColumn handles the data of the last column in a non-bar line / event line
func (p *body) handleLastColumn(data string) error {
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
func (p *body) parseEventsLine(tabs []string) (err error) {
	//	fmt.Printf("parseEventsLine in score %q\n", p.Score.FileName)

	if p.numTracks == -1 {
		p.numTracks = len(tabs) - 2
	}

	// fmt.Printf("num instruments: %v\n", p.numTracks)

	instrData := p.getInstrData(tabs)
	firstColumn := strings.TrimSpace(tabs[0])

	for i, data := range instrData {
		if i >= p.numTracks || i >= len(tabs)-2 {
			break
		}
		// fmt.Printf("column: %v\n", i)
		instr := p.getTrack(i)
		data = strings.TrimSpace(data)

		switch strings.ToLower(firstColumn) {
		case "":
			err = p.setTrackName(i, instr, data)
		case "pbrange", "pitchbendrange":
			err = p.setTrackPitchbendrange(i, instr, data)
		case "ch", "channel":
			err = p.setTrackChannel(i, instr, data)
		case "prog", "program":
			err = p.setTrackProgram(i, instr, data)
		case "vol", "volume":
			err = p.setTrackVolume(i, instr, data)
		case "trans", "transpose":
			err = p.setTrackTranspose(i, instr, data)
		case "bank":
			err = p.setTrackBank(i, instr, data)
		case "file":
			err = p.setTrackFileGroup(i, instr, data)
		case "delay":
			err = p.setDelay(i, instr, data)
		default:
			if p.jumpInLineBefore { //&& p.firstBarSet {
				//				fmt.Println("new bar in parse events")
				p.newBar(score.NewBar())
			}
			p.jumpInLineBefore = false
			p.Score.Bars[p.currentBarNo].IsNaked = false
			if i == 0 {
				p.lastPosition, p.currentPosIn32ths, err = items.PositionTo32th(p.lastPosition, firstColumn)
				if err != nil {
					return err
				}

				p.Score.Bars[p.currentBarNo].OriginalPositions = append(p.Score.Bars[p.currentBarNo].OriginalPositions, firstColumn)
				p.Score.Bars[p.currentBarNo].Positions = append(p.Score.Bars[p.currentBarNo].Positions, p.currentPosIn32ths)
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
func (p *body) parseLine(line string) error {
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
func (p *body) addComment(relLine int, line string) {
	p.Score.BodyComments[relLine] = line
}
