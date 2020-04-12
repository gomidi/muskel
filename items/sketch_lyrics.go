package items

import (
	"fmt"
	"strings"
)

type SketchLyrics struct {
	column Lyricser
	lyrics *LyricsTable
}

func NewSketchLyrics(column Lyricser, lyrics *LyricsTable) *SketchLyrics {
	return &SketchLyrics{
		column: column,
		lyrics: lyrics,
	}
}

type Lyricser interface {
	Lyric(name string, from, to int) ([]string, error)
}

func (c *SketchLyrics) ApplyLyrics(evts []*Event) ([]*Event, error) {
	cc := c.lyrics

	idx := strings.Index(cc.Name, ".")
	if idx == -1 {
		return nil, fmt.Errorf("invalid lyrics call name %q (missing dot)", cc.Name)
	}

	if idx >= len(cc.Name)-1 {
		return nil, fmt.Errorf("invalid lyrics call name %q", cc.Name)
	}

	l, err := c.column.Lyric(cc.Name, cc.Slice[0], cc.Slice[1])
	if err != nil {
		return nil, err
	}
	return ApplyLyrics(evts, l), nil
}
