package muskel

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"

	"gitlab.com/gomidi/midi/mid"
	"gitlab.com/gomidi/midi/smf"
	"gitlab.com/gomidi/midi/smf/smfwriter"
)

type Score struct {
	Instruments []*Instrument
	Bars        []*Bar

	Meta               map[string]string
	Temperament        map[string]string
	PatternDefinitions map[string]*PatternDefinition
	Parts              map[string][2]int // part name to start bar no and last bar
	BodyComments       map[int]string    // line calculated from the start of the system
	HeaderComments     []string          // comments in the header, as they come

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
		Meta:               map[string]string{},
		Temperament:        map[string]string{},
		PatternDefinitions: map[string]*PatternDefinition{},
		Parts:              map[string][2]int{},
		BodyComments:       map[int]string{},
		isUnrolled:         false,
	}
}

func (s *Score) isPartial() bool {
	return s.Meta["partial"] != ""
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
func (s *Score) WriteSMF(midifile string) error {

	if !s.isUnrolled {
		ur, err := s.Unroll()
		if err != nil {
			return err
		}
		return ur.WriteSMF(midifile)
	}

	numTracks := uint16(1) // first track is for tempo and time signatures

	for _, instr := range s.Instruments {
		if instr.MIDIChannel >= 0 {
			numTracks++
		}
	}

	sw := NewSMFWriter(s)

	return mid.NewSMFFile(midifile, numTracks, sw.Write,
		smfwriter.TimeFormat(smf.MetricTicks(960)),
		smfwriter.Debug(log.New(os.Stdout, "write MIDI", log.Lshortfile)),
	)
}

// WriteTo writes the score to the given writer (in a formatted way)
func (s *Score) WriteTo(wr io.Writer) (err error) {

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	var bf bytes.Buffer

	fm := NewFormatter(s)
	fm.Format(&bf)

	_, err = bf.WriteTo(wr)
	return
}

// WriteTo writes the score to the given file (in a formatted way)
// It only writes to the file if the formatting was successful
func (s *Score) WriteToFile(filepath string) (err error) {
	dir, err := ioutil.TempDir(".", "muskel-fmt")
	if err != nil {
		return err
	}

	base := path.Base(filepath)

	f, err := os.Create(path.Join(dir, base))
	if err != nil {
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			f.Close()
			err = fmt.Errorf("%v", r)
		}
		os.RemoveAll(dir)
	}()

	err = s.WriteTo(f)
	if err != nil {
		return
	}
	f.Close()

	err = os.Rename(path.Join(dir, base), filepath)
	return
}
