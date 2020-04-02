package items

import (
	"fmt"
	"regexp"
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
	itemGroupModifier
	//Exploded bool // e.g. =patt(\major^c'...) will result in  =patt(c',d',e',f',g',a',b')
}

func (s *Scale) Dup() Item {
	return &Scale{
		Mode:     s.Mode,
		BaseNote: s.BaseNote,
		Name:     s.Name,
		//Exploded: s.Exploded,
		itemGroupModifier: s.itemGroupModifier,
	}
}

func (s *Scale) All() []uint8 {
	return s.Mode.All(s.BaseNote)
}

var regExScaleNum = regexp.MustCompile("^([a-zA-Z" + regexp.QuoteMeta("\"") + regexp.QuoteMeta("'") + "#]+)(.*)")

func (s *Scale) Parse(data string, pos uint) error {

	data = data[1:]

	s.itemGroupModifier.Reset()
	if idx := strings.Index(data, "..."); idx > 0 {
		s.Exploded = true
		data = data[:idx]
	}
	//fmt.Printf("parsing scale %q\n", data)
	if idx := strings.Index(data, ")"); idx > 0 && data[0] == '(' {
		//fmt.Printf("parsing InplaceScale: %q\n", data)
		var is inplaceScale
		err := is.Parse(data[1:idx], 0)
		s.Mode = &is
		if len(data) > idx+1 {
			s.itemGroupModifier.parse(data[idx+1:])
		}
		return err
	}

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
	mt := regExScaleNum.FindStringSubmatch(data[idx+1:])
	//nt, err := ParseNote(data[idx+1:])
	nt, err := ParseNote(mt[1])
	if err != nil {
		return fmt.Errorf("invalid base note for scale: %q", data[idx+1:])
	}
	s.BaseNote = nt.ToMIDI()
	//s.Mode = md
	err = s.itemGroupModifier.parse(mt[2])
	return err
}

func (s *Scale) String() string {
	var nt Note
	nt.Letter, nt.Augmenter, nt.Octave = KeyToNote(s.BaseNote)
	name := s.Name
	if name == "" {
		name = s.Mode.Name()
	}

	if _, isInPlace := s.Mode.(*inplaceScale); isInPlace {
		return fmt.Sprintf("\\%s", name)
	}
	return fmt.Sprintf("\\%s^%s", name, nt.String()) + s.itemGroupModifier.string()
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
