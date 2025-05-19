package lilypond

import (
	"gitlab.com/gomidi/lilypond"
	"gitlab.com/gomidi/muskel/smf"
)

type Track struct {
	voices []*voice
	staff  *lilypond.Staff
	name   string
	lyrics map[uint]string // barnumber to lyrics
	texts  map[smf.TicksAbsPos]string
}

func (t *Track) addVoice() (v *voice) {
	v = &voice{}
	v.track = t
	v.no = len(t.voices)
	v.Voice = t.staff.NewVoice()
	v.notes = map[uint][]note{}
	t.voices = append(t.voices, v)
	return
}

type str string

func (s str) String() string {
	return string(s)
}

func (t *Track) registerAllMyNotes(s *Score) {
	haslyrics := len(t.lyrics) > 0

	if haslyrics {
		// TODO: find the bars and add the barchanges
		//fmt.Printf("adding lyrics")
		ly := t.voices[0].NewLyrics()

		var elms []lilypond.Element

		for _, b := range s.bars {
			if b.timeSignatureChange {
				elms = append(elms, lilypond.TimeSignature{uint(b.num), uint(b.denom)})
			}

			if t.lyrics[b.no] != "" {
				elms = append(elms, lilypond.String(t.lyrics[b.no]))
			}
			elms = append(elms, lilypond.BarChange)
		}

		ly.Add(lilypond.ElementGroup(elms...))
	}

	for _, v := range t.voices {
		v.registerAllMyNotes(s)
	}

}

func (t *Track) findFreeVoice() *voice {
	for _, v := range t.voices {
		if !v.hasNote() {
			return v
		}
	}

	return t.addVoice()
}

func (t *Track) closeOpenedNote(key uint8) (v *voice) {
	for _, v := range t.voices {
		if v.openNotes[key] {
			v.openNotes[key] = false
			return v
		}
	}
	return nil
}

func (t *Track) openNote(key, velocity uint8, pos smf.TicksAbsPos) (v *voice) {
	v = t.findFreeVoice()
	v.openNotes[key] = true
	v.notePositions[key] = pos
	v.noteVelocities[key] = velocity
	return v
}
