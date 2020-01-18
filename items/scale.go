package items

import (
	"fmt"
	"strings"
)

type Mode interface {
	Name() string
	NoteToStep(basenote, note uint8) (step int8)
	StepToNote(basenote uint8, step int8) (note uint8)
	All(basenote uint8) (notes []uint8)
}

type Scale struct {
	Mode
	BaseNote uint8
	Name     string
	Exploded bool // e.g. =patt(\major^c'...) will result in  =patt(c',d',e',f',g',a',b')
}

func (s *Scale) Dup() Item {
	return &Scale{
		Mode:     s.Mode,
		BaseNote: s.BaseNote,
		Name:     s.Name,
		Exploded: s.Exploded,
	}
}

func (s *Scale) All() []uint8 {
	return s.Mode.All(s.BaseNote)
}

func (s *Scale) Parse(data string, pos uint) error {
	idx := strings.Index(data, "^")
	if idx < 1 {
		return fmt.Errorf("invalid scale syntax: \\%s", data)
	}

	mod := data[:idx]
	s.Name = mod
	md, has := Modes[mod]
	if has {
		s.Mode = md
	}
	nt, err := ParseNote(data[idx+1:])
	if err != nil {
		return fmt.Errorf("invalid base note for scale: %q", data[idx+1:])
	}
	s.BaseNote = nt.ToMIDI()
	//s.Mode = md
	return nil
}

func (s *Scale) String() string {
	var nt Note
	nt.Letter, nt.Augmenter, nt.Octave = KeyToNote(s.BaseNote)
	name := s.Name
	if name == "" {
		name = s.Mode.Name()
	}
	return fmt.Sprintf("\\%s^%s", name, nt.String())
}

func (s *Scale) WriteMIDI(SMFWriter) []uint8 {
	panic("don't call me")
}

func (s Scale) NoteToStep(note uint8) int8 {
	if s.Mode == nil {
		panic("mode not set / known")
	}
	return s.Mode.NoteToStep(s.BaseNote, note)
}

func (s Scale) StepToNote(step int8) uint8 {
	if s.Mode == nil {
		panic("mode not set / known")
	}
	return s.Mode.StepToNote(s.BaseNote, step)
}

type mode struct {
	steps [7]uint8
	name  string
}

func (m mode) All(basenote uint8) (notes []uint8) {
	for step := range m.steps {
		notes = append(notes, m.StepToNote(basenote, int8(step)))
	}
	return
}

func (m mode) Name() string {
	return m.name
}

func (m mode) NoteToStep(basenote, note uint8) (step int8) {
	// fmt.Printf("NoteToStep(%v,%v) in mode %q\n", basenote, note, m.name)
	var diff int8
	if basenote > note {
		diff = int8(basenote-note) * (-1)
	} else {
		diff = int8(note) - int8(basenote)
	}

	diffPlusOne := diff + 1
	octaves := diff / 12
	octavesPlusOne := diffPlusOne / 12
	diffPlusOne = diffPlusOne % 12
	diff = diff % 12

	if diff >= 0 {
		for i, v := range m.steps {
			if int8(v) == diff {
				return int8(i) + octaves*7
			}
		}

		for i, v := range m.steps {
			if int8(v) == diffPlusOne {
				return int8(i) + octavesPlusOne*7
			}
		}

		panic("must not happen 1")
	}

	// fmt.Printf("diff is: %v diffPlusOne is %v octavesPlusOne: %v\n", diff, diffPlusOne, octavesPlusOne)

	for i, v := range m.steps {
		if int8(v) == (diff+12)%12 {
			// fmt.Printf("A [%v] step: %v\n", i, v)
			if diff == 0 {
				return int8(i) + octaves*7
			}
			return int8(i) - 7 + octaves*7
		}
	}

	for i, v := range m.steps {
		if int8(v) == (diffPlusOne+12)%12 {
			// fmt.Printf("B [%v] step: %v\n", i, v)
			if diffPlusOne == 0 {
				return int8(i) + octavesPlusOne*7
			}
			return int8(i) - 7 + octavesPlusOne*7
		}
	}
	panic("must not happen 2")
}

func (m mode) StepToNote(basenote uint8, step int8) uint8 {
	octaves := step / 7
	step = step % 7

	if step >= 0 {
		return uint8(int8(basenote+m.steps[step]) + octaves*12)
	}

	return uint8(int8(basenote) - (12 - int8(m.steps[step+7])) + octaves*12)
}

var Ionian = mode{
	name:  "ionian",
	steps: [7]uint8{0, 2, 4, 5, 7, 9, 11},
}

var Major = mode{
	name:  "major",
	steps: [7]uint8{0, 2, 4, 5, 7, 9, 11},
}

var Aeolian = mode{
	name:  "aeolian",
	steps: [7]uint8{0, 2, 3, 5, 7, 8, 10},
}

var Minor = mode{
	name:  "minor",
	steps: [7]uint8{0, 2, 3, 5, 7, 8, 10},
}

var Dorian = mode{
	name:  "dorian",
	steps: [7]uint8{0, 2, 3, 5, 7, 9, 10},
}

var Phrygian = mode{
	name:  "phrygian",
	steps: [7]uint8{0, 1, 3, 5, 7, 8, 10},
}

var Lydian = mode{
	name:  "lydian",
	steps: [7]uint8{0, 2, 4, 6, 7, 9, 11},
}

var Mixolydian = mode{
	name:  "mixolydian",
	steps: [7]uint8{0, 2, 4, 5, 7, 9, 10},
}

var Locrian = mode{
	name:  "locrian",
	steps: [7]uint8{0, 1, 3, 5, 6, 8, 10},
}

/*
Dorian Phrygian  Lydian  Mixolydian Aeolian Locrian Ionian
d      e         f       g          a       b       c
*/

var Modes = map[string]Mode{
	"dorian":     Dorian,
	"phrygian":   Phrygian,
	"lydian":     Lydian,
	"mixolydian": Mixolydian,
	"aeolian":    Aeolian,
	"locrian":    Locrian,
	"ionian":     Ionian,
	"minor":      Minor,
	"major":      Major,
}
