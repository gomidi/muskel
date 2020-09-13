package table

import (
	"gitlab.com/gomidi/muskel/filter"
	"gitlab.com/gomidi/muskel/items"
	"gitlab.com/gomidi/muskel/timbre"
	"gitlab.com/gomidi/muskel/track"
	"gitlab.com/gomidi/muskel/tuning"
)

type Formatter interface {
	WriteLine(line string) error
}

type Score interface {
	AddTrack(*track.Track)
	AddTuning(t *tuning.Tuning)
	AddFilter(t *filter.Filter)
	AddScale(t items.Mode)
	AddTimbre(t *timbre.Timbre)
	AddSketch(name string) interface {
		ParseLine([]string) error
		AddColumn(string)
		SetRealColNum(n int)
		SetGroupCol(pos int, subcols []string)
	}
	AddProperty(key, value string)
	AddLyrics(map[string][]string)
	AddInclude(file, sketch string, params []string) error
	AddToken(key, value string)
}

type sketch interface {
	ParseLine([]string) error
	AddColumn(name string)
	SetRealColNum(int)
	SetGroupCol(pos int, subcols []string)
}
