package sketch

import (
	"gitlab.com/golang-utils/fs/path"
	"gitlab.com/gomidi/muskel/items"
	"gitlab.com/gomidi/muskel/track"
)

type Score interface {
	GetSketch(sketch_table string) (*Sketch, error)
	GetIncludedSketch(filename, sketch_table string, params []string) (*Sketch, error)
	GetExternalSketch(filename path.Relative, sketch_table string, params []string) (*Sketch, error)
	HasTrack(name string) bool
	GetTrack(track string) (*track.Track, error)
	GetToken(name string) (string, error)
	GetExternalToken(file path.Relative, token string) (string, error)
	Embed(patterns ...string) error
	GetMode(name string) items.Mode
	FilterTrack(colName string, events []*items.Event) []*items.Event
	Properties() map[string]string
	Lyric(part string, fromLine, toLine int) (tokens []string, err error)
	SetStartBar(uint)
	SetEndBar(uint)
}
