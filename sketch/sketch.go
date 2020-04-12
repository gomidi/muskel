package sketch

import (
	"fmt"
	"strconv"
	"strings"

	"gitlab.com/gomidi/muskel/items"
	"gitlab.com/gomidi/muskel/track"
)

/*

Folgende Schwierigkeiten:

1. Die Sketchspalte muss parametrisiert werden, bevor irgendetwas anderes passiert
2. erst dann werden die items geparst und dabei die zufälligkeiten der Spalte gelost.
3. erst jetzt werden die weiteren Sketchspaltenaufrufe ersetzt und ggf. wiederholt
4. erst jetzt werden die Teil-Wiederholungen durchgeführt
5. erst jetzt werden imports durchgeführt
6. erst jetzt auf der obersten ebene werden die Trackeigenschaften assoziiert
7. erst jetzt werden die verbliebenen zufälligkeiten gelost
8. erst jetzt werden die skalentöne durch absolute töne ersetzt
9. erst jetzt werden die filter angewendet
10. erst jetzt wird in MIDI umgewandelt
*/

/*
Phase 0: Tabelle wird eingelesen und Spalten stehen unverändert, Taktpositionen und Teile werden aber schon identifiziert
Phase 1: Parameter werden eingefügt
Phase 2: Items werden geparst, Zufälligkeiten gelost
Phase 3: "Templates" und imports werden ersetzt (triggert beim Template Phase 1-4)
Phase 4: Teil-Wiederholungen werden durchgeführt
Phase 5: auf der obersten Ebene werden Spalten mit Trackeigenschaften assoziiert
Phase 6: die verbliebenen Zufälligkeiten werden gelost
Phase 7: Skalentöne werden aufgelöst
Phase 8: Filter werden angewendet
Phase 9: jetzt kann das Erbegnis als einfache tabelle entrollt geschrieben werden
Phase 10: Umwandlung nach MIDI
*/

type include struct {
	position uint
	sketch   *Sketch
}

type Sketch struct {
	Name              string
	Bars              []*Bar
	Positions         [][2]uint // bar+position
	Columns           map[string][]string
	Parts             map[string][2]uint // absolute positions
	Score             Score
	File              string
	barChangeRequired bool

	includes            []*include
	colOrder            []string
	inPart              string
	jumpInLineBefore    bool
	currentBarNo        int
	currentPosIn32ths   uint
	projectedBarEnd     uint
	currentTimeSignatur [2]uint8
	currentTempo        float64
	lastPosition        string
	currentScale        *items.Scale
	currentBeat         uint
}

func NewSketch(name string, sc Score) *Sketch {
	sk := &Sketch{
		Score:   sc,
		Name:    name,
		Columns: map[string][]string{},
		Parts:   map[string][2]uint{},

		currentBarNo:        -1,
		barChangeRequired:   true,
		currentTimeSignatur: [2]uint8{4, 4},
		currentTempo:        120,
		currentBeat:         0,
	}

	if sk.isScore() {
		scale := &items.Scale{}
		scale.Reset()
		scale.BaseNote = 60
		scale.Mode = items.Major
		sk.currentScale = scale
	}
	return sk
}

func (s *Sketch) calcColWidth(colName string) (colWidth int) {
	colWidth = len(colName)

	data := s.Columns[colName]

	if data == nil {
		return
	}

	for _, d := range data {
		if l := len(d); l > colWidth {
			colWidth = l
		}
	}

	return
}

func (s *Sketch) getRangeInBars(startbar, endBar int) (startPos, endPos int) {
	if startbar == len(s.Bars) {
		return -1, -1
	}

	startPos = int(s.Bars[startbar].Position)

	if endBar >= len(s.Bars) {
		endPos = int(s.Bars[len(s.Bars)-1].Position)
	} else {
		endPos = int(s.Bars[endBar].Position + s.Bars[endBar].Length32th())
	}
	return
}

func (s *Sketch) getEventsInBars(startbar, endBar int, evts []*items.Event) []*items.Event {
	startPos, endPos := s.getRangeInBars(startbar, endBar)
	if startPos < 0 {
		return nil
	}

	return items.GetEventsInPosRange(uint(startPos), uint(endPos), evts)
}

// gets the bar idx of an event at the given position
// returns -1, if no bar could be found
func (s *Sketch) getBarIdxOf(pos uint) (baridx int) {
	baridx = -1

	for idx, b := range s.Bars {
		if b.Position > pos {
			break
		}
		baridx = idx
	}

	return baridx
}

// loop = 0: column is not looped; loop > 0: column is looped n times
func (s *Sketch) parseEvents(data []string, origEndPos uint) (res []*items.Event, loop uint, err error) {
	var lastNoRestItem items.Item

	for _, dat := range data {
		var ev items.Event
		err = ev.Parse(dat)
		if err != nil {
			return nil, 0, err
		}

		if ev.Item == items.RepeatLastEvent {
			ev.Item = lastNoRestItem
		}

		if ev.Item != items.Rest {
			lastNoRestItem = ev.Item
		}

		if inc, isInc := ev.Item.(items.Include); isInc {
			_ = inc
			panic("includes not allowed inside columns")
		}

		il := items.IsLoop(ev.Item)

		if il >= 0 {
			loop = uint(il)
		}

		ev.Item = items.RollTheDiceForAnItem(ev.Item)
		res = append(res, &ev)
	}
	// parse items and resolve randomness at this level
	// TODO check, if it works
	return
}

func (s *Sketch) findBarAtPosition(pos uint) (b *Bar) {
	for _, bb := range s.Bars {
		if bb.Position > pos {
			return
		}
		b = bb
	}
	return
}

func (s *Sketch) getAbsPos(bar, pos32ths uint) uint {
	var pos uint

	for i, b := range s.Bars {
		if uint(i) == bar {
			pos = b.Position
			break
		}
	}

	return pos + pos32ths
}

func (s *Sketch) newCol(tr *track.Track, colName string) *column {
	return &column{
		sketch: s,
		name:   colName,
		track:  tr,
	}
}

func (s *Sketch) Unroll(_tr *track.Track, colName string, params ...string) ([]*items.Event, error) {
	col := s.newCol(_tr, colName)
	events, _, err := col.Call(0, false, params...)
	if err != nil {
		return nil, err
	}
	if events == nil {
		return nil, nil
	}

	events = items.RollTheDice(events)
	events = s.Score.FilterTrack(colName, events)
	return events, nil
}

func (s *Sketch) repeatBars(repevts []*items.Event, diff uint) (out []*items.Event, nextBarPos uint) {
	lastPos := repevts[len(repevts)-1].Position
	bidx := s.getBarIdxOf(lastPos)
	nextBarPos = s.Bars[bidx].Position + uint(s.Bars[bidx].Length32th()) + diff

	for _, rev := range repevts {
		nev := rev.Dup()
		nev.Position += diff
		out = append(out, nev)
	}

	return
}

func (t *Sketch) isScore() bool {
	return t.Name == "=SCORE"
}

// parseBarLine parses a bar line (i.e. either simple bar change, or jump, or time signature change or tempo change and combinations of the last two
func (p *Sketch) parseBarLine(data string) error {
	if data == "" {
		return fmt.Errorf("empty lines are allowed within a table")
	}

	var comment string
	var part string

	if idx := strings.Index(data, "//"); idx > -1 {
		comment = strings.TrimSpace(data[idx+2:])
		if idx2 := strings.Index(comment, "##"); idx2 > -1 {
			comment = comment[:idx2+2]
		}
		data = strings.TrimSpace(data[:idx])
	}

	if data == "#" {
		p.handleEmptyBarChange(comment, part)
		p.jumpInLineBefore = false
		p.barChangeRequired = false
		return nil
	}

	if data[0] == '$' {
		p.barChangeRequired = true
		return p.parseCommandLF(data)
	}

	// its a jump
	if data[0] == '[' {
		p.barChangeRequired = true
		if p.currentBarNo == -1 {
			return fmt.Errorf("can't start with a jump: we need bars and parts first")
		}
		return p.handleJump(data)

	}

	if data[0] == '*' {
		p.barChangeRequired = true
		num, err := strconv.Atoi(data[1:])
		if err != nil {
			return err
		}
		for n := 0; n < num; n++ {
			b := NewBar() //score.NewBar()
			b.TimeSig = p.currentTimeSignatur
			b.TempoChange = p.currentTempo
			b.Scale = p.currentScale
			if n == 0 {
				b.Comment = comment
			}
			p.newBar(b)
		}
		p.jumpInLineBefore = false
		return nil
	}

	if data[0] != '#' {
		return fmt.Errorf("invalid bar change: %q", data)
	}

	idx := strings.Index(data, " ")
	if idx > 1 {
		part = data[1:idx]
		data = "#" + data[idx:]
	}

	if idx == -1 {
		part = data[1:]
		data = "#"
	}

	if data == "#" {
		p.handleEmptyBarChange(comment, part)
		p.jumpInLineBefore = false
		p.barChangeRequired = false
		p.handlePart(part)
		return nil
	}

	p.barChangeRequired = false

	data = strings.TrimSpace(data[1:])

	var b = NewBar()
	b.Part = part
	p.jumpInLineBefore = false
	b.TimeSig = p.currentTimeSignatur
	b.TempoChange = p.currentTempo
	b.Scale = p.currentScale
	b.Comment = comment

	if idx := strings.Index(data, "\\"); idx >= 0 {
		err := p.parseScale(data[idx:], b)
		if err != nil {
			return err
		}
		data = strings.TrimSpace(data[:idx])
		if data == "" {
			// add bar
			p.newBar(b)
			p.handlePart(part)
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
			p.newBar(b)
			p.handlePart(part)
			return nil
		}
	}

	err := p.handleTimeSigChange(b, data)
	if err != nil {
		return err
	}
	p.handlePart(part)
	return nil
}

func (s *Sketch) unrollPartBars(bars []*Bar) ([]*Bar, error) {
	var res []*Bar
	var lastBarEnd uint
	var i int

	for _, bar := range bars {
		lastBarEnd = bar.Position + uint(bar.Length32th())

		if bar.JumpTo != "" {

			part, has := s.Parts[bar.JumpTo]
			if !has {
				return nil, fmt.Errorf("can't jump to part %q from bar %v: part is not defined", bar.JumpTo, bar.No+1)
			}
			startPos := part[0]
			endPos := part[1]
			partBars := getBarsInPosRange(startPos, endPos, bars)

			var nbars []*Bar

			for kk, b := range partBars {
				nub := b.Dup()
				nub.Position = nub.Position + bar.Position
				nub.Comment = fmt.Sprintf("[%v.%v]", bar.JumpTo, kk)
				nub.No = i
				nub.Part = ""
				nub.Position -= startPos
				nbars = append(nbars, nub)
				i++
			}

			res = append(res, nbars...)

		} else {
			endPos := bar.Position + uint(bar.Length32th())
			if endPos > s.projectedBarEnd {
				endPos = s.projectedBarEnd
			}

			brs := getBarsInPosRange(bar.Position, endPos, bars)
			var nbars []*Bar

			for _, b := range brs {
				nub := b.Dup()
				nub.No = i
				nbars = append(nbars, nub)
				i++
			}
			res = append(res, nbars...)
		}

	}

	nbars := getBarsInPosRange(lastBarEnd, s.projectedBarEnd, bars)
	res = append(res, nbars...)

	return res, nil
}

func (s *Sketch) unrollIncludedBars() (unrolled []*Bar, err error) {
	var no int

	for _, bar := range s.Bars {
		if bar.Include != nil {
			var sk *Sketch
			sk, err = s.Score.GetIncludedSketch(bar.Include.File, bar.Include.Sketch, bar.Include.Params)
			if err != nil {
				return
			}

			skname := sk.Name
			if skname == "" {
				skname = "=SCORE"
			}

			skBars, err := sk.UnrolledBars()

			if err != nil {
				return nil, fmt.Errorf("can't get unrolled bars for include %q: %s", bar.Include.File+"#"+skname, err.Error())
			}

			for _, incbar := range skBars {
				b := incbar.Dup()
				b.No = no
				b.Position += bar.Position
				b.Comment = fmt.Sprintf("from %s.%s#%v", bar.Include.File, skname, incbar.No+1)
				//b.Part = ""
				unrolled = append(unrolled, b)
				no++
			}
		} else {
			b := bar.Dup()
			b.No = no
			unrolled = append(unrolled, b)
			no++
		}
	}

	if no != len(unrolled) {
		panic("problems with numbering")
	}

	return
}

func (s *Sketch) UnrolledBars() (unrolled []*Bar, err error) {
	unrolled, err = s.unrollIncludedBars()
	if err != nil {
		return nil, err
	}
	unrolled, err = s.unrollPartBars(unrolled)
	if err != nil {
		return nil, err
	}
	return
}

func (p *Sketch) newBar(b *Bar) {
	p.currentBarNo++
	b.No = p.currentBarNo
	b.Position = p.projectedBarEnd
	p.Bars = append(p.Bars, b)
	p.currentPosIn32ths = p.projectedBarEnd
	p.projectedBarEnd += uint(b.Length32th())
	p.currentBeat = 0
}

// handleEmptyLine handles an empty line
func (p *Sketch) handleEmptyBarChange(comment, part string) {
	b := NewBar()
	b.Comment = comment
	b.TimeSig = p.currentTimeSignatur
	b.TempoChange = p.currentTempo
	b.Scale = p.currentScale
	b.Part = part
	p.newBar(b)
}

func (p *Sketch) finishPart(end uint) {
	if p.inPart != "" {
		if DEBUG {
			fmt.Printf("finishPart %q at %v\n", p.inPart, end)
		}
		if old := p.Parts[p.inPart]; old[1] == 0 {
			old[1] = end
			p.Parts[p.inPart] = old
		}
		p.inPart = ""
	}
}

func (t *Sketch) isTemplate() bool {
	return !t.isScore()
}

// handleJump handles a jump
func (p *Sketch) handleJump(data string) error {
	p.finishPart(p.projectedBarEnd)
	b := NewBar()
	part := strings.TrimSpace(strings.Trim(data, "[]"))

	if p.inPart != "" {
		old := p.Parts[p.inPart]
		old[1] = p.projectedBarEnd
		p.Parts[p.inPart] = old
		p.inPart = ""
	}

	b.JumpTo = part
	end := p.projectedBarEnd
	p.newBar(b)
	p.jumpInLineBefore = true
	p.projectedBarEnd = end + p.Parts[part][1] - p.Parts[part][0]

	return nil
}

// handleTempoChange handles a tempo change
func (p *Sketch) handleTempoChange(b *Bar, data string) error {
	if idx := strings.Index(data, "~"); idx > 0 {
		b.Tilde = data[idx:]
		data = data[:idx]
	}

	bpm, err := strconv.ParseFloat(data, 64)
	if err != nil {
		return fmt.Errorf("error in tempo change %#v. must be a number", data)
	}
	b.TempoChange = bpm
	p.currentTempo = bpm
	return nil
}

// handleTimeSigChange handles a time signature change
func (p *Sketch) handleTimeSigChange(b *Bar, data string) error {
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
	b.TimeSig = b.TimeSigChange

	p.currentTimeSignatur = b.TimeSigChange

	p.newBar(b)
	p.jumpInLineBefore = false
	return nil
}

func (s *Sketch) explodeParam(params []string) (res []string) {

	for _, param := range params {
		it, err := items.Parse(param, 0)

		if err == nil {
			switch v := it.(type) {

			case *items.Scale:
				if v.Exploded {
					if v.Mode == nil {
						v.Mode = s.Score.GetMode(v.Name)
					}
					if v.Mode != nil {
						for _, key := range v.All() {
							var note items.Note
							note.Letter, note.Augmenter, note.Octave = items.KeyToNote(key)
							res = append(res, note.String())
						}
					} else {
						res = append(res, param)
					}
				} else {
					res = append(res, param)
				}

			case *items.Pattern:
				res = append(res, param)
			case *items.Token:
				if v.Exploded {
					tok, errTok := s.Score.GetToken(v.Name)
					if errTok == nil {
						itt, errItt := items.Parse(tok, 0)
						if errItt == nil {
							switch vv := itt.(type) {
							case *items.MultiItem:
								for _, iiit := range vv.Events {
									res = append(res, iiit.String())
								}
							default:
								res = append(res, param)
							}
						} else {
							res = append(res, param)
						}
					} else {
						res = append(res, param)
					}
				} else {
					res = append(res, param)
				}
			case *items.MultiItem:
				if v.Exploded {
					for _, iit := range v.Events {
						res = append(res, iit.String())
					}
				} else {
					res = append(res, param)
				}
			default:
				res = append(res, param)
			}
		} else {
			res = append(res, param)
		}
	}

	return
}

func (s *Sketch) injectParams(col []string, params []string) (res []string) {
	params = s.explodeParam(params)

	// replace placeholders with params inside the col strings
	for _, c := range col {
		res = append(res, items.ReplaceParams(c, params))
	}

	return
}

func (p *Sketch) parseScale(data string, b *Bar) error {
	var sc items.Scale

	err := sc.Parse(data, 0)
	if err != nil {
		return err
	}

	b.Scale = &sc
	p.currentScale = &sc
	return nil
}

func (p *Sketch) getColumnData(tabs []string) (colData []string, lastColumn string) {
	numCols := len(p.Columns)
	if len(tabs)-2 > numCols {
		numCols = len(tabs) - 2
	}
	to := numCols + 1
	if len(tabs)-1 < to {
		to = len(tabs) - 1
	}

	if len(tabs)-1 >= to {
		lastColumn = strings.TrimSpace(tabs[len(tabs)-1])
	}
	colData = make([]string, len(tabs[1:to]))
	copy(colData, tabs[1:to])

	if missingTabs := numCols - (len(tabs) - 2); missingTabs > 0 {
		for mt := 0; mt < missingTabs; mt++ {
			colData = append(colData, "")
		}
	}
	return colData, lastColumn
}

// handleLastColumn handles the data of the last column in a non-bar line / event line
func (p *Sketch) handlePart(data string) error {

	if data == "" {
		return nil
	}

	// part definition
	if _, has := p.Parts[data]; has {
		return nil
	}

	p.finishPart(p.Bars[p.currentBarNo].Position)

	p.Parts[data] = [2]uint{p.Bars[p.currentBarNo].Position, 0}
	p.inPart = data
	return nil
}

func (s *Sketch) FirstColumn() (colName string) {
	for i, col := range s.colOrder {
		if i == 0 || col[0] == '!' {
			colName = col
		}
	}
	return
}

func (s *Sketch) _includeCol(column string, inc items.Include) (evts []*items.Event, end uint, err error) {
	sk, err := s.Score.GetIncludedSketch(inc.File, inc.Sketch, inc.Params)
	if err != nil {
		return nil, 0, fmt.Errorf("can't find sketch %q in file %q", inc.Sketch, inc.File)
	}
	if !sk.Score.HasTrack(column) {
		return nil, 0, nil
	}

	tr, err := sk.Score.GetTrack(column)
	if err != nil {
		return nil, 0, nil
	}
	var patt = sk.newCol(tr, column)
	return patt.Call(0, false, inc.Params...)
}

func (s *Sketch) includeCol(start uint, column string, inc items.Include) (evts []*items.Event, err error) {
	evts, _, err = s._includeCol(column, inc)
	if err != nil {
		return
	}

	evts = items.ForwardEvents(evts, start)
	printEvents("after including in col "+column, evts)
	return
}

func (p *Sketch) parseCommandLF(data string) error {
	p.barChangeRequired = true
	var c items.Command
	err := c.Parse(data, 0)
	if err != nil {
		return err
	}
	switch c.Name {
	case "$embed":
		err = p.Score.Embed(c.Params...)
		if err != nil {
			return err
		}
	case "$include":
		it, err := items.Parse(data, 0)
		if err != nil {
			return fmt.Errorf("can't parse include: $%q", data)
		}

		if inc, isInc := it.(items.Include); isInc {
			sk, err := p.Score.GetIncludedSketch(inc.File, inc.Sketch, inc.Params)
			if err != nil {
				return fmt.Errorf("can't find sketch table %q in include %q: %s", inc.Sketch, inc.File, err.Error())
			}

			inc.Length32ths = sk.projectedBarEnd
			end := p.projectedBarEnd
			b := NewBar() //score.NewBar()
			b.Include = &inc
			p.newBar(b)
			p.projectedBarEnd = end + uint(inc.Length32ths)
		} else {
			return fmt.Errorf("not a proper include: $%q", data)
		}
		return nil
	default:
	}
	return fmt.Errorf("unsupported command in body: %q", c.Name)
}

func (s *Sketch) parsePosition(firstColumn string) (err error) {
	firstColumn = strings.TrimSpace(firstColumn)

	if firstColumn == "" && s.lastPosition != "" {
		firstColumn = s.lastPosition
	}

	if firstColumn == "" {
		return fmt.Errorf("can't start with an empty position")
	}

	p := strings.Split(firstColumn, " ")
	var posRaw string
	var tempoChange float64
	var scale string

	for _, _p := range p {

		_p = strings.TrimSpace(_p)

		switch {
		case _p == "":
			// ignore
		case _p[0] == '\\':
			scale = _p
		case _p[0] == '@':
			tempoChange, err = strconv.ParseFloat(_p[1:], 64)
			if err != nil {
				return fmt.Errorf("position %q contains invalid tempo", firstColumn)
			}
			s.currentTempo = tempoChange
		default:
			posRaw = _p
		}
	}

	pos, pos32ths, err := items.PositionTo32th(s.currentBeat, posRaw)

	if err != nil {
		return err
	}
	s.currentBeat, _ = items.GetQNNumberFromPos(pos)

	if pos != "" {
		s.lastPosition = pos
	}

	s.currentPosIn32ths += pos32ths
	s.Positions = append(s.Positions, [2]uint{uint(s.currentBarNo), pos32ths})
	if tempoChange > 0 {
		s.Bars[s.currentBarNo].InnerTempoChanges[pos32ths] = tempoChange
	}
	if scale != "" {
		var sc items.Scale

		err := sc.Parse(scale, 0)
		if err != nil {
			return fmt.Errorf("position %q contains invalid scale: %v", firstColumn, err)
		}

		s.currentScale = &sc
		s.Bars[s.currentBarNo].InnerScales[pos32ths] = &sc
	}
	return nil
}

// parseEventsLine parses a non-bar line / event line
func (s *Sketch) parseEventsLine(tabs []string) error {
	if s.barChangeRequired {
		if len(s.Bars) == 0 {
			b := NewBar()
			b.TempoChange = 120
			s.currentTempo = b.TempoChange
			s.currentTimeSignatur = b.TimeSig
			s.newBar(b)
			s.barChangeRequired = false
		} else {
			return fmt.Errorf("bar change # needed")
		}
	}

	if len(tabs) == 1 {
		return nil
	}

	colsData, _ := s.getColumnData(tabs)
	firstColumn := strings.TrimSpace(tabs[0])
	err := s.parsePosition(firstColumn)

	if err != nil {
		return err
	}

	for i, data := range colsData {
		data = strings.TrimSpace(data)
		if i < len(s.colOrder) {
			colName := s.colOrder[i]
			if s.isScore() && !s.Score.HasTrack(colName) {
				return fmt.Errorf("can't find track %q", colName)
			}
			s.Columns[colName] = append(s.Columns[colName], data)
		}
	}

	return nil
}

func (p *Sketch) AddColumn(name string) {
	if disallowedColNames[name] {
		panic(fmt.Sprintf("%q is not allowed as a column name of a sketch", name))
	}
	p.colOrder = append(p.colOrder, name)
	p.Columns[name] = []string{}
}

// parseLine parses a line in the body (everything after the =)
func (p *Sketch) ParseLine(tabs []string) error {
	if len(tabs) == 0 {
		return fmt.Errorf("must have tabs")
	}

	first := strings.TrimSpace(tabs[0])

	if first != "" {
		switch first[0] {
		case '1', '2', '3', '4', '5', '6', '7', '8', '9', '&':
			return p.parseEventsLine(tabs)
		}
	}

	if len(tabs) == 1 || (first != "" && first[0] == '#') {
		return p.parseBarLine(first)
	}

	return p.parseEventsLine(tabs)
}
