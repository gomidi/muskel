package items

import (
	"fmt"
	"strings"
)

type sketchLyrics struct {
	column lyricser
	lyrics *LyricsTable
}

type lyricser interface {
	Lyric(name string, from, to int) ([]string, error)
}

func (c *sketchLyrics) applyLyrics(evts []*Event) ([]*Event, error) {
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
