package muskel

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"

	"math/rand"

	"gitlab.com/gomidi/midi/mid"
	"gitlab.com/gomidi/midi/smf"
	"gitlab.com/gomidi/midi/smf/smfwriter"
)

type Piece struct {
	currentBar        int
	numInstruments    int
	Instruments       []*Instrument
	rd                *bufio.Reader
	props             map[string]string
	temperament       map[string]string
	patterns          map[string]*PatternDefinition
	parts             map[string]int // part name to bar no
	jumps             map[int]string // bar no at the end of which the jump should be done mapped to the jump definition
	bars              []*Bar
	comments          map[int]string // line calculated from the start of the system
	headerComments    []string       // comments in the header, as they come
	inComment         bool
	inHeader          bool
	lineNo            int
	lineNosystemStart int
	writerSysLine     int
}

// TODO implement parseMIDICC
func (p *itemParser) parseMIDICC(data string) (cc MIDICC, err error) {
	return cc, fmt.Errorf("parseMIDICC is not implemented yet")
}

// TODO implement parseMIDIPitchbend
func (p *itemParser) parseMIDIPitchbend(string) (pb MIDIPitchbend, err error) {
	return pb, fmt.Errorf("parseMIDIPitchbend is not implemented yet")
}

// TODO implement parseMIDIAftertouch
func (p *itemParser) parseMIDIAftertouch(string) (at MIDIAftertouch, err error) {
	return at, fmt.Errorf("parseMIDIAftertouch is not implemented yet")
}

// TODO implement parseMIDIPolyAftertouch
func (p *itemParser) parseMIDIPolyAftertouch(string) (pt MIDIPolyAftertouch, err error) {
	return pt, fmt.Errorf("parseMIDIPolyAftertouch is not implemented yet")
}

func (p *itemParser) parseNote(data string) (item interface{}, err error) {
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

func (p *itemParser) parsePattern(data string) (*PatternCall, error) {
	if p.Getter == nil {
		return nil, fmt.Errorf("could not embed a pattern inside a resolved pattern")
	}
	var pc PatternCall
	err := pc.Parse(data)

	if err != nil {
		return nil, fmt.Errorf("could not parse pattern call: %s", err)
	}

	pd := p.GetPatternDefinition(pc.Name)

	if pd == nil {
		return nil, fmt.Errorf("could not find definition for pattern %s", pc.Name)
	}

	ppc := &pc

	pc.result, err = pd.Call(ppc, p)
	if err != nil {
		panic(fmt.Sprintf("could no call pattern %s with %q: %s", pc.Name, data, err))
	}

	err = pc.parseBars(pc.result)

	if err != nil {
		panic(fmt.Sprintf("could no call pattern %s with %q: %s", pc.Name, data, err))
	}

	return ppc, err
}

// ntuple has the form {c,e,d}3&
// where 3& is the ending position that defines the total length
func (p *itemParser) parseNTuple(data string) (nt NTuple, err error) {

	dd := strings.Split(data, "}")
	if len(dd) != 2 {
		return nt, fmt.Errorf("invalid n-tuple: %#v", "{"+data)
	}

	completedEndPos, endPos := positionTo32th("", dd[1])

	_ = completedEndPos

	data = strings.TrimSpace(dd[0])
	if data == "" {
		return
	}

	d := strings.Split(data, ",")

	var ntp NTuple
	ntp.endPos = endPos

	for pos, it := range d {
		itt, err := p.parseItem(it)
		if err != nil {
			return nt, fmt.Errorf("ERROR invalid n-tuple at position %v: %#v: %s", pos, it, err)
		}
		ntp.items = append(ntp.items, itt)
	}
	return ntp, nil
}

type itemParser struct {
	Getter
}

type RandomProbability struct {
	prob   uint8
	item   interface{}
	chosen bool
}

type RandomChooser struct {
	alternatives []interface{}
	chosen       int
}

func (p *itemParser) parseRandom(data string) (item interface{}, err error) {
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

		rp.item, err = p.parseItem(data[idx+1:])

		if err != nil {
			return nil, fmt.Errorf("invalid random value item: ?%s", data)
		}
		return &rp, nil
	case '(':
		alternatives := strings.Trim(data, "()")
		alt := strings.Split(alternatives, ",")
		var r RandomChooser
		for _, a := range alt {
			a = strings.TrimSpace(a)
			if a != "" {
				item, err = p.parseItem(a)

				if err != nil {
					return nil, fmt.Errorf("invalid random value item: %s", a)
				}
				r.alternatives = append(r.alternatives, item)
			}
		}
		return &r, nil
	default:
		return nil, fmt.Errorf("invalid random value: ?%s", data)
	}
}

func (p *itemParser) parseItem(data string) (interface{}, error) {
	data = strings.TrimSpace(data)
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
			return p.parseNTuple(data[1:])
		case '$':
			return p.parsePattern(data[1:])
		case 'S':
			return p.parseNote(data[1:])
		case 'Z':
			return p.parseNote(data[1:])
		case '?':
			return p.parseRandom(data[1:])
		case 'O':
			return p.parseOSC(data[1:])
		case '.':
			if data == "..." {
				return RepeatLastBarUntilChange{}, nil
			}
			if data == "./." {
				return RepeatLastBar{}, nil
			}
			if strings.Index("123456789", data[1:2]) != -1 && data[2] == '.' {
				n, _ := strconv.Atoi(data[1:2])
				//				fmt.Printf("RepeatLastNBarsUntilChange %v found\n", n)
				return RepeatLastNBarsUntilChange(n), nil
			}
		default:
			if len(data) > 1 {
				switch data[0:2] {
				case "CC":
					return p.parseMIDICC(data[2:])
				case "PB":
					return p.parseMIDIPitchbend(data[2:])
				case "AT":
					return p.parseMIDIAftertouch(data[2:])
				case "PT":
					return p.parseMIDIPolyAftertouch(data[2:])
				default:
				}
			}
			return p.parseNote(data)
		}
	}

	return nil, nil

}

func (p *Piece) parseData(data string) (interface{}, error) {
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
			it, err := (&itemParser{p}).parseItem(d[0])
			if err != nil {
				return nil, err
			}
			return it, nil
		default:
			var m MultiItem

			for _, dd := range d {
				it, err := (&itemParser{p}).parseItem(strings.TrimSpace(dd))

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

func (p *Piece) GetPatternDefinition(name string) *PatternDefinition {
	return p.patterns[name]
}

func (p *Piece) writeFirstTrack(wr *mid.SMFWriter) error {
	num := uint8(4)
	denom := uint8(4)

	for i, b := range p.bars {
		b.positions = make([]uint8, len(b.originalPositions))
		var oldPos string

		for pi, pos := range b.originalPositions {
			oldPos, b.positions[pi] = positionTo32th(oldPos, pos)
		}

		fmt.Printf("bar %v: %#v\n", i, b)

		if i > 0 {
			fmt.Printf("[firstrack] Forward(1, 0, 0)\n")
			wr.Forward(1, 0, 0)
			/*
				fmt.Printf("Forward(0, %v, %v)\n", uint32(num), uint32(denom))
				wr.Forward(0, uint32(num), uint32(denom))
			*/
		}

		if b.timeSigChange[0] > 0 {
			num = b.timeSigChange[0]
			denom = b.timeSigChange[1]
		}
		fmt.Printf("write meter %v/%v\n", num, denom)

		if b.tempoChange > 0 {
			wr.TempoBPM(b.tempoChange)
			fmt.Printf("write tempo change %v\n", b.tempoChange)
		}
		wr.Meter(num, denom)
		//wr.Forward(1, 0, 0)
	}

	return nil
}

func (p *Piece) trackBarNumbers() {
	for _, instr := range p.Instruments {
		var events []BarEvents

		for barNo, bar := range instr.events {
			es := make(BarEvents, len(bar))
			events = append(events, es)

			//			if bar.

			for barLine, ev := range bar {
				//
				/*
					set the bar number and position
				*/
				ev.BarNo = barNo
				ev.OriginalBarNo = barNo
				ev.DistanceToStartOfBarIn32th = p.bars[barNo].positions[barLine]

				events[barNo][barLine] = ev
			}
		}

		instr.events = events
		//instr.unrolled = events
	}
}

func (p *Piece) unrollBarsAndJumps() {
	var unrolledBars []*Bar
	var instrBarevents = map[int][]BarEvents{}
	var num, denom = uint8(4), uint8(4)

	var newBarNo int
	//var barMoved int

	for _, bar := range p.bars {

		bar.barNo = newBarNo
		if bar.timeSigChange[0] > 0 {
			num, denom = bar.timeSigChange[0], bar.timeSigChange[1]
		}

		bar.timeSig[0] = num
		bar.timeSig[1] = denom

		unrolledBars = append(unrolledBars, bar)
		newBarNo++

		for ii, instr := range p.Instruments {
			instrBarevents[ii] = append(instrBarevents[ii], instr.events[bar.originalBarNo])
		}

		if jump, has := p.jumps[bar.originalBarNo]; has {
			startJump := bar.originalBarNo

			// unroll the repeating bars in p.Bars for the first track and for reference
			//		(and renumber them properly)
			for j, bar2ndloop := range p.bars[p.parts[jump]:startJump] {
				newBarNo++
				nub := bar2ndloop.Dup()
				if bar2ndloop.timeSigChange[0] > 0 {
					num, denom = bar2ndloop.timeSigChange[0], bar2ndloop.timeSigChange[1]
				}
				nub.barNo = newBarNo
				nub.timeSig[0] = num
				nub.timeSig[1] = denom
				unrolledBars = append(unrolledBars, nub)
				for ii, instr := range p.Instruments {
					instrBarevents[ii] = append(instrBarevents[ii], instr.events[p.parts[jump]+j])
				}
			}
		}

		/*
			maybe we want an additional field for the original bar as in the notes (and the original line)
		*/

		/*
					the unrolled events arent any longer in the synced line
			an therefor need the reference to the original bar number
			and then the distance from this bars start. other than that, they are sorted
			by this distance. So before unrolling, we simply tag the events
		*/

		/*
			TODO go through each event
			on the first run, pick up every event that fits into its bar
			when we reach the end of a bar where the jump happens (jumpingBar)
			we restart from the bar were we jump to and set the bars of the
			copied events to the new real bar.

			maybe we want an additional field for the original bar as in the notes (and the original line)
		*/

		// unroll the repeating bars in p.Bars for the first track and for reference
		//		(and renumber them properly)

		/*
			now unroll the random elements and selections indicated by ? to allow for randomized pattern selection
		*/

		/*
			for _, instr := range p.Instruments {
				var unrolled []*Event
				for _, ev := range instr.unrolled {
					// handle patterns
					_ = ev
				}

				instr.unrolled = unrolled
			}
		*/
	}

	p.bars = unrolledBars

	for ii, instr := range p.Instruments {
		instr.events = instrBarevents[ii]
	}
}

func (p *Piece) flattenInstrumentEvents() {
	var currentlyRepeatingBars []BarEvents
	var indexWithinCurrentlyRepeatingBars int

	for _, instr := range p.Instruments {
		var events []*Event
		for barNo, bar := range instr.events {

			numBar, denomBar := p.bars[barNo].timeSig[0], p.bars[barNo].timeSig[1]
			lenBar := length32ths(numBar, denomBar)

			numRep, untilNext := bar.RepeatingBars()

			if numRep > 0 {
				if !untilNext {
					for _, ev := range instr.events[barNo-1] {
						eNu := ev.Dup()
						eNu.BarNo = barNo
						// skip empty items
						if eNu.Item != nil && int(eNu.DistanceToStartOfBarIn32th) < lenBar {
							events = append(events, eNu)
						}
					}
				} else {
					currentlyRepeatingBars = instr.events[barNo-numRep : barNo]
					indexWithinCurrentlyRepeatingBars = 0
					for _, ev := range instr.events[barNo-1] {
						eNu := ev.Dup()
						eNu.BarNo = barNo
						// skip empty items
						if eNu.Item != nil && int(eNu.DistanceToStartOfBarIn32th) < lenBar {
							events = append(events, eNu)
						}
					}
				}
				continue
			}

			if bar.isEmpty() {
				switch len(currentlyRepeatingBars) {
				case 0:
					continue
				case 1:
					for _, ev := range currentlyRepeatingBars[0] {
						eNu := ev.Dup()
						eNu.BarNo = barNo

						// skip empty items
						if eNu.Item != nil && int(eNu.DistanceToStartOfBarIn32th) < lenBar {
							events = append(events, eNu)
						}
					}
					continue
				default:
					indexWithinCurrentlyRepeatingBars = (indexWithinCurrentlyRepeatingBars + 1) % len(currentlyRepeatingBars)
					for _, ev := range currentlyRepeatingBars[indexWithinCurrentlyRepeatingBars] {
						eNu := ev.Dup()
						eNu.BarNo = barNo
						// skip empty items
						if eNu.Item != nil && int(eNu.DistanceToStartOfBarIn32th) < lenBar {
							events = append(events, eNu)
						}
					}
					continue
				}
			}

			for barLine, ev := range bar {
				//
				/*
					set the bar number and position
				*/
				eNu := ev.Dup()
				eNu.BarNo = barNo
				eNu.DistanceToStartOfBarIn32th = p.bars[barNo].positions[barLine]

				// skip empty items
				if ev.Item != nil && int(eNu.DistanceToStartOfBarIn32th) < lenBar {
					events = append(events, ev)
				}
			}
		}

		instr.unrolled = events
	}

}

func (p *Piece) evalRandomItems() {
	for _, instr := range p.Instruments {
		var unrolled []*Event
		for _, ev := range instr.unrolled {
			switch v := ev.Item.(type) {
			case *RandomChooser:
				v.chosen = rand.Intn(len(v.alternatives))
				nuEv := ev.Dup()
				nuEv.Item = v.alternatives[v.chosen]
				unrolled = append(unrolled, nuEv)
			case *RandomProbability:
				if got := rand.Intn(int(v.prob)); got >= 50 {
					nuEv := ev.Dup()
					nuEv.Item = v.item
					unrolled = append(unrolled, nuEv)
				}

			default:
				unrolled = append(unrolled, ev)
			}
			// handle patterns
			_ = ev
		}

		instr.unrolled = unrolled
	}
}

func (p *Piece) unfoldPatterns() {
	for _, instr := range p.Instruments {
		var unrolled []*Event
		for idx, ev := range instr.unrolled {
			// handle patterns

			switch v := ev.Item.(type) {
			case *PatternCall:
				/*
						TODO
						1. calc the distance of the next event in 32ths (respecting bar changes etc)
						2. for each pattern event:
						   a. calc the distance to the previous pattern event in 32ths
						   b. if it is smaller than the next normal event, write it to unrolled
						   c. otherwise skip the pattern event loop

						ad 1:
						  we need the distance between the bars and corresponding from the start of the bar to
						  the event. then for each bar the length in 32ths has to be calculated + distance from start
						  of event - distance from start of previous

					    ad 2:
						  a for the first pattern event it is either 0 (sync start) or based on the first position. for
						    the following pattern events it is the position related to current position minus the relative position
							of the previous pattern event

					Question: what to do, if there is no next event?
					I think then the pattern can be fully unrolled
				*/
				// there is a following event
				if idx+1 < len(instr.unrolled) {
					// 1. calc the distance of the next event in 32ths (respecting bar changes etc)
					diffBars := instr.unrolled[idx+1].BarNo - ev.BarNo
					var diff32ths int
					for didx := 0; didx < diffBars; didx++ {
						diff32ths += p.bars[ev.BarNo+didx].length32th()
					}

					diff32ths += int(instr.unrolled[idx+1].DistanceToStartOfBarIn32th)
					diff32ths -= int(ev.DistanceToStartOfBarIn32th)

					// 2a. calc the distance to the previous pattern event in 32ths
					preventRelPosition := uint8(0)
					_ = preventRelPosition
					prevPos := ""
					barLenInBetween := 0
				pattenEventLoop:
					for idxBarPev, barpev := range v.Events {
						currentBarNo := ev.BarNo + idxBarPev

						for idxPev, pev := range barpev {
							_ = pev
							nuev := ev.Dup()
							nuev.BarNo = currentBarNo
							nuev.Item = pev.item

							if idxBarPev == 0 && idxPev == 0 {
								//dist := pev.position
								if !v.SyncFirst {
									prevPos, nuev.DistanceToStartOfBarIn32th = positionTo32th(prevPos, pev.position)
								}
								preventRelPosition = nuev.DistanceToStartOfBarIn32th

								// 2c. otherwise skip the pattern event loop
								if int(nuev.DistanceToStartOfBarIn32th) >= diff32ths {
									break pattenEventLoop
								}

								// 2b. if it is smaller than the next normal event, write it to unrolled
								unrolled = append(unrolled, nuev)

							} else {
								prevPos, nuev.DistanceToStartOfBarIn32th = positionTo32th(prevPos, pev.position)

								// 2c. otherwise skip the pattern event loop
								if barLenInBetween+int(nuev.DistanceToStartOfBarIn32th) >= diff32ths {
									break pattenEventLoop
								}

								// 2b. if it is smaller than the next normal event, write it to unrolled
								unrolled = append(unrolled, nuev)
							}
						}

						prevPos = ""
						barLenInBetween += p.bars[currentBarNo].length32th()

					}
				}

				//v.Events
			default:
				unrolled = append(unrolled, ev)
			}
		}

		instr.unrolled = unrolled
	}
}

func (p *Piece) unrollInstruments() error {

	// attach the barnumbers and position in 32th relative to start to the event
	p.trackBarNumbers()

	// unroll the repetition within an instrument and the total jumps
	// thereby creating new bars and attaching those new bar numbers in addition
	// to the original bar numbers to the events
	p.unrollBarsAndJumps()

	// since now we have the barnumbers properly attached to events, we can flatten
	// the events so that we have a single line of events where we could easily look
	// for the next event
	p.flattenInstrumentEvents()

	// evaluate randomness here before the pattern unfolding, to be able to randomly choose
	// patterns
	p.evalRandomItems()

	// unfold the patterns in a way that they are interrupted by the next non empty event
	p.unfoldPatterns()

	// evalute randomness a second time, so that randomness produced by patterns can be evaluated
	p.evalRandomItems()

	return nil
}

func (p *Piece) writeSMF(wr *mid.SMFWriter) error {

	err := p.unrollInstruments()
	if err != nil {
		return err
	}

	err = p.writeFirstTrack(wr)

	if err != nil {
		return err
	}

	for _, instr := range p.Instruments {
		// ignore instruments without a MIDIchannel
		if instr.MIDIChannel < 0 {
			continue
		}

		fmt.Println("EOT")
		wr.EndOfTrack()
		iw := newInstrumentSMFWriter(p, wr, instr)
		err = iw.writeUnrolled()
		//err = iw.writeTrack()
		//		err = p.writeInstrumentTrack(wr, instr)
		if err != nil {
			return err
		}
	}

	fmt.Println("EOT")
	wr.EndOfTrack()
	return nil
}

func (p *Piece) WriteSMF(midifile string) error {

	numTracks := uint16(1) // first track is for tempo and time signatures

	for _, instr := range p.Instruments {
		if instr.MIDIChannel >= 0 {
			numTracks++
		}
	}

	return mid.NewSMFFile(midifile, numTracks, p.writeSMF,
		smfwriter.TimeFormat(smf.MetricTicks(960)),
		//		smfwriter.Debug(log.New(os.Stdout, "write MIDI", log.Lshortfile)),
	)
}

func (p *Piece) WriteFormatted(filepath string) (err error) {
	dir, err := ioutil.TempDir(".", "muskel-fmt")
	if err != nil {
		return err
	}

	base := path.Base(filepath)

	f, err := os.Create(path.Join(dir, base))
	if err != nil {
		return err
	}
	var bf bytes.Buffer
	p.writerSysLine = 0

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error: %v", r)
		}
		f.Close()
		os.RemoveAll(dir)
	}()

	p.format(&bf)

	_, err = bf.WriteTo(f)
	if err != nil {
		return
	}

	err = os.Rename(path.Join(dir, base), filepath)
	return
}

func (p *Piece) writeSystemLine(bf *bytes.Buffer, line string) {
	if comment, has := p.comments[p.writerSysLine]; has {
		p.writerSysLine++
		bf.WriteString(comment + "\n")
		p.writeSystemLine(bf, line)
		return
	}
	p.writerSysLine++
	bf.WriteString(line + "\n")
}

func (p *Piece) format(bf *bytes.Buffer) {
	//bf.WriteString("\n// info\n")

	//fmt.Fprintf(bf, "%s %s\n", pad("version:", 15), time.Now().Format(time.RFC822))

	for k, v := range p.props {
		fmt.Fprintf(bf, "%s %s\n", pad(k+":", 15), v)
		//fmt.Fprintf(bf, "%15s+ %s\n", k+":", v)
	}

	//bf.WriteString("\n// temperament\n")

	bf.WriteString("\n\n")

	for k, v := range p.temperament {
		fmt.Fprintf(bf, "/%s/%s\n", k, v)
	}

	bf.WriteString("\n\n")

	//bf.WriteString("\n// pattern definitions\n")

	for k, v := range p.patterns {
		//fmt.Fprintf(bf, "%s %s\n", pad("$"+k+":", 15), v)
		fmt.Fprintf(bf, "%s %s\n", pad("$"+k+":", 15), v.Original)
		//fmt.Fprintf(bf, "%15s+ %s\n", "$"+k+":", v.Original)
	}

	bf.WriteString("\n\n")

	for _, comment := range p.headerComments {
		bf.WriteString(comment + "\n")
	}

	//fmt.Fprintf(bf, "\n\nBPM: %0.2f\n", p.tempoBPM)

	bf.WriteString("\n=\n")

	// 9 whitespace to first pipe
	l := "         |"

	for _, instr := range p.Instruments {
		instr.calcColWidth()
		//instrColWidths[i] = instr.colWidth
		l += fmt.Sprintf(" %s |", instr.pad(instr.Name))
	}

	p.writeSystemLine(bf, l)

	l = "Ch       |"

	for _, instr := range p.Instruments {
		if instr.MIDIChannel == -1 {
			l += fmt.Sprintf(" %v |", instr.pad("-"))
		} else {
			l += fmt.Sprintf(" %v |", instr.pad(fmt.Sprintf("%v", instr.MIDIChannel+1)))
		}
	}

	p.writeSystemLine(bf, l)

	l = "Bank     |"

	for _, instr := range p.Instruments {
		if instr.MIDIBank == -1 {
			l += fmt.Sprintf(" %v |", instr.pad("-"))
		} else {
			l += fmt.Sprintf(" %v |", instr.pad(fmt.Sprintf("%v", instr.MIDIBank+1)))
		}
	}

	p.writeSystemLine(bf, l)

	l = "Prog     |"

	for _, instr := range p.Instruments {
		if instr.MIDIProgram == -1 {
			l += fmt.Sprintf(" %v |", instr.pad("-"))
		} else {
			l += fmt.Sprintf(" %v |", instr.pad(fmt.Sprintf("%v", instr.MIDIProgram+1)))
		}
	}

	p.writeSystemLine(bf, l)

	l = "Vol      |"

	for _, instr := range p.Instruments {
		if instr.MIDIVolume == -1 {
			l += fmt.Sprintf(" %v |", instr.pad("-"))
		} else {
			l += fmt.Sprintf(" %v |", instr.pad(fmt.Sprintf("%v", instr.MIDIVolume+1)))
		}
	}

	p.writeSystemLine(bf, l)

	for i, bar := range p.bars {
		if bar.timeSigChange[0] != 0 {
			l = fmt.Sprintf("%v/%v", bar.timeSigChange[0], bar.timeSigChange[1])
		} else {
			l = ""
		}

		if bar.tempoChange > 0 {
			l += fmt.Sprintf("@%g", bar.tempoChange)
		}

		p.writeSystemLine(bf, l)

		for pi, pos := range bar.originalPositions {
			switch {
			case pos[0] == '&':
				l = fmt.Sprintf("     %s |", pad(pos, 3))
				//l = fmt.Sprintf("     %3s+ |")
			case strings.Index("123456789", pos[0:1]) != -1:
				l = fmt.Sprintf("    %s |", pad(pos, 4))
				//l= fmt.Sprintf("    %4s+ |")
			case pos[0] == '.':
				l = fmt.Sprintf("     %s |", pad(pos, 3))
				//l = fmt.Sprintf("     %3s+ |")
			case pos[0] == ';':
				l = fmt.Sprintf("     %s |", pad(pos, 3))
				//l = fmt.Sprintf("     %3s+ |")
			}

			for _, instr := range p.Instruments {
				be := instr.events[i]

				l += fmt.Sprintf(" %s |", instr.pad(be[pi].originalData))

			}

			if pi == 0 {
				for prt, br := range p.parts {
					if br == i {
						l += fmt.Sprintf(" %s", prt)
					}
				}
			}

			if pi == len(bar.originalPositions)-1 {
				for br, jmp := range p.jumps {
					if br == i {
						l += fmt.Sprintf(" %s", jmp)
					}
				}
			}

			p.writeSystemLine(bf, l)
		}

	}
	//bf.WriteString("// END\n")
	bf.WriteString("\n")

}

func (p *Piece) newBar(b *Bar) {
	b.barNo = p.currentBar
	b.originalBarNo = p.currentBar
	p.bars = append(p.bars, b)
	p.currentBar++
}

func (p *Piece) parseSystemLine(line string) error {
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
				if _, has := p.parts[part]; !has {
					return fmt.Errorf("could not jump to part %#v: not found", part)
				}
				p.jumps[p.currentBar] = part
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
			if i > len(p.Instruments)-1 {
				instr = &Instrument{}
				instr.MIDIProgram = -1
				instr.MIDIChannel = -1
				instr.MIDIVolume = -1
				instr.MIDIBank = -1
				p.Instruments = append(p.Instruments, instr)
			} else {
				instr = p.Instruments[i]
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
					p.bars[p.currentBar].originalPositions = append(p.bars[p.currentBar].originalPositions, position)
				}
			}
		}

		part := strings.TrimSpace(tabs[len(tabs)-1])

		switch len(part) {
		case 1:
			// part definition
			if _, hasPart := p.parts[part]; hasPart {
				return fmt.Errorf("part %#v already exists", part)
			} else {
				p.parts[part] = p.currentBar
			}

		case 0:
		// do nothing
		default:
			// probably a jump; check
			p.jumps[p.currentBar] = part
			//fmt.Printf("line: %v, position: %v part: %v\n", lineNo, position, part)
		}

	}

	return nil
}

func (p *Piece) parseHeader(lines string) error {
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

			p.temperament[strings.TrimSpace(keyVal[0])] = strings.TrimSpace(keyVal[1])
		case '$':
			var pd PatternDefinition
			err := pd.Parse(l[1:])
			if err != nil {
				return fmt.Errorf("invalid pattern definition line %#v: %s", l, err)
			}
			p.patterns[pd.Name] = &pd
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

			p.props[key] = val

		}

	}

	//fmt.Printf("properties: %#v\n", props)
	//fmt.Printf("temperament: %#v\n", temperament)
	//fmt.Printf("patterns: %#v\n", patterns)

	return nil
}

func (p *Piece) addSystemComment(relLine int, line []byte) {
	p.comments[relLine] = string(line)
}

func (p *Piece) addHeaderComment(line []byte) {
	p.headerComments = append(p.headerComments, string(line))
}

func (p *Piece) addComment(line []byte) {
	if p.inHeader {
		p.addHeaderComment(line)
		return
	}
	p.addSystemComment(p.lineNo-p.lineNosystemStart, line)
}

func (p *Piece) Parse() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error: %v", r)
		}
	}()
	err = p.parse()
	return
}

func (p *Piece) parse() (err error) {
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

func NewPiece(r io.Reader) *Piece {
	p := &Piece{}
	p.currentBar = -1
	p.numInstruments = -1
	p.rd = bufio.NewReader(r)
	p.props = map[string]string{}
	p.temperament = map[string]string{}
	p.patterns = map[string]*PatternDefinition{}
	p.parts = map[string]int{}
	p.jumps = map[int]string{}
	p.comments = map[int]string{}
	return p
}
