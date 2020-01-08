package sketch

import (
	"gitlab.com/gomidi/muskel/items"
	"gitlab.com/gomidi/muskel/track"
)

type Score interface {
	GetSketch(sketch_table string) (*Sketch, error)
	GetIncludedSketch(filename, sketch_table string, params []string) (*Sketch, error)
	HasTrack(name string) bool
	GetTrack(track string) (*track.Track, error)
	GetToken(name string) (string, error)
	GetMode(name string) items.Mode
	FilterTrack(colName string, events []*Event) []*Event
	Properties() map[string]interface{}
	Lyric(part string, fromLine, toLine int) (tokens []string)
}
