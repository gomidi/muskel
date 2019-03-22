package muskel

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"

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
	prefill := map[string]string{
		"composer":  "",
		"artist":    "",
		"title":     "",
		"date":      "",
		"version":   "",
		"copyright": "",
		"genre":     "",
		"tags":      "",
	}

	for k, v := range prefill {
		if _, has := s.Meta[k]; !has {
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

func NewInstrument(name string) *Instrument {
	instr := &Instrument{}
	instr.MIDIProgram = -1
	instr.MIDIChannel = -1
	instr.MIDIVolume = -1
	instr.MIDIBank = -1
	return instr
}

func (s *Score) RmInstrument(name string) error {
	var instrs []*Instrument

	name = strings.TrimSpace(name)
	var found bool

	for _, instr := range s.Instruments {
		if instr.Name != name {
			instrs = append(instrs, instr)
		} else {
			found = true
		}
	}

	if found {
		s.Instruments = instrs
		return nil
	}

	return fmt.Errorf("could not find instrument %q", name)
}

func (s *Score) getInstrument(name string) *Instrument {
	for _, instr := range s.Instruments {
		if instr.Name == name {
			return instr
		}
	}
	return nil
}

func (s *Score) RenameInstrument(old, nu string) error {
	in := s.getInstrument(old)
	if in == nil {
		return fmt.Errorf("instrument with name %q could not be found", old)
	}

	in.Name = nu
	err := s.WriteToFile(s.FileName)

	var includes = map[string]bool{}
	for _, b := range s.Bars {
		if b.include != "" {

			if includes[b.include] {
				continue
			}

			target, err := s.include(b.include)
			if err != nil {
				return err
			}

			includes[b.include] = true

			target.RenameInstrument(old, nu)
		}
	}
	return err
}

func (s *Score) syncInclude(include string) error {

	target, err := s.include(include)
	if err != nil {
		// fmt.Printf("ERROR while including %q\n", include)
		return err
	}

	// fmt.Printf("including %q\n", include)

	var changed bool

	for _, instr := range s.Instruments {
		if !target.hasInstrument(instr.Name) {
			target.AddInstrument(instr)
			changed = true
		} else {
			ti := target.getInstrument(instr.Name)

			if ti == nil {
				return fmt.Errorf("could not get instrument %q from include %q although it says, is has this instrument", instr.Name, include)
			}

			ti.MIDIBank = instr.MIDIBank
			ti.MIDIChannel = instr.MIDIChannel
			ti.MIDIProgram = instr.MIDIProgram
			ti.MIDITranspose = instr.MIDITranspose
			ti.MIDIVolume = instr.MIDIVolume
			ti.FileGroup = instr.FileGroup
			ti.PitchbendRange = instr.PitchbendRange
			changed = true
		}
	}

	if changed {
		// fmt.Printf("writing include %q to %q\n", include, target.FileName)
		target.WriteToFile(target.FileName)
	}

	return target.SyncInstruments()
}

func (s *Score) SyncInstruments() error {
	var includes = map[string]bool{}
	for _, b := range s.Bars {
		if b.include != "" {

			if includes[b.include] {
				continue
			}

			includes[b.include] = true

			err := s.syncInclude(b.include)

			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Score) AddInstrument(i *Instrument) error {
	i.Name = strings.TrimSpace(i.Name)
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
func (s *Score) WriteSMF(midifile string, options ...smfwriter.Option) (err error) {
	defer func() {
		if !DEBUG {
			if r := recover(); r != nil {
				err = fmt.Errorf("%v", r)
			}
		}
	}()

	if !s.isUnrolled {
		ur, err := s.Unroll()
		if err != nil {
			return err
		}
		return ur.WriteSMF(midifile)
	}

	hasPlaceholder := strings.Index(midifile, "%s") > -1

	if !hasPlaceholder {
		return s.writeToFile(midifile, "*", options...)
	}

	var fileGroups = map[string]string{}
	for _, instr := range s.Instruments {
		fileGroups[instr.FileGroup] = fmt.Sprintf(midifile, instr.FileGroup)
	}

	var errs errors

	for grp, fl := range fileGroups {
		err := s.writeToFile(fl, grp, options...)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) == 0 {
		return nil
	}

	return errs
}

type errors []error

func (e errors) Error() string {
	var bf strings.Builder
	bf.WriteString("The following errors happened:\n")
	for _, err := range e {
		bf.WriteString(err.Error() + "\n")
	}

	return bf.String()
}

func (s *Score) writeToFile(midifile, filegroup string, options ...smfwriter.Option) error {

	numTracks := uint16(2) // first track is for time signatures, second track is for tempo changes

	for _, instr := range s.Instruments {
		if instr.MIDIChannel >= 0 && (filegroup == "*" || instr.FileGroup == filegroup) {
			numTracks++
		}
	}

	sw := NewSMFWriter(s, filegroup)

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

func (s *Score) Formatter() *Formatter {
	return NewFormatter(s)
}

// WriteTo writes the score to the given writer (in a formatted way)
func (s *Score) WriteTo(wr io.Writer) (err error) {

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	fm := s.Formatter()
	fm.WriteTo(wr)

	return nil
}

func tempDir(prefix string) (dir string, err error) {
	dir = fmt.Sprintf("%s-%v", prefix, time.Now().Unix())
	err = os.Mkdir(dir, 0775)
	return
}

func (s *Score) renamePatternInCalls(old, nu string) (err error) {
	for _, instr := range s.Instruments {
		for ii, bev := range instr.events {
			for i, ev := range bev {
				switch v := ev.Item.(type) {
				case *PatternCall:
					if v.Name == old {
						v.Name = nu
					}
					ev.Item = v
					ev.originalData = strings.Replace(ev.originalData, old, nu, 1)
					bev[i] = ev
				}
			}

			instr.events[ii] = bev
		}
	}

	return nil

}

func (s *Score) hasPattern(name string) bool {
	_, hasOwn := s.PatternDefinitions[name]
	_, hasIncluded := s.IncludedPatternDefinitions[name]

	return hasOwn || hasIncluded
}

func (s *Score) renamePatternDefinition(old, nu string) error {
	_, hasOwn := s.PatternDefinitions[old]
	_, hasIncluded := s.IncludedPatternDefinitions[old]

	if !s.hasPattern(old) {
		return fmt.Errorf("could not find a pattern with the name %q", old)
	}

	if s.hasPattern(nu) {
		return fmt.Errorf("a pattern with name %q already exists", nu)
	}

	if hasOwn {
		pt := s.PatternDefinitions[old]

		delete(s.PatternDefinitions, old)

		pt.Name = nu

		s.PatternDefinitions[nu] = pt
		return s.renamePatternInCalls(old, nu)
	}

	if hasIncluded {
		pt := s.IncludedPatternDefinitions[old]

		delete(s.IncludedPatternDefinitions, old)

		pt.Name = nu

		s.IncludedPatternDefinitions[nu] = pt
	}

	return nil
}

func (s *Score) RenamePattern(old, nu string) (err error) {
	err = s.renamePatternDefinition(old, nu)

	if err != nil {
		return
	}

	var includes = map[string]bool{}

	for _, b := range s.Bars {
		if b.include != "" {

			if includes[b.include] {
				continue
			}

			target, err := s.include(b.include)
			if err != nil {
				return err
			}

			target.RenamePattern(old, nu)
			//target.WriteToFile(target.FileName)

			includes[b.include] = true
		}
	}

	s.renamePatternInCalls(old, nu)

	for _, patt := range s.PatternDefinitions {
		it := strings.Split(patt.Original, " ")

		var changed bool = false

		for i, thing := range it {
			stripped := strings.TrimSpace(thing)
			stripped = strings.TrimRight(stripped, ":")
			stripped = strings.TrimRight(stripped, "+")
			stripped = strings.TrimRight(stripped, "+")
			stripped = strings.TrimRight(stripped, "+")
			stripped = strings.TrimRight(stripped, "-")
			stripped = strings.TrimRight(stripped, "-")
			stripped = strings.TrimRight(stripped, "-")
			stripped = strings.TrimRight(stripped, "*")
			stripped = strings.TrimLeft(stripped, "1234567890")
			stripped = strings.TrimLeft(stripped, "1234567890")
			stripped = strings.TrimLeft(stripped, "!")

			if stripped == old {
				it[i] = strings.Replace(thing, old, nu, 1)
				changed = true
			}
		}

		if changed {
			patt.Original = strings.Join(it, " ")
		}
	}

	return s.WriteToFile(s.FileName)

}

// WriteTo writes the score to the given file (in a formatted way)
// It only writes to the file if the formatting was successful
func (s *Score) WriteToFile(filep string) (err error) {

	// dir, err := ioutil.TempDir(".", "muskel-fmt")
	dir, err := tempDir("muskel-fmt")
	if err != nil {
		return fmt.Errorf("can't create tempdir: %v", err)
	}

	base := path.Base(filep)

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

	err = s.WriteTo(f)
	if err != nil {
		return
	}
	err = f.Close()
	if err != nil {
		return fmt.Errorf("can't close file %q: %v", path.Join(dir, base), err)
	}

	err = os.Rename(path.Join(dir, base), filep)
	if err != nil {
		err = fmt.Errorf("can't move %q to %q", path.Join(dir, base), filep)
	}
	return
}
