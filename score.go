package muskel

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"

	"gitlab.com/gomidi/midi/mid"
	"gitlab.com/gomidi/midi/smf"
	"gitlab.com/gomidi/midi/smf/smfwriter"
)

type Score struct {
	FileName    string
	Instruments []*Instrument
	Bars        []*Bar

	Meta                       map[string]string
	HeaderIncludes             []string
	Temperament                map[string]string
	PatternDefinitions         map[string]*PatternDefinition
	IncludedPatternDefinitions map[string]*PatternDefinition
	Parts                      map[string][2]int // part name to start bar no and last bar
	BodyComments               map[int]string    // line calculated from the start of the system
	HeaderComments             []string          // comments in the header, as they come

	SmallColumns        bool // if the formatting should have no distance to the column lines
	isUnrolled          bool
	barNumbersTracked   bool
	lineWhereBodyStarts int
	lastLine            int
}

func (s *Score) AddMissingProperties() {
	if s.isPartial() {
		return
	}
	prefill := map[string]string{
		"composer":  "[insert composer here]",
		"title":     "[insert title here]",
		"date":      "[insert date here]",
		"version":   "[insert version here]",
		"copyright": "[insert copyright/license notice here]",
		"genre":     "[insert genre here]",
		"tags":      "[insert tags here] (comma separated)",
	}

	for k, v := range prefill {
		if s.Meta[k] == "" {
			s.Meta[k] = v
		}
	}
}

func NewScore() *Score {
	return &Score{
		Meta:                       map[string]string{},
		Temperament:                map[string]string{},
		PatternDefinitions:         map[string]*PatternDefinition{},
		Parts:                      map[string][2]int{},
		BodyComments:               map[int]string{},
		isUnrolled:                 false,
		IncludedPatternDefinitions: map[string]*PatternDefinition{},
	}
}

func (s *Score) isPartial() bool {
	return s.Meta["partial"] != ""
}

func (p *Score) include(file string) (*Score, error) {
	//	wd, _ := os.Getwd()
	//	fmt.Println(wd)
	//	sc, err := ParseFile(filepath.Join(wd, file+".mski"))

	found, err := findInclude(file + ".mskl")
	if err != nil {
		return nil, fmt.Errorf("Error while including file %q: file not found", file)
	}

	if found == p.FileName {
		return nil, fmt.Errorf("file can't include itself")
	}
	b, err := ioutil.ReadFile(found)

	if err != nil {
		return nil, fmt.Errorf("Error while reading include file %q: %s", file, err.Error())
	}

	var sc *Score
	sc, err = Parse(bytes.NewReader(b), found)

	if err != nil {
		return nil, fmt.Errorf("Error while including file %q: %s", file, err.Error())
	}

	return sc, nil
}

// trackBarNumbers tracks the bar numbers and positions for each event
func (p *ScoreUnroller) trackBarNumbers() {

	for _, instr := range p.dest.Instruments {
		var events []BarEvents

		for barNo, bar := range instr.events {
			es := make(BarEvents, len(bar))
			events = append(events, es)

			//			if bar.

			for barLine, ev := range bar {
				nuev := ev.Dup()
				//
				/*
					set the bar number and position
				*/
				nuev.BarNo = barNo
				nuev.OriginalBarNo = barNo
				//fmt.Printf("barNo: %v, BarLine: %v p.Bar: %#v\n", barNo, barLine, p.Bars[barNo])
				nuev.DistanceToStartOfBarIn32th = p.dest.Bars[barNo].positions[barLine]

				events[barNo][barLine] = nuev
			}
		}

		//		nu := instr.Dup()

		instr.events = events

		//		p.dest.Instruments = append(p.dest.Instruments, nu)
		//instr.unrolled = events
	}
	//	p.dest.barNumbersTracked = true
}

// Unroll unrolls repetitions, patterns and randomness and returns a score
// without repetition, randomness and patterns
func (s *Score) Unroll() (dest *Score, err error) {

	for _, b := range s.Bars {
		b.positions = make([]uint, len(b.originalPositions))
		var oldPos string

		for pi, pos := range b.originalPositions {
			oldPos, b.positions[pi], err = positionTo32th(oldPos, pos)
			if err != nil {
				return nil, err
			}
		}
	}

	ur := newScoreUnroller(s)
	ur.dest.SmallColumns = s.SmallColumns
	err = ur.unrollInstruments(s.Bars, s.Instruments)
	return ur.dest, err
}

func (s *Score) AddInstrument(i *Instrument) error {
	if s.hasInstrument(i.Name) {
		return fmt.Errorf("instrument %q already defined", i.Name)
	}
	s.Instruments = append(s.Instruments, i)
	return nil
}

func (p *Score) hasInstrument(name string) bool {
	for _, instr := range p.Instruments {
		if instr.Name == name {
			return true
		}
	}
	return false
}

func (p *Score) enroll() {

	// enroll the instrument events again
	for _, instr := range p.Instruments {
		instr.events = []BarEvents{}
		var be BarEvents
		var lastBarNo int

		//		var lastBar = p.Bars[len(p.Bars)-1]

		for _, ev := range instr.unrolled {
			//				fmt.Printf("event: %#v\n", ev)
			//				fmt.Printf("ev.BarNo: %v  len(p.score.Bars): %v\n", ev.BarNo, len(p.score.Bars))
			p.Bars[ev.BarNo].ensurePositionExist(ev.DistanceToStartOfBarIn32th)

			diff := ev.BarNo - lastBarNo

			for d := 0; d < diff; d++ {
				instr.events = append(instr.events, be)
				be = BarEvents{}
			}

			lastBarNo = ev.BarNo
			be = append(be, ev)
		}

		//		fmt.Printf("last bar with event: %v, last bar total: %v len bars: %v\n", lastBarNo, p.Bars[len(p.Bars)-1].barNo, len(p.Bars))

		//			fmt.Printf("len bars in instrument: %v\n", len(be))
		instr.events = append(instr.events, be)

		if lastBarNo < p.Bars[len(p.Bars)-1].barNo {
			instr.events = append(instr.events, BarEvents{})
		}

		instr.calcColWidth(p.isUnrolled)
	}
}

// WriteSMF writes the score to the given SMF file
func (s *Score) WriteSMF(midifile string, options ...smfwriter.Option) error {

	if !s.isUnrolled {
		ur, err := s.Unroll()
		if err != nil {
			return err
		}
		return ur.WriteSMF(midifile)
	}

	numTracks := uint16(2) // first track is for time signatures, second track is for tempo changes

	for _, instr := range s.Instruments {
		if instr.MIDIChannel >= 0 {
			numTracks++
		}
	}

	sw := NewSMFWriter(s)

	options = append(
		[]smfwriter.Option{
			smfwriter.TimeFormat(smf.MetricTicks(960)),
		}, options...)

	if DEBUG {
		options = append(options, smfwriter.Debug(debugLog{}))
	}

	return mid.NewSMFFile(midifile, numTracks, sw.Write, options...)

}

type debugLog struct{}

func (debugLog) Printf(format string, vals ...interface{}) {
	fmt.Fprintf(os.Stdout, format+"\n", vals...)
}

// WriteTo writes the score to the given writer (in a formatted way)
func (s *Score) WriteTo(wr io.Writer) (n int64, err error) {

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	var bf bytes.Buffer

	fm := NewFormatter(s)
	fm.Format(&bf)

	n, err = bf.WriteTo(wr)
	return
}

// WriteTo writes the score to the given file (in a formatted way)
// It only writes to the file if the formatting was successful
func (s *Score) WriteToFile(filepath string) (err error) {
	dir, err := ioutil.TempDir(".", "muskel-fmt")
	if err != nil {
		return fmt.Errorf("can't create tempdir: %v", err)
	}

	base := path.Base(filepath)

	f, err := os.Create(path.Join(dir, base))
	if err != nil {
		return fmt.Errorf("can't create tempdir: %v", err)
	}

	defer func() {
		if r := recover(); r != nil {
			f.Close()
			err = fmt.Errorf("%v", r)
		}
		os.RemoveAll(dir)
	}()

	_, err = s.WriteTo(f)
	if err != nil {
		return
	}
	err = f.Close()
	if err != nil {
		return fmt.Errorf("can't close file %q: %v", path.Join(dir, base), err)
	}

	err = os.Rename(path.Join(dir, base), filepath)
	if err != nil {
		err = fmt.Errorf("can't move %q to %q", path.Join(dir, base), filepath)
	}
	return
}
