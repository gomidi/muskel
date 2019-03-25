package score

import (
	"fmt"
	"strings"

	"gitlab.com/gomidi/muskel/template"
)

type Score struct {
	FileName    string
	Instruments []*Instrument
	Bars        []*Bar

	Meta                        map[string]string
	HeaderIncludes              []string
	Temperament                 map[string]string
	TemplateDefinitions         map[string]*template.Definition
	IncludedTemplateDefinitions map[string]*template.Definition
	Parts                       map[string][2]int // part name to start bar no and last bar
	BodyComments                map[int]string    // line calculated from the start of the system
	HeaderComments              []string          // comments in the header, as they come

	SmallColumns        bool // if the formatting should have no distance to the column lines
	IsUnrolled          bool
	BarNumbersTracked   bool
	LineWhereBodyStarts int
	LastLine            int
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

func New() *Score {
	return &Score{
		Meta:                        map[string]string{},
		Temperament:                 map[string]string{},
		TemplateDefinitions:         map[string]*template.Definition{},
		Parts:                       map[string][2]int{},
		BodyComments:                map[int]string{},
		IsUnrolled:                  false,
		IncludedTemplateDefinitions: map[string]*template.Definition{},
	}
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

func (s *Score) GetInstrument(name string) *Instrument {
	for _, instr := range s.Instruments {
		if instr.Name == name {
			return instr
		}
	}
	return nil
}

func (s *Score) AddInstrument(i *Instrument) error {
	i.Name = strings.TrimSpace(i.Name)
	if s.HasInstrument(i.Name) {
		return fmt.Errorf("instrument %q already defined", i.Name)
	}
	s.Instruments = append(s.Instruments, i)
	return nil
}

func (p *Score) HasInstrument(name string) bool {
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
		instr.Events = []BarEvents{}
		var be BarEvents
		var lastBarNo int

		//		var lastBar = p.Bars[len(p.Bars)-1]

		for _, ev := range instr.Unrolled {
			//				fmt.Printf("event: %#v\n", ev)
			//				fmt.Printf("ev.BarNo: %v  len(p.score.Bars): %v\n", ev.BarNo, len(p.score.Bars))
			p.Bars[ev.BarNo].ensurePositionExist(ev.DistanceToStartOfBarIn32th)

			diff := ev.BarNo - lastBarNo

			for d := 0; d < diff; d++ {
				instr.Events = append(instr.Events, be)
				be = BarEvents{}
			}

			lastBarNo = ev.BarNo
			be = append(be, ev)
		}

		//		fmt.Printf("last bar with event: %v, last bar total: %v len bars: %v\n", lastBarNo, p.Bars[len(p.Bars)-1].barNo, len(p.Bars))

		//			fmt.Printf("len bars in instrument: %v\n", len(be))
		instr.Events = append(instr.Events, be)

		if lastBarNo < p.Bars[len(p.Bars)-1].BarNo {
			instr.Events = append(instr.Events, BarEvents{})
		}

		instr.CalcColWidth(p.IsUnrolled)
	}
}

var DEBUG bool

func (s *Score) HasTemplate(name string) bool {
	_, hasOwn := s.TemplateDefinitions[name]
	_, hasIncluded := s.IncludedTemplateDefinitions[name]

	return hasOwn || hasIncluded
}
